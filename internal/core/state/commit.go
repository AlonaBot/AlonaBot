package state

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CommitCandidate 在同一事务中复核候选，提交 Active、COMMITTED 与 Journal 阶段。
// prepared 表示配置快照及 secret 路径已整体保存；文件持久化与 Session 激活由调用方负责。
func (s *Store) CommitCandidate(ctx context.Context, opID, pluginID, expectedOld, newGen string) error {
	var committed Journal
	err := s.tx(ctx, func(tx *gorm.DB) error {
		// 1. 复核 Active 与操作身份，拒绝并发修改后的陈旧候选。
		var current pluginState
		if err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Take(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("plugin %s: %w", pluginID, ErrNotFound)
			}
			return err
		}
		if current.ActiveGeneration.String != expectedOld {
			return fmt.Errorf("%w: expected %q, got %q", ErrStaleGeneration, expectedOld, current.ActiveGeneration.String)
		}
		var operation journalRecord
		if err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: opID}).Take(&operation).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("journal %s: %w", opID, ErrNotFound)
			}
			return err
		}
		if operation.PluginID != pluginID || operation.ExpectedOld != expectedOld || operation.NewGeneration != newGen {
			return journalPrecondition("journal operation does not match candidate identity")
		}
		if operation.Phase != "prepared" {
			return journalPrecondition("candidate journal is not prepared")
		}
		var candidate generation
		if err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: newGen}).Take(&candidate).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("generation %s: %w", newGen, ErrNotFound)
			}
			return err
		}
		if candidate.PluginID != pluginID || candidate.CommittedAt != nil {
			return journalPrecondition("candidate generation has a different owner or is already committed")
		}
		if !validate.IsJSONObject([]byte(candidate.ConfigSnapshot)) {
			return journalPrecondition("candidate config snapshot is missing or invalid")
		}

		// 2. 更新或回退的程序身份与代次一起切换，后续 restart 据此选择新程序。
		if operation.Kind == "update" || operation.Kind == "rollback" {
			var installed installation
			if err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Take(&installed).Error; err != nil {
				return err
			}
			installed.Entrypoint, installed.Version = operation.Entrypoint, operation.Version
			installed.BuildDigest, installed.PublisherKeyID = operation.BuildDigest, operation.PublisherKeyID
			if err := tx.Save(&installed).Error; err != nil {
				return err
			}
		}

		// 3. 任一元数据写入失败均回滚，事务中只访问状态库。
		now := time.Now().UnixMilli()
		candidate.CommittedAt = &now
		if err := tx.Save(&candidate).Error; err != nil {
			return err
		}
		current.ActiveGeneration = sql.NullString{String: newGen, Valid: true}
		if err := tx.Save(&current).Error; err != nil {
			return err
		}
		operation.Phase, operation.UpdatedAtMillis = "committed", now
		row := journalRow(operation.record())
		if err := tx.Save(&row).Error; err != nil {
			return err
		}
		committed = operation.record()
		return nil
	})
	if err == nil {
		logJournal(committed)
	}
	return err
}
