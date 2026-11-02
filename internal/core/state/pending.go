package state

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PendingConfig 记录尚未激活的配置编辑，BaseGeneration 是它写入时插件所处的 ActiveGeneration。
type PendingConfig struct {
	PluginID       string    // 所属插件，主键
	BaseGeneration string    // 所基于的代次；插件尚无 Active 时为空串
	ConfigJSON     string    // 已规范化的完整 JSON 对象
	UpdatedAt      time.Time // 最后一次写入时间，按毫秒存取
}

// GetPendingConfig 读取插件的 Pending 配置；没有记录时返回 ErrNotFound。
func (s *Store) GetPendingConfig(ctx context.Context, pluginID string) (PendingConfig, error) {
	p := PendingConfig{PluginID: pluginID}
	var row pendingConfig
	err := s.db.WithContext(ctx).Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return p, fmt.Errorf("plugin %s: pending config: %w", pluginID, ErrNotFound)
	}
	if err != nil {
		return p, err
	}
	p.BaseGeneration = row.BaseGeneration
	p.ConfigJSON = row.ConfigJSON
	p.UpdatedAt = time.UnixMilli(row.UpdatedAt)
	return p, nil
}

// PutPendingConfig 覆盖写入插件的 Pending 配置。
func (s *Store) PutPendingConfig(ctx context.Context, p PendingConfig) error {
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: columnPluginID}},
		DoUpdates: clause.AssignmentColumns([]string{columnBaseGeneration, columnConfigJSON, columnUpdatedAt}),
	}).Create(&pendingConfig{PluginID: p.PluginID, BaseGeneration: p.BaseGeneration, ConfigJSON: p.ConfigJSON, UpdatedAt: p.UpdatedAt.UnixMilli()}).Error
}

// DeletePendingConfig 删除插件的 Pending 配置；没有记录时不报错。
func (s *Store) DeletePendingConfig(ctx context.Context, pluginID string) error {
	return s.db.WithContext(ctx).Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Delete(&pendingConfig{}).Error
}

// GenerationConfig 读取某个代次保存的配置快照；代次不存在时返回 ErrNotFound。
func (s *Store) GenerationConfig(ctx context.Context, generationID string) (string, error) {
	var row generation
	err := s.db.WithContext(ctx).Where(clause.Eq{Column: clause.PrimaryColumn, Value: generationID}).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("generation %s: %w", generationID, ErrNotFound)
	}
	return row.ConfigSnapshot, err
}
