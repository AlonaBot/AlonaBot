package manage

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

type jobSummary struct {
	ID        string     `json:"job_id"`
	Trigger   string     `json:"trigger"`
	Paused    bool       `json:"paused"`
	LastRunAt *time.Time `json:"last_run_at,omitempty"`
	NextRunAt *time.Time `json:"next_run_at,omitempty"`
}

func (s *Service) jobRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/plugins/{id}/jobs", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if _, err := s.installed(r.Context(), id); err != nil {
			respond(w, nil, err)
			return
		}
		result := make([]jobSummary, 0)
		for _, job := range s.opt.Scheduler.Jobs(id) {
			item := jobSummary{ID: job.ID, Trigger: job.Trigger, Paused: job.Paused}
			if !job.LastRunAt.IsZero() {
				item.LastRunAt = &job.LastRunAt
			}
			if !job.NextRunAt.IsZero() {
				item.NextRunAt = &job.NextRunAt
			}
			result = append(result, item)
		}
		respond(w, struct {
			Jobs []jobSummary `json:"jobs"`
		}{result}, nil)
	})
}

// setJobPaused 在路径规范化前读取不透明 job_id；Require 已完成认证和 CSRF 检查。
func (s *Service) setJobPaused(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.EscapedPath(), "/"), "/")
	if len(parts) != 7 || parts[0] != "api" || parts[1] != "v1" || parts[2] != "plugins" || parts[4] != "jobs" || (parts[6] != "pause" && parts[6] != "resume") {
		return false
	}
	id, err := url.PathUnescape(parts[3])
	if err != nil {
		respond(w, nil, invalid("invalid plugin path"))
		return true
	}
	jobID, err := url.PathUnescape(parts[5])
	if err != nil {
		respond(w, nil, invalid("invalid job path"))
		return true
	}
	unlock := s.lock(id)
	defer unlock()
	if _, err := s.installed(r.Context(), id); err != nil {
		respond(w, nil, err)
		return true
	}
	respond(w, struct{}{}, s.opt.Scheduler.SetJobPaused(r.Context(), id, jobID, parts[6] == "pause"))
	return true
}
