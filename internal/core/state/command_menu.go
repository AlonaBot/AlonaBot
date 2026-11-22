package state

import (
	"context"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CommandMenuAccount 保存一个平台账号已接纳的模式、完成状态与同步错误。
type CommandMenuAccount struct {
	ID             string `gorm:"column:id;primaryKey"`
	AdapterID      string `gorm:"column:adapter_id"`
	AccountID      string `gorm:"column:account_id"`
	Mode           string `gorm:"column:mode"`
	Ready          bool   `gorm:"column:ready"`
	LastError      string `gorm:"column:last_error"`
	ClearAutomatic bool   `gorm:"column:clear_automatic"` // 进入手动模式时的清理决定，重入继续处理离群位置。
}

// TableName 指定账号模式表。
func (CommandMenuAccount) TableName() string { return "command_menu_account" }

// CommandMenuChat 保存账号与聊天的独立关系，所有话题共用一行。
type CommandMenuChat struct {
	ID        string `gorm:"column:id;primaryKey"`
	AdapterID string `gorm:"column:adapter_id"`
	AccountID string `gorm:"column:account_id"`
	ChatID    string `gorm:"column:chat_id"`
	Active    bool   `gorm:"column:active"`
}

// TableName 指定账号聊天关系表。
func (CommandMenuChat) TableName() string { return "command_menu_chat" }

// CommandMenuPosition 只保存位置与最后成功写入的整份菜单所有者。
type CommandMenuPosition struct {
	ID            string `gorm:"column:id;primaryKey"`
	AdapterID     string `gorm:"column:adapter_id"`
	AccountID     string `gorm:"column:account_id"`
	Scope         string `gorm:"column:scope"`
	ChatID        string `gorm:"column:chat_id"`
	UserID        string `gorm:"column:user_id"`
	LanguageCode  string `gorm:"column:language_code"`
	Mode          string `gorm:"column:mode"`
	OwnerPluginID string `gorm:"column:owner_plugin_id"`
}

// TableName 指定菜单位置表。
func (CommandMenuPosition) TableName() string { return "command_menu_position" }

// CommandMenuKey 用完整有序字段生成可精确匹配空语言、空成员的位置主键。
func CommandMenuKey(parts ...string) string {
	encoded, _ := json.Marshal(parts) // 字符串切片始终可编码。
	return string(encoded)
}

// MenuAccount 读取账号已完成的模式，首次使用返回空模式。
func (s *Store) MenuAccount(ctx context.Context, adapterID, accountID string) (CommandMenuAccount, error) {
	row := CommandMenuAccount{ID: CommandMenuKey(adapterID, accountID), AdapterID: adapterID, AccountID: accountID}
	err := s.db.WithContext(ctx).Where(clause.Eq{Column: clause.PrimaryColumn, Value: row.ID}).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, nil
	}
	return row, err
}

// SaveMenuAccount 持久化模式及实际错误，保留 false 与空串。
func (s *Store) SaveMenuAccount(ctx context.Context, row CommandMenuAccount) error {
	row.ID = CommandMenuKey(row.AdapterID, row.AccountID)
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error
}

// MenuChats 读取账号的聊天关系，包含已离开的目标供重入恢复。
func (s *Store) MenuChats(ctx context.Context, adapterID, accountID string) ([]CommandMenuChat, error) {
	var rows []CommandMenuChat
	err := s.db.WithContext(ctx).Where(&CommandMenuChat{AdapterID: adapterID, AccountID: accountID}).Find(&rows).Error
	return rows, err
}

// ObserveMenuChat 首次登记账号内的聊天关系，已存在时保留最新成员状态。
func (s *Store) ObserveMenuChat(ctx context.Context, row CommandMenuChat) error {
	return observeMenuChat(s.db.WithContext(ctx), row, nil)
}

// MenuPositions 读取账号的全部受管位置。
func (s *Store) MenuPositions(ctx context.Context, adapterID, accountID string) ([]CommandMenuPosition, error) {
	var rows []CommandMenuPosition
	err := s.db.WithContext(ctx).Where(&CommandMenuPosition{AdapterID: adapterID, AccountID: accountID}).Find(&rows).Error
	return rows, err
}

// SaveMenuPosition 在远端写入成功后原子保存位置所有权与首次见到的聊天关系。
func (s *Store) SaveMenuPosition(ctx context.Context, row CommandMenuPosition) error {
	row.ID = CommandMenuKey(row.AdapterID, row.AccountID, row.Scope, row.ChatID, row.UserID, row.LanguageCode)
	return s.tx(ctx, func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error; err != nil {
			return err
		}
		if row.ChatID == "" {
			return nil
		}
		return observeMenuChat(tx, CommandMenuChat{AdapterID: row.AdapterID, AccountID: row.AccountID, ChatID: row.ChatID, Active: true}, nil)
	})
}

// DeleteMenuPosition 在远端删除成功后移除精确位置记录。
func (s *Store) DeleteMenuPosition(ctx context.Context, row CommandMenuPosition) error {
	row.ID = CommandMenuKey(row.AdapterID, row.AccountID, row.Scope, row.ChatID, row.UserID, row.LanguageCode)
	return s.db.WithContext(ctx).Delete(&row).Error
}

// observeMenuChat 原子登记首次观察，只有明确成员事件才更新已有关系。
func observeMenuChat(tx *gorm.DB, row CommandMenuChat, membership *bool) error {
	row.ID = CommandMenuKey(row.AdapterID, row.AccountID, row.ChatID)
	if membership == nil {
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
	}
	row.Active = *membership
	return tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error
}

// MenuChat 精确读取账号内聊天关系；首次手动设置的位置可尚未出现过入站事件。
func (s *Store) MenuChat(ctx context.Context, adapterID, accountID, chatID string) (CommandMenuChat, bool, error) {
	var row CommandMenuChat
	err := s.db.WithContext(ctx).Where(clause.Eq{Column: clause.PrimaryColumn, Value: CommandMenuKey(adapterID, accountID, chatID)}).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, false, nil
	}
	return row, err == nil, err
}
