package sdk

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlonaBot/AlonaBot/internal/oslock"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// acquireRuntimeLock 在任何用户回调或受管 DB 打开之前取得 <plugins>/<id>/runtime.lock（CONTRACTS §12.2），
// 由 Run 持有至进程退出。取不到锁说明旧写入者（或 Core 的停机复制）仍在，必须拒绝启动，
// 不能删除锁文件或等待重试。
func acquireRuntimeLock() (*oslock.Lock, error) {
	path := os.Getenv(wire.EnvRuntimeLock)
	switch {
	case path == "":
		return nil, fmt.Errorf("missing environment variable %s (the plugin must be started by Core)", wire.EnvRuntimeLock)
	case !filepath.IsAbs(path):
		return nil, fmt.Errorf("%s must be an absolute path, got %q", wire.EnvRuntimeLock, path)
	}
	l, err := oslock.TryLock(path)
	if errors.Is(err, oslock.ErrLocked) {
		return nil, fmt.Errorf("another instance or Core still holds runtime.lock, refusing to start: %w", err)
	}
	return l, err
}
