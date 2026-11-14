package manage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adminauth"
	"github.com/AlonaBot/AlonaBot/internal/core/config"
	"github.com/AlonaBot/AlonaBot/internal/core/options"
	"github.com/AlonaBot/AlonaBot/internal/core/secret"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"golang.org/x/text/language"
)

const configSnapshotFile = "config.json"

type pendingConfig struct {
	BaseGeneration string          `json:"base_generation"`
	Config         json.RawMessage `json:"config"`
}

type pluginConfig struct {
	ActiveGeneration string                     `json:"active_generation"`
	Active           json.RawMessage            `json:"active"`
	Pending          *pendingConfig             `json:"pending"`
	Schema           json.RawMessage            `json:"schema"`
	I18n             map[string]json.RawMessage `json:"i18n"`
}

func (s *Service) configurations() *config.Service {
	return config.New(s.opt.Store, filepath.Dir(s.opt.Packages))
}

func (s *Service) configRoutes(mux *http.ServeMux, auth *adminauth.Auth) {
	mux.HandleFunc("GET /api/v1/plugins/{id}/config", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.readConfig(r.Context(), r.PathValue("id"))
		respond(w, result, err)
	})
	for _, method := range []string{"PUT", "PATCH"} {
		mux.HandleFunc(method+" /api/v1/plugins/{id}/config/pending", func(w http.ResponseWriter, r *http.Request) {
			var request struct {
				BaseGeneration *string         `json:"base_generation"`
				Config         json.RawMessage `json:"config"`
				Patch          json.RawMessage `json:"patch"`
			}
			if err := decode(r, &request); err != nil {
				respond(w, nil, err)
				return
			}
			if request.BaseGeneration == nil {
				respond(w, nil, invalid("base_generation is required"))
				return
			}
			raw := request.Config
			if method == "PATCH" {
				raw = request.Patch
			}
			result, err := s.writeConfig(r.Context(), r.PathValue("id"), request.BaseGeneration, raw, method == "PATCH")
			respond(w, result, err)
		})
	}
	mux.HandleFunc("DELETE /api/v1/plugins/{id}/config/pending", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		unlock := s.lock(id)
		defer unlock()
		_, err := s.installed(r.Context(), id)
		if err == nil {
			err = s.configurations().Discard(r.Context(), id)
		}
		respond(w, struct{}{}, err)
	})
	mux.HandleFunc("POST /api/v1/plugins/{id}/reload", func(w http.ResponseWriter, r *http.Request) {
		respond(w, struct{}{}, s.Reload(r.Context(), r.PathValue("id")))
	})
	mux.HandleFunc("POST /api/v1/plugins/{id}/config/import", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.writeConfig(r.Context(), r.PathValue("id"), nil, nil, false)
		respond(w, result, err)
	})
	mux.HandleFunc("POST /api/v1/secrets/{plugin}/{path}/reveal", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Token string `json:"token"`
		}
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		if !auth.Reauthenticate(r, request.Token) {
			respond(w, nil, denied("secret reauthentication failed"))
			return
		}
		value, err := s.reveal(r.Context(), r.PathValue("plugin"), r.PathValue("path"))
		if err == nil {
			w.Header().Set("Cache-Control", "no-store")
			slog.Info("secret revealed", "plugin_id", r.PathValue("plugin"), "path", r.PathValue("path"))
		}
		respond(w, struct {
			Value string `json:"value"`
		}{value}, err)
	})
}

