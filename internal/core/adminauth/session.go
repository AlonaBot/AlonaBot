package adminauth

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func (a *Auth) currentSession(w http.ResponseWriter, r *http.Request) {
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
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	confirmed, err := a.confirmed(credential.AuthGeneration)
	if err != nil {
		a.respondError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(struct {
		BootstrapPending bool   `json:"bootstrap_pending"`
		CSRFToken        string `json:"csrf_token"`
		ExpiresAt        int64  `json:"expires_at_unix_ms"`
	}{!confirmed && a.bootstrapConnection(r), session.CSRFToken, session.ExpiresAt}); err != nil {
		slog.Error(fmt.Sprintf("Failed to write the admin Session response: %v", err), "err", err)
	}
}
