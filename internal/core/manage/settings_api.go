package manage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/config"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

type commandPrefixes struct {
	Global     []string            `json:"command_prefixes"`
	ByInstance map[string][]string `json:"command_prefixes_by_instance"`
}

type globalDenies struct {
	Users []string `json:"users"`
	Chats []string `json:"chats"`
}

var escapeChat = strings.NewReplacer("%", "%25", "#", "%23")
var unescapeChat = strings.NewReplacer("%25", "%", "%23", "#")

func (s *Service) settingsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/core/admins", func(w http.ResponseWriter, r *http.Request) { respond(w, s.Admins(r.Context()), nil) })
	for _, method := range []string{"PUT", "DELETE"} {
		mux.HandleFunc(method+" /api/v1/core/admins", func(w http.ResponseWriter, r *http.Request) {
			var users []string
			if err := decode(r, &users); err != nil {
				respond(w, nil, err)
				return
			}
			respond(w, struct{}{}, s.updateAdmins(users, method == "DELETE"))
		})
	}
	mux.HandleFunc("GET /api/v1/core/default-locale", func(w http.ResponseWriter, r *http.Request) {
		respond(w, struct {
			Locale string `json:"default_locale"`
		}{s.opt.Settings.Settings().DefaultLocale}, nil)
	})
	mux.HandleFunc("PUT /api/v1/core/default-locale", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Locale *string `json:"default_locale"`
		}
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		if request.Locale == nil || *request.Locale == "" {
			respond(w, nil, invalid("default_locale is required"))
			return
		}
		expanded, err := expandConfiguration(*request.Locale)
		if err != nil {
			respond(w, nil, err)
			return
		}
		if expanded == "" {
			respond(w, nil, invalid("default_locale must not be empty"))
			return
		}
		err = s.opt.Settings.Update(func(doc *settings.Document) error { return doc.Set([]string{"default_locale"}, *request.Locale) })
		respond(w, struct{}{}, err)
	})
	mux.HandleFunc("GET /api/v1/core/command-prefixes", func(w http.ResponseWriter, r *http.Request) {
		current := s.opt.Settings.Settings()
		byInstance := current.CommandPrefixesByInstance
		if byInstance == nil {
			byInstance = map[string][]string{}
		}
		respond(w, commandPrefixes{current.CommandPrefixes, byInstance}, nil)
	})
	mux.HandleFunc("PUT /api/v1/core/command-prefixes", func(w http.ResponseWriter, r *http.Request) {
		var request commandPrefixes
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		if len(request.Global) == 0 || request.ByInstance == nil {
			respond(w, nil, invalid("command prefixes and instance overrides are required"))
			return
		}
		for _, prefixes := range request.ByInstance {
			if len(prefixes) == 0 {
				respond(w, nil, invalid("instance command prefixes must not be empty"))
				return
			}
			for _, prefix := range prefixes {
				if _, err := expandConfiguration(prefix); err != nil {
					respond(w, nil, err)
					return
				}
			}
		}
		for _, prefix := range request.Global {
			if _, err := expandConfiguration(prefix); err != nil {
				respond(w, nil, err)
				return
			}
		}
		err := s.opt.Settings.Update(func(doc *settings.Document) error {
			if err := doc.Set([]string{"command_prefixes"}, request.Global); err != nil {
				return err
			}
			return doc.Set([]string{"command_prefixes_by_instance"}, request.ByInstance)
		})
		respond(w, struct{}{}, err)
	})
	mux.HandleFunc("GET /api/v1/core/denies", func(w http.ResponseWriter, r *http.Request) {
		denies, err := s.readDenies(r.Context())
		respond(w, denies, err)
	})
	for _, method := range []string{"PUT", "DELETE"} {
		mux.HandleFunc(method+" /api/v1/core/denies", func(w http.ResponseWriter, r *http.Request) {
			var request globalDenies
			if err := decode(r, &request); err != nil {
				respond(w, nil, err)
				return
			}
			respond(w, struct{}{}, s.updateDenies(r.Context(), request, method == "DELETE"))
		})
	}
}