func (s *Service) configResources(ctx context.Context, id string) (json.RawMessage, map[string]json.RawMessage, error) {
	in, err := s.installed(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if in.Role == state.RoleWebUI {
		return nil, nil, invalid("webui has no config")
	}
	var raw []byte
	var files map[string][]byte
	if in.Source == "dev" {
		raw, files, err = s.opt.DevelopmentConfig(id)
		if err != nil {
			return nil, nil, err
		}
	} else {
		directory := s.directory(in)
		raw, err = os.ReadFile(filepath.Join(directory, "config_schema.json"))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, nil, err
		}
		files = make(map[string][]byte)
		entries, err := os.ReadDir(filepath.Join(directory, "i18n"))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
				continue
			}
			content, err := os.ReadFile(filepath.Join(directory, "i18n", entry.Name()))
			if err != nil {
				return nil, nil, err
			}
			files[entry.Name()] = content
		}
	}
	var schema json.RawMessage
	if len(raw) > 0 {
		parsed, problems := configschema.Parse(raw)
		if len(problems) != 0 {
			return nil, nil, invalid(fmt.Sprintf("invalid config schema: %v", problems))
		}
		schema = parsed.Canonical()
	}
	translations := make(map[string]json.RawMessage)
	for name, content := range files {
		var roots map[string]json.RawMessage
		if err := json.Unmarshal(content, &roots); err != nil {
			return nil, nil, fmt.Errorf("read config translations %s: %w", name, err)
		}
		if config, ok := roots["config"]; ok {
			translations[strings.TrimSuffix(name, ".json")] = config
		}
	}
	return schema, translations, nil
}

// configSnapshot 在包管理锁内读取 Active 与有效 Pending，写入时优先编辑 Pending。
func (s *Service) configSnapshot(ctx context.Context, id string) (active string, snapshot []byte, pending *pendingConfig, err error) {
	active, err = s.opt.Store.Active(ctx, id)
	if err != nil {
		return
	}
	snapshot = []byte("{}")
	if active != "" {
		var stored string
		stored, err = s.opt.Store.GenerationConfig(ctx, active)
		if err != nil {
			return
		}
		snapshot = []byte(stored)
	}
	stored, found, err := s.configurations().Pending(ctx, id)
	if err != nil {
		return
	}
	if found {
		pending = &pendingConfig{BaseGeneration: stored.BaseGeneration, Config: stored.ConfigJSON}
	}
	return
}

func (s *Service) readConfig(ctx context.Context, id string) (pluginConfig, error) {
	unlock := s.lock(id)
	defer unlock()
	schema, translations, err := s.configResources(ctx, id)
	if err != nil {
		return pluginConfig{}, err
	}
	active, snapshot, pending, err := s.configSnapshot(ctx, id)
	if err != nil {
		return pluginConfig{}, err
	}
	redacted, err := s.redactConfig(ctx, active, schema, snapshot)
	if err != nil {
		return pluginConfig{}, err
	}
	if pending != nil {
		pending.Config, err = s.redactConfig(ctx, active, schema, pending.Config)
		if err != nil {
			return pluginConfig{}, err
		}
	}
	return pluginConfig{active, redacted, pending, schema, translations}, nil
}

// redactConfig 同时隐藏引用和旧导入入口尚未抽取的明文 Pending，保留旧代次的秘密路径。
func (s *Service) redactConfig(ctx context.Context, active string, schema, snapshot []byte) ([]byte, error) {
	redacted, err := secret.New(s.opt.Store).Redact(snapshot)
	if err != nil {
		return nil, err
	}
	var paths [][]string
	if len(schema) != 0 {
		parsed, _ := configschema.Parse(schema) // configResources 已校验安装包中的声明。
		paths = parsed.SecretPaths()
	}
	if active != "" {
		kept, err := s.opt.Store.GenerationSecretPaths(ctx, active)
		if err != nil {
			return nil, err
		}
		for _, path := range kept {
			parts := strings.Split(path, ".")
			for index, part := range parts {
				parts[index] = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "."), "~0", "~")
			}
			paths = append(paths, parts)
		}
	}
	var value map[string]any
	if err := decodeConfigObject(redacted, &value); err != nil {
		return nil, err
	}
	for _, path := range paths {
		redactSecretPath(value, path)
	}
	return json.Marshal(value)
}

