// Package gormdb 提供插件受管数据库的 GORM 入口与迁移接口（CONTRACTS §6）。
package gormdb

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// SelfMigrator 由自行管理迁移的插件实现：作者自行管理迁移版本与幂等性。
type SelfMigrator interface {
	// Migrate 把 db 迁移到本版本需要的结构；必须可重复执行。
	Migrate(ctx context.Context, db *gorm.DB) error
}

// Migration 是辅助迁移的一步；ID 非零、唯一、递增，发布后不得改作他用。
type Migration struct {
	// ID 是迁移编号，非零、唯一、递增。
	ID uint32
	// Up 在事务中执行这一步迁移。
	Up func(context.Context, *gorm.DB) error
}

// VersionedMigrator 由按版本迁移的插件实现：SDK 只执行尚未完成的 ID，每步 Up 与完成记录在同一事务提交。
type VersionedMigrator interface {
	// Migrations 返回全部迁移步骤，按 ID 递增排列。
	Migrations() []Migration
}

// 迁移模式，与 manifest 的 migration_mode 取值一致。
const (
	ModeNone      = "none"
	ModeSelf      = "self"
	ModeVersioned = "versioned"
)

// CheckMigrators 返回根插件声明的迁移模式；两种接口同时实现或迁移列表非法时报错。
func CheckMigrators(root any) (string, error) {
	_, isSelf := root.(SelfMigrator)
	versioned, isVersioned := root.(VersionedMigrator)
	switch {
	case isSelf && isVersioned:
		return "", fmt.Errorf("%T 同时实现了 Migrate 和 Migrations，两者互斥", root)
	case isSelf:
		return ModeSelf, nil
	case !isVersioned:
		return ModeNone, nil
	}
	var last uint32
	for i, m := range versioned.Migrations() {
		switch {
		case m.ID == 0:
			return "", fmt.Errorf("第 %d 个迁移的 ID 为 0", i)
		case m.ID <= last:
			return "", fmt.Errorf("迁移 ID %d 不大于前一个 ID %d（必须唯一且递增）", m.ID, last)
		case m.Up == nil:
			return "", fmt.Errorf("迁移 %d 缺少 Up", m.ID)
		}
		last = m.ID
	}
	return ModeVersioned, nil
}
