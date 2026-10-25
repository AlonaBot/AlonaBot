//go:build unix

// 维护者：FlanChanXwO
// 创建日期：2026-10-21
// 职责：用 flock 实现 Linux/macOS 上的独占文件锁。

package oslock

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

// tryLock 以非阻塞方式对 f 加独占锁；被信号打断时重试，已被占用时返回 ErrLocked。
func tryLock(f *os.File) error {
	for {
		err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, unix.EINTR):
			continue
		case errors.Is(err, unix.EWOULDBLOCK):
			return ErrLocked
		default:
			return err
		}
	}
}

// unlock 释放 f 上的锁。
func unlock(f *os.File) error {
	return unix.Flock(int(f.Fd()), unix.LOCK_UN)
}
