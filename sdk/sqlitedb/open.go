// Package sqlitedb 提供受管 SQLite 的标准库 database/sql 入口。
//
// [Open] 返回当前 Session 复用的连接，由 SDK 在停止或准备失败时关闭。
// [OpenFile] 打开独立文件，由调用方关闭。[ReadPragmas] 回读连接实际生效的设置。
// 两个入口均采用单连接、WAL、FULL、外键和 5 秒忙等待，见 CONTRACTS §6.1。
//
// 包的组成：[SelfMigrator] 承载自行管理的迁移，[VersionedMigrator] 提供逐步迁移；
// [RunVersioned] 与 [AppliedMigrations] 执行迁移并查询与 GORM 入口共用的账本。
package sqlitedb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/sdk/internal/dbhook"
	_ "modernc.org/sqlite" // 注册 database/sql 驱动，GORM 保持按需链接。
)

// ErrNoDatabase 表示当前插件未配置受管数据库。
var ErrNoDatabase = errors.New("this plugin has no database")

// 受管连接的 PRAGMA 值由 CONTRACTS §6.1 规定。
const (
	journalModeWAL    = "wal"
	synchronousFull   = 2
	foreignKeysOn     = 1
	busyTimeoutMillis = 5000
)

// sqlitePragmas 随 DSN 应用到每个新物理连接，连接被替换后设置仍然生效。
var sqlitePragmas = "_pragma=journal_mode(WAL)&_pragma=synchronous(FULL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(" + strconv.Itoa(busyTimeoutMillis) + ")"

// Open 返回当前 Session 的受管连接，同一 Session 重复调用复用同一个 *sql.DB。
// 迁移阶段开始前及 Session 停止后返回 FAILED_PRECONDITION；无数据库时返回 ErrNoDatabase。
// 只有 first_install 允许创建文件；连接由 SDK 关闭，调用方须保持它打开。
// 事务回调须使用传入的 tx，在事务内等待此单连接池的独立查询会自锁。
func Open(ctx context.Context) (*sql.DB, error) {
	return dbhook.Connection(ctx, func(ctx context.Context, descriptor *pb.DatabaseDescriptor) (*sql.DB, error) {
		switch descriptor.GetBackend() {
		case pb.DatabaseBackend_DATABASE_BACKEND_NONE:
			return nil, ErrNoDatabase
		case pb.DatabaseBackend_DATABASE_BACKEND_SQLITE:
			return openFile(ctx, descriptor.GetSqlitePath(), descriptor.GetActivationReason() == pb.ActivationReason_ACTIVATION_REASON_FIRST_INSTALL)
		default:
			return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "invalid plugin database backend"}
		}
	})
}

// OpenFile 打开 path 处的 SQLite，文件缺失时创建；返回连接由调用方关闭。
// 连接采用单连接、WAL、FULL、外键和 5000ms 忙等待，并回读核对实际值。
// 事务回调须使用传入的 tx，在事务内等待此单连接池的独立查询会自锁。
func OpenFile(ctx context.Context, path string) (*sql.DB, error) {
	return openFile(ctx, path, true)
}

func openFile(ctx context.Context, path string, create bool) (*sql.DB, error) {
	// 1. 将文件路径编码为 URI，mode=rw 在 SQLite 打开时阻止丢失文件被重新创建。
	if path == "" || strings.ContainsAny(path, "?#") {
		return nil, fmt.Errorf("SQLite path %q is empty or contains ? or #", path)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	p := filepath.ToSlash(absolute)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p // Windows 盘符需写为 file:///C:/...，与 Core 状态库相同。
	}
	query := sqlitePragmas
	if !create {
		query += "&mode=rw"
	}
	dsn := (&url.URL{Scheme: "file", Path: p, RawQuery: query}).String()
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// 2. 限制物理连接数并回读；查询会实际打开连接，失败时保留原始错误与关闭错误。
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := checkPragmas(ctx, db); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return db, nil
}

// PragmaValues 记录受管 SQLite 连接实际生效的设置。
type PragmaValues struct {
	JournalMode string // wal
	Synchronous int    // 2（FULL）
	ForeignKeys int    // 1（开启）
	BusyTimeout int    // 忙等待毫秒数，5000
}

// ReadPragmas 回读 db 上实际生效的 PRAGMA 值，并保留所有查询错误。
func ReadPragmas(ctx context.Context, db *sql.DB) (PragmaValues, error) {
	var values PragmaValues
	err := errors.Join(
		db.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&values.JournalMode),
		db.QueryRowContext(ctx, "PRAGMA synchronous").Scan(&values.Synchronous),
		db.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&values.ForeignKeys),
		db.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&values.BusyTimeout),
	)
	return values, err
}

func checkPragmas(ctx context.Context, db *sql.DB) error {
	values, err := ReadPragmas(ctx, db)
	if err != nil {
		return err
	}
	want := PragmaValues{JournalMode: journalModeWAL, Synchronous: synchronousFull, ForeignKeys: foreignKeysOn, BusyTimeout: busyTimeoutMillis}
	if values != want {
		return fmt.Errorf("SQLite settings not applied: got %+v, want %+v", values, want)
	}
	return nil
}
