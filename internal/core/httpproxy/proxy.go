// Package httpproxy 将 Core 的插件与 Adapter HTTP 入口代理到当前已激活 Session 的 loopback Listener。
//
// 装配方在激活后调用 [Server.Bind]，排空时调用 [Server.Drain]，停止或 Session 失效时调用 [Server.Unbind]；
// 每次请求读取一个不可变路由快照，网络转发期间不持有映射锁。
package httpproxy

import (
	"bufio"
	"context"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"sync"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
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
	RequestTimeout time.Duration // 既有 REQUEST_TIMEOUT，Core 在开放入口前注入；零值沿用父请求期限。
	auth           Authorizer    // nil 时拒绝全部私有路由
	mu             sync.RWMutex  // 保护 mappings 及 binding 的排空状态；调用授权方与网络转发前释放
	mappings       map[string]*binding
	snapshot       func(string) (string, []*pb.RouteDescriptor, bool)
	ensure         func(context.Context, string, int64) (func(), error)
}

type binding struct {
	generationID string
	role         state.Role
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
	method        string
	path          string
	public        bool
	platformToken bool
}

// New 构造统一插件代理；auth 为 nil 时私有路由返回 401。
func New(auth Authorizer) *Server {
	return &Server{auth: auth, mappings: make(map[string]*binding)}
}