func (s *Service) updateAdmins(users []string, remove bool) error {
	if users == nil {
		return invalid("admins must be a string array")
	}
	resolved := make([]string, 0, len(users))
	for _, user := range users {
		expanded, err := expandConfiguration(user)
		if err != nil {
			return err
		}
		identity := expanded.(string)
		if _, _, err := parseSubject(identity, false); err != nil {
			return err
		}
		if !remove && slices.Contains(resolved, identity) {
			return invalid("duplicate admin: " + identity)
		}
		resolved = append(resolved, identity)
	}
	// 同 CLI 的增删共用管理员锁；整份替换也由同一个设置文件与策略发布。
	s.admins.Lock()
	defer s.admins.Unlock()
	if remove {
		// 保留未删除用户的环境变量原文，避免从有效快照回写展开后的身份。
		raw, err := os.ReadFile(filepath.Join(filepath.Dir(s.opt.Packages), settings.FileName))
		if err != nil {
			return err
		}
		var document struct {
			Admins []string `json:"admins"`
		}
		if err := json.Unmarshal(raw, &document); err != nil {
			return err
		}
		remaining := make([]string, 0, len(document.Admins))
		for _, original := range document.Admins {
			expanded, err := expandConfiguration(original)
			if err != nil {
				return err
			}
			if !slices.Contains(resolved, expanded.(string)) {
				remaining = append(remaining, original)
			}
		}
		users = remaining
	}
	err := s.opt.Settings.Update(func(doc *settings.Document) error { return doc.Set([]string{"admins"}, users) })
	admins := make([]policy.Admin, 0)
	for _, admin := range s.opt.Settings.Settings().Admins {
		admins = append(admins, policy.Admin{AdapterID: admin.AdapterID, UserID: admin.UserID})
	}
	s.opt.Policy.SetAdmins(admins)
	return err
}

func (s *Service) readDenies(ctx context.Context) (globalDenies, error) {
	entries, err := s.opt.Policy.Denies(ctx)
	result := globalDenies{Users: []string{}, Chats: []string{}}
	for _, entry := range entries {
		if entry.Kind == policy.DenyUser {
			result.Users = append(result.Users, entry.AdapterID+":"+entry.SubjectID)
		} else {
			result.Chats = append(result.Chats, entry.AdapterID+":"+escapeChat.Replace(entry.SubjectID))
		}
	}
	return result, err
}

func parseSubject(value string, chat bool) (string, string, error) {
	adapterID, subject, found := strings.Cut(value, ":")
	if !found || subject == "" {
		return "", "", invalid("subject must be <adapter_id>:<external_id>")
	}
	if _, err := validate.PluginID(adapterID); err != nil || strings.HasPrefix(adapterID, validate.MarketPrefix) {
		return "", "", invalid("invalid subject adapter ID")
	}
	if chat {
		if strings.Contains(subject, "#") {
			return "", "", invalid("chat deny must not contain a thread")
		}
		decoded := unescapeChat.Replace(subject)
		if escapeChat.Replace(decoded) != subject {
			return "", "", invalid("invalid chat string encoding")
		}
		subject = decoded
	}
	return adapterID, subject, nil
}

