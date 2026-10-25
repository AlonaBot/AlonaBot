// 维护者：FlanChanXwO
// 创建日期：2026-10-21
// 职责：在真实 SQLite 上测试 PRAGMA、CRUD 与事务、外键、WAL 在线备份、versioned 账本同事务，以及 SQL 日志不含参数值。

package gormdb

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/gorm"
	moderncsqlite "modernc.org/sqlite"
)

// item 用作测试的 GORM 模型。
type item struct {
	ID    uint `gorm:"primaryKey"`
	Name  string
	Owner *uint // 可空的所有者 ID
}

// openTemp 在临时目录打开名为 name 的受管 SQLite，测试结束时关闭。
func openTemp(t *testing.T, name string) (*gorm.DB, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	db, err := OpenSQLite(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return db, path
}

// 打开后 WAL、FULL、外键与 busy timeout 实际生效。
func TestOpenSQLitePragmas(t *testing.T) {
	db, _ := openTemp(t, "p.sqlite")
	v, err := ReadPragmas(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	// 回读的实际值（CONTRACTS §6.1）。
	if v.JournalMode != "wal" || v.Synchronous != 2 || v.ForeignKeys != 1 || v.BusyTimeout != 5000 {
		t.Fatalf("PRAGMA 实际值 %+v", v)
	}
	sqlDB, _ := db.DB()
	if got := sqlDB.Stats().MaxOpenConnections; got != 1 {
		t.Fatalf("连接池上限 = %d，应为 1", got)
	}
	if _, err := OpenSQLite(context.Background(), "a?b"); err == nil {
		t.Fatal("含 ? 的路径应被拒绝")
	}
}

// 基本 CRUD 可用；事务回调返回错误时写入不可见，正常返回时提交。
func TestCRUDAndTx(t *testing.T) {
	ctx := context.Background()
	db, _ := openTemp(t, "crud.sqlite")
	if err := db.AutoMigrate(&item{}); err != nil {
		t.Fatal(err)
	}
	it := item{Name: "a"}
	if err := db.WithContext(ctx).Create(&it).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&it).Update("name", "b").Error; err != nil {
		t.Fatal(err)
	}
	var got item
	if err := db.First(&got, it.ID).Error; err != nil || got.Name != "b" {
		t.Fatalf("读取 %+v, %v", got, err)
	}

	// 事务回滚：回调返回错误后写入不可见。
	errBoom := errors.New("boom")
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&item{Name: "rolled-back"}).Error; err != nil {
			return err
		}
		return errBoom
	})
	if !errors.Is(err, errBoom) {
		t.Fatalf("事务应返回回调错误，实际 %v", err)
	}
	var n int64
	db.Model(&item{}).Where("name = ?", "rolled-back").Count(&n)
	if n != 0 {
		t.Fatal("回滚后的行仍然存在")
	}

	// 事务提交。
	if err := db.Transaction(func(tx *gorm.DB) error { return tx.Create(&item{Name: "committed"}).Error }); err != nil {
		t.Fatal(err)
	}
	db.Model(&item{}).Where("name = ?", "committed").Count(&n)
	if n != 1 {
		t.Fatal("提交后的行不存在")
	}

	if err := db.Delete(&item{}, it.ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.First(&got, it.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("删除后应找不到记录，实际 %v", err)
	}
}

