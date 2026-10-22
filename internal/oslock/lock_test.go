// 本文件测试独占文件锁：同进程与跨进程的互斥、重复解锁、锁文件保留，以及持有进程退出后锁被释放。
package oslock

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestMain 在设置 OSLOCK_HOLD 时作为子进程运行：持有锁，输出一行 "locked"，直到 stdin 关闭后退出（父进程不用 sleep 判断状态）。
func TestMain(m *testing.M) {
	if path := os.Getenv("OSLOCK_HOLD"); path != "" {
		l, err := TryLock(path)
		if err != nil {
			os.Stdout.WriteString("error " + err.Error() + "\n")
			os.Exit(2)
		}
		os.Stdout.WriteString("locked\n")
		_, _ = io.Copy(io.Discard, os.Stdin)
		_ = l.Unlock()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// 同一进程再次加锁失败；重复 Unlock 无副作用；解锁后锁文件保留且可以再次加锁。
func TestSameProcessSecondLockFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.lock")
	l, err := TryLock(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := TryLock(path); !errors.Is(err, ErrLocked) {
		t.Fatalf("第二个句柄应得到 ErrLocked，实际 %v", err)
	}
	if err := l.Unlock(); err != nil {
		t.Fatal(err)
	}
	if err := l.Unlock(); err != nil {
		t.Fatalf("重复 Unlock 应无副作用: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("解锁后锁文件必须保留: %v", err)
	}
	l2, err := TryLock(path)
	if err != nil {
		t.Fatalf("释放后应能再次加锁: %v", err)
	}
	_ = l2.Unlock()
}

// 另一进程持锁时加锁失败；持有进程确认退出后可以加锁。
func TestCrossProcessLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.lock")
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = append(os.Environ(), "OSLOCK_HOLD="+path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	line, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil || line != "locked\n" {
		t.Fatalf("子进程未能持锁: %q %v", line, err)
	}
	if _, err := TryLock(path); !errors.Is(err, ErrLocked) {
		t.Fatalf("另一进程持锁时应得到 ErrLocked，实际 %v", err)
	}
	_ = stdin.Close()
	if err := cmd.Wait(); err != nil {
		t.Fatalf("子进程退出异常: %v", err)
	}
	// 持有者进程已确认退出（Wait 返回），锁由 OS 释放。
	l, err := TryLock(path)
	if err != nil {
		t.Fatalf("持有进程退出后应能加锁: %v", err)
	}
	_ = l.Unlock()
}
