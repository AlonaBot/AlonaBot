// Package webhost 按受控主机名分派管理入口和插件路由，并统一隔离插件响应。
//
// [Host] 将 /plugins/ 请求交给代理、/media/ 交给媒体处理器，其余交给管理处理器。
// WebUI 后台与管理入口同源，管理认证由代理按 Session 执行。
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
	Media   http.Handler // 自行校验签发 Host；nil 时保留的 /media/ 路径返回 404。
	Proxy   *httpproxy.Server
	Now     func() time.Time
}

// Host 持有受控主机与响应策略；方法可并发调用。
type Host struct{ opt Options }

// New 构造受控 Host 入口；opt 的 Base URL 已由 settings 校验。
func New(opt Options) *Host { return &Host{opt: opt} }

// ServeHTTP 分派媒体取件与受控主机名下的管理、插件请求。
func (h *Host) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 媒体处理器按令牌校验签发主机，允许协议端使用独立的容器取件地址。
	if strings.HasPrefix(r.URL.Path, "/media/") && h.opt.Media != nil {
		h.opt.Media.ServeHTTP(w, r)
		return
	}
	pluginPath := strings.HasPrefix(r.URL.Path, pluginPrefix)
	if !strings.EqualFold(requestHostname(r.Host), h.opt.BaseURL.Hostname()) {
		if pluginPath {
			w = h.pluginWriter(w, r)
		}
		http.Error(w, http.StatusText(http.StatusMisdirectedRequest), http.StatusMisdirectedRequest)
		return
	}
	if pluginPath {
		h.opt.Proxy.ServeWithIsolation(w, r, func(w http.ResponseWriter) (http.ResponseWriter, func()) {
			plugin := h.pluginWriter(w, r)
			return plugin, func() { plugin.after(plugin.Header()) }
		})
		return
	}
	if strings.HasPrefix(r.URL.Path, "/media/") {
		http.NotFound(w, r)
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
