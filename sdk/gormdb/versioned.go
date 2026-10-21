package gormdb

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// LedgerTable 是框架保留的迁移完成表名，业务代码不得挪作他用。
const LedgerTable = "_alonabot_migrations"

type ledgerRow struct {
	ID        uint32 `gorm:"primaryKey;autoIncrement:false"`
	AppliedAt int64  `gorm:"not null"`
}

func (ledgerRow) TableName() string { return LedgerTable }

// AppliedMigrations 返回账本中已完成的迁移 ID（升序）；账本不存在时返回空。
func AppliedMigrations(ctx context.Context, db *gorm.DB) ([]uint32, error) {
	q := db.WithContext(ctx)
	if !q.Migrator().HasTable(LedgerTable) {
		return nil, nil
	}
	var ids []uint32
	err := q.Model(&ledgerRow{}).Order("id").Pluck("id", &ids).Error
	return ids, err
}

// RunVersioned 依次执行尚未完成的迁移。每一步 Up 与账本记录在插件 DB 的同一事务中提交；
// Up 失败时整步回滚、不记录成功，并停止执行后续迁移。返回本次新完成的 ID。
func RunVersioned(ctx context.Context, db *gorm.DB, migrations []Migration) ([]uint32, error) {
	if _, err := CheckMigrators(versionedList(migrations)); err != nil {
		return nil, err
	}
	q := db.WithContext(ctx)
	if err := q.AutoMigrate(&ledgerRow{}); err != nil {
		return nil, fmt.Errorf("创建迁移账本失败: %w", err)
	}
	applied, err := AppliedMigrations(ctx, db)
	if err != nil {
		return nil, err
	}
	done := make(map[uint32]bool, len(applied))
	for _, id := range applied {
		done[id] = true
	}
	var ran []uint32
	for _, m := range migrations {
		if done[m.ID] {
			continue
		}
		err := q.Transaction(func(tx *gorm.DB) error {
			if err := m.Up(ctx, tx); err != nil {
				return err
			}
			return tx.Create(&ledgerRow{ID: m.ID, AppliedAt: time.Now().UnixMilli()}).Error
		})
		if err != nil {
			return ran, fmt.Errorf("迁移 %d 失败: %w", m.ID, err)
		}
		ran = append(ran, m.ID)
	}
	return ran, nil
}

type versionedList []Migration

func (v versionedList) Migrations() []Migration { return v }
