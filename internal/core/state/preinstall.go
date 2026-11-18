package state

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const coreInitializationID = 1

type coreInitialization struct {
	ID           int  `gorm:"column:id;primaryKey;autoIncrement:false"`
	Preinstalled bool `gorm:"column:preinstalled"`
}

func (coreInitialization) TableName() string { return "core_initialization" }

// Preinstalled 读取首次预装完成标记；尚无初始化记录时返回 false。
func (s *Store) Preinstalled(ctx context.Context) (bool, error) {
	var row coreInitialization
	err := s.db.WithContext(ctx).Where(clause.Eq{Column: clause.PrimaryColumn, Value: coreInitializationID}).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return row.Preinstalled, err
}

// MarkPreinstalled 持久化首次预装完成标记，重复调用保持完成状态。
func (s *Store) MarkPreinstalled(ctx context.Context) error {
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&coreInitialization{ID: coreInitializationID, Preinstalled: true}).Error
}
