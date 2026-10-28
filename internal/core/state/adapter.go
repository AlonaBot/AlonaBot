package state

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Role 标记可信安装记录中的进程角色。
type Role string

const (
	// RolePlugin 指定插件进程。
	RolePlugin Role = "plugin"
	// RoleAdapter 指定承载平台账号实例的 Adapter 进程。
	RoleAdapter Role = "adapter"
)

// AdapterCursor 读取实例的已持久化游标；没有记录时返回 nil, nil。
func (s *Store) AdapterCursor(ctx context.Context, instanceID string) ([]byte, error) {
	var cursor []byte
	err := s.db.QueryRowContext(ctx, "SELECT cursor FROM adapter_cursors WHERE instance_id = ?", instanceID).Scan(&cursor)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return cursor, err
}

// PutAdapterCursor 持久化实例交付后的游标，保留完整的不透明字节。
func (s *Store) PutAdapterCursor(ctx context.Context, pluginID, instanceID string, cursor []byte) error {
	// 空游标对无重放能力的平台有效，以空 BLOB 保存，满足 NOT NULL 契约。
	if cursor == nil {
		cursor = []byte{}
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO adapter_cursors (instance_id, plugin_id, cursor, updated_at) VALUES (?, ?, ?, ?)
 ON CONFLICT(instance_id) DO UPDATE SET plugin_id = excluded.plugin_id, cursor = excluded.cursor, updated_at = excluded.updated_at`, instanceID, pluginID, cursor, time.Now().UnixMilli())
	return err
}
