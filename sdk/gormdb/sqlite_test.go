package gormdb

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlonaBot/AlonaBot/sdk/sqlitedb"
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
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	v, err := sqlitedb.ReadPragmas(context.Background(), sqlDB)
	if err != nil {
		t.Fatal(err)
	}
	// 回读的实际值（CONTRACTS §6.1）。
	if v.JournalMode != "wal" || v.Synchronous != 2 || v.ForeignKeys != 1 || v.BusyTimeout != 5000 {
		t.Fatalf("PRAGMA values %+v", v)
	}
	if got := sqlDB.Stats().MaxOpenConnections; got != 1 {
		t.Fatalf("connection pool limit = %d, want 1", got)
	}
	if _, err := OpenSQLite(context.Background(), "a?b"); err == nil {
		t.Fatal("path with ? should be rejected")
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
		t.Fatalf("read %+v, %v", got, err)
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
		t.Fatalf("transaction should return the callback error, got %v", err)
	}
	var n int64
	db.Model(&item{}).Where("name = ?", "rolled-back").Count(&n)
	if n != 0 {
		t.Fatal("rolled-back row still exists")
	}

	// 事务提交。
	if err := db.Transaction(func(tx *gorm.DB) error { return tx.Create(&item{Name: "committed"}).Error }); err != nil {
		t.Fatal(err)
	}
	db.Model(&item{}).Where("name = ?", "committed").Count(&n)
	if n != 1 {
		t.Fatal("committed row does not exist")
	}

	if err := db.Delete(&item{}, it.ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.First(&got, it.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("record should be gone after delete, got %v", err)
	}
}

// 违反外键约束的写入被拒绝。
func TestForeignKeysEnforced(t *testing.T) {
	db, _ := openTemp(t, "fk.sqlite")
	if err := db.Exec(`CREATE TABLE parent(id INTEGER PRIMARY KEY); CREATE TABLE child(id INTEGER PRIMARY KEY, parent_id INTEGER NOT NULL REFERENCES parent(id))`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO child(id, parent_id) VALUES (1, 99)`).Error; err == nil {
		t.Fatal("inserting a missing parent should fail with foreign_keys=ON")
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
			return errors.New("underlying connection does not support Backup")
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
		t.Fatalf("expected a non-empty WAL file: %v", err)
	}

	sqlDB, _ := db.DB()
	candidate := filepath.Join(t.TempDir(), "candidate.sqlite")
	if err := backupTo(ctx, sqlDB, candidate); err != nil {
		t.Fatalf("Backup failed: %v", err)
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
		t.Fatalf("candidate copy content = %v", names)
	}
	// 候选可写，且写入不影响旧库。
	if err := cand.Create(&item{Name: "candidate-only"}).Error; err != nil {
		t.Fatal(err)
	}
	var n int64
	db.Model(&item{}).Where("name = ?", "candidate-only").Count(&n)
	if n != 0 {
		t.Fatal("candidate write appeared in the old database")
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
		t.Fatalf("ran=%v err=%v, want only 1 completed and the Up error returned", ran, err)
	}
	applied, _ := AppliedMigrations(ctx, db)
	if len(applied) != 1 || applied[0] != 1 {
		t.Fatalf("ledger = %v, a failed migration must not be recorded as done", applied)
	}
	if db.Migrator().HasTable("b") {
		t.Fatal("DDL in the failed migration was not rolled back")
	}

	// 修复后重跑：只执行未完成的 ID。
	ran, err = RunVersioned(ctx, db, []Migration{{1, createA}, {2, createB}})
	if err != nil || len(ran) != 1 || ran[0] != 2 {
		t.Fatalf("rerun ran=%v err=%v, want only 2 executed", ran, err)
	}
	if !db.Migrator().HasTable("b") {
		t.Fatal("migration 2 did not take effect")
	}
	if ran, err = RunVersioned(ctx, db, []Migration{{1, createA}, {2, createB}}); err != nil || len(ran) != 0 {
		t.Fatalf("no migration should run once all are done: %v %v", ran, err)
	}
	if _, err := RunVersioned(ctx, db, []Migration{{2, createB}, {1, createA}}); err == nil {
		t.Fatal("a non-increasing migration list should be rejected")
	}
}

// 失败 SQL 的日志只能含占位符，不能含绑定参数的值。
func TestGormLogOmitsParameters(t *testing.T) {
	db, _ := openTemp(t, "log.sqlite")
	if err := db.Exec(`CREATE TABLE parent(id INTEGER PRIMARY KEY); CREATE TABLE child(id INTEGER PRIMARY KEY, parent_id INTEGER NOT NULL REFERENCES parent(id), note TEXT)`).Error; err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	quiet := db.Session(&gorm.Session{Logger: newGormLogger(slog.New(slog.NewJSONHandler(&buf, nil)))})
	const secret = "SECRET-VALUE-42"
	if err := quiet.Exec("INSERT INTO child(id, parent_id, note) VALUES (?, ?, ?)", 1, 99, secret).Error; err == nil {
		t.Fatal("foreign key constraint should make the insert fail")
	}
	out := buf.String()
	if !strings.Contains(out, "INSERT INTO child") {
		t.Fatalf("failing SQL should be logged, got log: %q", out)
	}
	if strings.Contains(out, secret) {
		t.Fatalf("log leaked a parameter value: %q", out)
	}
}
