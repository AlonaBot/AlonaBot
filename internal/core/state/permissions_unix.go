//go:build unix

package state

import (
	"errors"
	"fmt"
	"os"
)

func protectStateDirectory(string) error { return nil }

func checkStatePermissions(path string) error {
	for _, file := range []string{path, path + "-wal", path + "-shm"} {
		info, err := os.Stat(file)
		if errors.Is(err, os.ErrNotExist) && file != path {
			continue
		}
		if err != nil {
			return fmt.Errorf("inspect state file %q: %w", file, err)
		}
		if info.Mode().Perm() != 0o600 {
			return fmt.Errorf("state file %q has mode %04o, want 0600", file, info.Mode().Perm())
		}
	}
	return nil
}
