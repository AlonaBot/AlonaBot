package adminauth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/adminauth"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// TestLocalBootstrap 验证真实 SQLite、文件和管理会话跨领取、确认、重启、轮换的共同生命周期。
func TestLocalBootstrap(t *testing.T) {
	dir := t.TempDir()
	st, err := state.Open(t.Context(), dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := st.Close(); err != nil {
			t.Error(err)
		}
	}()
	token, err := adminauth.InitToken(t.Context(), st, dir)
	if err != nil {
		t.Fatal(err)
	}
	base, _ := url.Parse("http://localhost:7650")
	var mux *http.ServeMux
	rebuild := func() { mux = http.NewServeMux(); adminauth.New(st, time.Now, base, dir).Routes(mux) }
	rebuild()
	call := func(path string, cookie *http.Cookie, csrf, remote, host, origin, forwarded string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest("POST", "http://localhost:7650"+path, nil)
		req.RemoteAddr = remote
		req.Host = host
		req.Header.Set("Origin", origin)
		req.Header.Set("X-Alonabot-CSRF", csrf)
		if cookie != nil {
			req.AddCookie(cookie)
		}
		if forwarded != "" {
			req.Header.Set("Forwarded", forwarded)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		return w
	}
	claim := func() *httptest.ResponseRecorder {
		return call("/api/auth/bootstrap", nil, "", "127.0.0.1:1234", base.Host, base.String(), "")
	}
	for _, c := range []struct{ remote, host, origin, forwarded string }{
		{"192.0.2.1:42", base.Host, base.String(), ""}, {"127.0.0.1:42", "evil.test", base.String(), ""},
		{"127.0.0.1:42", base.Host, "http://evil.test", ""}, {"127.0.0.1:42", base.Host, base.String(), "for=127.0.0.1"},
	} {
		if w := call("/api/auth/bootstrap", nil, "", c.remote, c.host, c.origin, c.forwarded); w.Code != 403 {
			t.Fatalf("unsafe claim: %d %s", w.Code, w.Body.String())
		}
	}
	w := claim()
	if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("claim: %d %s", w.Code, w.Body.String())
	}
	var body struct {
		Token string `json:"token"`
		CSRF  string `json:"csrf_token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Token != token || body.CSRF == "" {
		t.Fatal("claim did not deliver current credential and CSRF")
	}
	cookie := w.Result().Cookies()[0]
	if w := claim(); w.Code != 200 {
		t.Fatalf("unconfirmed refresh: %d", w.Code)
	}
	if w := call("/api/auth/bootstrap/confirm", cookie, "wrong", "127.0.0.1:42", base.Host, base.String(), ""); w.Code != 403 {
		t.Fatal("confirmation accepted invalid CSRF")
	}
	if w := call("/api/auth/bootstrap/confirm", cookie, body.CSRF, "127.0.0.1:42", base.Host, "http://evil.test", ""); w.Code != 403 {
		t.Fatalf("confirmation accepted foreign Origin: %d", w.Code)
	}
	confirm := func() *httptest.ResponseRecorder {
		return call("/api/auth/bootstrap/confirm", cookie, body.CSRF, "127.0.0.1:42", base.Host, base.String(), "")
	}
	// 确认文件写入故障保留待确认状态和已领取 token。
	if err := os.Mkdir(filepath.Join(dir, "admin_token.confirmed"), 0700); err != nil {
		t.Fatal(err)
	}
	if w := confirm(); w.Code != 500 {
		t.Fatalf("confirmation write failure: %d", w.Code)
	}
	if err := os.Remove(filepath.Join(dir, "admin_token.confirmed")); err != nil {
		t.Fatal(err)
	}
	if w := confirm(); w.Code != 204 {
		t.Fatalf("confirm: %d %s", w.Code, w.Body.String())
	}
	rebuild()
	if w := claim(); w.Code != 409 {
		t.Fatalf("confirmed restart: %d", w.Code)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "admin_token"))
	if err != nil || strings.TrimSpace(string(raw)) != token {
		t.Fatalf("persisted token: %v", err)
	}
	rotated, err := adminauth.ResetToken(t.Context(), st, dir)
	if err != nil {
		t.Fatal(err)
	}
	if rotated == token {
		t.Fatal("reset reused token")
	}
	if w := claim(); w.Code != 200 {
		t.Fatalf("reset did not reopen claim: %d", w.Code)
	}
	if w := confirm(); w.Code != 403 {
		t.Fatalf("old session confirmed rotated token: %d", w.Code)
	}
	// 在线轮换复用同一持久化入口，文件失败保留当前凭据，成功后撤销会话。
	rotate := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", base.String()+"/api/auth/token/rotate", strings.NewReader(`{"token":"`+rotated+`"}`))
		req.Header.Set("Authorization", "Bearer "+rotated)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		return w
	}
	name := filepath.Join(dir, "admin_token")
	if err := os.Rename(name, name+".test-backup"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(name, 0700); err != nil {
		t.Fatal(err)
	}
	if w := rotate(); w.Code != 500 {
		t.Fatalf("rotation accepted invalid file: %d", w.Code)
	}
	if err := os.Remove(name); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(name+".test-backup", name); err != nil {
		t.Fatal(err)
	}
	if w := rotate(); w.Code != 200 {
		t.Fatalf("rotation failed after file repair: %d %s", w.Code, w.Body.String())
	}
	if w := claim(); w.Code != 200 {
		t.Fatalf("rotation did not reopen claim: %d", w.Code)
	}

}
