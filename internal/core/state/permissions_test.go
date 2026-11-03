//go:build unix

package state

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestStateFilePermissions 在独立进程的 umask 022 下核对真实 SQLite 及其 WAL/SHM 权限。
func TestStateFilePermissions(t *testing.T) {
	if os.Getenv("STATE_PERMISSION_PROBE") != "1" {
		// 仅改变测试子进程的 umask，不影响其他测试或调用者的文件创建权限。
		command := exec.Command("sh", "-c", `umask 022; exec "$@"`, "permissions", os.Args[0], "-test.run=^TestStateFilePermissions$", "-test.v")
		command.Env = append(os.Environ(), "STATE_PERMISSION_PROBE=1")
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("permission subprocess: %v\n%s", err, output)
		}
		return
	}
	dir := t.TempDir()
	store, err := Open(t.Context(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.AddInstall(t.Context(), Install{Role: RolePlugin, PluginID: "permissions", Source: "local", Entrypoint: "plugin", BuildDigest: []byte{1}}); err != nil {
		store.Close()
		t.Fatal(err)
	}
	for _, name := range []string{stateFile, stateFile + "-wal", stateFile + "-shm"} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			t.Error(err)
			continue
		}
		if info.Mode().Perm() != 0o600 {
			t.Errorf("%s mode=%04o, want 0600", name, info.Mode().Perm())
		}
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, stateFile)
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	rejected, err := Open(t.Context(), dir)
	if err == nil {
		rejected.Close()
		t.Fatal("unsafe state permissions were accepted")
	}
	if !strings.Contains(err.Error(), path) {
		t.Fatalf("permission error omitted path: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o644 {
		t.Fatalf("unsafe permissions were silently changed: %v", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	// 纠正测试文件后可再次取得 core.lock，证明失败路径释放了锁和数据库连接。
	reopened, err := Open(t.Context(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := reopened.Close(); err != nil {
		t.Fatal(err)
	}
}
