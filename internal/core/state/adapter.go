package state

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AdapterCursor 读取实例的已持久化游标；没有记录时返回 nil, nil。
func (s *Store) AdapterCursor(ctx context.Context, instanceID string) ([]byte, error) {
	var row adapterCursor
	err := s.db.WithContext(ctx).Where(&adapterCursor{InstanceID: instanceID}, "InstanceID").Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return row.Cursor, err
}

// PutAdapterCursor 持久化实例交付后的游标，保留完整的不透明字节。
func (s *Store) PutAdapterCursor(ctx context.Context, pluginID, instanceID string, cursor []byte) error {
	// 空游标对无重放能力的平台有效，以空 BLOB 保存，满足 NOT NULL 契约。
	if cursor == nil {
		cursor = []byte{}
	}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: columnInstanceID}},
		DoUpdates: clause.AssignmentColumns([]string{columnPluginID, columnCursor, columnUpdatedAt}),
	}).Create(&adapterCursor{InstanceID: instanceID, PluginID: pluginID, Cursor: cursor, UpdatedAt: time.Now().UnixMilli()}).Error
}