func (s *Service) updateDenies(ctx context.Context, request globalDenies, remove bool) error {
	if request.Users == nil || request.Chats == nil {
		return invalid("users and chats arrays are required")
	}
	wanted := make([]policy.Deny, 0, len(request.Users)+len(request.Chats))
	for _, group := range []struct {
		kind   policy.DenyKind
		values []string
	}{{policy.DenyUser, request.Users}, {policy.DenyChat, request.Chats}} {
		for _, value := range group.values {
			adapter, subject, err := parseSubject(value, group.kind == policy.DenyChat)
			if err != nil {
				return err
			}
			wanted = append(wanted, policy.Deny{Kind: group.kind, AdapterID: adapter, SubjectID: subject})
		}
	}
	s.admins.Lock()
	defer s.admins.Unlock()
	if remove {
		for _, entry := range wanted {
			if err := s.opt.Policy.Undeny(ctx, entry.Kind, entry.AdapterID, entry.SubjectID); err != nil {
				return err
			}
		}
		return nil
	}
	// 先增加限制再解除旧项；任一持久化错误立即返回，保留已经提交的实际状态。
	current, err := s.opt.Policy.Denies(ctx)
	if err != nil {
		return err
	}
	for _, entry := range wanted {
		if err := s.opt.Policy.Deny(ctx, entry.Kind, entry.AdapterID, entry.SubjectID); err != nil {
			return err
		}
	}
	for _, entry := range current {
		if !slices.Contains(wanted, entry) {
			if err := s.opt.Policy.Undeny(ctx, entry.Kind, entry.AdapterID, entry.SubjectID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) instanceRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/adapters/{adapter_id}/instance-schema", func(w http.ResponseWriter, r *http.Request) {
		owner, err := s.adapterPackage(r.Context(), r.PathValue("adapter_id"))
		if err != nil {
			respond(w, nil, err)
			return
		}
		schema, err := s.opt.InstanceSchema(r.Context(), owner)
		if err != nil {
			respond(w, nil, err)
			return
		}
		respond(w, json.RawMessage(schema.Canonical()), nil)
	})
	mux.HandleFunc("GET /api/v1/adapters/{adapter_id}/instances/{instance_id}", func(w http.ResponseWriter, r *http.Request) {
		s.instances.Lock()
		defer s.instances.Unlock()
		adapterID, instanceID := r.PathValue("adapter_id"), r.PathValue("instance_id")
		owner, err := s.adapterPackage(r.Context(), adapterID)
		if err != nil {
			respond(w, nil, err)
			return
		}
		group := s.opt.Settings.Settings().Adapters[adapterID]
		raw, found := group.Instances[instanceID]
		if !found {
			respond(w, nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "adapter instance not found"})
			return
		}
		schema, err := s.opt.InstanceSchema(r.Context(), owner)
		if err != nil {
			respond(w, nil, err)
			return
		}
		var value map[string]any
		if err := decodeConfigObject(raw, &value); err != nil {
			respond(w, nil, err)
			return
		}
		// 旧版本的秘密可能已改名，先迁移读取副本才能使用当前声明脱敏。
		if err := schema.Migrate(value, group.Version); err != nil {
			respond(w, nil, precondition(fmt.Sprintf("$.adapters[%q].instances[%q]: %s", adapterID, instanceID, err)))
			return
		}
		for _, path := range schema.SecretPaths() {
			redactSecretPath(value, path)
		}
		w.Header().Set("Cache-Control", "no-store")
		respond(w, struct {
			Config map[string]any `json:"config"`
		}{value}, nil)
	})
	mux.HandleFunc("PUT /api/v1/adapters/{adapter_id}/instances/{instance_id}", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Config json.RawMessage `json:"config"`
		}
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		respond(w, struct{}{}, s.writeInstance(r.Context(), r.PathValue("adapter_id"), r.PathValue("instance_id"), request.Config))
	})
	mux.HandleFunc("DELETE /api/v1/adapters/{adapter_id}/instances/{instance_id}", func(w http.ResponseWriter, r *http.Request) {
		respond(w, struct{}{}, s.removeInstance(r.Context(), r.PathValue("adapter_id"), r.PathValue("instance_id")))
	})
}

func (s *Service) adapterPackage(ctx context.Context, adapterID string) (string, error) {
	group := s.opt.Settings.Settings().Adapters[adapterID]
	installs, err := s.opt.Store.Installs(ctx)
	if err != nil {
		return "", err
	}
	for _, in := range installs {
		if in.Role == state.RoleAdapter && in.AdapterID == adapterID && !in.Uninstalled {
			if group.Package != "" && group.Package != in.PluginID {
				return "", precondition("adapter configuration belongs to another component")
			}
			return in.PluginID, nil
		}
	}
	return "", &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "adapter not found"}
}

