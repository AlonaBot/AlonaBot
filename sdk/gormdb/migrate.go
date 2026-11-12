// Package gormdb 为插件打开受管数据库，并提供迁移接口。
//
// [Open] 包装当前 Session 的受管连接；[OpenSQLite] 打开由调用方关闭的独立连接。
// 两者均返回原生的 *gorm.DB，并核对连接上的关键设置确实生效。
// 插件在 New 中调用 [Migrate] 或 [Migrations] 将迁移登记到插件对象。
// 数据库的所有权、连接与迁移规则见 spec/CONTRACTS-v0.1.md §6。
//
// 包的组成：[Migrate] 登记自行管理的迁移，[Migrations] 登记逐步迁移；
// [RunVersioned] 与 [AppliedMigrations] 执行迁移并查询两个数据库入口共用的账本。
package gormdb

import (
	"context"
	"fmt"
	"slices"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/internal/dbhook"
	"gorm.io/gorm"
)

// Migration 描述辅助迁移的一步；ID 非零、唯一、递增，发布后不得改作他用。
type Migration struct {
	ID uint32 // 迁移编号，非零、唯一、递增
	// Up 须使用传入的 tx，在单连接池的事务内等待独立查询会自锁。
	Up func(context.Context, *gorm.DB) error
}

// 迁移模式与 manifest 的 migration_mode 取值一致。
const (
	ModeNone      = "none"
	ModeSelf      = "self"
	ModeVersioned = "versioned"
)

func checkMigrations(migrations []Migration) error {
	var last uint32
	for _, m := range migrations {
		if m.ID == 0 || m.ID <= last || m.Up == nil {
			return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: fmt.Sprintf("migration %d must have a nonzero, strictly increasing ID and an Up function (previous ID %d)", m.ID, last)}
		}
		last = m.ID
	}
	return nil
}

// Migrate 登记自行管理版本与幂等性的数据库迁移。
func Migrate(p *sdk.Plugin, fn func(context.Context, *gorm.DB) error) {
	dbhook.Register(p, dbhook.Migrator{Mode: pb.MigrationMode_MIGRATION_MODE_SELF, Check: func() error {
		if fn == nil {
			return fmt.Errorf("migration function is nil")
		}
		return nil
	}, Run: func(ctx context.Context) (*pb.MigrationState, error) {
		db, err := Open(ctx)
		if err != nil {
			return nil, err
		}
		if err := fn(ctx, db); err != nil {
			return nil, err
		}
		return &pb.MigrationState{Mode: pb.MigrationMode_MIGRATION_MODE_SELF, SelfMigrateCompleted: true}, nil
	}})
}

// Migrations 登记按 ID 严格递增的迁移列表，Collect 时核对全部步骤。
func Migrations(p *sdk.Plugin, list []Migration) {
	list = slices.Clone(list)
	dbhook.Register(p, dbhook.Migrator{Mode: pb.MigrationMode_MIGRATION_MODE_VERSIONED, Check: func() error { return checkMigrations(list) }, Run: func(ctx context.Context) (*pb.MigrationState, error) {
		db, err := Open(ctx)
		if err != nil {
			return nil, err
		}
		if _, err := RunVersioned(ctx, db, list); err != nil {
			return nil, err
		}
		ids, err := AppliedMigrations(ctx, db)
		return &pb.MigrationState{Mode: pb.MigrationMode_MIGRATION_MODE_VERSIONED, CompletedMigrationIds: ids}, err
	}})
}
