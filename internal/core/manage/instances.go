package manage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/config"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

var environmentReference = regexp.MustCompile(`^\$\{[A-Za-z_][A-Za-z0-9_]*\}$`)
var instanceName = regexp.MustCompile(fmt.Sprintf(`^[A-Za-z0-9_-]{1,%d}$`, validate.MaxPluginIDLen))

// CreateInstance 补全并校验实例配置，将凭证或环境变量引用原样写入 Core 设置。
func (s *Service) CreateInstance(ctx context.Context, adapterPackageID, instanceID string, raw json.RawMessage) error {
	if !instanceName.MatchString(instanceID) {
		return invalid("invalid adapter instance name")
	}
	// 实例名跨 Adapter 唯一，先串行化实例编辑，再取得包锁；删除沿用同一顺序。
	s.instances.Lock()
	defer s.instances.Unlock()
	unlock := s.lock(adapterPackageID)
	defer unlock()
	installed, err := s.opt.Store.GetInstall(ctx, adapterPackageID)
	if err != nil {
		return failure(err)
	}
	if installed.Role != state.RoleAdapter || installed.Uninstalled {
		return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "component is not an installed adapter"}
	}
	for _, group := range s.opt.Settings.Settings().Adapters {
		if _, exists := group.Instances[instanceID]; exists {
			return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, Message: "adapter instance already exists"}
		}
	}
	schema, err := s.opt.InstanceSchema(ctx, adapterPackageID)
	if err != nil {
		return failure(err)
	}
	canonical, err := config.Canonical(raw)
	if err != nil {
		return err
	}
	var original map[string]any
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.UseNumber()
	if err := decoder.Decode(&original); err != nil {
		return failure(err)
	}
	schema.Complete(original)
	expanded, err := expandConfiguration(original)
	if err != nil {
		return err
	}
	if err := schema.Validate(expanded.(map[string]any)); err != nil {
		return err
	}
	// 旧实例先按已有迁移入口准备，避免新实例使旧实例的版本标记提前升级。
	_, problems, err := s.opt.Settings.PrepareInstances(installed.AdapterID, schema)
	if err != nil {
		return failure(err)
	}
	if len(problems) != 0 {
		return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: fmt.Sprintf("adapter instances require attention: %v", problems)}
	}
	err = s.opt.Settings.Update(func(doc *settings.Document) error {
		if err := doc.Set([]string{"adapters", installed.AdapterID, "$version"}, schema.Version()); err != nil {
			return err
		}
		return doc.Set([]string{"adapters", installed.AdapterID, "instances", instanceID}, original)
	})
	s.opt.Adapters.SetDefaults(s.opt.Settings.Settings().DefaultInstances())
	if err != nil {
		return failure(err)
	}
	return nil
}

// DeleteInstance 停止仍在运行的实例，再删除设置中的实例及指向它的缺省选择。
func (s *Service) DeleteInstance(ctx context.Context, instanceID string) error {
	s.instances.Lock()
	defer s.instances.Unlock()
	var adapterID string
	for id, group := range s.opt.Settings.Settings().Adapters {
		if _, exists := group.Instances[instanceID]; exists {
			adapterID = id
			break
		}
	}
	if adapterID == "" {
		return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "adapter instance not found"}
	}
	installs, err := s.opt.Store.Installs(ctx)
	if err != nil {
		return failure(err)
	}
	var owner string
	for _, in := range installs {
		if in.Role == state.RoleAdapter && in.AdapterID == adapterID {
			owner = in.PluginID
			if !in.Uninstalled {
				break
			}
		}
	}
	if owner == "" {
		return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "adapter installation record is missing"}
	}
	unlock := s.lock(owner)
	defer unlock()
	if running, ok := s.opt.Adapters.Get(instanceID); ok {
		if err := running.Stop(ctx); err != nil {
			return failure(err)
		}
	}
	defaultInstance := s.opt.Settings.Settings().Adapters[adapterID].DefaultInstance
	err = s.opt.Settings.Update(func(doc *settings.Document) error {
		if err := doc.Delete([]string{"adapters", adapterID, "instances", instanceID}); err != nil {
			return err
		}
		if defaultInstance == instanceID {
			return doc.Delete([]string{"adapters", adapterID, "default_instance"})
		}
		return nil
	})
	s.opt.Adapters.SetDefaults(s.opt.Settings.Settings().DefaultInstances())
	if err != nil {
		return failure(err)
	}
	return nil
}

