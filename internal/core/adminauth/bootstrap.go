package adminauth

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

func (a *Auth) bootstrapConnection(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || !net.ParseIP(host).IsLoopback() {
		return false
	}
	target := a.baseURL
	target.Host = r.Host
	if !strings.EqualFold(target.Hostname(), "localhost") && !net.ParseIP(target.Hostname()).IsLoopback() {
		return false
	}
	if !strings.EqualFold(target.Hostname(), a.baseURL.Hostname()) || originPort(&target) != originPort(&a.baseURL) {
		return false
	}
	for name := range r.Header {
		if strings.EqualFold(name, "Forwarded") || strings.HasPrefix(strings.ToLower(name), "x-forwarded-") || strings.EqualFold(name, "X-Real-IP") {
			return false
		}
	}
	return true
}

func (a *Auth) localBootstrap(r *http.Request) bool {
	return a.bootstrapConnection(r) && a.validOrigin(r)
}

func (a *Auth) bootstrap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if !a.localBootstrap(r) {
		http.Error(w, "local same-origin connection required", http.StatusForbidden)
		return
	}
	a.credentials.Lock()
	defer a.credentials.Unlock()
	credential, err := a.credential(r.Context())
	if err != nil {
		a.respondError(w, err)
		return
	}
	confirmed, err := a.confirmed(credential.AuthGeneration)
	if err != nil {
		a.respondError(w, err)
		return
	}
	if confirmed {
		http.Error(w, "admin token already confirmed", http.StatusConflict)
		return
	}
	raw, err := os.ReadFile(a.firstTokenPath)
	if err != nil {
		a.respondError(w, fmt.Errorf("read admin token; run admin-token reset if missing: %w", err))
		return
	}
	token := strings.TrimSpace(string(raw))
	if !matchesToken(token, credential.TokenSHA256) {
		a.respondError(w, fmt.Errorf("admin token file does not match current credential; run admin-token reset"))
		return
	}
	now := a.now()
	session := state.AdminSession{SessionID: randomToken(), CSRFToken: randomToken(), AuthGeneration: credential.AuthGeneration, ExpiresAt: now.Add(sessionLifetime).UnixMilli()}
	if err := a.store.CreateAdminSession(r.Context(), session, now.UnixMilli()); err != nil {
		a.respondError(w, err)
		return
	}
	cookie := a.cookie(session.SessionID)
	cookie.Expires = time.UnixMilli(session.ExpiresAt)
	http.SetCookie(w, cookie)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(struct {
		Token     string `json:"token"`
		CSRFToken string `json:"csrf_token"`
	}{token, session.CSRFToken}); err != nil {
		slog.Error("write bootstrap response", "err", err)
	}
}

func (a *Auth) confirmBootstrap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	a.credentials.Lock()
	defer a.credentials.Unlock()
	credential, err := a.credential(r.Context())
	if err != nil {
		a.respondError(w, err)
		return
	}
	session, ok, err := a.cookieSession(r, credential.AuthGeneration)
	if err != nil {
		a.respondError(w, err)
		return
	}
	if !ok || !a.validOrigin(r) || subtle.ConstantTimeCompare([]byte(r.Header.Get(csrfHeader)), []byte(session.CSRFToken)) != 1 {
		http.Error(w, "session and CSRF required", http.StatusForbidden)
		return
	}
	if err := a.confirmGeneration(session.AuthGeneration); err != nil {
		slog.Error("confirm admin token failed", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// confirmed 将磁盘确认记录与当前代次精确比较，旧记录在轮换后自然失效。
func (a *Auth) confirmed(generation int64) (bool, error) {
	raw, err := os.ReadFile(a.firstTokenPath + ".confirmed")
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read admin token confirmation: %w", err)
	}
	confirmed, err := strconv.ParseInt(strings.TrimSpace(string(raw)), 10, 64)
	if err != nil {
		return false, fmt.Errorf("parse admin token confirmation: %w", err)
	}
	return confirmed == generation, nil
}

func (a *Auth) confirmGeneration(generation int64) (err error) {
	file, err := os.CreateTemp(filepath.Dir(a.firstTokenPath), ".admin-confirm-*")
	if err != nil {
		return fmt.Errorf("prepare admin token confirmation: %w", err)
	}
	defer func() { err = errors.Join(err, removeTemporary(file.Name())) }()
	_, writeErr := fmt.Fprintln(file, generation)
	syncErr := file.Sync()
	if err := errors.Join(writeErr, syncErr, file.Close()); err != nil {
		return fmt.Errorf("write admin token confirmation: %w", err)
	}
	return os.Rename(file.Name(), a.firstTokenPath+".confirmed")
}

func removeTemporary(name string) error {
	err := os.Remove(name)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
