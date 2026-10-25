package gormdb

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// LedgerTable 指定框架保留的迁移完成表名，业务代码不得挪作他用。
const LedgerTable = "_alonabot_migrations"

// ledgerRow 对应迁移账本中的一行。
type ledgerRow struct {
	ID        uint32 `gorm:"primaryKey;autoIncrement:false"` // 已完成的迁移 ID
	AppliedAt int64  `gorm:"not null"`                       // 完成时刻，UTC 毫秒
}

// TableName 返回固定的账本表名。
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
	// 1. 校验迁移列表，确保账本存在并读出已完成的 ID。
	if _, err := CheckMigrators(versionedList(migrations)); err != nil {
		return nil, err
	}
	q := db.WithContext(ctx)
	if err := q.AutoMigrate(&ledgerRow{}); err != nil {
		return nil, fmt.Errorf("create migration ledger: %w", err)
	}
	applied, err := AppliedMigrations(ctx, db)
	if err != nil {
		return nil, err
	}
	done := make(map[uint32]bool, len(applied))
	for _, id := range applied {
		done[id] = true
	}
	// 2. 逐个执行未完成的迁移，每步与账本记录同事务提交；失败即停止。
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
			return ran, fmt.Errorf("migration %d failed: %w", m.ID, err)
		}
		ran = append(ran, m.ID)
	}
	return ran, nil
}

// versionedList 把迁移列表适配为 VersionedMigrator，以复用 CheckMigrators 的校验。
type versionedList []Migration

// Migrations 返回列表本身。
func (v versionedList) Migrations() []Migration { return v }
