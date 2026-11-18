package adminauth

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

func (a *Auth) rotateToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	// 1. 在凭证锁外读取完整请求，管理认证和 CSRF 已由 Require 检查。
	var input struct {
		Token string `json:"token"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		rotationError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "invalid request JSON: "+err.Error())
		return
	}
	if _, err := decoder.Token(); err != io.EOF {
		rotationError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "request must contain one JSON object")
		return
	}
	// 2. 再认证与替换共用一把锁，并发请求必须核对轮换后的当前令牌。
	a.credentials.Lock()
	defer a.credentials.Unlock()
	if !a.Reauthenticate(r, input.Token) {
		rotationError(w, http.StatusForbidden, "PERMISSION_DENIED", "admin token reauthentication failed")
		return
	}
	token, err := ResetToken(r.Context(), a.store)
	if err != nil {
		a.respondError(w, err)
		return
	}
	// 3. 事务已撤销全部会话，只在本次响应交付明文令牌。
	cookie := a.cookie("")
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(struct {
		Token string `json:"token"`
	}{token}); err != nil {
		slog.Error("write admin token rotation response", "err", err)
	}
}

func rotationError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(struct {
		Code    string `json:"code"`
		Message string `json:"safe_message"`
	}{code, message}); err != nil {
		slog.Error("write admin token rotation error", "err", err)
	}
}
