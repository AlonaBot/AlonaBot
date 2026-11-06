//go:build unix

package candidate

import (
	"errors"
	"os"
	"path/filepath"
)

// publishDatabase 在文件内容落盘后发布文件名，再同步目录以持久化重命名。
func publishDatabase(from, to string) (err error) {
	if err := os.Rename(from, to); err != nil {
		return err
	}
	directory, err := os.Open(filepath.Dir(to))
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := directory.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()
	return directory.Sync()
}