func (s *Service) writeConfig(ctx context.Context, id string, base *string, raw []byte, patch bool) (*pendingConfig, error) {
	unlock := s.lock(id)
	defer unlock()
	schema, _, err := s.configResources(ctx, id)
	if err != nil {
		return nil, err
	}
	active, previous, pending, err := s.configSnapshot(ctx, id)
	if err != nil {
		return nil, err
	}
	if base != nil && *base != active {
		return nil, &configConflict{precondition("base generation is not the current active generation")}
	}
	if pending != nil {
		previous = pending.Config
	}
	// 1. 导入也走同一秘密抽取路径，磁盘文件中的明文不会直接进入 Pending。
	if base == nil {
		raw, err = os.ReadFile(filepath.Join(filepath.Dir(s.opt.Packages), "plugins", id, configSnapshotFile))
		if errors.Is(err, os.ErrNotExist) {
			if _, development := s.opt.Development[id]; development {
				raw, err = previous, nil
			} else {
				return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "config snapshot not found"}
			}
		}
		if err != nil {
			return nil, err
		}
	}
	candidate, err := config.Canonical(raw)
	if err != nil {
		return nil, err
	}
	var submitted any
	if err := json.Unmarshal(candidate, &submitted); err != nil {
		return nil, err
	}
	if err := rejectSecretMarkers(submitted); err != nil {
		return nil, err
	}
	if _, development := s.opt.Development[id]; development {
		if base == nil {
			redacted, err := s.redactConfig(ctx, active, schema, candidate)
			return &pendingConfig{BaseGeneration: active, Config: redacted}, err
		}
		return s.writeDevelopmentConfig(ctx, id, active, schema, previous, candidate, patch)
	}
	secrets := secret.New(s.opt.Store)
	var paths []string
	if active != "" {
		paths, err = s.opt.Store.GenerationSecretPaths(ctx, active)
		if err != nil {
			return nil, err
		}
	}
	// 2. 先统一旧导入的明文和引用，保留提交对象中省略的秘密；显式 null 继续表示删除。
	plain, err := secrets.Resolve(ctx, id, previous)
	if err != nil {
		return nil, err
	}
	previous, _, err = secrets.Extract(ctx, id, schema, plain, previous, paths)
	if err != nil {
		return nil, err
	}
	var current, prior map[string]any
	if err := decodeConfigObject(candidate, &current); err != nil {
		return nil, err
	}
	if err := decodeConfigObject(previous, &prior); err != nil {
		return nil, err
	}
	preserveSecrets(current, prior)
	candidate, err = json.Marshal(current)
	if err != nil {
		return nil, err
	}
	candidate, err = secrets.Resolve(ctx, id, candidate)
	if err != nil {
		return nil, err
	}
	if patch {
		candidate, err = config.MergePatch(plain, candidate)
		if err != nil {
			return nil, err
		}
	}
	snapshot, _, err := secrets.Extract(ctx, id, schema, candidate, previous, paths)
	if err != nil {
		return nil, err
	}
	// 3. Replace 再次检查 Active，防止候选提交期间的陈旧编辑。
	saved, err := s.configurations().Replace(ctx, id, active, snapshot)
	if err != nil {
		var detail *validate.Error
		if errors.As(err, &detail) && detail.Code == pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION {
			return nil, &configConflict{err}
		}
		return nil, err
	}
	redacted, err := secrets.Redact(saved.ConfigJSON)
	return &pendingConfig{BaseGeneration: saved.BaseGeneration, Config: redacted}, err
}

// configConflict 只把配置编辑的陈旧代次映射到 409，保留其他端点已有的 412。
type configConflict struct{ error }

func (e *configConflict) Unwrap() error { return e.error }

