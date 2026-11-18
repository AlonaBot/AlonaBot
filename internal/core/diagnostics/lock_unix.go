//go:build unix

package diagnostics

import (
	"errors"
	"os"

	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"golang.org/x/sys/unix"
)

// lockDirectory 对只读句柄采用与 oslock 相同的 flock；关闭句柄时释放。
func lockDirectory(f *os.File) error {
	for {
		err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
		switch {
		case errors.Is(err, unix.EINTR):
			continue
		case errors.Is(err, unix.EWOULDBLOCK):
			return state.ErrDataDirLocked
		default:
			return err
		}
	}
}
