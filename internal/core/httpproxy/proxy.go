// Package httpproxy 将 Core 的插件与 Adapter HTTP 入口代理到当前已激活 Session 的 loopback Listener。
//
// 装配方在激活后调用 [Server.Bind]，排空时调用 [Server.Drain]，停止或 Session 失效时调用 [Server.Unbind]；
// 每次请求读取一个不可变路由快照，网络转发期间不持有映射锁。代理身份隔离见 CONTRACTS §7.2。
package httpproxy

import (
	"bufio"
	"crypto/subtle"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"sync"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

const pluginPrefix = "/plugins/"

// Authorizer 检查 Core 是否允许请求访问指定插件的私有 HTTP 路由。
type Authorizer interface {
	Authorize(r *http.Request, pluginID string) bool
}

// Server 保存当前已激活 Session 的代理映射；Bind、Drain、Unbind 与 ServeHTTP 可并发调用。
// Drain 冻结新派发并按停止截止等待在途请求；Unbind 撤销匹配 Session 的映射。
type Server struct {
	auth     Authorizer   // nil 时拒绝全部私有路由
	mu       sync.RWMutex // 保护 mappings 及 binding 的排空状态；调用授权方与网络转发前释放
	mappings map[string]*binding
}

type binding struct {
	generationID string
	token        []byte // Bind 时复制的 Session 内部代理凭证，后续只读。
	routes       []route
	proxy        *httputil.ReverseProxy
	draining     bool
	inflight     int
	drained      chan struct{} // 冻结后最后一个在途请求返回时关闭，由 Server.mu 保护
	mu           sync.Mutex    // 保护 upgraded 与 closed；连接在锁外关闭
	upgraded     map[*upgradedConnection]struct{}
	closed       bool
}

type route struct {
	method string
	path   string
	public bool
}

// New 构造统一插件代理；auth 为 nil 时私有路由返回 401。
func New(auth Authorizer) *Server {
	return &Server{auth: auth, mappings: make(map[string]*binding)}
}

// Bind 用已通过注册校验、且已激活的 Session 替换插件的旧映射。
// endpoint、token 与 routes 在调用时复制为该 Session 的不可变快照。
func (s *Server) Bind(pluginID, generationID string, endpoint *pb.HttpListener, token []byte, routes []*pb.RouteDescriptor, isAdapter bool) {
	host := net.JoinHostPort(endpoint.GetHost(), strconv.FormatUint(uint64(endpoint.GetPort()), 10))
	auth := hex.EncodeToString(token)
	prefix := pluginPrefix + pluginID
	current := &binding{generationID: generationID, token: append([]byte(nil), token...), drained: make(chan struct{}), upgraded: make(map[*upgradedConnection]struct{})}
	for _, descriptor := range routes {
		if descriptor.GetKind() == pb.RouteKind_ROUTE_KIND_HTTP {
			spec := descriptor.GetHttp()
			current.routes = append(current.routes, route{method: spec.GetMethod(), path: spec.GetPath(), public: spec.GetIsPublic()})
		}
	}
	current.proxy = &httputil.ReverseProxy{
		ErrorLog: slog.NewLogLogger(slog.With("plugin_id", pluginID).Handler(), slog.LevelError),
		// 禁用连接复用，避免 Transport 对已复用连接上的幂等请求自动重试；内部 loopback 直连不走出站代理。
		Transport: &http.Transport{DisableKeepAlives: true},
		Rewrite: func(request *httputil.ProxyRequest) {
			request.Out.URL.Scheme = "http"
			request.Out.URL.Host = host
			request.Out.Host = host
			// 与准入匹配使用同一份未清理 Path；重新编码剥前缀后的路径，避免残留完整 RawPath。
			request.Out.URL.Path = strings.TrimPrefix(request.In.URL.Path, prefix)
			request.Out.URL.RawPath = ""
			request.Out.URL.RawQuery = request.In.URL.RawQuery
			stripIdentity(request.Out, isAdapter)
			request.Out.Header.Set(wire.HeaderProxyAuth, auth)
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Error("plugin HTTP proxy failed", "plugin_id", pluginID, "err", err)
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		},
	}
	s.mu.Lock()
	previous := s.mappings[pluginID]
	s.mappings[pluginID] = current
	s.mu.Unlock()
	if previous != nil {
		previous.closeUpgrades()
	}
}

// Drain 拒绝该插件的新请求，并等待已经派发的请求结束，最多等到 deadline。
// 截止后由装配方继续 Unbind 和停止进程，余留请求随插件 Listener 关闭而结束。
func (s *Server) Drain(pluginID string, deadline time.Time) {
	// 1. 冻结与派发共用映射锁，保证冻结后不再增加在途请求。
	s.mu.Lock()
	current := s.mappings[pluginID]
	if current == nil {
		s.mu.Unlock()
		return
	}
	if !current.draining {
		current.draining = true
		if current.inflight == 0 {
			close(current.drained)
		}
	}
	s.mu.Unlock()

	// 2. 等待期间释放锁，让请求完成和其他插件的启停继续推进。
	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()
	select {
	case <-current.drained:
	case <-timer.C:
	}
}

// Unbind 仅撤销代次与内部代理凭证均匹配的 Session 映射。
func (s *Server) Unbind(pluginID, generationID string, token []byte) {
	s.mu.Lock()
	current := s.mappings[pluginID]
	if current == nil || current.generationID != generationID || subtle.ConstantTimeCompare(current.token, token) != 1 {
		s.mu.Unlock()
		return
	}
	delete(s.mappings, pluginID)
	s.mu.Unlock()
	current.closeUpgrades()
}

// ServeHTTP 匹配 /plugins/{id}/ 下的原始请求路径并授权，再向绑定的固定端点转发。
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. 仅处理插件命名空间，并取得当前 Session 的路由快照。
	rest, ok := strings.CutPrefix(r.URL.Path, pluginPrefix)
	if !ok {
		http.NotFound(w, r)
		return
	}
	pluginID, path, ok := strings.Cut(rest, "/")
	if !ok || pluginID == "" {
		http.NotFound(w, r)
		return
	}
	path = "/" + path
	s.mu.RLock()
	current := s.mappings[pluginID]
	unavailable := current == nil || current.draining
	s.mu.RUnlock()
	if unavailable {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	// 2. 同一方法下选择最长的匹配路径；精确路径与尾斜杠前缀保持各自含义。
	var selected *route
	pathExists := false
	for i := range current.routes {
		candidate := &current.routes[i]
		if path != candidate.path && !(strings.HasSuffix(candidate.path, "/") && strings.HasPrefix(path, candidate.path)) {
			continue
		}
		pathExists = true
		if r.Method == candidate.method && (selected == nil || len(candidate.path) > len(selected.path)) {
			selected = candidate
		}
	}
	if selected == nil {
		if pathExists {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		} else {
			http.NotFound(w, r)
		}
		return
	}
	// 3. 私有路由由管理认证检查会话与 CSRF；公开回调的业务签名由插件检查。
	if !selected.public && (s.auth == nil || !s.auth.Authorize(r, pluginID)) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	// 4. 授权期间可能已冻结或换代；在同一锁内复核映射并登记实际派发。
	s.mu.Lock()
	if s.mappings[pluginID] != current || current.draining {
		s.mu.Unlock()
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	current.inflight++
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		current.inflight--
		if current.draining && current.inflight == 0 {
			close(current.drained) // 唤醒 Drain，完成路径始终操作本次派发的 Session 快照。
		}
	}()
	// 每次请求复制代理配置，响应策略不会改写其他并发请求的快照。
	proxy := *current.proxy
	if policy, ok := w.(interface{ ModifyResponse(*http.Response) error }); ok {
		proxy.ModifyResponse = policy.ModifyResponse
	}
	proxy.ServeHTTP(&upgradeWriter{ResponseWriter: w, binding: current}, r)
}

func stripIdentity(r *http.Request, isAdapter bool) {
	for name := range r.Header {
		if strings.EqualFold(name, wire.HeaderProxyAuth) || (!isAdapter && strings.EqualFold(name, "Authorization")) || strings.EqualFold(name, "Cookie") || strings.EqualFold(name, "X-Alonabot-CSRF") || strings.EqualFold(name, "Forwarded") || strings.HasPrefix(strings.ToLower(name), "x-forwarded-") {
			delete(r.Header, name)
		}
	}
}

// upgradeWriter 登记 ReverseProxy 劫持的客户端连接；关闭客户端也会结束代理的双向复制。
type upgradeWriter struct {
	http.ResponseWriter
	binding *binding
}

func (w *upgradeWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *upgradeWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, rw, err := http.NewResponseController(w.ResponseWriter).Hijack()
	if err != nil {
		return nil, nil, err
	}
	tracked := &upgradedConnection{Conn: conn, binding: w.binding}
	w.binding.mu.Lock()
	closed := w.binding.closed
	if !closed {
		w.binding.upgraded[tracked] = struct{}{}
	}
	w.binding.mu.Unlock()
	if closed {
		_ = conn.Close() // 撤销先于 Hijack 完成时，已经劫持的连接同样由本绑定回收。
		return nil, nil, net.ErrClosed
	}
	return tracked, rw, nil
}

type upgradedConnection struct {
	net.Conn
	binding *binding
}

func (c *upgradedConnection) Close() error {
	c.binding.mu.Lock()
	delete(c.binding.upgraded, c)
	c.binding.mu.Unlock()
	return c.Conn.Close()
}

func (b *binding) closeUpgrades() {
	b.mu.Lock()
	b.closed = true
	connections := make([]*upgradedConnection, 0, len(b.upgraded))
	for conn := range b.upgraded {
		connections = append(connections, conn)
	}
	b.mu.Unlock()
	for _, conn := range connections {
		_ = conn.Close()
	}
}
