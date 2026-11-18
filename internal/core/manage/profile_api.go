package manage

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/internal/core/profile"
)

// profiles 在现有目录锁与实例锁内建立管理快照；插件写入还持有相应包锁。
func (s *Service) profiles(ctx context.Context, lockedID string) (*profile.Service, error) {
	return profile.New(ctx, profile.Options{Store: s.opt.Store, Settings: s.opt.Settings, Schema: func(ctx context.Context, id string) (*configschema.Schema, error) {
		if id != lockedID {
			unlock := s.lock(id)
			defer unlock()
		}
		active, err := s.opt.Store.Active(ctx, id)
		if err != nil {
			return nil, err
		}
		if active == "" {
			return nil, nil
		}
		raw, _, err := s.configResources(ctx, id)
		if err != nil {
			return nil, err
		}
		if len(raw) == 0 {
			return nil, nil
		}
		schema, _ := configschema.Parse(raw) // configResources 已校验声明。
		return schema, nil
	}})
}

func (s *Service) profileRoutes(mux *http.ServeMux) {
	register := func(pattern string, action func(*profile.Service, http.ResponseWriter, *http.Request)) {
		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			// 与卸载保持 catalog → instances → 包锁的顺序，覆盖读取与写入整个操作。
			s.catalog.Lock()
			defer s.catalog.Unlock()
			s.instances.Lock()
			defer s.instances.Unlock()
			id := r.PathValue("plugin_id")
			if id != "" {
				unlock := s.lock(id)
				defer unlock()
			}
			profiles, err := s.profiles(r.Context(), id)
			if err != nil {
				respond(w, nil, err)
				return
			}
			action(profiles, w, r)
		})
	}
	register("GET /api/v1/profiles", func(p *profile.Service, w http.ResponseWriter, r *http.Request) {
		result, err := p.List(r.Context())
		respond(w, result, err)
	})
	register("POST /api/v1/profiles", func(p *profile.Service, w http.ResponseWriter, r *http.Request) {
		var request struct {
			Name     string `json:"name"`
			Color    string `json:"color"`
			CopyFrom string `json:"copy_from"`
		}
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		result, err := p.Create(r.Context(), request.Name, request.Color, request.CopyFrom)
		respond(w, struct {
			ID string `json:"id"`
		}{result.ID}, err)
	})
	register("PATCH /api/v1/profiles/{id}", func(p *profile.Service, w http.ResponseWriter, r *http.Request) {
		var request struct {
			Name  *string `json:"name"`
			Color *string `json:"color"`
		}
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		respond(w, struct{}{}, p.Rename(r.Context(), r.PathValue("id"), request.Name, request.Color))
	})
	register("DELETE /api/v1/profiles/{id}", func(p *profile.Service, w http.ResponseWriter, r *http.Request) {
		respond(w, struct{}{}, p.Delete(r.Context(), r.PathValue("id")))
	})
	register("GET /api/v1/profiles/assignments", func(p *profile.Service, w http.ResponseWriter, r *http.Request) { respond(w, p.Assignments(), nil) })
	register("PUT /api/v1/profiles/assignments/instances/{instance}", func(p *profile.Service, w http.ResponseWriter, r *http.Request) {
		var request *struct {
			Profile json.RawMessage `json:"profile"`
		}
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		var id *string
		if request != nil {
			var err error
			id, err = profileAssignmentID(request.Profile)
			if err != nil {
				respond(w, nil, err)
				return
			}
		}
		respond(w, struct{}{}, p.Assign(r.Context(), r.PathValue("instance"), "", id))
	})
	register("PUT /api/v1/profiles/assignments/sessions", func(p *profile.Service, w http.ResponseWriter, r *http.Request) {
		var request struct {
			Instance string          `json:"instance"`
			Session  string          `json:"session"`
			Profile  json.RawMessage `json:"profile"`
		}
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		if request.Session == "" {
			respond(w, nil, invalid("session is required"))
			return
		}
		id, err := profileAssignmentID(request.Profile)
		if err != nil {
			respond(w, nil, err)
			return
		}
		respond(w, struct{}{}, p.Assign(r.Context(), request.Instance, request.Session, id))
	})
	register("GET /api/v1/profiles/{id}/plugins", func(p *profile.Service, w http.ResponseWriter, r *http.Request) {
		result, err := p.Plugins(r.PathValue("id"))
		respond(w, result, err)
	})
	register("PATCH /api/v1/profiles/{id}/plugins/{plugin_id}", func(p *profile.Service, w http.ResponseWriter, r *http.Request) {
		var request profile.PluginPatch
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		respond(w, struct{}{}, p.PatchPlugin(r.Context(), r.PathValue("id"), r.PathValue("plugin_id"), request))
	})
	register("POST /api/v1/profiles/{id}/plugins/{plugin_id}/prune", func(p *profile.Service, w http.ResponseWriter, r *http.Request) {
		removed, err := p.Prune(r.Context(), r.PathValue("id"), r.PathValue("plugin_id"))
		respond(w, struct {
			Removed []string `json:"removed"`
		}{removed}, err)
	})
	register("GET /api/v1/profiles/{id}/platform", func(p *profile.Service, w http.ResponseWriter, r *http.Request) {
		result, err := p.Platform(r.PathValue("id"))
		respond(w, result, err)
	})
	register("PUT /api/v1/profiles/{id}/platform", func(p *profile.Service, w http.ResponseWriter, r *http.Request) {
		var request profile.Platform
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		respond(w, struct{}{}, p.PutPlatform(r.Context(), r.PathValue("id"), request))
	})
	register("GET /api/v1/profiles/{id}/tools", func(p *profile.Service, w http.ResponseWriter, r *http.Request) {
		result, err := p.Tools(r.PathValue("id"))
		respond(w, result, err)
	})
	register("PUT /api/v1/profiles/{id}/tools", func(p *profile.Service, w http.ResponseWriter, r *http.Request) {
		var request profile.ToolPolicies
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		respond(w, struct{}{}, p.PutTools(r.Context(), r.PathValue("id"), request))
	})
}

func profileAssignmentID(raw json.RawMessage) (*string, error) {
	if len(raw) == 0 {
		return nil, invalid("profile is required")
	}
	var id *string
	if err := json.Unmarshal(raw, &id); err != nil {
		return nil, invalid("profile must be a string or null")
	}
	return id, nil
}
