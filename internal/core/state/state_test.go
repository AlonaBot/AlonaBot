package state

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// TestMain 在设置 STATE_HOLD_DIR 时作为子进程运行：以另一个 Core 的身份打开 data_dir 并持有，直到 stdin 关闭。
func TestMain(m *testing.M) {
	if dir := os.Getenv("STATE_HOLD_DIR"); dir != "" {
		s, err := Open(context.Background(), dir)
		if err != nil {
			os.Stdout.WriteString("error " + err.Error() + "\n")
			os.Exit(2)
		}
		os.Stdout.WriteString("opened\n")
		_, _ = io.Copy(io.Discard, os.Stdin)
		_ = s.Close()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// mustOpen 打开 dir 的状态库，测试结束时关闭。
func mustOpen(t *testing.T, dir string) *Store {
	t.Helper()
	s, err := Open(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// 连接设置实际生效；关闭重开后记录保留，schema 为最新版本。
func TestOpenPragmasAndReopen(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	s, err := Open(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	p, err := s.ReadPragmas(ctx)
	if err != nil || p != (Pragmas{"wal", 2, 1, 5000}) {
		t.Fatalf("PRAGMA 实际值 %+v, %v", p, err)
	}
	if err := s.AddInstall(ctx, Install{PluginID: "demo", Source: "local", Entrypoint: "bin/demo", BuildDigest: []byte{1}}); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s2 := mustOpen(t, dir)
	if in, err := s2.GetInstall(ctx, "demo"); err != nil || in.Entrypoint != "bin/demo" {
		t.Fatalf("重开后记录丢失: %+v %v", in, err)
	}
	var v int
	if err := s2.db.QueryRow("SELECT version FROM schema_version").Scan(&v); err != nil || v != len(migrations) {
		t.Fatalf("schema 版本 = %d, %v", v, err)
	}
}

// 同一进程内再次打开同一 data_dir 被拒绝。
func TestSecondCoreRejectedInProcess(t *testing.T) {
	dir := t.TempDir()
	mustOpen(t, dir)
	if _, err := Open(context.Background(), dir); !errors.Is(err, ErrDataDirLocked) {
		t.Fatalf("同一 data_dir 的第二个 Core 应被拒绝，实际 %v", err)
	}
}

// 另一进程持有 data_dir 时打开被拒绝；持有者退出后可以打开。
func TestSecondCoreRejectedAcrossProcesses(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = append(os.Environ(), "STATE_HOLD_DIR="+dir)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if line, err := bufio.NewReader(stdout).ReadString('\n'); err != nil || line != "opened\n" {
		t.Fatalf("子进程 Core 未能打开: %q %v", line, err)
	}
	if _, err := Open(context.Background(), dir); !errors.Is(err, ErrDataDirLocked) {
		t.Fatalf("另一进程持有 data_dir 时应被拒绝，实际 %v", err)
	}
	_ = stdin.Close()
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
	mustOpen(t, dir) // 持有者确认退出后可以启动
}

// CommitActive 核对预期旧代次，拒绝陈旧候选、重复提交、他人或不存在的代次；启用状态可持久化。
func TestCommitActiveChecksExpectedOld(t *testing.T) {
	ctx := context.Background()
	s := mustOpen(t, t.TempDir())
	for _, id := range []string{"a", "b"} {
		if err := s.AddInstall(ctx, Install{PluginID: id, Source: "local", Entrypoint: "x", BuildDigest: []byte{1}}); err != nil {
			t.Fatal(err)
		}
	}
	for _, g := range []struct {
		plugin string // 代次所属插件
		gen    string // 代次 ID
	}{{"a", "a-g1"}, {"a", "a-g2"}, {"a", "a-g3"}, {"b", "b-g1"}} {
		if err := s.AddGeneration(ctx, g.plugin, g.gen, `{}`, 1); err != nil {
			t.Fatal(err)
		}
	}
	if g, _ := s.Active(ctx, "a"); g != "" {
		t.Fatalf("未提交前 Active 应为空，实际 %q", g)
	}
	if err := s.CommitActive(ctx, "a", "", "a-g1"); err != nil {
		t.Fatal(err)
	}
	// 两个候选都以「无 Active」为前提：后到的那个是陈旧候选。
	if err := s.CommitActive(ctx, "a", "", "a-g2"); !errors.Is(err, ErrStaleGeneration) {
		t.Fatalf("陈旧的预期旧代次应被拒绝，实际 %v", err)
	}
	if g, _ := s.Active(ctx, "a"); g != "a-g1" {
		t.Fatalf("陈旧提交失败后 Active 应保持 a-g1，实际 %q", g)
	}
	if err := s.CommitActive(ctx, "a", "a-g1", "a-g2"); err != nil {
		t.Fatal(err)
	}
	if err := s.CommitActive(ctx, "a", "a-g2", "a-g1"); err == nil {
		t.Fatal("已提交过的代次不能再次提交（恢复必须使用新代次）")
	}
	if err := s.CommitActive(ctx, "a", "a-g2", "b-g1"); err == nil {
		t.Fatal("不能把其他插件的代次提交为 Active")
	}
	if err := s.CommitActive(ctx, "a", "a-g2", "missing"); err == nil {
		t.Fatal("不存在的代次不能提交")
	}
	if g, _ := s.Active(ctx, "a"); g != "a-g2" {
		t.Fatalf("失败的提交不应改变 Active，实际 %q", g)
	}
	if err := s.CommitActive(ctx, "nope", "", "a-g3"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("未安装的插件应返回 ErrNotFound，实际 %v", err)
	}
	if err := s.SetEnabled(ctx, "a", true); err != nil {
		t.Fatal(err)
	}
	if e, _ := s.Enabled(ctx, "a"); !e {
		t.Fatal("启用状态未持久化")
	}
}

// 垃圾内容、空文件或其他 SQLite 库都拒绝打开，文件不被改写，且失败后释放 core.lock。
func TestCorruptStateNotRebuilt(t *testing.T) {
	foreign := func(t *testing.T, path string) {
		db, err := sql.Open("sqlite", path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec("CREATE TABLE other(x int)"); err != nil {
			t.Fatal(err)
		}
		_ = db.Close()
	}
	cases := map[string]func(t *testing.T, path string){
		"垃圾内容": func(t *testing.T, path string) {
			_ = os.WriteFile(path, bytes.Repeat([]byte("not sqlite "), 500), 0o600)
		},
		"空文件":         func(t *testing.T, path string) { _ = os.WriteFile(path, nil, 0o600) },
		"其他 SQLite 库": foreign,
	}
	for name, prepare := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, stateFile)
			prepare(t, path)
			before, _ := os.ReadFile(path)
			_, err := Open(context.Background(), dir)
			if err == nil {
				t.Fatal("损坏或陌生的状态库不应打开成功")
			}
			if errors.Is(err, ErrDataDirLocked) {
				t.Fatalf("失败原因不应是锁: %v", err)
			}
			after, _ := os.ReadFile(path)
			if !bytes.Equal(before, after) {
				t.Fatal("状态库文件被改写（不允许重建为空库）")
			}
			// 打开失败必须释放 core.lock，否则管理员无法修复后重试。
			if _, err2 := Open(context.Background(), dir); errors.Is(err2, ErrDataDirLocked) {
				t.Fatal("打开失败后 core.lock 没有释放")
			}
		})
	}
}

// schema 版本高于本程序支持的版本时拒绝打开。
func TestNewerSchemaRejected(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec("UPDATE schema_version SET version = 99"); err != nil {
		t.Fatal(err)
	}
	_ = s.Close()
	if _, err := Open(context.Background(), dir); err == nil || !strings.Contains(err.Error(), "高于") {
		t.Fatalf("更高的 schema 版本应被拒绝，实际 %v", err)
	}
}

// 同一插件的管理锁互斥，不同插件互不阻塞。
func TestLockPluginSerializesSamePlugin(t *testing.T) {
	s := mustOpen(t, t.TempDir())
	var inside, maxInside atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unlock := s.LockPlugin("demo")
			defer unlock()
			n := inside.Add(1)
			for {
				m := maxInside.Load()
				if n <= m || maxInside.CompareAndSwap(m, n) {
					break
				}
			}
			inside.Add(-1)
		}()
	}
	wg.Wait()
	if maxInside.Load() != 1 {
		t.Fatalf("同一插件的管理操作并发度 = %d，应为 1", maxInside.Load())
	}
	// 不同插件互不阻塞：持有 a 的同时可以取得 b。
	ua := s.LockPlugin("a")
	ub := s.LockPlugin("b")
	ub()
	ua()
}

// 回归（REVIEW-4）：相对路径的 data_dir 必须能反复打开，且读到之前写入的数据。
func TestRelativeDataDirReopens(t *testing.T) {
	ctx := context.Background()
	t.Chdir(t.TempDir())
	for i := 0; i < 2; i++ {
		s, err := Open(ctx, "reldata")
		if err != nil {
			t.Fatalf("第 %d 次以相对路径打开失败: %v", i+1, err)
		}
		if i == 0 {
			if err := s.AddInstall(ctx, Install{PluginID: "demo", Source: "local", Entrypoint: "x", BuildDigest: []byte{1}}); err != nil {
				t.Fatal(err)
			}
		} else if _, err := s.GetInstall(ctx, "demo"); err != nil {
			t.Fatalf("重新打开后读不到之前的记录: %v", err)
		}
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
