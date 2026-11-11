// Package httpproxy 将 Core 的插件 HTTP 入口代理到当前已激活 Session 的 loopback Listener。
//
// 装配方在激活后调用 [Server.Bind]，停止或 Session 失效时调用 [Server.Unbind]；
// 每次请求读取一个不可变路由快照，网络转发期间不持有映射锁。代理身份隔离见 CONTRACTS §7.2。
package httpproxy

import (
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"sync"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

const pluginPrefix = "/plugins/"

// Authorizer 检查 Core 是否允许请求访问指定插件的私有 HTTP 路由。
type Authorizer interface {
	Authorize(r *http.Request, pluginID string) bool
}

// Server 保存当前已激活 Session 的代理映射；Bind、Unbind 与 ServeHTTP 可并发调用。
// 映射撤销阻止新的请求；已经取得快照的在途请求由 SDK 的门禁与停止预算收尾。
type Server struct {
	auth     Authorizer   // nil 时拒绝全部私有路由
	mu       sync.RWMutex // 保护 mappings；调用授权方与网络转发前释放
	mappings map[string]*binding
}

type binding struct {
	generationID string
	routes       []route
	proxy        *httputil.ReverseProxy
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
func (s *Server) Bind(pluginID, generationID string, endpoint *pb.HttpListener, token []byte, routes []*pb.RouteDescriptor) {
	host := net.JoinHostPort(endpoint.GetHost(), strconv.FormatUint(uint64(endpoint.GetPort()), 10))
	auth := hex.EncodeToString(token)
	prefix := pluginPrefix + pluginID
	current := &binding{generationID: generationID}
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
			stripIdentity(request.Out)
			request.Out.Header.Set(wire.HeaderProxyAuth, auth)
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Error("plugin HTTP proxy failed", "plugin_id", pluginID, "err", err)
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		},
	}
	s.mu.Lock()
	s.mappings[pluginID] = current
	s.mu.Unlock()
}

// Unbind 仅撤销 generationID 仍与当前映射一致的插件，旧代次清理不会删除新映射。
func (s *Server) Unbind(pluginID, generationID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if current := s.mappings[pluginID]; current != nil && current.generationID == generationID {
		delete(s.mappings, pluginID)
	}
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
	s.mu.RUnlock()
	if current == nil {
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
	// 每次请求复制代理配置，响应策略不会改写其他并发请求的快照。
	proxy := *current.proxy
	if policy, ok := w.(interface{ ModifyResponse(*http.Response) error }); ok {
		proxy.ModifyResponse = policy.ModifyResponse
	}
	proxy.ServeHTTP(w, r)
}

func stripIdentity(r *http.Request) {
	for name := range r.Header {
		if strings.EqualFold(name, wire.HeaderProxyAuth) || strings.EqualFold(name, "Authorization") || strings.EqualFold(name, "X-Alonabot-CSRF") || strings.EqualFold(name, "Forwarded") || strings.HasPrefix(strings.ToLower(name), "x-forwarded-") {
			delete(r.Header, name)
		}
	}
	cookies := r.Cookies()
	r.Header.Del("Cookie")
	for _, cookie := range cookies {
		if !strings.HasPrefix(cookie.Name, "alonabot_") && !strings.HasPrefix(cookie.Name, "__Host-alonabot_") {
			r.AddCookie(cookie)
		}
	}
}