// SetWake 接入已提交路由和当前 Core 的唯一唤醒入口。
func (s *Server) SetWake(snapshot func(pluginID string) (generationID string, routes []*pb.RouteDescriptor, ok bool), ensure func(context.Context, string, int64) (func(), error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot, s.ensure = snapshot, ensure
}

func httpRoutes(routes []*pb.RouteDescriptor, role state.Role) []route {
	var result []route
	for _, descriptor := range routes {
		if descriptor.GetKind() == pb.RouteKind_ROUTE_KIND_HTTP {
			spec := descriptor.GetHttp()
			result = append(result, route{method: spec.GetMethod(), path: spec.GetPath(), public: spec.GetIsPublic(),
				platformToken: role == state.RoleAdapter && spec.GetIsPublic() && spec.GetAuthMode() == pb.HttpAuthMode_HTTP_AUTH_MODE_PLATFORM_TOKEN})
		}
	}
	return result
}

func matchRoute(routes []route, method, path string) (selected *route, pathExists bool) {
	for i := range routes {
		candidate := &routes[i]
		if path != candidate.path && !(strings.HasSuffix(candidate.path, "/") && strings.HasPrefix(path, candidate.path)) {
			continue
		}
		pathExists = true
		if method == candidate.method && (selected == nil || len(candidate.path) > len(selected.path)) {
			selected = candidate
		}
	}
	return selected, pathExists
}

// Bind 用已通过注册校验、且已激活的 Session 替换插件的旧映射。
// 角色、端点、凭证与路由组成该 Session 的不可变快照。
func (s *Server) Bind(pluginID, generationID string, endpoint *pb.HttpListener, token []byte, routes []*pb.RouteDescriptor, role state.Role) {
	host := net.JoinHostPort(endpoint.GetHost(), strconv.FormatUint(uint64(endpoint.GetPort()), 10))
	auth := hex.EncodeToString(token)
	prefix := pluginPrefix + pluginID
	current := &binding{generationID: generationID, role: role, token: append([]byte(nil), token...), drained: make(chan struct{}), upgraded: make(map[*upgradedConnection]struct{})}
	current.routes = httpRoutes(routes, role)
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
			request.Out.Header.Set(wire.HeaderProxyAuth, auth)
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Error(fmt.Sprintf("HTTP proxy request to component %s failed: %v", pluginID, err), "plugin_id", pluginID, "err", err)
			if role == state.RoleWebUI {
				w.Header().Set("Cache-Control", "no-store")
			}
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
	s.ServeWithIsolation(w, r, nil)
}

// ServeWithIsolation 使用同一 Session 快照决定认证、响应隔离和转发端点。
// isolate 包装普通插件的响应，并返回写完 trailer 后执行的清理函数。
func (s *Server) ServeWithIsolation(w http.ResponseWriter, r *http.Request, isolate func(http.ResponseWriter) (http.ResponseWriter, func())) {
	parent := r.Context()
	if s.RequestTimeout > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), s.RequestTimeout)
		defer cancel()
		r = r.WithContext(ctx)
	}
	// 1. 仅处理插件命名空间，并取得当前 Session 的路由快照。
	rest, ok := strings.CutPrefix(r.URL.Path, pluginPrefix)
	if !ok {
		http.NotFound(w, r)
		return
	}
	pluginID, path, ok := strings.Cut(rest, "/")
	s.mu.RLock()
	current := s.mappings[pluginID]
	unavailable := current == nil || current.draining
	snapshot, ensure := s.snapshot, s.ensure
	s.mu.RUnlock()
	var generation string
	var routes []route
	if current != nil {
		generation, routes = current.generationID, current.routes
	}
	committedSnapshot := false
	if snapshot != nil {
		id, descriptors, known := snapshot(pluginID)
		committedSnapshot = known
		if unavailable && known {
			generation, routes = id, httpRoutes(descriptors, state.RolePlugin)
			unavailable = false
		}
	}
	webUI := ok && current != nil && current.role == state.RoleWebUI
	if webUI {
		w.Header().Set("Cache-Control", "no-store")
	} else if isolate != nil {
		var cleanup func()
		w, cleanup = isolate(w)
		defer cleanup()
	}
	if !ok || pluginID == "" {
		http.NotFound(w, r)
		return
	}
	if unavailable {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	if webUI && !s.authorized(w, r, pluginID) {
		return
	}
	// 2. 同一方法下选择最长的匹配路径；精确路径与尾斜杠前缀保持各自含义。
	path = "/" + path
	selected, pathExists := matchRoute(routes, r.Method, path)
	if selected == nil {
		if pathExists {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		} else {
			http.NotFound(w, r)
		}
		return
	}
	// 3. 普通插件的私有路由也用管理认证；公开回调的业务签名由插件检查。
	if !webUI && !selected.public && !s.authorized(w, r, pluginID) {
		return
	}
	if ensure != nil {
		// 正文仍由 ReverseProxy 流式读取；此处尚未在 Core 缓存请求正文。
		release, err := ensure(r.Context(), pluginID, 0)
		if err != nil {
			status := http.StatusServiceUnavailable
			switch wire.CodeOf(err) {
			case pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED:
				status = http.StatusTooManyRequests
			case pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED:
				status = http.StatusGatewayTimeout
			}
			slog.Warn(fmt.Sprintf("Failed to wake component %s for an HTTP request: %v", pluginID, err), "plugin_id", pluginID, "error", err)
			http.Error(w, http.StatusText(status), status)
			return
		}
		defer release()
		s.mu.RLock()
		current = s.mappings[pluginID]
		s.mu.RUnlock()
		if current == nil || current.generationID != generation {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		confirmed, _ := matchRoute(current.routes, r.Method, path)
		if snapshot != nil {
			id, descriptors, known := snapshot(pluginID)
			if known {
				committed, _ := matchRoute(httpRoutes(descriptors, current.role), r.Method, path)
				if id != generation || committed == nil || *committed != *selected {
					http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
					return
				}
			} else if committedSnapshot {
				http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
				return
			}
		}
		if confirmed == nil || *confirmed != *selected || r.Context().Err() != nil {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		if !webUI && !confirmed.public && !s.authorized(w, r, pluginID) {
			return
		}
		selected = confirmed
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
	rewrite := proxy.Rewrite
	proxy.Rewrite = func(request *httputil.ProxyRequest) {
		stripIdentity(request.Out, selected.platformToken)
		rewrite(request)
	}
	if webUI {
		proxy.ModifyResponse = func(response *http.Response) error {
			// ReverseProxy 追加上游头；统一由上游响应副本写出一条 no-store。
			w.Header().Del("Cache-Control")
			response.Header.Set("Cache-Control", "no-store")
			return nil
		}
	} else if policy, ok := w.(interface{ ModifyResponse(*http.Response) error }); ok {
		proxy.ModifyResponse = policy.ModifyResponse
	}
	writer := &upgradeWriter{ResponseWriter: w, binding: current}
	if s.RequestTimeout > 0 && r.Header.Get("Upgrade") != "" {
		// 握手前沿用请求期限；成功劫持后停止期限传播，父请求取消仍关闭连接。
		ctx, cancel := context.WithCancel(parent)
		defer cancel()
		writer.stopTimeout = context.AfterFunc(r.Context(), cancel)
		defer writer.stopTimeout()
		r = r.WithContext(ctx)
	}
	proxy.ServeHTTP(writer, r)
}

func (s *Server) authorized(w http.ResponseWriter, r *http.Request, pluginID string) bool {
	if s.auth != nil && s.auth.Authorize(r, pluginID) {
		return true
	}
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	return false
}

func stripIdentity(r *http.Request, platformToken bool) {
	for name := range r.Header {
		if strings.EqualFold(name, wire.HeaderProxyAuth) || (!platformToken && strings.EqualFold(name, "Authorization")) || strings.EqualFold(name, "Cookie") || strings.EqualFold(name, "X-Alonabot-CSRF") || strings.EqualFold(name, "Forwarded") || strings.HasPrefix(strings.ToLower(name), "x-forwarded-") {
			delete(r.Header, name)
		}
	}
}

// upgradeWriter 登记 ReverseProxy 劫持的客户端连接；关闭客户端也会结束代理的双向复制。
type upgradeWriter struct {
	http.ResponseWriter
	binding     *binding
	stopTimeout func() bool // 仅在 ReverseProxy 校验上游升级响应并成功劫持后停止请求期限传播。
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
	if w.stopTimeout != nil {
		w.stopTimeout()
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
