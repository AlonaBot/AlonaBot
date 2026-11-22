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
	err := s.db.WithContext(ctx).Where(clause.Eq{Column: clause.PrimaryColumn, Value: instanceID}).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return row.Cursor, err
}

// PutAdapterCursor 在同一事务中持久化游标、聊天与账号菜单关系；nil 项保持原记录。
// 空名称与未知类型保留已有值，首次活跃时刻在聊天更新时保持不变。
func (s *Store) PutAdapterCursor(ctx context.Context, pluginID, instanceID string, cursor []byte, chat *Chat, menu *CommandMenuChat, membership *bool) error {
	// 空游标对无重放能力的平台有效，以空 BLOB 保存，满足 NOT NULL 契约。
	if cursor == nil {
		cursor = []byte{}
	}
	return s.recordAdapterEvent(ctx, pluginID, instanceID, cursor, chat, menu, membership, true)
}

// ObserveAdapterEvent 保存管理注入的聊天与菜单关系，保留真实 Adapter 接收游标。
func (s *Store) ObserveAdapterEvent(ctx context.Context, pluginID, instanceID string, chat *Chat, menu *CommandMenuChat, membership *bool) error {
	return s.recordAdapterEvent(ctx, pluginID, instanceID, nil, chat, menu, membership, false)
}

func (s *Store) recordAdapterEvent(ctx context.Context, pluginID, instanceID string, cursor []byte, chat *Chat, menu *CommandMenuChat, membership *bool, saveCursor bool) error {
	now := time.Now().UTC()
	return s.tx(ctx, func(tx *gorm.DB) error {
		if saveCursor {
			if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&adapterCursor{InstanceID: instanceID, PluginID: pluginID, Cursor: cursor, UpdatedAt: now.UnixMilli()}).Error; err != nil {
				return err
			}
		}
		if chat != nil {
			if err := putChat(tx, chat, now); err != nil {
				return err
			}
		}
		if menu != nil {
			return observeMenuChat(tx, *menu, membership)
		}
		return nil
	})
}
