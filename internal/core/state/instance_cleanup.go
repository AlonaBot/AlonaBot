package state

import (
	"context"
	"slices"

	"gorm.io/gorm"
)

// DeleteInstanceState 原子清除已删除实例的游标、Profile 指定和会话访问名单引用。
// 名单的访问模式保持原值，空 allow 名单仍拒绝全部会话。
func (s *Store) DeleteInstanceState(ctx context.Context, instances []string) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		for _, instance := range instances {
			if err := tx.Delete(&adapterCursor{InstanceID: instance}).Error; err != nil {
				return err
			}
			if err := tx.Where(&ProfileAssignment{Instance: instance}).Delete(&ProfileAssignment{}).Error; err != nil {
				return err
			}
		}
		var access []ProfileSessionAccess
		if err := tx.Find(&access).Error; err != nil {
			return err
		}
		for _, row := range access {
			before := len(row.Sessions)
			row.Sessions = slices.DeleteFunc(row.Sessions, func(session ProfileSession) bool {
				return slices.Contains(instances, session.Instance)
			})
			if len(row.Sessions) != before {
				if err := tx.Save(&row).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
