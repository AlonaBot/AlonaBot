package state

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	devSource          = "dev"
	columnGenerationID = "generation_id"
)

// Installs 按安装顺序读取全部记录，供启动装配与身份冲突检查使用。
func (s *Store) Installs(ctx context.Context) ([]Install, error) {
	var rows []installation
	if err := s.db.WithContext(ctx).Order(clause.OrderByColumn{Column: clause.Column{Name: columnInstallSeq}}).Find(&rows).Error; err != nil {
		return nil, err
	}
	installs := make([]Install, 0, len(rows))
	for _, row := range rows {
		installs = append(installs, row.Install)
	}
	return installs, nil
}

// SetDevBuild 刷新已存在的 dev 安装记录的版本与摘要，保留身份、安装顺序、启用状态与代次。
// 记录不存在时返回 ErrNotFound，来源不同返回错误。
func (s *Store) SetDevBuild(ctx context.Context, pluginID, version string, digest []byte) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		var row installation
		err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Take(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("plugin %s: %w", pluginID, ErrNotFound)
		}
		if err != nil {
			return err
		}
		if row.Source != devSource {
			return fmt.Errorf("plugin %s has source %q, expected dev", pluginID, row.Source)
		}
		row.BuildDigest = digest
		row.Version = version
		return tx.Save(&row).Error
	})
}

// SetPin 保存允许更新到的精确版本；空串解除固定。
func (s *Store) SetPin(ctx context.Context, id, version string) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		var row installation
		if err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: id}).Take(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		row.Pin = version
		return tx.Save(&row).Error
	})
}

// SetPublisherKeyID 保存管理服务已确认并验证的发布者指纹。
func (s *Store) SetPublisherKeyID(ctx context.Context, id, key string) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		result := tx.Model(&installation{}).Where(clause.Eq{Column: clause.PrimaryColumn, Value: id}).Update("publisher_key_id", key)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

type cleanupObject struct {
	model     interface{ TableName() string }
	condition clause.Expression
}

// cleanupObjects 集中列出卸载清理对象；数据与配置的删除、预览使用同一份清单。
func cleanupObjects(db *gorm.DB, id string, all bool) ([]cleanupObject, error) {
	owned := clause.Eq{Column: columnPluginID, Value: id}
	var generations []generation
	if err := db.Where(owned).Find(&generations).Error; err != nil {
		return nil, err
	}
	var journals []journalRecord
	if err := db.Where(owned).Find(&journals).Error; err != nil {
		return nil, err
	}
	for _, record := range journals {
		if record.Phase != "cleaned" && record.Phase != "committed" {
			return nil, journalPrecondition("candidate operation " + record.OpID + " requires recovery before purge")
		}
	}
	objects := []cleanupObject{{&journalRecord{}, owned}, {&ButtonToken{}, owned}, {&conversationEntry{}, owned}, {&adapterCursor{}, owned}, {&dynamicTask{}, owned}, {&scheduledJob{}, owned}}
	var installed installation
	if err := db.Where(owned).Take(&installed).Error; err != nil {
		return nil, err
	}
	if installed.Role == RoleAdapter {
		objects = append(objects, cleanupObject{&chatRecord{}, clause.Eq{Column: "adapter_id", Value: installed.AdapterID}})
	}
	for _, row := range generations {
		objects = append(objects, cleanupObject{&generationSecretPath{}, clause.Eq{Column: columnGenerationID, Value: row.GenerationID}})
	}
	objects = append(objects, cleanupObject{&generation{}, owned})
	if all {
		objects = append(objects, cleanupObject{&pendingConfig{}, owned}, cleanupObject{&Secret{}, owned},
			cleanupObject{&CommandPolicy{}, owned}, cleanupObject{&ToolPolicy{}, owned},
			cleanupObject{&adapterPermission{}, owned}, cleanupObject{&adapterRestriction{}, owned})
	}
	return objects, nil
}

// PurgeObjects 列出插件实际存在的受管数据或配置表，供管理层确认清理。
func (s *Store) PurgeObjects(ctx context.Context, id string, all bool) ([]string, error) {
	objects, err := cleanupObjects(s.db.WithContext(ctx), id, all)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, object := range objects {
		var count int64
		if err := s.db.WithContext(ctx).Model(object.model).Where(object.condition).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			name := object.model.TableName()
			if _, chat := object.model.(*chatRecord); chat {
				name = fmt.Sprintf("%s (%d)", name, count)
			}
			names = append(names, name)
		}
	}
	return names, nil
}

// Uninstall 在一个事务中停用、标记卸载并执行选定范围的状态清理。
// data 清除全部代次与运行数据；all 还清除配置和安装绑定。
func (s *Store) Uninstall(ctx context.Context, id string, data, all bool) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		var installed installation
		if err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: id}).Take(&installed).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		var current pluginState
		if err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: id}).Take(&current).Error; err != nil {
			return err
		}
		current.Enabled = false
		if data || all {
			current.ActiveGeneration = sql.NullString{}
		}
		if err := tx.Save(&current).Error; err != nil {
			return err
		}
		if data || all {
			objects, err := cleanupObjects(tx, id, all)
			if err != nil {
				return err
			}
			for _, object := range objects {
				if err := tx.Where(object.condition).Delete(object.model).Error; err != nil {
					return err
				}
			}
		}
		if all {
			if err := tx.Delete(&current).Error; err != nil {
				return err
			}
			return tx.Delete(&installed).Error
		}
		installed.Uninstalled = true
		return tx.Save(&installed).Error
	})
}
