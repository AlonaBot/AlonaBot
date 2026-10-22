package sdk

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlonaBot/AlonaBot/internal/oslock"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

func TestRuntimeLockSecondInstanceFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), wire.RuntimeLockFile)
	t.Setenv(wire.EnvRuntimeLock, path)
	first, err := acquireRuntimeLock()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := acquireRuntimeLock(); !errors.Is(err, oslock.ErrLocked) {
		t.Fatalf("runtime.lock 被占用时第二个实例应拒绝启动，实际 %v", err)
	}
	_ = first.Unlock()
	if _, err := os.Stat(path); err != nil {
		t.Fatal("锁文件必须保留")
	}
	again, err := acquireRuntimeLock()
	if err != nil {
		t.Fatalf("前一个实例释放后应能取锁: %v", err)
	}
	_ = again.Unlock()
}

func TestRuntimeLockRequiresAbsolutePathFromCore(t *testing.T) {
	t.Setenv(wire.EnvRuntimeLock, "")
	if _, err := acquireRuntimeLock(); err == nil {
		t.Fatal("缺少环境变量时应失败")
	}
	t.Setenv(wire.EnvRuntimeLock, "relative/runtime.lock")
	if _, err := acquireRuntimeLock(); err == nil {
		t.Fatal("相对路径应被拒绝")
	}
}

// 旧实例仍持有 runtime.lock 时，Run 在调用任何用户代码（包括注册入口 Commands）之前就失败。
func TestRunTakesLockBeforeRegistration(t *testing.T) {
	path := filepath.Join(t.TempDir(), wire.RuntimeLockFile)
	t.Setenv(wire.EnvRuntimeLock, path)
	held, err := oslock.TryLock(path)
	if err != nil {
		t.Fatal(err)
	}
	defer held.Unlock()
	called := false
	if err := Run(&root{commands: func(*CommandRegistry) { called = true }}); !errors.Is(err, oslock.ErrLocked) {
		t.Fatalf("锁被占用时应返回 ErrLocked，实际 %v", err)
	}
	if called {
		t.Fatal("取锁失败后仍调用了用户的 Commands 注册入口")
	}
}
