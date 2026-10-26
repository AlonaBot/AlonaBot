package gormdb

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"regexp"
	"strconv"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// maxIdentLen 是 PostgreSQL 标识符的最大字节数（NAMEDATALEN - 1）。
const maxIdentLen = 63

var pgIdent = regexp.MustCompile(fmt.Sprintf(`^[a-z_][a-z0-9_]{0,%d}$`, maxIdentLen-1))

// OpenPostgres 用 Bootstrap 下发的插件专属 role/schema 连接 PostgreSQL（CONTRACTS §6.3）。
// search_path 只设为该插件 schema（pg_catalog 始终隐式可见），连接后核对 current_user 与
// current_schemas(false)；任何不一致都返回错误，不会改用其他 schema 或后端。
func OpenPostgres(ctx context.Context, c *pb.PostgresConnection) (*gorm.DB, error) {
	switch {
	case c == nil:
		return nil, fmt.Errorf("missing PostgreSQL connection parameters")
	case !pgIdent.MatchString(c.GetSchema()):
		return nil, fmt.Errorf("schema %q is not a valid lowercase identifier", c.GetSchema())
	case c.GetHost() == "" || c.GetDatabase() == "" || c.GetUser() == "":
		return nil, fmt.Errorf("PostgreSQL connection parameters are missing host, database or user")
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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: newGormLogger(slog.Default())})
	if err != nil {
		// 不把 DSN（含密码）放进错误信息。
		return nil, fmt.Errorf("connect to PostgreSQL %s/%s: %w", host, c.GetDatabase(), err)
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
		return nil, closeWith(db, fmt.Errorf("PostgreSQL session check failed: current_user=%q current_schemas=%v, want %q/[%s] (schema missing or no USAGE privilege)",
			user, schemas, c.GetUser(), c.GetSchema()))
	}
	return db, nil
}

// closeWith 关闭 db 的底层连接池并原样返回 err，用于打开失败后的清理。
func closeWith(db *gorm.DB, err error) error {
	if sqlDB, e := db.DB(); e == nil {
		_ = sqlDB.Close()
	}
	return err
}
