// 本文件测试 runtime.lock：第二实例拒绝启动、锁文件保留、路径必须由 Core 以绝对路径给出，以及 Run 先取锁再执行用户代码。
package sdk

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlonaBot/AlonaBot/internal/oslock"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// 锁被占用时第二个实例失败；释放后锁文件仍在，可以再次取锁。
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

// 缺少环境变量或给出相对路径时拒绝取锁。
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
