package manage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
)

const logDateLayout = "2006-01-02"

type logLevels struct {
	Global  string            `json:"global"`
	Sources map[string]string `json:"sources"`
}

type logLevelUpdate struct {
	Global  *string           `json:"global"`
	Sources map[string]string `json:"sources"`
}

type logFilter struct {
	source string
	level  slog.Level
	trace  string
}

func parseLogLevel(value string, source bool) (slog.Level, error) {
	switch value {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	case "off":
		if source {
			return logging.LevelOff, nil
		}
	case "inherit":
		if source {
			return 0, nil
		}
	}
	return 0, invalid("invalid log level: " + value)
}

func logQuery(r *http.Request) (logFilter, error) {
	query := r.URL.Query()
	level := query.Get("level")
	if level == "" {
		level = "debug"
	}
	parsed, err := parseLogLevel(level, false)
	return logFilter{source: query.Get("source"), level: parsed, trace: query.Get("trace")}, err
}
func (f logFilter) match(record logging.Record) bool {
	return (f.source == "" || f.source == record.Source) && record.Level >= f.level && (f.trace == "" || f.trace == record.TraceID)
}

func (s *Service) logRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/logs/recent", s.recentLogs)
	mux.HandleFunc("GET /api/v1/logs/history", s.historyLogs)
	mux.HandleFunc("GET /api/v1/logs/stream", s.streamLogs)
	mux.HandleFunc("GET /api/v1/logs/levels", func(w http.ResponseWriter, r *http.Request) {
		s.logs.Lock()
		result := s.logLevels()
		s.logs.Unlock()
		respond(w, result, nil)
	})
	mux.HandleFunc("PUT /api/v1/logs/levels", func(w http.ResponseWriter, r *http.Request) {
		var request logLevelUpdate
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		result, err := s.updateLogLevels(r.Context(), request)
		respond(w, result, err)
	})
}

func (s *Service) recentLogs(w http.ResponseWriter, r *http.Request) {
	filter, err := logQuery(r)
	if err != nil {
		respond(w, nil, err)
		return
	}
	records := s.opt.Logs.Recent(filter.source)
	records = slices.DeleteFunc(records, func(record logging.Record) bool { return !filter.match(record) })
	respond(w, records, nil)
}
func (s *Service) historyLogs(w http.ResponseWriter, r *http.Request) {
	filter, err := logQuery(r)
	if err != nil {
		respond(w, nil, err)
		return
	}
	date, err := time.ParseInLocation(logDateLayout, r.URL.Query().Get("date"), time.Local)
	if err != nil {
		respond(w, nil, invalid("date must be YYYY-MM-DD: "+err.Error()))
		return
	}
	records, err := logging.Query(filepath.Join(filepath.Dir(s.opt.Packages), "logs"), date, filter.source, filter.level, filter.trace)
	respond(w, records, err)
}
func (s *Service) streamLogs(w http.ResponseWriter, r *http.Request) {
	filter, err := logQuery(r)
	if err != nil {
		respond(w, nil, err)
		return
	}
	stream := s.opt.Logs.Subscribe(r.Context())
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	controller := http.NewResponseController(w)
	if err := controller.Flush(); err != nil {
		return
	}
	var last time.Time
	for {
		select {
		case <-r.Context().Done():
			return
		case record, ok := <-stream:
			if !ok {
				if r.Context().Err() == nil {
					encoded, err := json.Marshal(last)
					if err != nil {
						slog.Error("encode log overflow failed", "error", err)
						return
					}
					if _, err := fmt.Fprintf(w, "event: overflow\ndata: %s\n\n", encoded); err != nil {
						return
					}
					_ = controller.Flush()
				}
				return
			}
			if !filter.match(record) {
				continue
			}
			encoded, err := json.Marshal(record)
			if err != nil {
				slog.Error("encode log stream failed", "error", err)
				return
			}
			if _, err := fmt.Fprintf(w, "event: log\ndata: %s\n\n", encoded); err != nil {
				return
			}
			if err := controller.Flush(); err != nil {
				return
			}
			last = record.Time
		}
	}
}

