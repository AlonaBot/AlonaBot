package state

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type adapterRestriction struct {
	PluginID string `gorm:"column:plugin_id;primaryKey"`
}

func (adapterRestriction) TableName() string { return "plugin_adapter_restriction" }

// AdapterInstanceAllowed 按当前许可判定目标实例；插件未收紧时允许全部实例。
func (s *Store) AdapterInstanceAllowed(ctx context.Context, pluginID, instanceID string) (bool, error) {
	all, instances, err := s.AdapterInstanceAccess(ctx, pluginID)
	if err != nil {
		return false, err
	}
	return all || slices.Contains(instances, instanceID), nil
}

// AdapterInstanceAccess 在同一快照中读取插件的许可；全部实例时返回 nil 集合，否则按字典序。
func (s *Store) AdapterInstanceAccess(ctx context.Context, pluginID string) (all bool, instances []string, err error) {
	if pluginID == "" {
		return false, nil, errEmptyPluginID
	}
	err = s.tx(ctx, func(tx *gorm.DB) error {
		var restriction adapterRestriction
		err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Take(&restriction).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			all = true
			return nil
		}
		if err != nil {
			return err
		}
		var rows []adapterPermission
		if err := tx.Where(&adapterPermission{PluginID: pluginID}).Find(&rows).Error; err != nil {
			return err
		}
		for _, row := range rows {
			instances = append(instances, row.AdapterInstanceID)
		}
		slices.Sort(instances)
		return nil
	})
	return all, instances, err
}

// SetAdapterInstanceAccess 在一个事务中整体替换已安装插件的许可。
// all 为 true 时清除收紧与实例行；false 配合空集合时拒绝全部实例。
func (s *Store) SetAdapterInstanceAccess(ctx context.Context, pluginID string, all bool, instances []string) error {
	if pluginID == "" {
		return errEmptyPluginID
	}
	return s.tx(ctx, func(tx *gorm.DB) error {
		// 1. 确认安装记录，使恢复全部实例的操作同样拒绝未知插件。
		var installed installation
		if err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Take(&installed).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("plugin %s: %w", pluginID, ErrNotFound)
			}
			return err
		}
		// 2. 清除旧集合并写入新许可，任一步失败均回滚整次替换。
		if err := tx.Where(&adapterPermission{PluginID: pluginID}).Delete(&adapterPermission{}).Error; err != nil {
			return err
		}
		if all {
			return tx.Delete(&adapterRestriction{PluginID: pluginID}).Error
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&adapterRestriction{PluginID: pluginID}).Error; err != nil {
			return err
		}
		for _, instanceID := range instances {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&adapterPermission{PluginID: pluginID, AdapterInstanceID: instanceID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
