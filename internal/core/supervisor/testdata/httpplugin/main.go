// httpplugin 命令运行经 SDK 托管 HTTP 的真实集成测试插件。
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk"
)

const callbackState = "test-state"

type plugin struct {
	sdk.Base
	calls    atomic.Int32
	stopping chan struct{} // OnStop放行已开始的请求，让测试确定请求横跨停止流程
}

// HTTP 注册私有入口、带state验证的公开回调和前缀路由。
func (p *plugin) HTTP(r *sdk.HTTPRegistry) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls := p.calls.Add(1)
		if r.URL.Path == "/hook" && r.URL.Query().Get("state") != callbackState {
			http.Error(w, "invalid callback state", http.StatusUnauthorized)
			return
		}
		if r.URL.Query().Get("wait") == "1" {
			if _, err := fmt.Fprintln(w, "started"); err != nil {
				slog.Error("write HTTP start failed", "err", err)
				return
			}
			w.(http.Flusher).Flush()
			select {
			case <-p.stopping:
			case <-r.Context().Done():
				return
			}
			if _, err := fmt.Fprintln(w, "finished"); err != nil {
				slog.Error("write HTTP finish failed", "err", err)
			}
			return
		}
		switch r.URL.Path {
		case "/static/page":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if _, err := fmt.Fprint(w, "<html><body>plugin page</body></html>"); err != nil {
				slog.Error("write plugin page failed", "err", err)
			}
			return
		case "/static/cookies":
			w.Header().Add("Set-Cookie", "parent=secret; Domain=.test; Path=/")
			w.Header().Add("Set-Cookie", "alonabot_admin=secret; Path=/")
			w.Header().Add("Set-Cookie", "__Host-alonabot_admin=secret; Secure; Path=/")
			w.Header().Add("Set-Cookie", "plugin=keep; HttpOnly; Path=/")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("X-Content-Type-Options", "plugin-value")
			w.Header().Set("Content-Security-Policy", "default-src *")
		}
		w.Header().Set("Content-Type", "application/json")
		reply := struct {
			Path, Query string
			Header      http.Header
			Calls       int32
			Environment map[string]string
		}{r.URL.Path, r.URL.RawQuery, r.Header, calls, map[string]string{}}
		for _, key := range []string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY"} {
			if value, ok := os.LookupEnv(key); ok {
				reply.Environment[key] = value
			}
		}
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			slog.Error("encode HTTP reply failed", "err", err)
		}
	})
	r.Handle("GET", "/hello", handler)
	r.Handle("POST", "/hook", handler, sdk.Public(pb.HttpAuthMode_HTTP_AUTH_MODE_CALLBACK_STATE))
	r.Handle("GET", "/static/", handler, sdk.Public(pb.HttpAuthMode_HTTP_AUTH_MODE_CALLBACK_STATE))
	r.Handle("GET", "/static/private", handler)
}

// OnStop 放行等待停止的在途HTTP请求。
func (p *plugin) OnStop(context.Context) error {
	close(p.stopping)
	return nil
}

func main() {
	if err := sdk.Run(&plugin{stopping: make(chan struct{})}); err != nil {
		slog.Error("HTTP plugin failed", "err", err)
		os.Exit(1)
	}
}
