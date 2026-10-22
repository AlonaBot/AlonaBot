//go:build windows

// 本文件用 LockFileEx 实现 Windows 上的独占文件锁。
package oslock

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

// tryLock 以非阻塞方式独占锁住文件起始的一个字节（与文件实际长度无关）；已被占用时返回 ErrLocked。
func tryLock(f *os.File) error {
	ol := new(windows.Overlapped)
	err := windows.LockFileEx(windows.Handle(f.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, ol)
	if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		return ErrLocked
	}
	return err
}

// unlock 释放 tryLock 锁住的字节。
func unlock(f *os.File) error {
	return windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, new(windows.Overlapped))
}