func rejectSecretMarkers(value any) error {
	switch value := value.(type) {
	case map[string]any:
		if _, ok := value[secret.RefKey]; ok {
			return invalid("secret references cannot be submitted")
		}
		if _, ok := value["$secret_set"]; ok {
			return invalid("secret masks cannot be submitted")
		}
		for _, child := range value {
			if err := rejectSecretMarkers(child); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range value {
			if err := rejectSecretMarkers(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func decodeConfigObject(raw []byte, target *map[string]any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	return decoder.Decode(target)
}

func preserveSecrets(current, prior any) {
	switch current := current.(type) {
	case map[string]any:
		previous, _ := prior.(map[string]any)
		for key, old := range previous {
			next, present := current[key]
			if !present {
				if retained, ok := secretSubtree(old); ok {
					current[key] = retained
				}
				continue
			}
			preserveSecrets(next, old)
		}
	case []any:
		previous, _ := prior.([]any)
		for index, child := range current {
			if index < len(previous) {
				preserveSecrets(child, previous[index])
			}
		}
	}
}

// secretSubtree 只保留省略的秘密路径；整个数组省略时保留其原有行结构。
func secretSubtree(value any) (any, bool) {
	switch value := value.(type) {
	case map[string]any:
		if _, ok := value[secret.RefKey]; ok {
			return value, true
		}
		retained := make(map[string]any)
		for key, child := range value {
			if secret, ok := secretSubtree(child); ok {
				retained[key] = secret
			}
		}
		return retained, len(retained) != 0
	case []any:
		for _, child := range value {
			if _, ok := secretSubtree(child); ok {
				return value, true
			}
		}
	}
	return nil, false
}

func (s *Service) reveal(ctx context.Context, id, path string) (string, error) {
	unlock := s.lock(id)
	defer unlock()
	if _, err := s.installed(ctx, id); err != nil {
		return "", err
	}
	_, snapshot, pending, err := s.configSnapshot(ctx, id)
	if err != nil {
		return "", err
	}
	if pending != nil {
		snapshot = pending.Config
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(snapshot))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return "", err
	}
	// 路径沿 secret 路径编码，数组下标占一段，~1 与 ~0 分别还原点号与波浪线。
	parts := strings.Split(path, ".")
	for index, part := range parts {
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "."), "~0", "~")
		parts[index] = part
		switch current := value.(type) {
		case map[string]any:
			value = current[part]
		case []any:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(current) {
				return "", invalid("invalid secret path")
			}
			value = current[index]
		default:
			return "", invalid("invalid secret path")
		}
	}
	if _, development := s.opt.Development[id]; development {
		schema, _, err := s.configResources(ctx, id)
		if err != nil {
			return "", err
		}
		if len(schema) > 0 {
			parsed, _ := configschema.Parse(schema)
			for _, declared := range parsed.SecretPaths() {
				if len(declared) != len(parts) {
					continue
				}
				match := true
				for index, part := range declared {
					if part != "*" && part != parts[index] {
						match = false
						break
					}
				}
				if match {
					text, ok := value.(string)
					if ok {
						return text, nil
					}
				}
			}
		}
	}
	reference, ok := value.(map[string]any)
	if !ok {
		return "", invalid("path does not contain a secret")
	}
	ref, ok := reference[secret.RefKey].(string)
	if !ok {
		return "", invalid("path does not contain a secret")
	}
	stored, err := s.opt.Store.GetSecret(ctx, id, ref)
	return string(stored.Value), err
}

func (s *Service) optionRoutes(mux *http.ServeMux) {
	for _, method := range []string{"GET", "POST"} {
		suffix := ""
		if method == "POST" {
			suffix = "/resolve"
		}
		mux.HandleFunc(method+" /api/v1/plugins/{id}/config/options"+suffix, func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("id")
			if _, err := s.installed(r.Context(), id); err != nil {
				respond(w, nil, err)
				return
			}
			path, q, err := s.optionQuery(r)
			if err != nil {
				respond(w, nil, err)
				return
			}
			result, err := s.opt.Options.Query(r.Context(), id, options.ConfigSchema, path, q)
			respond(w, resolveOptionsPage(r, q, result), err)
		})
		mux.HandleFunc(method+" /api/v1/adapters/{adapter_id}/instances/{instance_id}/options"+suffix, func(w http.ResponseWriter, r *http.Request) {
			adapterID := r.PathValue("adapter_id")
			id, err := s.adapterPackage(r.Context(), adapterID)
			if err != nil {
				respond(w, nil, err)
				return
			}
			if _, found := s.opt.Settings.Settings().Adapters[adapterID].Instances[r.PathValue("instance_id")]; !found {
				respond(w, nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "adapter instance not found"})
				return
			}
			path, q, err := s.optionQuery(r)
			if err != nil {
				respond(w, nil, err)
				return
			}
			result, err := s.opt.Options.Query(r.Context(), id, options.AdapterInstanceSchema, path, q)
			respond(w, resolveOptionsPage(r, q, result), err)
		})
		mux.HandleFunc(method+" /api/v1/core/options/{source}"+suffix, func(w http.ResponseWriter, r *http.Request) {
			_, q, err := s.optionQuery(r)
			if err != nil {
				respond(w, nil, err)
				return
			}
			filter := make(map[string]string)
			for key, values := range r.URL.Query() {
				if strings.HasPrefix(key, "filter[") && strings.HasSuffix(key, "]") {
					filter[strings.TrimSuffix(strings.TrimPrefix(key, "filter["), "]")] = values[0]
				}
			}
			result, err := s.opt.Options.QueryBuiltin(r.Context(), r.PathValue("source"), filter, q)
			respond(w, resolveOptionsPage(r, q, result), err)
		})
	}
}

