package adminauth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

const (
	sessionLifetime   = 8 * time.Hour // CONTRACTS §7.4 的固定管理会话期限。
	adminCookie       = "alonabot_admin"
	secureAdminCookie = "__Host-" + adminCookie
	csrfHeader        = "X-Alonabot-CSRF"
)

// Auth 校验管理凭证、持久会话与浏览器 CSRF；同时提供派生授权的原会话检查。
// 构造后只读取配置，数据库操作由 state.Store 串行化。
type Auth struct {
	store          *state.Store
	now            func() time.Time
	baseURL        url.URL
	secure         bool
	firstTokenPath string
}

// New 使用已校验的管理 Base URL 构造认证入口。
// HTTPS 决定 Cookie 的 Secure 属性与 __Host- 前缀；now 提供会话时钟。
// 登录成功时删除 dataDir 下的首次 token 文件。
func New(st *state.Store, now func() time.Time, adminBaseURL *url.URL, dataDir string) *Auth {
	return &Auth{store: st, now: now, baseURL: *adminBaseURL, secure: adminBaseURL.Scheme == "https", firstTokenPath: filepath.Join(dataDir, FirstTokenFile)}
}

// Routes 注册登录与需认证的登出接口；装配方只在 admin host 调用这些路由。
func (a *Auth) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/login", a.login)
	mux.Handle("POST /api/auth/logout", a.Require(http.HandlerFunc(a.logout)))
}

// Require 允许当前 Bearer 凭证或有效管理 Cookie；Cookie 的变更请求须通过 CSRF 与 Origin 校验。
func (a *Auth) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		credential, err := a.credential(r.Context())
		if err != nil {
			a.respondError(w, err)
			return
		}
		ok, err := a.authorized(r, credential)
		if err != nil {
			a.respondError(w, err)
			return
		}
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Lookup 返回当前 Cookie 会话及其到期时刻，供 webhost 绑定派生授权。
// 无效凭证返回 ok=false，状态库故障用 slog 记录。
func (a *Auth) Lookup(r *http.Request) (sessionID string, expiresAt time.Time, ok bool) {
	credential, err := a.credential(r.Context())
	if err != nil {
		a.lookupError(err)
		return "", time.Time{}, false
	}
	session, ok, err := a.cookieSession(r, credential.AuthGeneration)
	if err != nil {
		a.lookupError(err)
	}
	if !ok || err != nil {
		return "", time.Time{}, false
	}
	return session.SessionID, time.UnixMilli(session.ExpiresAt), true
}

// Valid 检查会话存在、尚未到期且认证代次仍匹配，供 webhost 即时撤销派生授权。
func (a *Auth) Valid(sessionID string) bool {
	if sessionID == "" {
		return false
	}
	ctx := context.Background()
	credential, err := a.credential(ctx)
	if err != nil {
		a.lookupError(err)
		return false
	}
	_, ok, err := a.session(ctx, sessionID, credential.AuthGeneration)
	if err != nil {
		a.lookupError(err)
	}
	return ok && err == nil
}

// Reauthenticate 在请求具备有效管理认证时再校验当前管理 token，供查看 secret 前使用。
func (a *Auth) Reauthenticate(r *http.Request, token string) bool {
	credential, err := a.credential(r.Context())
	if err != nil {
		a.lookupError(err)
		return false
	}
	ok, err := a.authorized(r, credential)
	if err != nil {
		a.lookupError(err)
		return false
	}
	return ok && matchesToken(token, credential.TokenSHA256)
}

