package gormdb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/libtnb/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 受管 SQLite 的连接级设置（CONTRACTS §6.1）。写在 DSN 的 _pragma 中，驱动对每个新物理连接都会执行。
const sqlitePragmas = "_pragma=journal_mode(WAL)&_pragma=synchronous(FULL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"

// stdio 传输时 stdout 专用于协议帧，GORM 日志只能写 stderr。
var gormLogger = newGormLogger(os.Stderr)

// newGormLogger 记录慢查询与错误 SQL。ParameterizedQueries 使日志只含占位符 SQL，
// 绑定参数（可能是业务数据或 secret）不进入日志（CONTRACTS §5.3）。
func newGormLogger(w io.Writer) logger.Interface {
	return logger.New(log.New(w, "gorm ", log.LstdFlags), logger.Config{
		SlowThreshold:        200 * time.Millisecond,
		LogLevel:             logger.Warn,
		ParameterizedQueries: true,
	})
}

// OpenSQLite 以受管默认值打开 path 处的 SQLite：单连接、WAL、synchronous=FULL、外键开启、
// busy_timeout=5000，并回读核对实际生效值；连接或核对失败时返回错误，不会改用空库或其他后端。
// 事务回调必须使用传入的 tx：单连接池下在事务里等待另一条独立查询会自锁。
func OpenSQLite(ctx context.Context, path string) (*gorm.DB, error) {
	if path == "" || strings.ContainsAny(path, "?#") {
		return nil, fmt.Errorf("SQLite 路径 %q 为空或含 ?/#", path)
	}
	db, err := gorm.Open(sqlite.Open(path+"?"+sqlitePragmas), &gorm.Config{Logger: gormLogger})
	if err != nil {
		return nil, fmt.Errorf("打开 SQLite 失败: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	if err := checkPragmas(ctx, db); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return db, nil
}

// PragmaValues 是受管 SQLite 连接上实际生效的设置。
type PragmaValues struct {
	JournalMode string
	Synchronous int // FULL = 2
	ForeignKeys int
	BusyTimeout int
}

// ReadPragmas 读取当前连接上实际生效的 PRAGMA 值。
func ReadPragmas(ctx context.Context, db *gorm.DB) (PragmaValues, error) {
	var v PragmaValues
	q := db.WithContext(ctx)
	err := errors.Join(
		q.Raw("PRAGMA journal_mode").Scan(&v.JournalMode).Error,
		q.Raw("PRAGMA synchronous").Scan(&v.Synchronous).Error,
		q.Raw("PRAGMA foreign_keys").Scan(&v.ForeignKeys).Error,
		q.Raw("PRAGMA busy_timeout").Scan(&v.BusyTimeout).Error,
	)
	return v, err
}

func checkPragmas(ctx context.Context, db *gorm.DB) error {
	v, err := ReadPragmas(ctx, db)
	if err != nil {
		return fmt.Errorf("读取 PRAGMA 失败: %w", err)
	}
	want := PragmaValues{JournalMode: "wal", Synchronous: 2, ForeignKeys: 1, BusyTimeout: 5000}
	if v != want {
		return fmt.Errorf("SQLite 设置未生效：实际 %+v，期望 %+v", v, want)
	}
	return nil
}
