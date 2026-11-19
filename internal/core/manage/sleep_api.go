package manage

import (
	"context"
	"net/http"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// SleepPolicy 返回插件覆盖、生效策略与自动休眠的准入原因。
type SleepPolicy struct {
	Policy          *string  `json:"policy"` // nil 输出 null，表示没有逐插件覆盖。
	EffectivePolicy string   `json:"effective_policy"`
	SleepCapable    bool     `json:"sleep_capable"`
	Eligible        bool     `json:"eligible"`
	Reason          string   `json:"reason"`
	Warnings        []string `json:"warnings,omitempty"`
}

type sleepEvent struct {
	Reason string `json:"reason"`
	At     string `json:"at"`
}

type sleepRuntime struct {
	Enabled      bool               `json:"enabled"`
	GenerationID string             `json:"generation_id"`
	State        state.RuntimeState `json:"state"`
	Reason       string             `json:"reason"`
	ChangedAt    string             `json:"changed_at"`
	LastWake     *sleepEvent        `json:"last_wake,omitempty"`
	LastSuspend  *sleepEvent        `json:"last_suspend,omitempty"`
}

func (s *Service) sleepRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/core/sleep", func(w http.ResponseWriter, r *http.Request) {
		sleep := s.opt.Settings.Settings().Sleep
		respond(w, struct {
			Mode        string            `json:"mode"`
			IdleTimeout string            `json:"idle_timeout"`
			Plugins     map[string]string `json:"plugins"`
		}{sleep.Mode, sleep.IdleTimeout.String(), sleep.Plugins}, nil)
	})
	mux.HandleFunc("PUT /api/v1/core/sleep", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Mode        *string `json:"mode"`
			IdleTimeout *string `json:"idle_timeout"`
		}
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		if request.Mode == nil || request.IdleTimeout == nil {
			respond(w, nil, invalid("mode and idle_timeout are required"))
			return
		}
		err := s.opt.Settings.Update(func(doc *settings.Document) error {
			if err := doc.SetSleep(*request.Mode, *request.IdleTimeout); err != nil {
				return invalid(err.Error())
			}
			return nil
		})
		respond(w, struct{}{}, err)
	})
	mux.HandleFunc("GET /api/v1/plugins/{id}/sleep-policy", func(w http.ResponseWriter, r *http.Request) {
		in, err := s.sleepInstall(r.Context(), r.PathValue("id"))
		if err != nil {
			respond(w, nil, err)
			return
		}
		if s.opt.SleepPolicy == nil {
			respond(w, nil, precondition("sleep policy is unavailable"))
			return
		}
		policy, err := s.opt.SleepPolicy(r.Context(), in)
		respond(w, policy, err)
	})
	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		mux.HandleFunc(method+" /api/v1/plugins/{id}/sleep-policy", func(w http.ResponseWriter, r *http.Request) {
			in, err := s.sleepInstall(r.Context(), r.PathValue("id"))
			if err != nil {
				respond(w, nil, err)
				return
			}
			var request struct {
				Policy *string `json:"policy"`
			}
			if method == http.MethodPut {
				if err := decode(r, &request); err != nil {
					respond(w, nil, err)
					return
				}
				if request.Policy == nil {
					respond(w, nil, invalid("policy is required"))
					return
				}
			}
			// Update 的发布者在文件锁和生命周期循环外等待恢复，此处也不持逐包锁。
			err = s.opt.Settings.Update(func(doc *settings.Document) error {
				if err := doc.SetSleepPolicy(in.PluginID, request.Policy); err != nil {
					return invalid(err.Error())
				}
				return nil
			})
			respond(w, struct{}{}, err)
		})
	}
	mux.HandleFunc("GET /api/v1/plugins/{id}/runtime", s.sleepRuntime)
}

func (s *Service) sleepInstall(ctx context.Context, id string) (state.Install, error) {
	in, err := s.opt.Store.GetInstall(ctx, id)
	if err != nil {
		return state.Install{}, err
	}
	if in.Uninstalled || in.Role != state.RolePlugin && in.Role != state.RoleAgent {
		return state.Install{}, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "plugin not found"}
	}
	return in, nil
}

func (s *Service) sleepRuntime(w http.ResponseWriter, r *http.Request) {
	in, err := s.sleepInstall(r.Context(), r.PathValue("id"))
	if err != nil {
		respond(w, nil, err)
		return
	}
	enabled, err := s.opt.Store.Enabled(r.Context(), in.PluginID)
	if err != nil {
		respond(w, nil, err)
		return
	}
	generation, err := s.opt.Store.Active(r.Context(), in.PluginID)
	if err != nil {
		respond(w, nil, err)
		return
	}
	status, err := s.opt.Store.RuntimeStatus(r.Context(), in.PluginID)
	if err != nil {
		respond(w, nil, err)
		return
	}
	result := sleepRuntime{Enabled: enabled, GenerationID: generation, State: status.State, Reason: status.Reason, ChangedAt: status.ChangedAt.UTC().Format(time.RFC3339)}
	if !status.LastWakeAt.IsZero() {
		result.LastWake = &sleepEvent{Reason: status.LastWakeReason, At: status.LastWakeAt.UTC().Format(time.RFC3339)}
	}
	if !status.LastSuspendAt.IsZero() {
		result.LastSuspend = &sleepEvent{Reason: "idle", At: status.LastSuspendAt.UTC().Format(time.RFC3339)}
	}
	respond(w, result, nil)
}
