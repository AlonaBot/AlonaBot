package state

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// PendingConfig 是尚未激活的配置编辑，base_generation 记录它基于的 ActiveGeneration。
type PendingConfig struct {
	PluginID       string    // 所属插件，主键
	BaseGeneration string    // 所基于的代次；插件尚无 Active 时为空串
	ConfigJSON     string    // 已规范化的完整 JSON 对象
	UpdatedAt      time.Time // 最后一次写入时间，按毫秒存取
}

// GetPendingConfig 读取插件的 Pending 配置；没有记录时返回 ErrNotFound。
func (s *Store) GetPendingConfig(ctx context.Context, pluginID string) (PendingConfig, error) {
	p := PendingConfig{PluginID: pluginID}
	var updatedAt int64
	err := s.db.QueryRowContext(ctx,
		"SELECT base_generation, config_json, updated_at FROM pending_configs WHERE plugin_id = ?", pluginID).
		Scan(&p.BaseGeneration, &p.ConfigJSON, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return p, fmt.Errorf("plugin %s: pending config: %w", pluginID, ErrNotFound)
	}
	if err != nil {
		return p, err
	}
	p.UpdatedAt = time.UnixMilli(updatedAt)
	return p, nil
}

// PutPendingConfig 覆盖写入插件的 Pending 配置。
func (s *Store) PutPendingConfig(ctx context.Context, p PendingConfig) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO pending_configs (plugin_id, base_generation, config_json, updated_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT (plugin_id) DO UPDATE SET base_generation = excluded.base_generation,
		   config_json = excluded.config_json, updated_at = excluded.updated_at`,
		p.PluginID, p.BaseGeneration, p.ConfigJSON, p.UpdatedAt.UnixMilli())
	return err
}

// DeletePendingConfig 删除插件的 Pending 配置；没有记录时不报错。
func (s *Store) DeletePendingConfig(ctx context.Context, pluginID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM pending_configs WHERE plugin_id = ?", pluginID)
	return err
}

// GenerationConfig 读取某个代次保存的配置快照；代次不存在时返回 ErrNotFound。
func (s *Store) GenerationConfig(ctx context.Context, generationID string) (string, error) {
	var config string
	err := s.db.QueryRowContext(ctx, "SELECT config_snapshot FROM generations WHERE generation_id = ?", generationID).Scan(&config)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("generation %s: %w", generationID, ErrNotFound)
	}
	return config, err
}