// 违反外键约束的写入被拒绝。
func TestForeignKeysEnforced(t *testing.T) {
	db, _ := openTemp(t, "fk.sqlite")
	if err := db.Exec(`CREATE TABLE parent(id INTEGER PRIMARY KEY); CREATE TABLE child(id INTEGER PRIMARY KEY, parent_id INTEGER NOT NULL REFERENCES parent(id))`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO child(id, parent_id) VALUES (1, 99)`).Error; err == nil {
		t.Fatal("foreign_keys=ON 时插入不存在的父行应失败")
	}
}

// backupTo 用 modernc 的在线 Backup API 把 src 复制到 dst。
func backupTo(ctx context.Context, src *sql.DB, dst string) error {
	conn, err := src.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	return conn.Raw(func(dc any) error {
		b, ok := dc.(interface {
			// NewBackup 创建到 dst 的在线备份。
			NewBackup(string) (*moderncsqlite.Backup, error)
		})
		if !ok {
			return errors.New("底层连接不支持 Backup")
		}
		bk, err := b.NewBackup(dst)
		if err != nil {
			return err
		}
		if _, err := bk.Step(-1); err != nil {
			_ = bk.Finish()
			return err
		}
		return bk.Finish()
	})
}

// WAL 模式下在线备份得到的候选包含尚未 checkpoint 的写入；备份后旧库与候选的写入互不影响。
func TestWALBackupToCandidate(t *testing.T) {
	ctx := context.Background()
	db, path := openTemp(t, "active.sqlite")
	if err := db.AutoMigrate(&item{}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"x", "y", "z"} {
		if err := db.Create(&item{Name: name}).Error; err != nil {
			t.Fatal(err)
		}
	}
	// 数据仍在 WAL 中（尚未 checkpoint）：证明备份读的是运行中 WAL 库的一致视图，而不是主文件。
	if fi, err := os.Stat(path + "-wal"); err != nil || fi.Size() == 0 {
		t.Fatalf("预期存在非空 WAL 文件: %v", err)
	}

	sqlDB, _ := db.DB()
	candidate := filepath.Join(t.TempDir(), "candidate.sqlite")
	if err := backupTo(ctx, sqlDB, candidate); err != nil {
		t.Fatalf("Backup 失败: %v", err)
	}
	// 备份之后旧库的写入不影响候选副本。
	if err := db.Create(&item{Name: "after-backup"}).Error; err != nil {
		t.Fatal(err)
	}

	cand, err := OpenSQLite(ctx, candidate)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { c, _ := cand.DB(); _ = c.Close() }()
	var names []string
	if err := cand.Model(&item{}).Order("id").Pluck("name", &names).Error; err != nil {
		t.Fatal(err)
	}
	if len(names) != 3 || names[0] != "x" || names[2] != "z" {
		t.Fatalf("候选副本内容 = %v", names)
	}
	// 候选可写，且写入不影响旧库。
	if err := cand.Create(&item{Name: "candidate-only"}).Error; err != nil {
		t.Fatal(err)
	}
	var n int64
	db.Model(&item{}).Where("name = ?", "candidate-only").Count(&n)
	if n != 0 {
		t.Fatal("候选写入出现在旧库中")
	}
}

// 迁移失败时整步回滚、不记账并停止；再次执行只跑未完成的迁移。
func TestRunVersionedLedgerInSameTx(t *testing.T) {
	ctx := context.Background()
	db, _ := openTemp(t, "mig.sqlite")
	errUp := errors.New("up failed")
	createA := func(_ context.Context, tx *gorm.DB) error { return tx.Exec("CREATE TABLE a(id INTEGER)").Error }
	failingB := func(_ context.Context, tx *gorm.DB) error {
		if err := tx.Exec("CREATE TABLE b(id INTEGER)").Error; err != nil {
			return err
		}
		return errUp // 已执行的 DDL 必须随事务回滚
	}
	createB := func(_ context.Context, tx *gorm.DB) error { return tx.Exec("CREATE TABLE b(id INTEGER)").Error }

	ran, err := RunVersioned(ctx, db, []Migration{{1, createA}, {2, failingB}, {3, createA}})
	if !errors.Is(err, errUp) || len(ran) != 1 || ran[0] != 1 {
		t.Fatalf("ran=%v err=%v，期望只完成 1 并返回 Up 错误", ran, err)
	}
	applied, _ := AppliedMigrations(ctx, db)
	if len(applied) != 1 || applied[0] != 1 {
		t.Fatalf("账本 = %v，失败的迁移不得记录成功", applied)
	}
	if db.Migrator().HasTable("b") {
		t.Fatal("失败迁移中的 DDL 没有回滚")
	}

	// 修复后重跑：只执行未完成的 ID。
	ran, err = RunVersioned(ctx, db, []Migration{{1, createA}, {2, createB}})
	if err != nil || len(ran) != 1 || ran[0] != 2 {
		t.Fatalf("重跑 ran=%v err=%v，期望只执行 2", ran, err)
	}
	if !db.Migrator().HasTable("b") {
		t.Fatal("迁移 2 未生效")
	}
	if ran, err = RunVersioned(ctx, db, []Migration{{1, createA}, {2, createB}}); err != nil || len(ran) != 0 {
		t.Fatalf("全部完成后再跑不应执行任何迁移: %v %v", ran, err)
	}
	if _, err := RunVersioned(ctx, db, []Migration{{2, createB}, {1, createA}}); err == nil {
		t.Fatal("非递增的迁移列表应被拒绝")
	}
}

// 失败 SQL 的日志只能含占位符，不能含绑定参数的值。
func TestGormLogOmitsParameters(t *testing.T) {
	db, _ := openTemp(t, "log.sqlite")
	if err := db.Exec(`CREATE TABLE parent(id INTEGER PRIMARY KEY); CREATE TABLE child(id INTEGER PRIMARY KEY, parent_id INTEGER NOT NULL REFERENCES parent(id), note TEXT)`).Error; err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	quiet := db.Session(&gorm.Session{Logger: newGormLogger(&buf)})
	const secret = "SECRET-VALUE-42"
	if err := quiet.Exec("INSERT INTO child(id, parent_id, note) VALUES (?, ?, ?)", 1, 99, secret).Error; err == nil {
		t.Fatal("外键约束应使插入失败")
	}
	out := buf.String()
	if !strings.Contains(out, "INSERT INTO child") {
		t.Fatalf("错误 SQL 应被记录，实际日志: %q", out)
	}
	if strings.Contains(out, secret) {
		t.Fatalf("日志泄漏了参数值: %q", out)
	}
}
