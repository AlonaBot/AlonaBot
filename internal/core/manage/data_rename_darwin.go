package manage

import (
	"os"

	"golang.org/x/sys/unix"
)

func dataRenameAt(from *os.File, fromName string, to *os.File, toName string) error {
	return unix.RenameatxNp(int(from.Fd()), fromName, int(to.Fd()), toName, unix.RENAME_EXCL)
}
