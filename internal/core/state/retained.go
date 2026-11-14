package state

import (
	"cmp"
	"context"
	"errors"
	"slices"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RetainedGenerations 保留每个插件最近提交的三个代次，包含当前 Active。
const RetainedGenerations = 3

// Retained 读取插件最近提交的代次，按提交时间从新到旧排列。
func (s *Store) Retained(ctx context.Context, pluginID string) ([]string, error) {
	if pluginID == "" {
		return nil, errEmptyPluginID
	}
	return retained(s.db.WithContext(ctx), pluginID)
}

func retained(db *gorm.DB, pluginID string) ([]string, error) {
	var rows []generation
	if err := db.Where(&generation{PluginID: pluginID}).Find(&rows).Error; err != nil {
		return nil, err
	}
	rows = slices.DeleteFunc(rows, func(row generation) bool { return row.CommittedAt == nil })
	slices.SortFunc(rows, func(a, b generation) int {
		if order := cmp.Compare(*b.CommittedAt, *a.CommittedAt); order != 0 {
			return order
		}
		if order := cmp.Compare(b.CreatedAt, a.CreatedAt); order != 0 {
			return order
		}
		return cmp.Compare(b.GenerationID, a.GenerationID)
	})
	var ids []string
	for _, row := range rows[:min(len(rows), RetainedGenerations)] {
		ids = append(ids, row.GenerationID)
	}
	return ids, nil
}

// ReferencedGenerations 读取 Active、保留代次、Pending 基线与未完成 Journal 的代次引用。
// 调用方须串行化管理操作，覆盖读取引用和回收文件的整个过程。
func (s *Store) ReferencedGenerations(ctx context.Context, pluginID string) ([]string, error) {
	if pluginID == "" {
		return nil, errEmptyPluginID
	}
	return referencedGenerations(s.db.WithContext(ctx), pluginID)
}

func referencedGenerations(db *gorm.DB, pluginID string) ([]string, error) {
	ids, err := retained(db, pluginID)
	if err != nil {
		return nil, err
	}
	var current pluginState
	if err := db.Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Take(&current).Error; err != nil {
		return nil, err
	}
	ids = append(ids, current.ActiveGeneration.String)
	var pending pendingConfig
	err = db.Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Take(&pending).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err == nil {
		ids = append(ids, pending.BaseGeneration)
	}
	var operations []journalRecord
	if err := db.Where(&journalRecord{Journal: Journal{PluginID: pluginID}}).Find(&operations).Error; err != nil {
		return nil, err
	}
	for _, operation := range operations {
		if operation.Phase != "committed" && operation.Phase != "cleaned" {
			ids = append(ids, operation.ExpectedOld, operation.NewGeneration)
		}
	}
	ids = slices.DeleteFunc(ids, func(id string) bool { return id == "" })
	slices.Sort(ids)
	return slices.Compact(ids), nil
}

// DeleteUnreferencedGeneration 复核保护引用后删除代次及 secret 路径；缺记录时成功。
// 调用方持有插件操作锁，并负责随后删除代次目录和收集孤立 secret。
func (s *Store) DeleteUnreferencedGeneration(ctx context.Context, pluginID, generationID string) error {
	if pluginID == "" {
		return errEmptyPluginID
	}
	if generationID == "" {
		return errors.New("generation_id is required")
	}
	return s.tx(ctx, func(tx *gorm.DB) error {
		ids, err := referencedGenerations(tx, pluginID)
		if err != nil {
			return err
		}
		if slices.Contains(ids, generationID) {
			return journalPrecondition("generation is protected by a retained snapshot or operation")
		}
		var row generation
		err = tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: generationID}).Take(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if row.PluginID != pluginID {
			return journalPrecondition("generation belongs to a different plugin")
		}
		// 切换代次后按钮永久失效，先清除映射，避免外键阻止淘汰快照。
		if err := tx.Where(&ButtonToken{GenerationID: generationID}).Delete(&ButtonToken{}).Error; err != nil {
			return err
		}
		if err := tx.Where(&generationSecretPath{GenerationID: generationID}).Delete(&generationSecretPath{}).Error; err != nil {
			return err
		}
		return tx.Delete(&row).Error
	})
}
