// Package webhost 按受控主机名分派管理入口和插件路由，并统一隔离插件响应。
//
// [Host] 将 /plugins/ 请求交给代理、/media/ 交给媒体处理器，其余交给管理处理器。
// 管理认证由装配方接入。
// 插件响应经过沙箱 CSP、Cookie 与凭证 CORS 过滤，规则见 CONTRACTS §7.3。
package webhost

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/httpproxy"
)

const pluginPrefix = "/plugins/"

// Options 配置受控入口、管理与媒体处理器及插件代理。
type Options struct {
	BaseURL *url.URL     // 来自已校验的 settings，构造后不再修改
	Admin   http.Handler // nil 时管理入口返回 404
	Media   http.Handler // nil 时保留的 /media/ 路径返回 404
	Proxy   *httpproxy.Server
	Now     func() time.Time
}

// Host 持有受控主机与响应策略；方法可并发调用。
type Host struct{ opt Options }

// New 构造受控 Host 入口；opt 的 Base URL 已由 settings 校验。
func New(opt Options) *Host { return &Host{opt: opt} }

// ServeHTTP 按受控主机名分派管理、媒体与插件路由，其他 Host 返回 421。
func (h *Host) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, pluginPrefix) {
		plugin := h.pluginWriter(w, r)
		w = plugin
		// ReverseProxy 在正文结束后才复制 trailer，返回前还需清理一次。
		defer plugin.after(plugin.Header())
	}
	if !strings.EqualFold(requestHostname(r.Host), h.opt.BaseURL.Hostname()) {
		http.Error(w, http.StatusText(http.StatusMisdirectedRequest), http.StatusMisdirectedRequest)
		return
	}
	if strings.HasPrefix(r.URL.Path, pluginPrefix) {
		h.opt.Proxy.ServeHTTP(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/media/") {
		if h.opt.Media == nil {
			http.NotFound(w, r)
			return
		}
		h.opt.Media.ServeHTTP(w, r)
		return
	}
	if h.opt.Admin == nil {
		http.NotFound(w, r)
		return
	}
	h.opt.Admin.ServeHTTP(w, r)
}

func requestHostname(host string) string {
	if hostname, _, err := net.SplitHostPort(host); err == nil {
		return hostname
	}
	return strings.Trim(host, "[]")
}