func (s *Service) writeInstance(ctx context.Context, adapterID, instanceID string, raw []byte) error {
	if !instanceName.MatchString(instanceID) {
		return invalid("invalid adapter instance name")
	}
	s.instances.Lock()
	defer s.instances.Unlock()
	owner, err := s.adapterPackage(ctx, adapterID)
	if err != nil {
		return err
	}
	unlock := s.lock(owner)
	defer unlock()
	for id, group := range s.opt.Settings.Settings().Adapters {
		if _, exists := group.Instances[instanceID]; exists && id != adapterID {
			return existsInstance(instanceID)
		}
	}
	schema, err := s.opt.InstanceSchema(ctx, owner)
	if err != nil {
		return err
	}
	canonical, err := config.Canonical(raw)
	if err != nil {
		return err
	}
	var value map[string]any
	if err := decodeConfigObject(canonical, &value); err != nil {
		return err
	}
	if err := rejectSecretMarkers(value); err != nil {
		return err
	}
	// 1. 从未展开的磁盘文档保留省略的 secret，完整引用始终按原文写回。
	original, err := s.originalInstance(adapterID, instanceID)
	if err != nil {
		return err
	}
	group := s.opt.Settings.Settings().Adapters[adapterID]
	if _, exists := group.Instances[instanceID]; exists {
		if err := schema.Migrate(original, group.Version); err != nil {
			return precondition(fmt.Sprintf("$.adapters[%q].instances[%q]: %s", adapterID, instanceID, err))
		}
	}
	for _, path := range schema.SecretPaths() {
		preserveInstanceSecret(value, original, path)
	}
	schema.Complete(value)
	expanded, err := expandConfiguration(value)
	if err != nil {
		return err
	}
	if err := schema.Validate(expanded.(map[string]any)); err != nil {
		return err
	}
	// 2. 修复值和旧实例共用候选校验，成功后一次提交迁移与替换，保留迁移前备份。
	replacement, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, problems, err := s.opt.Settings.PrepareInstances(adapterID, schema, map[string]json.RawMessage{instanceID: replacement})
	if err != nil {
		return err
	}
	if len(problems) != 0 {
		return precondition(fmt.Sprintf("adapter instances require attention: %v", problems))
	}
	// 3. 提交后更新默认实例选择，运行实例仍由统一启停入口管理。
	err = s.opt.Settings.Update(func(doc *settings.Document) error {
		return doc.Set([]string{"adapters", adapterID, "package"}, owner)
	})
	s.opt.Adapters.SetDefaults(s.opt.Settings.Settings().DefaultInstances())
	if err != nil {
		return err
	}
	return s.opt.ForgetInstance(ctx, owner, instanceID)
}

func (s *Service) originalInstance(adapterID, instanceID string) (map[string]any, error) {
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(s.opt.Packages), settings.FileName))
	if err != nil {
		return nil, err
	}
	var document struct {
		Adapters map[string]struct {
			Instances map[string]json.RawMessage `json:"instances"`
		} `json:"adapters"`
	}
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	original := make(map[string]any)
	if instance, ok := document.Adapters[adapterID].Instances[instanceID]; ok {
		if err := decodeConfigObject(instance, &original); err != nil {
			return nil, err
		}
	}
	return original, nil
}

func preserveInstanceSecret(value, original any, path []string) {
	if array, ok := value.([]any); ok && path[0] == "*" {
		prior, _ := original.([]any)
		for index, child := range array {
			if index < len(prior) {
				preserveInstanceSecret(child, prior[index], path[1:])
			}
		}
		return
	}
	current, ok := value.(map[string]any)
	prior, wasObject := original.(map[string]any)
	if !ok || !wasObject {
		return
	}
	key := path[0]
	old, present := prior[key]
	if !present {
		return
	}
	next, present := current[key]
	if !present {
		current[key] = old
		return
	}
	if len(path) > 1 {
		preserveInstanceSecret(next, old, path[1:])
	}
}

func redactSecretPath(value any, path []string) {
	if len(path) == 0 {
		return
	}
	if array, ok := value.([]any); ok {
		for index, child := range array {
			if path[0] == "*" || path[0] == strconv.Itoa(index) {
				if len(path) == 1 {
					array[index] = secretState(child)
				} else {
					redactSecretPath(child, path[1:])
				}
			}
		}
		return
	}
	object, ok := value.(map[string]any)
	if !ok {
		return
	}
	child, found := object[path[0]]
	if !found {
		return
	}
	if len(path) == 1 {
		object[path[0]] = secretState(child)
		return
	}
	redactSecretPath(child, path[1:])
}

func secretState(value any) map[string]bool {
	if set, masked := secretMask(value); masked {
		return map[string]bool{"$secret_set": set}
	}
	return map[string]bool{"$secret_set": value != "" && value != nil}
}
