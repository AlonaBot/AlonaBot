package webhost

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	ticketLifetime     = 60 * time.Second // CONTRACTS §7.3 的一次性票据期限
	credentialBytes    = 32
	pluginCookie       = "alonabot_plugin"
	securePluginCookie = "__Host-alonabot_plugin"
)

type ticket struct {
	pluginID, adminSessionID  string
	expiresAt, adminExpiresAt time.Time
}

type pluginSession struct {
	pluginID, adminSessionID string
	expiresAt                time.Time
}

// IssueTicket 为已在 admin 请求中验证的管理会话签发绑定插件的 60s 一次性票据。
func (h *Host) IssueTicket(adminSessionID, pluginID string) (ticketURL string, err error) {
	if !h.opt.SecondHost {
		return "", errors.New("configure plugin_base_url to open plugin pages")
	}
	if h.opt.Sessions == nil || !h.opt.Sessions.Valid(adminSessionID) {
		return "", errors.New("invalid admin session")
	}
	now := h.opt.Now()
	h.mu.Lock()
	h.expire(now)
	expires, ok := h.adminExpires[adminSessionID]
	h.mu.Unlock()
	if !ok {
		return "", errors.New("admin session has not been verified on admin host")
	}
	value, err := randomCredential()
	if err != nil {
		return "", err
	}
	h.mu.Lock()
	h.tickets[value] = ticket{pluginID: pluginID, adminSessionID: adminSessionID, expiresAt: now.Add(ticketLifetime), adminExpiresAt: expires}
	h.mu.Unlock()
	target := *h.opt.PluginBaseURL
	target.Path = ticketPath
	target.RawPath = ""
	target.Fragment = ""
	target.RawQuery = url.Values{"t": {value}, "next": {pluginPrefix + pluginID + "/"}}.Encode()
	return target.String(), nil
}

func (h *Host) exchange(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Referrer-Policy", "no-referrer")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	// 1. 在锁内取出并删除，确保并发请求最多消费一次；票据从不交给代理。
	value := r.URL.Query().Get("t")
	h.mu.Lock()
	current, ok := h.tickets[value]
	delete(h.tickets, value)
	h.mu.Unlock()
	now := h.opt.Now()
	if !ok || !now.Before(current.expiresAt) || !now.Before(current.adminExpiresAt) || !h.opt.Sessions.Valid(current.adminSessionID) {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	next := r.URL.Query().Get("next")
	if !validNext(next, current.pluginID) {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	// 2. 派生会话使用原管理会话期限，不因兑换而延长有效期。
	id, err := randomCredential()
	if err != nil {
		slog.Error("create plugin session failed", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	h.mu.Lock()
	h.sessions[id] = pluginSession{pluginID: current.pluginID, adminSessionID: current.adminSessionID, expiresAt: current.adminExpiresAt}
	h.mu.Unlock()
	name, secure := h.cookieName()
	http.SetCookie(w, &http.Cookie{Name: name, Value: id, Path: "/", Expires: current.adminExpiresAt, HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode})
	http.Redirect(w, r, next, http.StatusSeeOther)
}

// Authorize 检查插件 Cookie 的绑定与期限，并实时复核原管理会话是否有效。
func (h *Host) Authorize(r *http.Request, pluginID string) bool {
	if h.opt.Sessions == nil {
		return false
	}
	name, _ := h.cookieName()
	cookie, err := r.Cookie(name)
	if err != nil {
		return false
	}
	h.mu.Lock()
	current, ok := h.sessions[cookie.Value]
	if ok && !h.opt.Now().Before(current.expiresAt) {
		delete(h.sessions, cookie.Value)
		ok = false
	}
	h.mu.Unlock()
	if !ok || current.pluginID != pluginID {
		return false
	}
	if !h.opt.Sessions.Valid(current.adminSessionID) {
		h.mu.Lock()
		delete(h.sessions, cookie.Value)
		h.mu.Unlock()
		return false
	}
	return true
}

func (h *Host) cookieName() (string, bool) {
	if h.opt.PluginBaseURL.Scheme == "https" {
		return securePluginCookie, true
	}
	return pluginCookie, false
}

func randomCredential() (string, error) {
	var value [credentialBytes]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate plugin credential: %w", err)
	}
	return hex.EncodeToString(value[:]), nil
}

func validNext(next, pluginID string) bool {
	target, err := url.Parse(next)
	if err != nil || target.IsAbs() || target.Host != "" || strings.Contains(target.Path, "\\") {
		return false
	}
	prefix := pluginPrefix + pluginID + "/"
	// 浏览器会清理点段，检查清理前后的路径以防跳到其他插件或内部端点。
	return strings.HasPrefix(target.Path, prefix) && strings.HasPrefix(path.Clean(target.Path)+"/", prefix)
}

// expire 在记录管理会话或签发新授权时回收已过期状态，无需常驻清理 goroutine。
// 调用者持有 mu。
func (h *Host) expire(now time.Time) {
	for id, expires := range h.adminExpires {
		if !now.Before(expires) {
			delete(h.adminExpires, id)
		}
	}
	for id, current := range h.tickets {
		if !now.Before(current.expiresAt) || !now.Before(current.adminExpiresAt) {
			delete(h.tickets, id)
		}
	}
	for id, current := range h.sessions {
		if !now.Before(current.expiresAt) {
			delete(h.sessions, id)
		}
	}
}
