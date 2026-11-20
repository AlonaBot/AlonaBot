package manage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/config"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

var environmentReference = regexp.MustCompile(`^\$\{[A-Za-z_][A-Za-z0-9_]*\}$`)

// 各名称入口先检查长度，避免编译展开有界重复。
var instanceName = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// CreateInstance 补全并校验实例配置，将凭证或环境变量引用原样写入 Core 设置。
// 更换配置块的所属组件须 confirm，确认后整块丢弃旧实例与缺省选择。
func (s *Service) CreateInstance(ctx context.Context, adapterPackageID, instanceID string, raw json.RawMessage, confirm bool) error {
	if len(instanceID) > validate.MaxPluginIDLen || !instanceName.MatchString(instanceID) {
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
	groups := s.opt.Settings.Settings().Adapters
	group := groups[installed.AdapterID]
	replacing := group.Package != "" && group.Package != adapterPackageID
	if replacing && !confirm {
		affected := []string{"adapter configuration owned by: " + group.Package}
		for id := range group.Instances {
			affected = append(affected, "adapter instance: "+id)
		}
		slices.Sort(affected)
		return &ErrNeedsConfirm{Affected: affected}
	}
	for adapterID, group := range groups {
		if replacing && adapterID == installed.AdapterID {
			continue
		}
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
	// 创建新实例前检查已有实例，保留既有的显式修订入口。
	if !replacing {
		_, problems, err := s.opt.Settings.PrepareInstances(installed.AdapterID, schema)
		if err != nil {
			return failure(err)
		}
		if len(problems) != 0 {
			return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: fmt.Sprintf("adapter instances require attention: %v", problems)}
		}
	}
	err = s.opt.Settings.Update(func(doc *settings.Document) error {
		if replacing {
			if err := doc.Delete([]string{"adapters", installed.AdapterID}); err != nil {
				return err
			}
		}
		if err := doc.Set([]string{"adapters", installed.AdapterID, "package"}, adapterPackageID); err != nil {
			return err
		}
		return doc.Set([]string{"adapters", installed.AdapterID, "instances", instanceID}, original)
	})
	s.opt.Adapters.SetDefaults(s.opt.Settings.Settings().DefaultInstances())
	current := s.opt.Settings.Settings().Adapters[installed.AdapterID]
	_, created := current.Instances[instanceID]
	if err != nil && (!created || current.Package != adapterPackageID) {
		return failure(err)
	}
	// Rename 后即使目录同步报错，也必须按已发布快照完成清理并保留原错误。
	ctx = context.WithoutCancel(ctx)
	if replacing {
		instances := make([]string, 0, len(group.Instances))
		for name := range group.Instances {
			instances = append(instances, name)
		}
		err = errors.Join(err, s.forgetInstances(ctx, group.Package, instances))
	}
	err = errors.Join(err, s.opt.ForgetInstance(ctx, adapterPackageID, instanceID))
	enabled, readErr := s.opt.Store.Enabled(ctx, adapterPackageID)
	if readErr != nil {
		return failure(errors.Join(err, readErr))
	}
	if !enabled {
		// 已持有包锁；内部 enable 启动 Adapter 时会读取刚提交的全部实例。
		return failure(errors.Join(err, s.enable(ctx, installed)))
	}
	// 已运行的 Adapter 沿用实例启动入口，只应用新提交的配置。
	configuration, marshalErr := json.Marshal(expanded)
	if marshalErr != nil {
		return failure(errors.Join(err, marshalErr))
	}
	return failure(errors.Join(err, s.opt.StartInstance(ctx, adapterPackageID, instanceID, configuration)))
}

// DeleteInstance 停止仍在运行的实例，再删除设置中的实例及指向它的缺省选择。
func (s *Service) DeleteInstance(ctx context.Context, instanceID string) error {
	return failure(s.removeInstance(ctx, "", instanceID))
}

func (s *Service) removeInstance(ctx context.Context, adapterID, instanceID string) error {
	s.instances.Lock()
	defer s.instances.Unlock()
	owner, actual, err := s.instanceOwner(ctx, instanceID)
	if err != nil {
		var typed *validate.Error
		if instanceID != "" && errors.As(err, &typed) && typed.Code == pb.ErrorCode_ERROR_CODE_NOT_FOUND {
			// 删除可重入：设置已消失时，重试仍可清除遗留指定，并保持 404 响应。
			if cleanupErr := s.opt.Store.DeleteProfileAssignments(context.WithoutCancel(ctx), []string{instanceID}); cleanupErr != nil {
				return cleanupErr
			}
		}
		return err
	}
	if adapterID != "" && actual != adapterID {
		return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "adapter instance not found"}
	}
	adapterID = actual
	unlock := s.lock(owner)
	defer unlock()
	if _, ok := s.opt.Adapters.Get(instanceID); ok {
		if err := s.opt.StopInstance(ctx, owner, instanceID); err != nil {
			return err
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
	_, retained := s.opt.Settings.Settings().Adapters[adapterID].Instances[instanceID]
	if err != nil && retained {
		return err
	}
	return errors.Join(err, s.forgetInstances(ctx, owner, []string{instanceID}))
}

// forgetInstances 在设置提交后清除指定与运行期选择；重复清理同一批实例仍然安全。
func (s *Service) forgetInstances(ctx context.Context, owner string, instances []string) error {
	ctx = context.WithoutCancel(ctx) // 请求结束后仍须完成跨文件与状态库的必要清理。
	err := s.opt.Store.DeleteProfileAssignments(ctx, instances)
	for _, instance := range instances {
		err = errors.Join(err, s.opt.ForgetInstance(ctx, owner, instance))
	}
	return err
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
			err := precondition("adapter instance configuration failed: " + problem.Path + ": " + problem.Reason)
			if s.opt.RecordInstanceFailure != nil {
				return errors.Join(err, s.opt.RecordInstanceFailure(ctx, owner, instanceID, err.Error()))
			}
			return err
		}
	}
	return s.opt.StartInstance(ctx, owner, instanceID, instances[instanceID])
}

func existsInstance(id string) error { return exists("adapter instance is already running: " + id) }

// StopInstance 停止运行中的实例，本次 Core 运行中的进程恢复继续保持停用。
// 配置保留，下次 Core 启动时重新启用。
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
	if _, ok := s.opt.Adapters.Get(instanceID); ok {
		return s.opt.StopInstance(ctx, owner, instanceID)
	}
	return precondition("adapter instance is not running")
}

func (s *Service) instanceOwner(ctx context.Context, instanceID string) (string, string, error) {
	var adapterID string
	var packageID string
	for id, group := range s.opt.Settings.Settings().Adapters {
		if _, ok := group.Instances[instanceID]; ok {
			adapterID = id
			packageID = group.Package
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
		if in.Role == state.RoleAdapter && in.AdapterID == adapterID && (packageID == in.PluginID || packageID == "" && !in.Uninstalled) {
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
