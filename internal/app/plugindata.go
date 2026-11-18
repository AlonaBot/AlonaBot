package app

import (
	"fmt"
	"os"
	"path/filepath"
)

// preparePluginDataDir 在 state.Open 保护过的 data_dir 下建目录；Windows 继承其账号专用 ACL。
func preparePluginDataDir(dataDir, id string) (string, error) {
	path, err := filepath.Abs(filepath.Join(dataDir, "plugin_data", id))
	if err != nil {
		return "", fmt.Errorf("resolve plugin data directory: %w", err)
	}
	if err := os.MkdirAll(path, 0o700); err != nil {
		return "", fmt.Errorf("create plugin data directory %q: %w", path, err)
	}
	if err := os.Chmod(path, 0o700); err != nil {
		return "", fmt.Errorf("protect plugin data directory %q: %w", path, err)
	}
	return path, nil
}
