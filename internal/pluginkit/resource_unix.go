//go:build unix

package pluginkit

import (
	"fmt"
	"os"
	"syscall"
)

func checkHardlink(path string, info os.FileInfo) error {
	if info.Mode().IsRegular() && info.Sys().(*syscall.Stat_t).Nlink > 1 {
		return fmt.Errorf("resource %q is a hard link", path)
	}
	return nil
}
