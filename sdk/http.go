package sdk

import (
	"context"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

const httpTokenBytes = 32
const httpCapabilityVersion = 1

// httpListener 持有一次 Run 的 HTTP 资源；准备期创建端点，激活后复用 activation 的 ready gate。
// routes/token 在 Serve 启动前写入；activation 通过原子指针发布，关闭按首次停止的截止执行一次。
type httpListener struct {
	runtime    *componentSession
	routes     []httpRoute
	listener   net.Listener
	server     *http.Server
	token      []byte // token 的十六进制，仅在进入用户 Handler 前比较
	activation atomic.Pointer[activation]
	mu         sync.Mutex // 保护 hijacked 与 closing；关闭连接时释放锁
	hijacked   map[*httpConnection]struct{}
	closing    bool
	closeOnce  sync.Once
}

func newHTTPListener(routes []httpRoute) *httpListener {
	if len(routes) == 0 {
		return nil
	}
	return &httpListener{routes: routes}
}

func (h *httpListener) start(boot *pb.Bootstrap) error {
	if h == nil {
		return nil
	}
	token := boot.GetHttpProxyAuth().GetToken()
	if len(token) != httpTokenBytes {
		return errors.New("Bootstrap is missing a valid HTTP proxy credential")
	}
	listener, err := wire.ListenLoopback()
	if err != nil {
		return fmt.Errorf("listen plugin HTTP: %w", err)
	}
	h.listener = listener
	h.hijacked = make(map[*httpConnection]struct{})
	h.token = []byte(hex.EncodeToString(token))
	h.server = &http.Server{Handler: http.HandlerFunc(h.serveHTTP), ErrorLog: slog.NewLogLogger(componentLogger(h.runtime.ctx).Handler(), slog.LevelError), ConnState: h.connectionState}
	go func() {
		if err := h.server.Serve(&httpNetworkListener{Listener: listener, owner: h}); err != nil && !errors.Is(err, http.ErrServerClosed) {
			componentLogger(h.runtime.ctx).Error("plugin HTTP listener failed", "err", err)
		}
	}()
	return nil
}

func (h *httpListener) report(r *pb.Register) {
	if h == nil {
		return
	}
	address := h.listener.Addr().(*net.TCPAddr)
	r.HttpListener = &pb.HttpListener{Host: address.IP.String(), Port: uint32(address.Port)}
}

func (h *httpListener) close(deadline time.Time) {
	if h == nil || h.server == nil {
		return
	}
	h.closeOnce.Do(func() {
		// Shutdown 不回收 Hijack 交给用户的连接；先冻结登记，再在锁外关闭快照。
		h.mu.Lock()
		h.closing = true
		connections := make([]*httpConnection, 0, len(h.hijacked))
		for conn := range h.hijacked {
			connections = append(connections, conn)
		}
		h.mu.Unlock()
		for _, conn := range connections {
			_ = conn.Close()
		}
		ctx, cancel := context.WithDeadline(h.runtime.ctx, deadline)
		defer cancel()
		if err := h.server.Shutdown(ctx); err != nil {
			componentLogger(h.runtime.ctx).Error("stop plugin HTTP listener failed", "err", err)
			if closeErr := h.server.Close(); closeErr != nil {
				componentLogger(h.runtime.ctx).Error("close plugin HTTP connections failed", "err", closeErr)
			}
		}
	})
}

func (h *httpListener) serveHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. 所有请求先验 Session 凭证；未激活端点同样拒绝伪造身份。
	if subtle.ConstantTimeCompare([]byte(r.Header.Get(wire.HeaderProxyAuth)), h.token) != 1 {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	a := h.activation.Load()
	if a == nil {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	// 与 ACTIVATED 写出同步，复用 RPC 门禁，避免首个代理请求读到过期的 ready 状态。
	a.readyMu.Lock()
	ready := a.ready.Load() && !a.stopping.Load()
	select {
	case <-a.sess.Done():
		ready = false
	default:
	}
	a.readyMu.Unlock()
	if !ready {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	// 2. 与 Core 一致：对未清理 Path 按 method 选择最长精确/尾斜杠前缀。
	var selected *httpRoute
	pathExists := false
	for i := range h.routes {
		candidate := &h.routes[i]
		if r.URL.Path != candidate.Path && !(strings.HasSuffix(candidate.Path, "/") && strings.HasPrefix(r.URL.Path, candidate.Path)) {
			continue
		}
		pathExists = true
		if r.Method == candidate.Method && (selected == nil || len(candidate.Path) > len(selected.Path)) {
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
	if !h.runtime.beginCallback() {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	defer h.runtime.callbacks.Done()
	r.Header.Del(wire.HeaderProxyAuth)
	selected.Handler.ServeHTTP(w, r.WithContext(a.l.runtime.context(r.Context())))
}

func appendHTTPRoutes(r *pb.Register, routes []httpRoute) {
	for _, route := range routes {
		r.Routes = append(r.Routes, &pb.RouteDescriptor{RouteId: "http." + route.Method + "." + hex.EncodeToString([]byte(route.Path)), Kind: pb.RouteKind_ROUTE_KIND_HTTP, DefaultEnabled: true, PageTitle: route.PageTitle, Spec: &pb.RouteDescriptor_Http{Http: &pb.HttpRouteSpec{Method: route.Method, Path: route.Path, IsPublic: route.Public, AuthMode: pb.HttpAuthMode(route.AuthMode)}}})
	}
}

func httpOffer(l launch) (supported, required []*pb.ProtocolCapability) {
	supported = append([]*pb.ProtocolCapability(nil), sdkOffer.Supported...)
	required = append([]*pb.ProtocolCapability(nil), sdkOffer.Required...)
	if l.http != nil {
		capability := &pb.ProtocolCapability{Name: wire.CapabilityHTTPListener, Version: httpCapabilityVersion}
		supported = append(supported, capability)
		required = append(required, capability)
	}
	return supported, required
}

func validHTTPMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace, http.MethodPatch:
		return true
	}
	return false
}

func canonicalHTTPPath(raw string) bool {
	canonical := path.Clean(raw)
	if strings.HasSuffix(raw, "/") && canonical != "/" {
		canonical += "/"
	}
	return strings.HasPrefix(raw, "/") && canonical == raw && !strings.ContainsAny(raw, "?#") && !strings.Contains(raw, "..")
}

// HTTPAuth 指定插件与 Adapter HTTP 路由的认证方式。
type HTTPAuth int

const (
	AuthCoreSession       HTTPAuth = iota + 1 // Core 会话认证。
	AuthCallbackState                         // 插件校验 state。
	AuthCallbackSignature                     // 插件校验签名。
	AuthCallbackNonce                         // 插件校验 nonce。
)

// AuthPlatformToken 声明由 Adapter 校验公开路由的平台凭证。
const AuthPlatformToken HTTPAuth = HTTPAuth(pb.HttpAuthMode_HTTP_AUTH_MODE_PLATFORM_TOKEN)

// httpNetworkListener 包装 SDK 接受的连接，使用户关闭 Hijack 连接时也能释放登记。
type httpNetworkListener struct {
	net.Listener
	owner *httpListener
}

func (l *httpNetworkListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &httpConnection{Conn: conn, owner: l.owner}, nil
}

type httpConnection struct {
	net.Conn
	owner *httpListener
}

func (c *httpConnection) Close() error {
	c.owner.mu.Lock()
	delete(c.owner.hijacked, c)
	c.owner.mu.Unlock()
	return c.Conn.Close()
}

func (h *httpListener) connectionState(conn net.Conn, state http.ConnState) {
	if state != http.StateHijacked {
		return
	}
	tracked := conn.(*httpConnection)
	h.mu.Lock()
	closing := h.closing
	if !closing {
		h.hijacked[tracked] = struct{}{}
	}
	h.mu.Unlock()
	if closing {
		_ = tracked.Close() // 停止快照之后才完成的 Hijack 同样回收。
	}
}
