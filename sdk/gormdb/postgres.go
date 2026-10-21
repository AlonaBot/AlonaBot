package gormdb

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"

	pb "github.com/AlonaBot/alonabot/contracts/pb/alonabotv1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var pgIdent = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,62}$`)

// OpenPostgres 用 Bootstrap 下发的插件专属 role/schema 连接 PostgreSQL（CONTRACTS §6.3）。
// search_path 只设为该插件 schema（pg_catalog 始终隐式可见），连接后核对 current_user 与
// current_schemas(false)；任何不一致都返回错误，不会改用其他 schema 或后端。
func OpenPostgres(ctx context.Context, c *pb.PostgresConnection) (*gorm.DB, error) {
	switch {
	case c == nil:
		return nil, fmt.Errorf("缺少 PostgreSQL 连接参数")
	case !pgIdent.MatchString(c.GetSchema()):
		return nil, fmt.Errorf("schema %q 不是合法的小写标识符", c.GetSchema())
	case c.GetHost() == "" || c.GetDatabase() == "" || c.GetUser() == "":
		return nil, fmt.Errorf("PostgreSQL 连接参数缺少 host、database 或 user")
	}
	q := url.Values{"search_path": {c.GetSchema()}}
	if c.GetSslMode() != "" {
		q.Set("sslmode", c.GetSslMode())
	}
	host := c.GetHost()
	if c.GetPort() != 0 {
		host = net.JoinHostPort(host, strconv.Itoa(int(c.GetPort())))
	}
	dsn := (&url.URL{
		Scheme: "postgres", User: url.UserPassword(c.GetUser(), c.GetPassword()),
		Host: host, Path: "/" + c.GetDatabase(), RawQuery: q.Encode(),
	}).String()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormLogger})
	if err != nil {
		// 不把 DSN（含密码）放进错误信息。
		return nil, fmt.Errorf("连接 PostgreSQL %s/%s 失败: %w", host, c.GetDatabase(), err)
	}
	var user string
	var schemas []string
	if err := db.WithContext(ctx).Raw("SELECT current_user").Scan(&user).Error; err != nil {
		return nil, closeWith(db, err)
	}
	if err := db.WithContext(ctx).Raw("SELECT unnest(current_schemas(false))").Scan(&schemas).Error; err != nil {
		return nil, closeWith(db, err)
	}
	if user != c.GetUser() || len(schemas) != 1 || schemas[0] != c.GetSchema() {
		return nil, closeWith(db, fmt.Errorf("PostgreSQL 会话核对失败：current_user=%q current_schemas=%v，期望 %q/[%s]（schema 不存在或无 USAGE 权限）",
			user, schemas, c.GetUser(), c.GetSchema()))
	}
	return db, nil
}

func closeWith(db *gorm.DB, err error) error {
	if sqlDB, e := db.DB(); e == nil {
		_ = sqlDB.Close()
	}
	return err
}
