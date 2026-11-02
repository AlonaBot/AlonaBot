package state

import (
	"context"
	"slices"

	"gorm.io/gorm/clause"
)

// AllowedAdapterInstances 按字典序读取插件当前获准发送的 Adapter 实例。
func (s *Store) AllowedAdapterInstances(ctx context.Context, pluginID string) ([]string, error) {
	var rows []adapterPermission
	err := s.db.WithContext(ctx).Where(&adapterPermission{PluginID: pluginID}).Find(&rows).Error
	var instances []string
	for _, row := range rows {
		instances = append(instances, row.AdapterInstanceID)
	}
	slices.Sort(instances)
	return instances, err
}

// GrantAdapterInstance 幂等地授予插件向指定 Adapter 实例发送的许可。
func (s *Store) GrantAdapterInstance(ctx context.Context, pluginID, instanceID string) error {
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: columnPluginID}, {Name: columnAdapterInstanceID}}, DoNothing: true,
	}).Create(&adapterPermission{PluginID: pluginID, AdapterInstanceID: instanceID}).Error
}

// RevokeAdapterInstance 幂等地撤销插件向指定 Adapter 实例发送的许可。
func (s *Store) RevokeAdapterInstance(ctx context.Context, pluginID, instanceID string) error {
	return s.db.WithContext(ctx).Delete(&adapterPermission{PluginID: pluginID, AdapterInstanceID: instanceID}).Error
}