// StartInstance 通过当前配置和 schema 启动实例，只影响本次运行。
func (s *Service) StartInstance(ctx context.Context, instanceID string) (err error) {
	defer finish(&err)
	s.instances.Lock()
	defer s.instances.Unlock()
	owner, adapterID, err := s.instanceOwner(ctx, instanceID)
	if err != nil {
		return err
	}
	unlock := s.lock(owner)
	defer unlock()
	if s.opt.State(owner) != "Activated" {
		return precondition("adapter is not activated")
	}
	if _, exists := s.opt.Adapters.Get(instanceID); exists {
		return existsInstance(instanceID)
	}
	schema, err := s.opt.InstanceSchema(ctx, owner)
	if err != nil {
		return err
	}
	instances, problems, err := s.opt.Settings.PrepareInstances(adapterID, schema)
	if err != nil {
		return err
	}
	for _, problem := range problems {
		if problem.InstanceID == instanceID {
			return precondition("adapter instance configuration failed: " + problem.Path + ": " + problem.Reason)
		}
	}
	return s.opt.StartInstance(ctx, owner, instanceID, instances[instanceID])
}

func existsInstance(id string) error { return exists("adapter instance is already running: " + id) }

// StopInstance 停止运行中的实例，保留配置供下次 Core 启动。
func (s *Service) StopInstance(ctx context.Context, instanceID string) (err error) {
	defer finish(&err)
	s.instances.Lock()
	defer s.instances.Unlock()
	owner, _, err := s.instanceOwner(ctx, instanceID)
	if err != nil {
		return err
	}
	unlock := s.lock(owner)
	defer unlock()
	if running, ok := s.opt.Adapters.Get(instanceID); ok {
		return running.Stop(ctx)
	}
	return precondition("adapter instance is not running")
}

func (s *Service) instanceOwner(ctx context.Context, instanceID string) (string, string, error) {
	var adapterID string
	for id, group := range s.opt.Settings.Settings().Adapters {
		if _, ok := group.Instances[instanceID]; ok {
			adapterID = id
			break
		}
	}
	if adapterID == "" {
		return "", "", &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "adapter instance not found"}
	}
	installs, err := s.opt.Store.Installs(ctx)
	if err != nil {
		return "", "", err
	}
	for _, in := range installs {
		if !in.Uninstalled && in.Role == state.RoleAdapter && in.AdapterID == adapterID {
			return in.PluginID, adapterID, nil
		}
	}
	return "", "", precondition("adapter installation record is missing")
}

// 只展开完整的环境变量引用；构造另一份对象，持久化的 original 保留原文。
func expandConfiguration(value any) (any, error) {
	switch value := value.(type) {
	case string:
		if environmentReference.MatchString(value) {
			name := strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")
			expanded, exists := os.LookupEnv(name)
			if !exists {
				return nil, invalid("environment variable " + name + " is not set")
			}
			return expanded, nil
		}
	case map[string]any:
		result := make(map[string]any, len(value))
		for key, child := range value {
			expanded, err := expandConfiguration(child)
			if err != nil {
				return nil, err
			}
			result[key] = expanded
		}
		return result, nil
	case []any:
		result := make([]any, len(value))
		for index, child := range value {
			expanded, err := expandConfiguration(child)
			if err != nil {
				return nil, err
			}
			result[index] = expanded
		}
		return result, nil
	}
	return value, nil
}
