package gormdb

import (
	"context"

	"github.com/AlonaBot/AlonaBot/sdk/internal/dbhook"
	"github.com/AlonaBot/AlonaBot/sdk/sqlitedb"
	"gorm.io/gorm"
)

// ErrNoDatabase 表示当前插件未配置受管数据库，与 sqlitedb 共用同一错误值。
var ErrNoDatabase = sqlitedb.ErrNoDatabase

// Open 在 ctx 所属 Session 的同一个受管 *sql.DB 上返回原生 GORM 包装，重复调用复用它。
// 可用期从迁移阶段开始至 Session 停止；底层连接由 SDK 关闭。
// 事务回调须使用传入的 tx，在事务内等待单连接池的独立查询会自锁。
func Open(ctx context.Context) (*gorm.DB, error) {
	db, err := sqlitedb.Open(ctx)
	if err != nil {
		return nil, err
	}
	wrapped, err := dbhook.GORM(ctx, db, func() (any, error) { return wrap(ctx, db) })
	if err != nil {
		return nil, err
	}
	return wrapped.(*gorm.DB), nil
}
