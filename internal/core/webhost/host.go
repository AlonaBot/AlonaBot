// Package webhost 按受控主机名分派管理入口和插件入口，并签发受管理会话约束的插件授权。
//
// [Host] 在管理请求进入页面处理器前通过 [AdminSessions] 记录会话期限，
// 页面调用 [Host.IssueTicket] 后，浏览器在插件主机兑换一次性票据。
// 管理认证由装配方提供；插件响应在写出前经过 Cookie、CORS 与单 Host 主动内容策略，
// 对应 CONTRACTS §7.3。
package webhost

import (
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/httpproxy"
)

const (
	pluginPrefix = "/plugins/"
	ticketPath   = "/_alonabot/ticket"
)

// AdminSessions 校验管理请求及其派生授权依赖的原会话。
type AdminSessions interface {
	Lookup(r *http.Request) (sessionID string, expiresAt time.Time, ok bool) // 只在 admin host 上调用
	Valid(sessionID string) bool
}

// Options 配置受控入口、代理和管理会话来源。
type Options struct {
	AdminBaseURL, PluginBaseURL *url.URL     // 来自已校验的 settings，构造后不再修改
	SecondHost                  bool         // 两个 Base URL 的主机名不同时为 true
	Admin                       http.Handler // nil 时管理入口返回 404
	Proxy                       *httpproxy.Server
	Sessions                    AdminSessions    // 未接入管理认证时为 nil，私有授权保持拒绝
	Now                         func() time.Time // nil 时使用 time.Now
}

// Host 持有主机分派和派生授权；方法可并发调用。
// mu 保护管理期限、待兑换票据和插件会话，调用外部管理会话来源前释放。
type Host struct {
	log          *slog.Logger
	opt          Options
	mu           sync.Mutex
	adminExpires map[string]time.Time
	tickets      map[string]ticket
	sessions     map[string]pluginSession
}

// New 构造受控 Host 入口；opt 的 Base URL 已由 settings 校验。
func New(log *slog.Logger, opt Options) *Host {
	if opt.Now == nil {
		opt.Now = time.Now
	}
	return &Host{log: log, opt: opt, adminExpires: make(map[string]time.Time), tickets: make(map[string]ticket), sessions: make(map[string]pluginSession)}
}

// ServeHTTP 仅使用请求 Host 分派管理入口、插件路由和票据兑换。
func (h *Host) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, pluginPrefix) {
		w = h.pluginWriter(w, r)
	}
	hostname := requestHostname(r.Host)
	if strings.EqualFold(hostname, h.opt.AdminBaseURL.Hostname()) {
		if strings.HasPrefix(r.URL.Path, pluginPrefix) {
			if h.opt.SecondHost {
				http.NotFound(w, r)
				return
			}
			if activeAccept(r.Header) {
				http.Error(w, strings.TrimSpace(configurePluginHost), http.StatusConflict)
				return
			}
			h.opt.Proxy.ServeHTTP(w, r)
			return
		}
		if h.opt.Admin == nil {
			http.NotFound(w, r)
			return
		}
		if h.opt.Sessions != nil {
			if id, expires, ok := h.opt.Sessions.Lookup(r); ok {
				h.mu.Lock()
				h.adminExpires[id] = expires
				h.mu.Unlock()
			}
		}
		if h.opt.SecondHost && pluginOrigin(r.Header.Get("Origin"), h.opt.PluginBaseURL) {
			w = &responseWriter{ResponseWriter: w, before: func(header http.Header) {
				header.Del("Access-Control-Allow-Origin")
				header.Del("Access-Control-Allow-Credentials")
			}}
		}
		h.opt.Admin.ServeHTTP(w, r)
		if policy, ok := w.(*responseWriter); ok && !policy.wrote {
			policy.before(policy.Header())
		}
		return
	}
	if h.opt.SecondHost && strings.EqualFold(hostname, h.opt.PluginBaseURL.Hostname()) {
		switch {
		case r.URL.Path == ticketPath:
			h.exchange(w, r)
		case strings.HasPrefix(r.URL.Path, pluginPrefix):
			h.opt.Proxy.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
		return
	}
	http.Error(w, http.StatusText(http.StatusMisdirectedRequest), http.StatusMisdirectedRequest)
}

func requestHostname(host string) string {
	if hostname, _, err := net.SplitHostPort(host); err == nil {
		return hostname
	}
	return strings.Trim(host, "[]")
}
