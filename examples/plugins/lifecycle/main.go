// lifecycle 是可以真实运行的最小插件示例。
//
// 它由 Core 启动，走完 HELLO、REGISTER、ACTIVATE、STOPPING 的完整生命周期，
// 在每个生命周期 Hook 中用 log/slog 记录一条 INFO。SDK 把记录以 JSON 行写到 stderr，由 Core 输出。
package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

const callbackStateHeader = "X-Lifecycle-Callback-State"

// Lifecycle 在每个生命周期 Hook 里记录一条日志。
type Lifecycle struct {
	// Base 提供根插件的默认实现。
	sdk.Base
}

// OnStart 记录 OnStart。
func (*Lifecycle) OnStart(context.Context) error {
	slog.Info("OnStart")
	return nil
}

// OnReady 记录 OnReady。
func (*Lifecycle) OnReady(context.Context) error {
	slog.Info("OnReady")
	return nil
}

// OnStop 记录 OnStop。
func (*Lifecycle) OnStop(context.Context) error {
	slog.Info("OnStop")
	return nil
}

// Commands 注册 /ping 命令，回复 pong。
func (*Lifecycle) Commands(r *sdk.CommandRegistry) {
	r.Command("ping").Handle(func(*sdk.Context) (message.Chain, error) {
		return message.Chain{message.Text("pong")}, nil
	})
}

// HTTP 注册私有页面、请求头回显与以随机 state 校验的公开回调。
func (*Lifecycle) HTTP(r *sdk.HTTPRegistry) {
	state := rand.Text()
	r.Handle(http.MethodGet, "/page", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("X-Content-Type-Options", "plugin-value")
		http.SetCookie(w, &http.Cookie{Name: "lifecycle", Value: "plugin-cookie-value", Path: "/"})
		// state 只向通过 Core 授权的调用者交付，公开回调再校验来源。
		w.Header().Set(callbackStateHeader, state)
		if _, err := io.WriteString(w, "<!doctype html><html><head><title>Lifecycle</title></head><body></body></html>"); err != nil {
			slog.Error("write plugin page", "err", err)
		}
	}))
	r.Handle(http.MethodPost, "/echo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(r.Header); err != nil {
			slog.Error("write request headers", "err", err)
		}
	}))
	r.Handle(http.MethodPost, "/callback", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare([]byte(r.Header.Get(callbackStateHeader)), []byte(state)) != 1 {
			http.Error(w, "invalid callback state", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}), sdk.Public(pb.HttpAuthMode_HTTP_AUTH_MODE_CALLBACK_STATE))
}

// main 运行插件；正常结束时记录 exiting normally，失败时记录错误并以 1 退出。
func main() {
	if err := sdk.Run(&Lifecycle{}); err != nil {
		slog.Error("plugin stopped with an error", "err", err)
		os.Exit(1)
	}
	slog.Info("exiting normally")
}
