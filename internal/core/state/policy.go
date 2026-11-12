package state

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	columnPluginID    = "plugin_id"
	columnScope       = "scope"
	columnAdapterID   = "adapter_id"
	columnDenyKind    = "kind"
	columnSubjectID   = "subject_id"
	columnNames       = "names"
	columnPermissions = "permissions"
	columnEnabled     = "enabled"
)

// CommandPolicy 保存插件一个作用域的显式覆盖；Scope 的编码由策略服务定义。
type CommandPolicy struct {
	PluginID    string     `gorm:"column:plugin_id;primaryKey"`
	Scope       string     `gorm:"column:scope;primaryKey"`
	Permissions []string   `gorm:"column:permissions;serializer:json"` // nil 为 NULL，空切片为 []
	Enabled     *bool      `gorm:"column:enabled"`                     // nil 表示没有覆盖
	Names       [][]string `gorm:"column:names;serializer:json"`       // nil 为 NULL；第一项为主名，仅路由作用域使用
}

// TableName 指定命令策略表。
func (CommandPolicy) TableName() string { return "command_policy" }

// GlobalDeny 保存 Adapter 作用域中的用户或会话禁止；Kind 由策略服务校验。
type GlobalDeny struct {
	Kind      string `gorm:"column:kind;primaryKey"`
	AdapterID string `gorm:"column:adapter_id;primaryKey"`
	SubjectID string `gorm:"column:subject_id;primaryKey"`
}

// TableName 指定全局禁止表。
func (GlobalDeny) TableName() string { return "global_deny" }

// PutCommandPolicy 保存权限与启用覆盖，保留命令名；调用方负责作用域和权限名校验。
func (s *Store) PutCommandPolicy(ctx context.Context, row CommandPolicy) error {
	return s.updateCommandPolicy(ctx, row.PluginID, row.Scope, func(stored *CommandPolicy) {
		stored.Permissions, stored.Enabled = row.Permissions, row.Enabled
	})
}

// PutCommandNames 只保存命令名覆盖；names 为 nil 时清空该列。
func (s *Store) PutCommandNames(ctx context.Context, pluginID, scope string, names [][]string) error {
	return s.updateCommandPolicy(ctx, pluginID, scope, func(stored *CommandPolicy) { stored.Names = names })
}

func (s *Store) updateCommandPolicy(ctx context.Context, pluginID, scope string, update func(*CommandPolicy)) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row := CommandPolicy{PluginID: pluginID, Scope: scope}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
			return err
		}
		if err := tx.Where(commandPolicyKey(pluginID, scope)).Take(&row).Error; err != nil {
			return err
		}
		update(&row)
		return tx.Save(&row).Error
	})
}

// DeleteCommandPolicy 按主键幂等删除权限、启用与命令名均未覆盖的行。
func (s *Store) DeleteCommandPolicy(ctx context.Context, pluginID, scope string) error {
	return s.db.WithContext(ctx).Where(commandPolicyKey(pluginID, scope)).Where(clause.And(
		clause.Eq{Column: columnPermissions, Value: nil},
		clause.Eq{Column: columnEnabled, Value: nil},
		clause.Eq{Column: columnNames, Value: nil},
	)).Delete(&CommandPolicy{}).Error
}

// AllCommandNames 读取全部带命令名覆盖的策略行。
func (s *Store) AllCommandNames(ctx context.Context) ([]CommandPolicy, error) {
	var rows []CommandPolicy
	err := s.db.WithContext(ctx).Where(clause.Neq{Column: columnNames, Value: nil}).Find(&rows).Error
	return rows, err
}

// CommandPolicies 读取插件全部覆盖；排序由策略服务按作用域语义完成。
func (s *Store) CommandPolicies(ctx context.Context, pluginID string) ([]CommandPolicy, error) {
	var rows []CommandPolicy
	err := s.db.WithContext(ctx).Where(clause.Eq{Column: columnPluginID, Value: pluginID}).Find(&rows).Error
	return rows, err
}

// CommandPolicySnapshot 在一个只读事务中检查 subjects 的禁止状态并读取插件覆盖。
// subjects 只包含实际存在的 Actor 与 Target；命中任一禁止时返回 true，无需读取覆盖。
func (s *Store) CommandPolicySnapshot(ctx context.Context, pluginID string, subjects []GlobalDeny) ([]CommandPolicy, bool, error) {
	var rows []CommandPolicy
	var denied bool
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 禁止判定先于策略，全部查询复用事务连接及同一快照。
		for _, subject := range subjects {
			found, err := hasGlobalDeny(tx, subject)
			if err != nil {
				return err
			}
			if found {
				denied = true
				return nil
			}
		}
		// 2. 同一事务读出全部覆盖，交给策略服务计算继承。
		return tx.Where(clause.Eq{Column: columnPluginID, Value: pluginID}).Find(&rows).Error
	}, &sql.TxOptions{ReadOnly: true})
	return rows, denied, err
}

// AddGlobalDeny 幂等添加禁止记录；调用方负责种类与身份校验。
func (s *Store) AddGlobalDeny(ctx context.Context, row GlobalDeny) error {
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}

// RemoveGlobalDeny 按完整主键幂等移除禁止记录。
func (s *Store) RemoveGlobalDeny(ctx context.Context, row GlobalDeny) error {
	return s.db.WithContext(ctx).Where(globalDenyKey(row)).Delete(&GlobalDeny{}).Error
}

// HasGlobalDeny 查询指定用户或会话是否被禁止。
func (s *Store) HasGlobalDeny(ctx context.Context, row GlobalDeny) (bool, error) {
	return hasGlobalDeny(s.db.WithContext(ctx), row)
}

// GlobalDenies 读取全部禁止记录；排序由策略服务完成。
func (s *Store) GlobalDenies(ctx context.Context) ([]GlobalDeny, error) {
	var rows []GlobalDeny
	err := s.db.WithContext(ctx).Find(&rows).Error
	return rows, err
}

func hasGlobalDeny(db *gorm.DB, row GlobalDeny) (bool, error) {
	var count int64
	err := db.Model(&GlobalDeny{}).Where(globalDenyKey(row)).Count(&count).Error
	return count > 0, err
}

// 使用等值条件保留零值主键的精确语义，避免 GORM 的结构体条件忽略空串。
func commandPolicyKey(pluginID, scope string) clause.Expression {
	return clause.And(clause.Eq{Column: columnPluginID, Value: pluginID}, clause.Eq{Column: columnScope, Value: scope})
}

func globalDenyKey(row GlobalDeny) clause.Expression {
	return clause.And(clause.Eq{Column: columnDenyKind, Value: row.Kind}, clause.Eq{Column: columnAdapterID, Value: row.AdapterID}, clause.Eq{Column: columnSubjectID, Value: row.SubjectID})
}
