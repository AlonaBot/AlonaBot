package sqlitedb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/sdk/internal/dbhook"
)

// SelfMigrator 由自行管理版本与幂等迁移的根插件实现。
type SelfMigrator interface {
	// Migrate 将 db 迁移到本版本需要的结构，须可重复执行。
	Migrate(ctx context.Context, db *sql.DB) error
}

// Migration 描述一个迁移步骤；ID 非零、唯一、严格递增，发布后保持含义不变。
type Migration struct {
	ID uint32
	// Up 须使用传入的 tx，在单连接池的事务内等待独立查询会自锁。
	Up func(context.Context, *sql.Tx) error
}

// VersionedMigrator 返回按 ID 递增的全部步骤，SDK 执行尚未完成的步骤。
type VersionedMigrator interface {
	// Migrations 返回全部步骤，按 ID 严格递增排列。
	Migrations() []Migration
}

// LedgerTable 指定框架保留的迁移完成表名。
const LedgerTable = dbhook.LedgerTable

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

func checkMigrations(migrations []Migration) error {
	var last uint32
	for _, migration := range migrations {
		if migration.ID == 0 || migration.ID <= last || migration.Up == nil {
			return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: fmt.Sprintf("migration %d must have a nonzero, strictly increasing ID and an Up function (previous ID %d)", migration.ID, last)}
		}
		last = migration.ID
	}
	return nil
}

// AppliedMigrations 返回账本中全部已完成 ID，按升序排列；账本不存在时返回空。
func AppliedMigrations(ctx context.Context, db *sql.DB) ([]uint32, error) {
	var exists bool
	if err := db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM sqlite_schema WHERE type = 'table' AND name = ?)", LedgerTable).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	rows, err := db.QueryContext(ctx, "SELECT id FROM "+LedgerTable+" ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uint32
	for rows.Next() {
		var id uint32
		if err := rows.Scan(&id); err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// RunVersioned 依次执行未完成的步骤，返回本次新完成的 ID。
// 每步 Up 与账本记录在同一事务提交，失败即回滚该步并停止。
// Up 须使用传入的 tx，在事务内等待此单连接池的独立查询会自锁。
func RunVersioned(ctx context.Context, db *sql.DB, migrations []Migration) ([]uint32, error) {
	// 1. 校验全部声明后建立账本并读取已完成的 ID。
	if err := checkMigrations(migrations); err != nil {
		return nil, err
	}
	if _, err := db.ExecContext(ctx, dbhook.LedgerDDL); err != nil {
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
	// 2. 每一步独立提交，失败时保留此前已完成的步骤。
	var ran []uint32
	for _, migration := range migrations {
		if done[migration.ID] {
			continue
		}
		if err := runMigration(ctx, db, migration); err != nil {
			return ran, fmt.Errorf("migration %d failed: %w", migration.ID, err)
		}
		ran = append(ran, migration.ID)
	}
	return ran, nil
}

func runMigration(ctx context.Context, db *sql.DB, migration Migration) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		// panic 同样触发回滚；成功提交或 context 自动回滚后，ErrTxDone 表示事务已结束。
		if rollbackError := tx.Rollback(); rollbackError != nil && !errors.Is(rollbackError, sql.ErrTxDone) {
			err = errors.Join(err, rollbackError)
		}
	}()
	if err := migration.Up(ctx, tx); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO "+LedgerTable+" (id, applied_at) VALUES (?, ?)", migration.ID, time.Now().UnixMilli()); err != nil {
		return err
	}
	return tx.Commit()
}
