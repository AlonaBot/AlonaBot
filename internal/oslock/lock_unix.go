//go:build unix

package oslock

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

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

func unlock(f *os.File) error {
	return unix.Flock(int(f.Fd()), unix.LOCK_UN)
}
