package sdk

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/AlonaBot/AlonaBot/internal/oslock"
)

// acquireRuntimeLock 在任何用户回调或受管 DB 打开之前取得 <plugins>/<id>/runtime.lock（CONTRACTS §12.2），
// 由 Run 持有至进程退出。取不到锁说明旧写入者（或 Core 的停机复制）仍在，必须拒绝启动，
// 不能删除锁文件或等待重试。
func acquireRuntimeLock(path string) (*oslock.Lock, error) {
	switch {
	case path == "":
		return nil, errors.New("missing runtime lock path (the plugin must be started by Core)")
	case !filepath.IsAbs(path):
		return nil, fmt.Errorf("runtime lock path must be absolute, got %q", path)
	}
	l, err := oslock.TryLock(path)
	if errors.Is(err, oslock.ErrLocked) {
		return nil, fmt.Errorf("another instance or Core still holds runtime.lock, refusing to start: %w", err)
	}
	return l, err
}
