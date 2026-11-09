//go:build unix

package settings

import (
	"errors"
	"os"
)

// syncDirectory 在重命名后同步目录，确认文件名替换已经持久化。
func syncDirectory(name string) error {
	directory, err := os.Open(name)
	if err != nil {
		return err
	}
	return errors.Join(directory.Sync(), directory.Close())
}
