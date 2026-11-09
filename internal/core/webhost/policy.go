package webhost

import (
	"log/slog"
	"net/http"
	"strings"
)

const (
	sandboxCSP         = "sandbox allow-scripts allow-forms allow-popups"
	untypedContentType = "application/octet-stream"
)

func (h *Host) pluginWriter(w http.ResponseWriter, r *http.Request) *responseWriter {
	rest := strings.TrimPrefix(r.URL.Path, pluginPrefix)
	pluginID, _, _ := strings.Cut(rest, "/")
	apply := func(header http.Header) { h.pluginHeaders(header, pluginID) }
	return &responseWriter{ResponseWriter: w, before: apply, after: func(header http.Header) {
		h.stripPluginHeaders(header, pluginID)
	}, modify: func(response *http.Response) error {
		apply(response.Header)
		return nil
	}}
}

func (h *Host) pluginHeaders(header http.Header, pluginID string) {
	h.stripPluginHeaders(header, pluginID)
	header.Set("Content-Security-Policy", sandboxCSP)
	header.Set("X-Content-Type-Options", "nosniff")
	if header.Get("Content-Type") == "" {
		// 阻止公开 Listener 将缺失 MIME 的正文重新嗅探成 HTML。
		header.Set("Content-Type", untypedContentType)
	}
}

func (h *Host) stripPluginHeaders(header http.Header, pluginID string) {
	// 遍历原始键，同时清除插件直接写入 map 的非规范大小写。
	for key, values := range header {
		name := strings.TrimPrefix(strings.ToLower(key), strings.ToLower(http.TrailerPrefix))
		switch name {
		case "content-security-policy", "x-content-type-options":
			delete(header, key)
		case "access-control-allow-credentials":
			delete(header, key)
			slog.Warn("removed plugin credential CORS header", "plugin_id", pluginID)
		case "set-cookie":
			delete(header, key)
			for _, value := range values {
				name, _, _ := strings.Cut(value, "=")
				slog.Warn("removed plugin cookie", "plugin_id", pluginID, "cookie_name", strings.TrimSpace(name))
			}
		}
	}
}

// responseWriter 在写头前执行策略，Unwrap 保留 ResponseController 的流式能力。
// ModifyResponse 由代理在上游响应写出前调用，也覆盖协议升级直接写头的路径。
type responseWriter struct {
	http.ResponseWriter
	before func(http.Header)
	after  func(http.Header)
	modify func(*http.Response) error
	wrote  bool
}

func (w *responseWriter) WriteHeader(status int) {
	if w.wrote {
		return
	}
	w.before(w.Header())
	if status >= http.StatusOK || status == http.StatusSwitchingProtocols {
		w.wrote = true
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(value []byte) (int, error) {
	if !w.wrote {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(value)
}

func (w *responseWriter) FlushError() error {
	if !w.wrote {
		w.WriteHeader(http.StatusOK)
	}
	return http.NewResponseController(w.ResponseWriter).Flush()
}

func (w *responseWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *responseWriter) ModifyResponse(response *http.Response) error {
	return w.modify(response)
}
