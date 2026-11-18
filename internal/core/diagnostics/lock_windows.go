//go:build windows

package diagnostics

import (
	"errors"
	"os"

	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"golang.org/x/sys/windows"
)

// lockDirectory 与 oslock 一样独占首字节，只读句柄关闭时释放锁。
func lockDirectory(f *os.File) error {
	err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
		0, 1, 0, new(windows.Overlapped))
	if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		return state.ErrDataDirLocked
	}
	return err
}
