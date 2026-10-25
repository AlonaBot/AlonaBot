package gormdb

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/proto"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// 需要 ALONABOT_TEST_PG_DSN（超级用户 URL，例如 postgres://postgres:pw@127.0.0.1:5432/postgres?sslmode=disable）；
// 未设置时跳过，证据中记为「未运行」。超级用户只用于测试中的预置与清理，插件连接只用专属 role。
func pgAdmin(t *testing.T) (*gorm.DB, *url.URL) {
	t.Helper()
	dsn := os.Getenv("ALONABOT_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("未设置 ALONABOT_TEST_PG_DSN，PostgreSQL 测试未运行")
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormLogger})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = closeWith(db, nil) })
	return db, u
}

// provision 按管理员预置流程创建插件专属 role 与 schema，并返回插件连接参数。
func provision(t *testing.T, admin *gorm.DB, u *url.URL) *pb.PostgresConnection {
	t.Helper()
	rnd := make([]byte, 4)
	_, _ = rand.Read(rnd)
	pluginID := "test-plugin-" + hex.EncodeToString(rnd) // 含 '-'，必须经过标识符清洗
	ident := manifest.PostgresIdentifier("ab_", pluginID)
	password := hex.EncodeToString(rnd) + "pw"
	stmts := []string{
		"CREATE ROLE " + ident + " LOGIN PASSWORD '" + password + "'",
		"CREATE SCHEMA " + ident + " AUTHORIZATION " + ident,
	}
	for _, s := range stmts {
		if err := admin.Exec(s).Error; err != nil {
			t.Fatalf("%s: %v", s, err)
		}
	}
	t.Cleanup(func() {
		admin.Exec("DROP SCHEMA IF EXISTS " + ident + " CASCADE")
		admin.Exec("DROP OWNED BY " + ident)
		admin.Exec("DROP ROLE IF EXISTS " + ident)
	})
	port, _ := strconv.Atoi(u.Port())
	return &pb.PostgresConnection{
		Host: u.Hostname(), Port: uint32(port), Database: strings.TrimPrefix(u.Path, "/"),
		User: ident, Password: password, Schema: ident, SslMode: u.Query().Get("sslmode"),
	}
}

// 插件 role 的 search_path 只有专属 schema，表落在其中；事务回滚与 versioned 账本同事务在 PG 上成立；读取其他 schema 被拒绝。
func TestPostgresPluginRole(t *testing.T) {
	ctx := context.Background()
	admin, u := pgAdmin(t)
	conn := provision(t, admin, u)

	// 另一个 schema 中的数据，插件 role 不应能访问。
	other := conn.Schema + "_other"
	if err := admin.Exec("CREATE SCHEMA " + other + "; CREATE TABLE " + other + ".secret(v text); INSERT INTO " + other + ".secret VALUES ('x')").Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Exec("DROP SCHEMA IF EXISTS " + other + " CASCADE") })

	db, err := OpenPostgres(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = closeWith(db, nil) })

	var searchPath string
	db.Raw("SHOW search_path").Scan(&searchPath)
	t.Logf("search_path=%s", searchPath)
	if searchPath != conn.Schema {
		t.Fatalf("search_path = %q，应只有插件 schema", searchPath)
	}

	// CRUD：未限定 schema 的表落在插件 schema。
	if err := db.AutoMigrate(&item{}); err != nil {
		t.Fatal(err)
	}
	var tableSchema string
	admin.Raw("SELECT table_schema FROM information_schema.tables WHERE table_name = 'items' AND table_schema = ?", conn.Schema).Scan(&tableSchema)
	if tableSchema != conn.Schema {
		t.Fatalf("items 表不在插件 schema 中（%q）", tableSchema)
	}
	it := item{Name: "a"}
	if err := db.Create(&it).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&it).Update("name", "b").Error; err != nil {
		t.Fatal(err)
	}
	var got item
	if err := db.First(&got, it.ID).Error; err != nil || got.Name != "b" {
		t.Fatalf("读取 %+v, %v", got, err)
	}

	// 事务失败后数据保持不变。
	errBoom := errors.New("boom")
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&item{}).Where("id = ?", it.ID).Update("name", "changed").Error; err != nil {
			return err
		}
		return errBoom
	})
	if !errors.Is(err, errBoom) {
		t.Fatalf("事务应返回回调错误，实际 %v", err)
	}
	db.First(&got, it.ID)
	if got.Name != "b" {
		t.Fatalf("回滚后值为 %q，应保持 b", got.Name)
	}

	// 版本化迁移在 PG 上同样与账本同事务。
	failing := func(_ context.Context, tx *gorm.DB) error {
		if err := tx.Exec("CREATE TABLE t2(id int)").Error; err != nil {
			return err
		}
		return errBoom
	}
	if _, err := RunVersioned(ctx, db, []Migration{{1, func(_ context.Context, tx *gorm.DB) error { return tx.Exec("CREATE TABLE t1(id int)").Error }}, {2, failing}}); !errors.Is(err, errBoom) {
		t.Fatalf("迁移 2 应失败，实际 %v", err)
	}
	if ids, _ := AppliedMigrations(ctx, db); len(ids) != 1 || ids[0] != 1 || db.Migrator().HasTable("t2") {
		t.Fatalf("PG 上失败迁移不得留下账本或 DDL：ledger=%v", ids)
	}

	// 越权：读取其他 schema 被拒绝。
	var v string
	if err := db.Raw("SELECT v FROM " + other + ".secret").Scan(&v).Error; err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("插件 role 不应能读取其他 schema，实际 v=%q err=%v", v, err)
	}
}

// clone 深拷贝连接参数。
func clone(c *pb.PostgresConnection) *pb.PostgresConnection {
	return proto.Clone(c).(*pb.PostgresConnection)
}

// schema 不存在、密码错误（错误信息不含密码）或 schema 名非法时打开失败。
func TestOpenPostgresRejects(t *testing.T) {
	ctx := context.Background()
	admin, u := pgAdmin(t)
	conn := provision(t, admin, u)

	wrongSchema := clone(conn)
	wrongSchema.Schema = conn.Schema + "_missing"
	if _, err := OpenPostgres(ctx, wrongSchema); err == nil || !strings.Contains(err.Error(), "会话核对失败") {
		t.Fatalf("不存在的 schema 应在核对时失败，实际 %v", err)
	}
	badPassword := clone(conn)
	badPassword.Password = "wrong"
	if _, err := OpenPostgres(ctx, badPassword); err == nil || strings.Contains(err.Error(), "wrong") {
		t.Fatalf("错误密码应失败且错误信息不含密码，实际 %v", err)
	}
	for _, s := range []string{"Bad-Name", "x;drop", ""} {
		c := clone(conn)
		c.Schema = s
		if _, err := OpenPostgres(ctx, c); err == nil {
			t.Fatalf("非法 schema %q 应被拒绝", s)
		}
	}
}
