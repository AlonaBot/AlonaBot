package state

import (
	"context"

	"gorm.io/gorm/clause"
)

// AllowedAdapterInstances 按字典序读取插件当前获准发送的 Adapter 实例。
func (s *Store) AllowedAdapterInstances(ctx context.Context, pluginID string) ([]string, error) {
	var rows []adapterPermission
	err := s.db.WithContext(ctx).Select("adapter_instance_id").Where("plugin_id = ?", pluginID).Order("adapter_instance_id").Find(&rows).Error
	var instances []string
	for _, row := range rows {
		instances = append(instances, row.AdapterInstanceID)
	}
	return instances, err
}

// GrantAdapterInstance 幂等地授予插件向指定 Adapter 实例发送的许可。
func (s *Store) GrantAdapterInstance(ctx context.Context, pluginID, instanceID string) error {
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "plugin_id"}, {Name: "adapter_instance_id"}}, DoNothing: true,
	}).Create(&adapterPermission{PluginID: pluginID, AdapterInstanceID: instanceID}).Error
}

// RevokeAdapterInstance 幂等地撤销插件向指定 Adapter 实例发送的许可。
func (s *Store) RevokeAdapterInstance(ctx context.Context, pluginID, instanceID string) error {
	return s.db.WithContext(ctx).Where("plugin_id = ? AND adapter_instance_id = ?", pluginID, instanceID).Delete(&adapterPermission{}).Error
}
