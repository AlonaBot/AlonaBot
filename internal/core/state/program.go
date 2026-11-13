package state

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReplaceProgram 替换已安装组件的程序身份；来源、pin、安装顺序、启用状态与 Active 保持不变。
func (s *Store) ReplaceProgram(ctx context.Context, pluginID, version, entrypoint, publisherKeyID string, buildDigest []byte) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		var installed installation
		if err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Take(&installed).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if installed.Source == devSource {
			return fmt.Errorf("development component %s cannot replace its installed program", pluginID)
		}
		installed.Version, installed.Entrypoint = version, entrypoint
		installed.PublisherKeyID, installed.BuildDigest = publisherKeyID, buildDigest
		return tx.Save(&installed).Error
	})
}
