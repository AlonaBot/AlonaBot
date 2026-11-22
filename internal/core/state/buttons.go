package state

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ButtonToken 保存按钮的用途、路由、代次、目标及插件 JSON 值或确认记录 ID。
// 期限以 Unix 毫秒保存，Session 重启不会改变记录。
type ButtonToken struct {
	Token               string `gorm:"column:token;primaryKey"`
	Purpose             string `gorm:"column:purpose"` // plugin 绑定插件路由，confirm 绑定 Core 确认记录
	PluginID            string `gorm:"column:plugin_id"`
	RouteID             string `gorm:"column:route_id"`
	ActionID            string `gorm:"column:action_id"`
	ActionSchemaVersion uint32 `gorm:"column:action_schema_version"`
	GenerationID        string `gorm:"column:generation_id"`
	AdapterID           string `gorm:"column:adapter_id"`
	InstanceID          string `gorm:"column:instance_id"`
	AccountID           string `gorm:"column:account_id"`
	ChatID              string `gorm:"column:chat_id"`
	ThreadID            string `gorm:"column:thread_id"` // 空串表示无话题，与 Target.GetThreadId 一致
	Value               string `gorm:"column:value"`
	ExpiresAt           int64  `gorm:"column:expires_at"` // Unix 毫秒
}

// TableName 指定按钮映射表。
func (ButtonToken) TableName() string { return "button_token" }

// PutButtonTokens 在同一事务中插入整条消息的按钮，任一失败均回滚。
func (s *Store) PutButtonTokens(ctx context.Context, tokens []ButtonToken) error {
	return s.tx(ctx, func(tx *gorm.DB) error { return tx.Create(&tokens).Error })
}

// GetButtonToken 按精确 token 读取映射，缺失时返回 ErrNotFound。
func (s *Store) GetButtonToken(ctx context.Context, token string) (ButtonToken, error) {
	var row ButtonToken
	err := s.db.WithContext(ctx).Where(clause.Eq{Column: clause.PrimaryColumn, Value: token}).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	return row, err
}

// PurgeButtonTokens 删除指定插件的全部按钮映射。
func (s *Store) PurgeButtonTokens(ctx context.Context, pluginID string) error {
	if pluginID == "" {
		return errEmptyPluginID
	}
	return s.db.WithContext(ctx).Where(&ButtonToken{PluginID: pluginID}).Delete(&ButtonToken{}).Error
}

// RevokeConfirmTokens 删除同一确认记录的两枚 token。
func (s *Store) RevokeConfirmTokens(ctx context.Context, recordID string) error {
	if recordID == "" {
		return errors.New("empty confirmation record ID")
	}
	return s.db.WithContext(ctx).Where(&ButtonToken{Purpose: "confirm", Value: recordID}).Delete(&ButtonToken{}).Error
}
