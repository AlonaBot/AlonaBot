package state

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const columnServiceName = "service_name"

// ToolPolicy 保存按插件与服务名指定的工具策略覆盖。
type ToolPolicy struct {
	PluginID string        `gorm:"column:plugin_id;primaryKey"`
	Service  string        `gorm:"column:service_name;primaryKey"`
	Policy   pb.ToolPolicy `gorm:"column:policy"`
}

// TableName 指定工具策略表。
func (ToolPolicy) TableName() string { return "tool_policy" }

// SetToolPolicy 在事务内核对安装记录并幂等保存覆盖；UNSPECIFIED 删除覆盖。
// 调用方负责策略取值校验；安装记录缺失时返回 ErrNotFound。
func (s *Store) SetToolPolicy(ctx context.Context, row ToolPolicy) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		// 1. 删除也要求存在安装记录，检查与写入共用事务。
		var installed installation
		err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: row.PluginID}).Take(&installed).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("plugin %s: %w", row.PluginID, ErrNotFound)
		}
		if err != nil {
			return err
		}
		// 2. 提交后才返回，下一次读取直接取得当前覆盖。
		if row.Policy == pb.ToolPolicy_TOOL_POLICY_UNSPECIFIED {
			return tx.Where(toolPolicyKey(row.PluginID, row.Service)).Delete(&ToolPolicy{}).Error
		}
		return tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error
	})
}

// GetToolPolicy 读取工具覆盖；未保存覆盖时返回 ErrNotFound。
func (s *Store) GetToolPolicy(ctx context.Context, pluginID, service string) (pb.ToolPolicy, error) {
	var row ToolPolicy
	err := s.db.WithContext(ctx).Where(toolPolicyKey(pluginID, service)).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row.Policy, fmt.Errorf("tool %s/%s: %w", pluginID, service, ErrNotFound)
	}
	return row.Policy, err
}

// ToolPolicies 读取全部保存的覆盖；排序由工具目录完成。
func (s *Store) ToolPolicies(ctx context.Context) ([]ToolPolicy, error) {
	var rows []ToolPolicy
	err := s.db.WithContext(ctx).Find(&rows).Error
	return rows, err
}

func toolPolicyKey(pluginID, service string) clause.Expression {
	return clause.And(clause.Eq{Column: columnPluginID, Value: pluginID}, clause.Eq{Column: columnServiceName, Value: service})
}