func (a *Auth) login(w http.ResponseWriter, r *http.Request) {
	// 1. 登录本身使用受控 Origin 防止跨站提交凭证。
	if !a.validOrigin(r) {
		a.loginFailed(r)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		a.loginFailed(r)
		http.Error(w, "expected application/json", http.StatusUnsupportedMediaType)
		return
	}
	credential, err := a.credential(r.Context())
	if err != nil {
		a.loginFailed(r)
		a.respondError(w, err)
		return
	}
	var input struct {
		Token string `json:"token"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&input); err != nil {
		a.loginFailed(r)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		a.loginFailed(r)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !matchesToken(input.Token, credential.TokenSHA256) {
		a.loginFailed(r)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. 固定期限和代次随会话持久化，成功后才向浏览器发出 Cookie。
	now := a.now()
	session := state.AdminSession{SessionID: randomToken(), CSRFToken: randomToken(), AuthGeneration: credential.AuthGeneration, ExpiresAt: now.Add(sessionLifetime).UnixMilli()}
	if err := a.store.CreateAdminSession(r.Context(), session, now.UnixMilli()); err != nil {
		a.loginFailed(r)
		a.respondError(w, err)
		return
	}
	// 3. 文件删除失败时撤销尚未交付的会话，保留 token 以便修复后重新登录。
	if err := os.Remove(a.firstTokenPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		a.loginFailed(r)
		a.respondError(w, errors.Join(fmt.Errorf("remove first admin token file: %w", err), a.store.DeleteAdminSession(r.Context(), session.SessionID)))
		return
	}
	cookie := a.cookie(session.SessionID)
	cookie.Expires = time.UnixMilli(session.ExpiresAt)
	http.SetCookie(w, cookie)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(struct {
		CSRFToken string `json:"csrf_token"`
	}{session.CSRFToken}); err != nil {
		slog.Error("write admin login response", "err", err)
	}
}

func (a *Auth) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(a.cookieName()); err == nil && cookie.Value != "" {
		if err := a.store.DeleteAdminSession(r.Context(), cookie.Value); err != nil {
			a.respondError(w, err)
			return
		}
	}
	cookie := a.cookie("")
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusNoContent)
}

func (a *Auth) credential(ctx context.Context) (state.AdminCredential, error) {
	credential, err := a.store.AdminCredential(ctx)
	if errors.Is(err, state.ErrNotFound) {
		return credential, ErrNotInitialized
	}
	return credential, err
}

func (a *Auth) authorized(r *http.Request, credential state.AdminCredential) (bool, error) {
	scheme, token, found := strings.Cut(r.Header.Get("Authorization"), " ")
	if found && strings.EqualFold(scheme, "Bearer") && matchesToken(token, credential.TokenSHA256) {
		return true, nil
	}
	session, ok, err := a.cookieSession(r, credential.AuthGeneration)
	if !ok || err != nil {
		return false, err
	}
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return true, nil
	}
	return a.validOrigin(r) && subtle.ConstantTimeCompare([]byte(r.Header.Get(csrfHeader)), []byte(session.CSRFToken)) == 1, nil
}

func (a *Auth) cookieSession(r *http.Request, generation int64) (state.AdminSession, bool, error) {
	cookie, err := r.Cookie(a.cookieName())
	if err != nil || cookie.Value == "" {
		return state.AdminSession{}, false, nil
	}
	return a.session(r.Context(), cookie.Value, generation)
}

func (a *Auth) session(ctx context.Context, id string, generation int64) (state.AdminSession, bool, error) {
	session, err := a.store.AdminSession(ctx, id)
	if errors.Is(err, state.ErrNotFound) {
		return session, false, nil
	}
	if err != nil {
		return session, false, err
	}
	return session, session.AuthGeneration == generation && a.now().UnixMilli() < session.ExpiresAt, nil
}

func matchesToken(token string, expected []byte) bool {
	digest := sha256.Sum256([]byte(token))
	return subtle.ConstantTimeCompare(digest[:], expected) == 1
}

func (a *Auth) validOrigin(r *http.Request) bool {
	origin, err := url.Parse(r.Header.Get("Origin"))
	if err != nil {
		return false
	}
	if origin.User != nil || origin.Path != "" || origin.RawQuery != "" || origin.Fragment != "" {
		return false
	}
	return origin.Scheme == a.baseURL.Scheme && strings.EqualFold(origin.Hostname(), a.baseURL.Hostname()) && originPort(origin) == originPort(&a.baseURL)
}

func originPort(u *url.URL) string {
	if u.Port() == "" {
		if u.Scheme == "https" {
			return "443"
		}
		if u.Scheme == "http" {
			return "80"
		}
		return ""
	}
	port, err := strconv.ParseUint(u.Port(), 10, 16)
	if err != nil {
		return ""
	}
	return strconv.FormatUint(port, 10)
}

func (a *Auth) cookieName() string {
	if a.secure {
		return secureAdminCookie
	}
	return adminCookie
}

func (a *Auth) cookie(value string) *http.Cookie {
	return &http.Cookie{Name: a.cookieName(), Value: value, Path: "/", Secure: a.secure, HttpOnly: true, SameSite: http.SameSiteStrictMode}
}

func (a *Auth) loginFailed(r *http.Request) {
	slog.Info("admin login failed", "remote_addr", r.RemoteAddr)
}

func (a *Auth) lookupError(err error) {
	if !errors.Is(err, ErrNotInitialized) {
		slog.Error("read admin authentication", "err", err)
	}
}

func (a *Auth) respondError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNotInitialized) {
		http.Error(w, ErrNotInitialized.Error(), http.StatusServiceUnavailable)
		return
	}
	slog.Error("admin authentication failed", "err", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
