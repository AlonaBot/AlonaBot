package state

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const devSource = "dev"

// Installs 按安装顺序读取全部记录，供启动装配与身份冲突检查使用。
func (s *Store) Installs(ctx context.Context) ([]Install, error) {
	var rows []installation
	if err := s.db.WithContext(ctx).Order(clause.OrderByColumn{Column: clause.Column{Name: columnInstallSeq}}).Find(&rows).Error; err != nil {
		return nil, err
	}
	installs := make([]Install, 0, len(rows))
	for _, row := range rows {
		installs = append(installs, row.Install)
	}
	return installs, nil
}

// SetDevBuildDigest 刷新已存在的 dev 安装记录的摘要，保留身份、安装顺序、启用状态与代次。
// 记录不存在时返回 ErrNotFound，来源不同返回错误。
func (s *Store) SetDevBuildDigest(ctx context.Context, pluginID string, digest []byte) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		var row installation
		err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Take(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("plugin %s: %w", pluginID, ErrNotFound)
		}
		if err != nil {
			return err
		}
		if row.Source != devSource {
			return fmt.Errorf("plugin %s has source %q, expected dev", pluginID, row.Source)
		}
		row.BuildDigest = digest
		return tx.Save(&row).Error
	})
}
