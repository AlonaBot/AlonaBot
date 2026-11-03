// Package gormdb 为插件打开受管数据库，并提供迁移接口。
//
// [Open] 包装当前 Session 的受管连接；[OpenSQLite] 打开由调用方关闭的独立连接。
// 两者均返回原生的 *gorm.DB，并核对连接上的关键设置确实生效。
// 插件可以自己实现 Migrate，也可以交出按编号排列的 Migrations 由 SDK 执行，两者只能选一种。
// 数据库的所有权、连接与迁移规则见 spec/CONTRACTS-v0.1.md §6。
//
// 包的组成：[SelfMigrator] 承载自行管理的迁移，[VersionedMigrator] 提供逐步迁移；
// [RunVersioned] 与 [AppliedMigrations] 执行迁移并查询两个数据库入口共用的账本。
package gormdb

import (
	"context"
	"fmt"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/sdk/internal/dbhook"
	"gorm.io/gorm"
)

// SelfMigrator 由自行管理迁移的插件实现：作者自行管理迁移版本与幂等性。
type SelfMigrator interface {
	// Migrate 把 db 迁移到本版本需要的结构；必须可重复执行。
	Migrate(ctx context.Context, db *gorm.DB) error
}

// Migration 描述辅助迁移的一步；ID 非零、唯一、递增，发布后不得改作他用。
type Migration struct {
	ID uint32 // 迁移编号，非零、唯一、递增
	// Up 须使用传入的 tx，在单连接池的事务内等待独立查询会自锁。
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
		return "", &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: fmt.Sprintf("%T implements both Migrate and Migrations, which are mutually exclusive", root)}
	case isSelf:
		return ModeSelf, nil
	case !isVersioned:
		return ModeNone, nil
	}
	var last uint32
	for i, m := range versioned.Migrations() {
		switch {
		case m.ID == 0:
			return "", &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: fmt.Sprintf("migration %d has ID 0", i)}
		case m.ID <= last:
			return "", &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: fmt.Sprintf("migration ID %d is not greater than the previous ID %d (must be unique and increasing)", m.ID, last)}
		case m.Up == nil:
			return "", &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: fmt.Sprintf("migration %d has no Up", m.ID)}
		}
		last = m.ID
	}
	return ModeVersioned, nil
}

func init() { dbhook.Register(recognize) }

func recognize(root any) []dbhook.Migrator {
	var migrators []dbhook.Migrator
	if self, ok := root.(SelfMigrator); ok {
		migrators = append(migrators, dbhook.Migrator{Mode: pb.MigrationMode_MIGRATION_MODE_SELF, Run: func(ctx context.Context) (*pb.MigrationState, error) {
			db, err := Open(ctx)
			if err != nil {
				return nil, err
			}
			if err := self.Migrate(ctx, db); err != nil {
				return nil, err
			}
			return &pb.MigrationState{Mode: pb.MigrationMode_MIGRATION_MODE_SELF, SelfMigrateCompleted: true}, nil
		}})
	}
	if versioned, ok := root.(VersionedMigrator); ok {
		migrators = append(migrators, dbhook.Migrator{Mode: pb.MigrationMode_MIGRATION_MODE_VERSIONED, Run: func(ctx context.Context) (*pb.MigrationState, error) {
			db, err := Open(ctx)
			if err != nil {
				return nil, err
			}
			if _, err := RunVersioned(ctx, db, versioned.Migrations()); err != nil {
				return nil, err
			}
			ids, err := AppliedMigrations(ctx, db)
			return &pb.MigrationState{Mode: pb.MigrationMode_MIGRATION_MODE_VERSIONED, CompletedMigrationIds: ids}, err
		}})
	}
	return migrators
}
