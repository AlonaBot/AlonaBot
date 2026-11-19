package manage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/config"
	"github.com/AlonaBot/AlonaBot/internal/core/profile"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// profiles 在现有目录锁与实例锁内建立管理快照；插件写入还持有相应包锁。
func (s *Service) profiles(ctx context.Context, lockedID string) (*profile.Service, error) {
	return profile.New(ctx, profile.Options{Store: s.opt.Store, Settings: s.opt.Settings, Schema: func(ctx context.Context, id string) (*configschema.Schema, error) {
		if id != lockedID {
			unlock := s.lock(id)
			defer unlock()
		}
		return s.profileSchema(ctx, id)
	}})
}

func (s *Service) profileSchema(ctx context.Context, id string) (*configschema.Schema, error) {
	active, err := s.opt.Store.Active(ctx, id)
	if err != nil {
		return nil, err
	}
	if active == "" {
		return nil, nil
	}
	raw, _, _, err := s.configResources(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, nil
	}
	schema, _ := configschema.Parse(raw) // configResources 已校验声明。
	return schema, nil
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
		var result profile.Profile
		err := s.changeProfileOverrides(r.Context(), p, "", func() error {
			var err error
			result, err = p.Create(r.Context(), request.Name, request.Color, request.CopyFrom)
			return err
		})
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
		respond(w, struct{}{}, s.changeProfileOverrides(r.Context(), p, "", func() error { return p.Delete(r.Context(), r.PathValue("id")) }))
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
		respond(w, struct{}{}, s.changeProfileOverrides(r.Context(), p, r.PathValue("plugin_id"), func() error { return p.PatchPlugin(r.Context(), r.PathValue("id"), r.PathValue("plugin_id"), request) }))
	})
	register("POST /api/v1/profiles/{id}/plugins/{plugin_id}/prune", func(p *profile.Service, w http.ResponseWriter, r *http.Request) {
		var removed []string
		err := s.changeProfileOverrides(r.Context(), p, r.PathValue("plugin_id"), func() error {
			var err error
			removed, err = p.Prune(r.Context(), r.PathValue("id"), r.PathValue("plugin_id"))
			return err
		})
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

// RegisterProfiles 将可信调用的配置文件服务接入固定服务分流器，写入复用管理互斥。
func (s *Service) RegisterProfiles(mux *router.Mux) {
	for _, name := range []string{wire.ServiceProfileList, wire.ServiceProfileCurrent, wire.ServiceProfileAssign, wire.ServiceProfileSet} {
		mux.Handle(name, s.callProfile)
	}
}

func (s *Service) callProfile(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	// 1. 固定服务使用各自的 Protobuf 载荷，先校验元数据再解码。
	var request proto.Message
	switch in.Target {
	case wire.ServiceProfileList:
		request = &pb.ProfileListRequest{}
	case wire.ServiceProfileCurrent:
		request = &pb.ProfileCurrentRequest{}
	case wire.ServiceProfileAssign:
		request = &pb.ProfileAssignRequest{}
	case wire.ServiceProfileSet:
		request = &pb.ProfileSetRequest{}
	}
	fail := func(err error) wire.Response {
		var detail *validate.Error
		if errors.As(err, &detail) {
			return wire.Response{Err: &pb.ErrorInfo{Code: detail.Code, SafeMessage: detail.Message}}
		}
		return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}}
	}
	if in.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.Payload.SchemaID != validate.SchemaID(request) || in.Payload.SchemaVersion != 1 {
		return fail(invalid("invalid profile service payload metadata"))
	}
	if err := proto.Unmarshal(in.Payload.Bytes, request); err != nil {
		return fail(invalid("decode profile service request: " + err.Error()))
	}
	// 2. 会话和管理员身份来自 Core 保留的父附件，不接受插件自报的 Target 或 Actor。
	parent, _ := in.ParentAttach.(*router.Invocation)
	if in.Target != wire.ServiceProfileList {
		if in.ParentCallID == 0 || parent == nil || parent.PluginID != caller.PluginID || parent.DefaultTarget == nil {
			return fail(precondition("profile service requires an active invocation with a default target"))
		}
		if in.Target == wire.ServiceProfileAssign || in.Target == wire.ServiceProfileSet {
			admin, err := s.opt.Policy.IsAdmin(ctx, parent.DefaultTarget.GetAdapterId(), parent.Event.GetActor())
			if err != nil {
				return fail(err)
			}
			if !admin {
				return fail(denied("profile changes require Core Admin"))
			}
		}
	}
	// 3. 查询只读快照；写操作沿原锁顺序尝试互斥，避免在途 Handler 等待管理操作排空自身。
	if in.Target == wire.ServiceProfileAssign || in.Target == wire.ServiceProfileSet {
		if !s.catalog.TryLock() {
			return fail(precondition("component management is in progress"))
		}
		defer s.catalog.Unlock()
		if !s.instances.TryLock() {
			return fail(precondition("adapter instance management is in progress"))
		}
		defer s.instances.Unlock()
	}
	lockedID := ""
	var schema *configschema.Schema
	if in.Target == wire.ServiceProfileSet {
		lockedID = caller.PluginID
		value, _ := s.packages.LoadOrStore(lockedID, &sync.Mutex{})
		mutex := value.(*sync.Mutex)
		if !mutex.TryLock() {
			return fail(precondition("plugin management is in progress: " + lockedID))
		}
		defer mutex.Unlock()
		var err error
		schema, err = s.profileSchema(ctx, lockedID)
		if err != nil {
			return fail(err)
		}
	}
	profiles, err := profile.New(ctx, profile.Options{Store: s.opt.Store, Settings: s.opt.Settings, Schema: func(_ context.Context, id string) (*configschema.Schema, error) {
		if id == lockedID {
			return schema, nil
		}
		return nil, nil
	}})
	if err != nil {
		return fail(err)
	}
	var selected profile.Profile
	if parent != nil && parent.DefaultTarget != nil {
		target := parent.DefaultTarget
		selected = profiles.Resolve(target.GetAdapterInstanceId(), profileSession(target))
	}
	summary := func() (*pb.ProfileSummary, error) {
		all, err := profiles.List(ctx)
		if err != nil {
			return nil, err
		}
		for _, p := range all {
			if p.ID == selected.ID {
				return &pb.ProfileSummary{Id: p.ID, Name: p.Name, Sessions: uint64(p.Sessions)}, nil
			}
		}
		return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "profile not found: " + selected.ID}
	}
	var response proto.Message
	switch request := request.(type) {
	case *pb.ProfileListRequest:
		all, err := profiles.List(ctx)
		if err != nil {
			return fail(err)
		}
		result := &pb.ProfileListResponse{}
		for _, p := range all {
			result.Profiles = append(result.Profiles, &pb.ProfileSummary{Id: p.ID, Name: p.Name, Sessions: uint64(p.Sessions)})
		}
		response = result
	case *pb.ProfileCurrentRequest:
		current, err := summary()
		if err != nil {
			return fail(err)
		}
		response = &pb.ProfileCurrentResponse{Profile: current}
	case *pb.ProfileAssignRequest:
		var id *string
		if request.ProfileId != "" {
			id = &request.ProfileId
		}
		if err := profiles.Assign(ctx, parent.DefaultTarget.GetAdapterInstanceId(), profileSession(parent.DefaultTarget), id); err != nil {
			return fail(err)
		}
		response = &pb.ProfileAssignResponse{}
	case *pb.ProfileSetRequest:
		if schema == nil {
			return fail(invalid("profile validation failed at " + request.Path + ": field is not marked for profiles"))
		}
		if !slices.Contains(schema.ProfilePaths(), request.Path) {
			return fail(invalid("profile validation failed at " + request.Path + ": field is not marked for profiles"))
		}
		value := json.RawMessage(request.ValueJson)
		remove := bytes.Equal(bytes.TrimSpace(value), []byte("null"))
		if !remove {
			if err := schema.ValidateProfileValue(request.Path, value); err != nil {
				return fail(err)
			}
		}
		reloading := selected.ID == profile.Default
		if reloading {
			configurations := s.configurations()
			if _, found, err := configurations.Pending(ctx, caller.PluginID); err != nil {
				return fail(err)
			} else if found {
				return fail(precondition("plugin already has pending config"))
			}
			active, err := s.opt.Store.Active(ctx, caller.PluginID)
			if err != nil {
				return fail(err)
			}
			if active == "" {
				return fail(precondition("plugin has no active generation"))
			}
			raw, err := s.opt.Store.GenerationConfig(ctx, active)
			if err != nil {
				return fail(err)
			}
			var changed []byte
			if remove {
				var patch any = json.RawMessage("null")
				parts := strings.Split(request.Path, ".")
				for i := len(parts) - 1; i >= 0; i-- {
					patch = map[string]any{parts[i]: patch}
				}
				encoded, err := json.Marshal(patch)
				if err != nil {
					return fail(err)
				}
				changed, err = config.MergePatch([]byte(raw), encoded)
				if err != nil {
					return fail(err)
				}
			} else {
				changed, err = schema.ApplyProfile(json.RawMessage(raw), map[string]json.RawMessage{request.Path: value})
				if err != nil {
					return fail(err)
				}
			}
			if _, err := configurations.Replace(ctx, caller.PluginID, active, changed); err != nil {
				return fail(err)
			}
			parent.AfterFinished(func() {
				// 父调用结束会取消请求 context；reload 沿独立管理操作完成并留存实际错误。
				go func() {
					if err := s.Reload(context.WithoutCancel(ctx), caller.PluginID, false); err != nil {
						slog.Error("profile config reload failed", "plugin_id", caller.PluginID, "error", err)
					}
				}()
			})
		} else {
			if err := profiles.PatchPlugin(ctx, selected.ID, caller.PluginID, profile.PluginPatch{Overrides: map[string]json.RawMessage{request.Path: value}}); err != nil {
				return fail(err)
			}
			if s.opt.NotifyProfileOverrides != nil {
				if err := s.opt.NotifyProfileOverrides(ctx, caller.PluginID); err != nil {
					return fail(err)
				}
			}
		}
		current, err := summary()
		if err != nil {
			return fail(err)
		}
		response = &pb.ProfileSetResponse{Profile: current, Reloading: reloading}
	}
	// 4. 返回同一服务声明的响应；异步 reload 的后续结果进入原状态与日志路径。
	encoded, err := proto.Marshal(response)
	if err != nil {
		return fail(err)
	}
	return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(response), SchemaVersion: 1, Bytes: encoded}}
}

func profileSession(target *pb.Target) string {
	escape := strings.NewReplacer("%", "%25", "#", "%23")
	session := target.GetAdapterId() + ":" + escape.Replace(target.GetChatId())
	if target.ThreadId != nil {
		session += "#" + escape.Replace(target.GetThreadId())
	}
	return session
}

func (s *Service) changeProfileOverrides(ctx context.Context, profiles *profile.Service, pluginID string, change func() error) error {
	ids := []string{pluginID}
	if pluginID == "" {
		plugins, err := profiles.Plugins(profile.Default)
		if err != nil {
			return err
		}
		ids = nil
		for _, plugin := range plugins {
			ids = append(ids, plugin.PluginID)
		}
	}
	before := make(map[string]map[string]map[string]json.RawMessage, len(ids))
	for _, id := range ids {
		before[id] = profiles.Overrides(id)
	}
	if err := change(); err != nil {
		return err
	}
	if s.opt.NotifyProfileOverrides == nil {
		return nil
	}
	var failures []error
	for _, id := range ids {
		if pluginID != "" || !reflect.DeepEqual(before[id], profiles.Overrides(id)) {
			failures = append(failures, s.opt.NotifyProfileOverrides(ctx, id))
		}
	}
	return errors.Join(failures...)
}
