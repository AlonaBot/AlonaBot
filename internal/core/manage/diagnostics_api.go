package manage

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/diagnostics"
)

func (s *Service) diagnostics(w http.ResponseWriter, r *http.Request) {
	// 1. 在生成文件之前校验参数；省略的字段使用 API 规定的缺省值。
	request := struct {
		Since      string `json:"since"`
		Logs       bool   `json:"logs"`
		Settings   bool   `json:"settings"`
		Components bool   `json:"components"`
		Redact     bool   `json:"redact"`
	}{"24h", true, true, true, true}
	if err := decode(r, &request); err != nil {
		respond(w, nil, err)
		return
	}
	since, err := time.ParseDuration(request.Since)
	if err != nil {
		respond(w, nil, invalid("invalid since: "+err.Error()))
		return
	}
	if since <= 0 {
		respond(w, nil, invalid("since must be a positive duration"))
		return
	}
	if !request.Logs && !request.Settings && !request.Components {
		respond(w, nil, invalid("select at least one diagnostic group"))
		return
	}
	version, _ := s.opt.CoreInfo()
	source := diagnostics.Running{
		Directory: diagnostics.Directory{DataDir: filepath.Dir(s.opt.Packages), Version: version},
		Store:     s.opt.Store,
		State:     s.opt.State,
		Schema: func(ctx context.Context, id string, instance bool) ([]byte, error) {
			if instance {
				schema, err := s.opt.InstanceSchema(ctx, id)
				if err != nil {
					return nil, err
				}
				return schema.Canonical(), nil
			}
			schema, _, _, err := s.configResources(ctx, id)
			return schema, err
		},
	}
	// 2. 完整生成临时归档后才发送响应，生成错误仍返回标准 JSON 错误。
	f, err := os.CreateTemp("", "alonabot-diagnostics-*.zip")
	if err != nil {
		respond(w, nil, err)
		return
	}
	defer func() {
		if err := errors.Join(f.Close(), os.Remove(f.Name())); err != nil {
			slog.Error("clean diagnostic archive failed", "error", err)
		}
	}()
	opts := diagnostics.Options{Since: since, Logs: request.Logs, Settings: request.Settings, Components: request.Components, Redact: request.Redact}
	if err := diagnostics.Write(r.Context(), f, source, opts); err != nil {
		respond(w, nil, err)
		return
	}
	name := "alonabot-diagnostics-" + time.Now().UTC().Format("20060102T150405.000000000Z") + ".zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	w.Header().Set("Cache-Control", "no-store")
	http.ServeContent(w, r, name, time.Time{}, f)
}