// logLevels 在 s.logs 下读取有效全局值和文件中的来源覆盖。
func (s *Service) logLevels() logLevels {
	sources := s.opt.Settings.Settings().LogSources
	if sources == nil {
		sources = map[string]string{}
	}
	return logLevels{Global: strings.ToLower(s.opt.Logs.Levels().Source("").String()), Sources: sources}
}

func (s *Service) updateLogLevels(ctx context.Context, request logLevelUpdate) (logLevels, error) {
	if request.Global != nil {
		if _, err := parseLogLevel(*request.Global, false); err != nil {
			return logLevels{}, err
		}
	}
	for id, value := range request.Sources {
		if _, err := validate.PluginID(id); err != nil {
			return logLevels{}, invalid(err.Error())
		}
		if _, err := parseLogLevel(value, true); err != nil {
			return logLevels{}, err
		}
	}
	s.logs.Lock() // 设置提交、内存阈值与一次通知共用顺序，防止后到的请求被旧通知覆盖。
	defer s.logs.Unlock()
	before := s.opt.Settings.Settings()
	saveErr := s.opt.Settings.Update(func(doc *settings.Document) error {
		if request.Global != nil {
			if err := doc.Set([]string{"log", "level"}, *request.Global); err != nil {
				return err
			}
		}
		for id, value := range request.Sources {
			if err := doc.Set([]string{"log", "sources", id}, value); err != nil {
				return err
			}
		}
		return nil
	})
	return s.applyLogLevels(ctx, request, before, saveErr)
}

func (s *Service) applyLogLevels(ctx context.Context, request logLevelUpdate, before settings.Settings, saveErr error) (logLevels, error) {
	current := s.opt.Settings.Settings()
	// File.Update 的目录同步错误明确标示提交后失败；同值写回也可能提交并替换本次启动覆盖。
	committed := saveErr == nil || strings.HasPrefix(saveErr.Error(), "settings saved but directory sync failed:")
	if !committed && reflect.DeepEqual(before.LogSources, current.LogSources) && before.LogLevel == current.LogLevel {
		return s.logLevels(), saveErr
	}
	levels := s.opt.Logs.Levels()
	globalChanged := request.Global != nil || before.LogLevel != current.LogLevel
	if globalChanged {
		levels.SetGlobal(current.LogLevel)
	}
	// 同一快照中的全部来源一起生效，保留未显式设置的开发来源 debug。
	for id := range before.LogSources {
		if _, exists := current.LogSources[id]; !exists {
			levels.InheritSource(id)
		}
	}
	for id, value := range current.LogSources {
		if value == "inherit" {
			levels.InheritSource(id)
			continue
		}
		level, _ := parseLogLevel(value, true)
		levels.SetSource(id, level)
	}
	for id := range s.opt.Development {
		if _, exists := current.LogSources[id]; !exists {
			levels.SetSource(id, slog.LevelDebug)
		}
	}
	installs, err := s.opt.Store.Installs(ctx)
	resultErr := errors.Join(saveErr, err)
	if err != nil {
		return s.logLevels(), resultErr
	}
	ids := map[string]bool{}
	for _, install := range installs {
		if !install.Uninstalled {
			ids[install.PluginID] = true
		}
	}
	for id := range request.Sources {
		ids[id] = true
	}
	for id := range s.opt.Development {
		ids[id] = true
	}
	ordered := make([]string, 0, len(ids))
	for id := range ids {
		ordered = append(ordered, id)
	}
	slices.Sort(ordered)
	for _, id := range ordered {
		// 单独来源修改只通知该来源；全局更新只额外通知实际继承来源。
		_, explicitlyChanged := request.Sources[id]
		configured, explicit := current.LogSources[id]
		_, development := s.opt.Development[id]
		inherited := configured == "inherit" || (!explicit && !development)
		if !explicitlyChanged && !(globalChanged && inherited) && before.LogSources[id] == configured {
			continue
		}
		if err := s.opt.NotifyLogLevel(ctx, id, logging.ProtoLevel(levels.Source(id))); err != nil {
			slog.Error("notify log level failed", "plugin_id", id, "error", err)
			resultErr = errors.Join(resultErr, fmt.Errorf("notify log level for %s: %w", id, err))
		}
	}
	return s.logLevels(), resultErr
}
