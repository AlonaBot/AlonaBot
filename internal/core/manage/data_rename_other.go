//go:build !linux && !darwin && !windows

package manage

import (
	"fmt"
	"os"
	"runtime"
)

func dataRenameAt(from *os.File, fromName string, to *os.File, toName string) error {
	return fmt.Errorf("atomic no-replace rename is not supported on %s", runtime.GOOS)
}
