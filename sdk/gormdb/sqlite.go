package gormdb

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/AlonaBot/AlonaBot/sdk/internal/dbhook"
	"github.com/AlonaBot/AlonaBot/sdk/sqlitedb"
	"github.com/libtnb/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// slowQueryThreshold 超过它的 SQL 会以慢查询记录。
const slowQueryThreshold = 200 * time.Millisecond

// newGormLogger 经所属 Session 的 Logger 记录慢查询与错误 SQL。
// ParameterizedQueries 使日志只含占位符 SQL，绑定参数（可能是业务数据或 secret）不进入日志。
func newGormLogger(l *slog.Logger) logger.Interface {
	return logger.NewSlogLogger(l, logger.Config{
		SlowThreshold:        slowQueryThreshold,
		LogLevel:             logger.Warn,
		ParameterizedQueries: true,
	})
}

// OpenSQLite 以受管默认值打开 path 处的 SQLite：单连接、WAL、synchronous=FULL、外键开启、
// busy_timeout=5000，并回读核对实际生效值。返回连接由调用方关闭。
// 事务回调必须使用传入的 tx：单连接池下在事务里等待另一条独立查询会自锁。
func OpenSQLite(ctx context.Context, path string) (*gorm.DB, error) {
	db, err := sqlitedb.OpenFile(ctx, path)
	if err != nil {
		return nil, err
	}
	wrapped, err := wrap(ctx, db)
	if err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return wrapped, nil
}

func wrap(ctx context.Context, db *sql.DB) (*gorm.DB, error) {
	// libtnb 初始化时以 Background 查询 SQLite 版本；先按调用方 ctx 取得物理连接，
	// 避免单连接池被事务占用时，初始化查询忽略迁移取消并一直等待。
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	dialector := sqlite.New(sqlite.Config{Conn: conn})
	wrapped, err := gorm.Open(dialector, &gorm.Config{Logger: newGormLogger(dbhook.Logger(ctx))})
	closeError := conn.Close()
	if err != nil || closeError != nil {
		return nil, errors.Join(err, closeError)
	}
	// 初始化专用的 *sql.Conn 已归还，之后全部操作使用原来的受管连接池。
	dialector.(*sqlite.Dialector).Conn = db
	wrapped.ConnPool = db
	wrapped.Statement.ConnPool = db
	return wrapped, nil
}
