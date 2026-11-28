package manage

import (
	"net/http"
	"time"
)

func (s *Service) coreInfo(w http.ResponseWriter, r *http.Request) {
	version, listen, startedAt := s.opt.CoreInfo()
	var startedText string
	if !startedAt.IsZero() {
		startedText = startedAt.UTC().Format(time.RFC3339)
	}
	respond(w, struct {
		Version       string `json:"version"`
		Listen        string `json:"listen"`
		AdminAPIMajor uint32 `json:"admin_api_major"`
		StartedAt     string `json:"started_at,omitempty"`
	}{version, listen, AdminAPIMajor, startedText}, nil)
}
