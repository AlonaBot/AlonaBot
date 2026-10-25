// 维护者：FlanChanXwO
// 创建日期：2026-10-21
// 职责：锁句柄类型，以及 TryLock、Lock、Unlock 的跨平台入口。

// Package oslock 提供跨进程的 OS 独占文件锁（CONTRACTS §12.2）：Linux/macOS 用 flock，
// Windows 用 LockFileEx。Core 的 core.lock 与 SDK 的 runtime.lock 共用本实现。
//
// 锁文件在解锁后保留，不删除或替换，防止并发进程锁住不同的文件；Go 打开的文件默认
// close-on-exec，锁句柄不会传给子进程。锁属于打开的文件句柄，同一进程内再次 TryLock 同一路径也会失败。
package oslock

import (
	"errors"
	"fmt"
	"os"
)

// ErrLocked 表示锁已被其他持有者占用。
var ErrLocked = errors.New("锁已被占用")

// Lock 代表一把已持有的独占锁。
type Lock struct {
	f *os.File // 持有锁的文件句柄；Unlock 后为 nil
}

// TryLock 非阻塞地获取 path 的独占锁；文件不存在时以 0600 创建。已被占用时返回包装了 ErrLocked 的错误。
func TryLock(path string) (*Lock, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, fmt.Errorf("打开锁文件 %s 失败: %w", path, err)
	}
	if err := tryLock(f); err != nil {
		_ = f.Close()
		if errors.Is(err, ErrLocked) {
			return nil, fmt.Errorf("%s: %w", path, ErrLocked)
		}
		return nil, fmt.Errorf("锁定 %s 失败: %w", path, err)
	}
	return &Lock{f: f}, nil
}

// Unlock 释放锁并关闭句柄；锁文件保留。重复调用无副作用。
func (l *Lock) Unlock() error {
	if l.f == nil {
		return nil
	}
	err := errors.Join(unlock(l.f), l.f.Close())
	l.f = nil
	return err
}
