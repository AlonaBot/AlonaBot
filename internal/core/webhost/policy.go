package webhost

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	singleHostCSP       = "default-src 'none'; frame-ancestors 'none'"
	configurePluginHost = "configure plugin_base_url to serve active plugin content\n"
	untypedContentType  = "application/octet-stream"
)

func (h *Host) pluginWriter(w http.ResponseWriter, r *http.Request) *responseWriter {
	rest := strings.TrimPrefix(r.URL.Path, pluginPrefix)
	pluginID, _, _ := strings.Cut(rest, "/")
	apply := func(header http.Header) { h.pluginHeaders(header, pluginID) }
	return &responseWriter{ResponseWriter: w, before: apply, modify: func(response *http.Response) error {
		apply(response.Header)
		if !h.opt.SecondHost && response.Header.Get("Content-Type") == "" {
			// 阻止公开 Listener 在写出缺失 MIME 的正文时重新嗅探成 HTML。
			response.Header.Set("Content-Type", untypedContentType)
		}
		if h.opt.SecondHost || !activeContent(response.Header.Get("Content-Type")) {
			return nil
		}
		if err := response.Body.Close(); err != nil {
			return fmt.Errorf("close blocked plugin response: %w", err)
		}
		response.StatusCode = http.StatusConflict
		response.Status = fmt.Sprintf("%d %s", http.StatusConflict, http.StatusText(http.StatusConflict))
		response.Body = io.NopCloser(strings.NewReader(configurePluginHost))
		response.ContentLength = int64(len(configurePluginHost))
		response.TransferEncoding = nil
		response.Trailer = nil
		response.Header.Del("Content-Encoding")
		response.Header.Del("Transfer-Encoding")
		response.Header.Del("Trailer")
		response.Header.Set("Content-Type", "text/plain; charset=utf-8")
		response.Header.Set("Content-Length", strconv.Itoa(len(configurePluginHost)))
		return nil
	}}
}

func (h *Host) pluginHeaders(header http.Header, pluginID string) {
	header.Set("X-Content-Type-Options", "nosniff")
	if !h.opt.SecondHost {
		header.Set("Content-Security-Policy", singleHostCSP)
	}
	if header.Get("Access-Control-Allow-Credentials") != "" {
		h.log.Warn("removed plugin credential CORS header", "plugin_id", pluginID)
	}
	header.Del("Access-Control-Allow-Credentials")
	cookies := header.Values("Set-Cookie")
	header.Del("Set-Cookie")
	for _, value := range cookies {
		// 按原始属性检查 Domain，连浏览器可能容忍的非规范 Cookie 也遵守父域限制。
		name, _, _ := strings.Cut(value, "=")
		name = strings.TrimSpace(name)
		domain := false
		for _, attribute := range strings.Split(value, ";")[1:] {
			key, _, _ := strings.Cut(attribute, "=")
			if strings.EqualFold(strings.TrimSpace(key), "Domain") {
				domain = true
			}
		}
		if domain || strings.HasPrefix(name, "alonabot_") || strings.HasPrefix(name, "__Host-alonabot_") {
			h.log.Warn("removed plugin cookie", "plugin_id", pluginID, "cookie_name", name)
			continue
		}
		header.Add("Set-Cookie", value)
	}
}

func activeContent(value string) bool {
	media, _, _ := strings.Cut(value, ";")
	media = strings.ToLower(strings.TrimSpace(media))
	switch media {
	case "text/html", "application/xhtml+xml", "image/svg+xml", "text/css", "text/javascript", "application/javascript", "application/x-javascript", "text/ecmascript", "application/ecmascript", "text/jscript", "text/livescript", "text/x-javascript", "text/x-ecmascript", "text/xml", "application/xml":
		return true
	}
	return strings.HasSuffix(media, "+xml") || strings.HasPrefix(media, "text/javascript")
}

func activeAccept(header http.Header) bool {
	for _, value := range header.Values("Accept") {
		for _, media := range strings.Split(value, ",") {
			if activeContent(media) {
				return true
			}
		}
	}
	return false
}

func pluginOrigin(origin string, base *url.URL) bool {
	source, err := url.Parse(origin)
	if err != nil || source.User != nil || source.Path != "" || source.RawQuery != "" || source.Fragment != "" {
		return false
	}
	return source.Scheme == base.Scheme && strings.EqualFold(source.Hostname(), base.Hostname()) && originPort(source) == originPort(base)
}

func originPort(u *url.URL) string {
	if port := u.Port(); port != "" {
		return port
	}
	if u.Scheme == "https" {
		return "443"
	}
	return "80"
}

// responseWriter 在写头前执行策略，Unwrap 保留 ResponseController 的流式能力。
// ModifyResponse 由代理在上游响应写出前调用，也覆盖协议升级直接写头的路径。
type responseWriter struct {
	http.ResponseWriter
	before func(http.Header)
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