func (s *Service) optionQuery(r *http.Request) (string, options.Query, error) {
	q := options.Query{Search: r.URL.Query().Get("q"), Locale: s.opt.Settings.Settings().DefaultLocale}
	if tags, _, err := language.ParseAcceptLanguage(r.Header.Get("Accept-Language")); err == nil && len(tags) > 0 {
		q.Locale = tags[0].String()
	}
	path := r.URL.Query().Get("path")
	if r.Method == http.MethodPost {
		var request struct {
			Path   string   `json:"path"`
			Values []string `json:"values"`
		}
		if err := decode(r, &request); err != nil {
			return "", q, err
		}
		path, q.Values = request.Path, request.Values
	}
	return path, q, nil
}

// resolveOptionsPage 保留空 resolve 的空集合语义；Service 的空 Values 默认表示列出。
func resolveOptionsPage(r *http.Request, q options.Query, page options.Page) options.Page {
	if r.Method == http.MethodPost && len(q.Values) == 0 {
		return options.Page{Options: []options.Option{}}
	}
	return page
}

func (s *Service) writeDevelopmentConfig(ctx context.Context, id, active string, schema, previous, candidate []byte, patch bool) (*pendingConfig, error) {
	// 1. 复用实例秘密保留规则，校验完成后才替换开发编辑文件。
	var current, prior map[string]any
	if err := decodeConfigObject(candidate, &current); err != nil {
		return nil, err
	}
	if err := decodeConfigObject(previous, &prior); err != nil {
		return nil, err
	}
	declaration := schema
	if len(declaration) == 0 {
		declaration = []byte("{}")
	}
	parsed, _ := configschema.Parse(declaration)
	for _, path := range parsed.SecretPaths() {
		preserveInstanceSecret(current, prior, path)
	}
	candidate, err := json.Marshal(current)
	if err != nil {
		return nil, err
	}
	if patch {
		candidate, err = config.MergePatch(previous, candidate)
		if err != nil {
			return nil, err
		}
		if err := decodeConfigObject(candidate, &current); err != nil {
			return nil, err
		}
	}
	parsed.Complete(current)
	if err := parsed.Validate(current); err != nil {
		return nil, err
	}
	candidate, err = json.Marshal(current)
	if err != nil {
		return nil, err
	}
	name := filepath.Join(filepath.Dir(s.opt.Packages), "plugins", id, configSnapshotFile)
	if err := replaceDevelopmentConfig(name, candidate); err != nil {
		return nil, err
	}
	// 2. 重启失败保留已保存文件并返回完整错误，调用方可修订后再次重启。
	if err := s.opt.RestartDevelopment(ctx, id); err != nil {
		return nil, err
	}
	redacted, err := s.redactConfig(ctx, active, schema, candidate)
	return &pendingConfig{BaseGeneration: active, Config: redacted}, err
}

func replaceDevelopmentConfig(name string, content []byte) (err error) {
	temporary, err := os.CreateTemp(filepath.Dir(name), ".config-*")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, os.Remove(temporary.Name()))
		}
	}()
	_, writeErr := temporary.Write(content)
	if writeErr == nil {
		writeErr = temporary.Sync()
	}
	if err := errors.Join(writeErr, temporary.Close()); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), name)
}
