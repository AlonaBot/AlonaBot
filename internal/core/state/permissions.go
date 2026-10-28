package state

import "context"

// AllowedAdapterInstances 按字典序读取插件当前获准发送的 Adapter 实例。
func (s *Store) AllowedAdapterInstances(ctx context.Context, pluginID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT adapter_instance_id FROM plugin_adapter_permissions WHERE plugin_id = ? ORDER BY adapter_instance_id", pluginID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var instances []string
	for rows.Next() {
		var instanceID string
		if err := rows.Scan(&instanceID); err != nil {
			return instances, err
		}
		instances = append(instances, instanceID)
	}
	return instances, rows.Err()
}

// GrantAdapterInstance 幂等地授予插件向指定 Adapter 实例发送的许可。
func (s *Store) GrantAdapterInstance(ctx context.Context, pluginID, instanceID string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO plugin_adapter_permissions (plugin_id, adapter_instance_id) VALUES (?, ?)
		 ON CONFLICT (plugin_id, adapter_instance_id) DO NOTHING`, pluginID, instanceID)
	return err
}

// RevokeAdapterInstance 幂等地撤销插件向指定 Adapter 实例发送的许可。
func (s *Store) RevokeAdapterInstance(ctx context.Context, pluginID, instanceID string) error {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM plugin_adapter_permissions WHERE plugin_id = ? AND adapter_instance_id = ?", pluginID, instanceID)
	return err
}
