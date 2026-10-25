//go:build unix

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
