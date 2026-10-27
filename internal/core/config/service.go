package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// Pending 是一份待激活的配置编辑，只在它所基于的代次仍是插件当前 Active 时有效。
type Pending struct {
	PluginID       string
	BaseGeneration string // 编辑所基于的代次；尚无 Active 时为空串
	ConfigJSON     []byte // 完整的规范 JSON 对象
	UpdatedAt      time.Time
}

// Service 读写插件的 Pending 配置；Active 代次的切换由 state.Store 负责。
type Service struct {
	store   *state.Store
	dataDir string
}

// New 返回使用 store 与 dataDir 下编辑快照的配置服务。
func New(store *state.Store, dataDir string) *Service {
	return &Service{store: store, dataDir: dataDir}
}

// Pending 返回插件的有效 Pending；base 已不等于当前 Active 时视为不存在。
func (s *Service) Pending(ctx context.Context, pluginID string) (Pending, bool, error) {
	unlock := s.store.LockPlugin(pluginID)
	defer unlock()
	active, err := s.active(ctx, pluginID)
	if err != nil {
		return Pending{}, false, err
	}
	return s.pending(ctx, pluginID, active)
}

// Replace 用 full 完整替换 Pending；baseGeneration 不是当前 Active 时拒绝陈旧提交。
func (s *Service) Replace(ctx context.Context, pluginID, baseGeneration string, full []byte) (Pending, error) {
	unlock := s.store.LockPlugin(pluginID)
	defer unlock()
	if err := s.checkBase(ctx, pluginID, baseGeneration); err != nil {
		return Pending{}, err
	}
	configJSON, err := Canonical(full)
	if err != nil {
		return Pending{}, err
	}
	return s.put(ctx, pluginID, baseGeneration, configJSON)
}

// Patch 按 RFC 7396 把 patch 合并到有效 Pending，没有 Pending 时合并到 Active 的配置；
// 插件尚无 Active 时以空对象为基础。
func (s *Service) Patch(ctx context.Context, pluginID, baseGeneration string, patch []byte) (Pending, error) {
	unlock := s.store.LockPlugin(pluginID)
	defer unlock()
	if err := s.checkBase(ctx, pluginID, baseGeneration); err != nil {
		return Pending{}, err
	}
	base, err := s.patchBase(ctx, pluginID, baseGeneration)
	if err != nil {
		return Pending{}, err
	}
	configJSON, err := MergePatch(base, patch)
	if err != nil {
		return Pending{}, err
	}
	return s.put(ctx, pluginID, baseGeneration, configJSON)
}

// ImportSnapshot 把 <dataDir>/plugins/<id>/config.json 导入为以当前 Active 为基础的 Pending；
// 不改变 Active。文件不存在返回 NOT_FOUND，内容不是 JSON 对象返回 INVALID_ARGUMENT 且不改已有 Pending。
func (s *Service) ImportSnapshot(ctx context.Context, pluginID string) (Pending, error) {
	unlock := s.store.LockPlugin(pluginID)
	defer unlock()
	active, err := s.active(ctx, pluginID)
	if err != nil {
		return Pending{}, err
	}
	path := filepath.Join(s.dataDir, "plugins", pluginID, "config.json")
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Pending{}, failure(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "config snapshot %s does not exist", path)
	} else if err != nil {
		return Pending{}, failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "read config snapshot %s: %v", path, err)
	}
	configJSON, err := Canonical(raw)
	if err != nil {
		return Pending{}, err
	}
	return s.put(ctx, pluginID, active, configJSON)
}

// Discard 删除插件的 Pending；没有 Pending 时什么也不做。
func (s *Service) Discard(ctx context.Context, pluginID string) error {
	unlock := s.store.LockPlugin(pluginID)
	defer unlock()
	if err := s.store.DeletePendingConfig(ctx, pluginID); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "delete pending config for plugin %s: %v", pluginID, err)
	}
	return nil
}

// active 返回插件当前 Active 代次，未激活时为空串。
func (s *Service) active(ctx context.Context, pluginID string) (string, error) {
	generation, err := s.store.Active(ctx, pluginID)
	if errors.Is(err, state.ErrNotFound) {
		return "", failure(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "plugin %s: %v", pluginID, err)
	} else if err != nil {
		return "", failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "read active generation of plugin %s: %v", pluginID, err)
	}
	return generation, nil
}

// checkBase 要求 baseGeneration 等于当前 Active，否则说明调用方基于旧代次。
func (s *Service) checkBase(ctx context.Context, pluginID, baseGeneration string) error {
	active, err := s.active(ctx, pluginID)
	if err != nil {
		return err
	}
	if baseGeneration != active {
		return failure(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION,
			"base generation %q is not the current active generation %q of plugin %s", baseGeneration, active, pluginID)
	}
	return nil
}

// pending 读取 Pending，并按 active 判断它是否仍然有效。
func (s *Service) pending(ctx context.Context, pluginID, active string) (Pending, bool, error) {
	stored, err := s.store.GetPendingConfig(ctx, pluginID)
	if errors.Is(err, state.ErrNotFound) {
		return Pending{}, false, nil
	} else if err != nil {
		return Pending{}, false, failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "read pending config for plugin %s: %v", pluginID, err)
	}
	if stored.BaseGeneration != active {
		return Pending{}, false, nil
	}
	return Pending{
		PluginID:       stored.PluginID,
		BaseGeneration: stored.BaseGeneration,
		ConfigJSON:     []byte(stored.ConfigJSON),
		UpdatedAt:      stored.UpdatedAt,
	}, true, nil
}

// patchBase 给出 Patch 的基础文档：有效 Pending，否则当前 Active 的配置，否则空对象。
func (s *Service) patchBase(ctx context.Context, pluginID, active string) ([]byte, error) {
	if pending, found, err := s.pending(ctx, pluginID, active); err != nil {
		return nil, err
	} else if found {
		return pending.ConfigJSON, nil
	}
	if active == "" {
		return []byte("{}"), nil
	}
	snapshot, err := s.store.GenerationConfig(ctx, active)
	if err != nil {
		return nil, failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "read config of generation %s: %v", active, err)
	}
	return []byte(snapshot), nil
}

// put 写入 Pending，updated_at 取当前时间。
func (s *Service) put(ctx context.Context, pluginID, baseGeneration string, configJSON []byte) (Pending, error) {
	updatedAt := time.Now()
	stored := state.PendingConfig{
		PluginID:       pluginID,
		BaseGeneration: baseGeneration,
		ConfigJSON:     string(configJSON),
		UpdatedAt:      updatedAt,
	}
	if err := s.store.PutPendingConfig(ctx, stored); err != nil {
		return Pending{}, failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "write pending config for plugin %s: %v", pluginID, err)
	}
	return Pending{
		PluginID:       pluginID,
		BaseGeneration: baseGeneration,
		ConfigJSON:     configJSON,
		UpdatedAt:      updatedAt,
	}, nil
}

// failure 构造带协议错误码的 *validate.Error；说明保留原始错误文本。
func failure(code pb.ErrorCode, format string, args ...any) error {
	return &validate.Error{Code: code, Message: fmt.Sprintf(format, args...)}
}
