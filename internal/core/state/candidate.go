package state

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrCommittedGeneration 表示代次已提交或正被 Active 引用，清理调用方须保留其资源。
var ErrCommittedGeneration = errors.New("generation is committed or active")

// PluginGenerationConfigs 读取插件全部代次的配置快照，包含尚未提交的代次。
func (s *Store) PluginGenerationConfigs(ctx context.Context, pluginID string) (map[string]string, error) {
	if pluginID == "" {
		return nil, errEmptyPluginID
	}
	var rows []generation
	if err := s.db.WithContext(ctx).Where(&generation{PluginID: pluginID}).Find(&rows).Error; err != nil {
		return nil, err
	}
	snapshots := make(map[string]string, len(rows))
	for _, row := range rows {
		snapshots[row.GenerationID] = row.ConfigSnapshot
	}
	return snapshots, nil
}

// DeleteCandidateGeneration 在同一事务中删除插件未提交代次及其 secret 路径；缺记录时返回 nil。
// 已提交或当前 Active 返回包装 ErrCommittedGeneration 的错误，归属不符时返回普通错误。
func (s *Store) DeleteCandidateGeneration(ctx context.Context, pluginID, generationID string) error {
	if pluginID == "" {
		return errEmptyPluginID
	}
	if generationID == "" {
		return errors.New("generation_id is required")
	}
	return s.tx(ctx, func(tx *gorm.DB) error {
		// 1. 核对归属与保护状态，任何拒绝都发生在删除路径之前。
		var row generation
		err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: generationID}).Take(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if row.PluginID != pluginID {
			return fmt.Errorf("generation %s does not belong to plugin %s", generationID, pluginID)
		}
		if row.CommittedAt != nil {
			return fmt.Errorf("generation %s: %w", generationID, ErrCommittedGeneration)
		}
		var current pluginState
		if err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Take(&current).Error; err != nil {
			return err
		}
		if current.ActiveGeneration.Valid && current.ActiveGeneration.String == generationID {
			return fmt.Errorf("generation %s: %w", generationID, ErrCommittedGeneration)
		}
		// 2. 先删引用路径，再删代次；任一步失败回滚整个事务。
		if err := tx.Where(&generationSecretPath{GenerationID: generationID}).Delete(&generationSecretPath{}).Error; err != nil {
			return err
		}
		return tx.Delete(&row).Error
	})
}
