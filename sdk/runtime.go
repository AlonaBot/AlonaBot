package sdk

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlonaBot/alonabot/internal/oslock"
	"github.com/AlonaBot/alonabot/internal/wire"
)

// acquireRuntimeLock 在任何用户回调或受管 DB 打开之前取得 <plugins>/<id>/runtime.lock（CONTRACTS §12.2），
// 由 Run 持有至进程退出。取不到锁说明旧写入者（或 Core 的停机复制）仍在，必须拒绝启动，
// 不能删除锁文件或等待重试。
func acquireRuntimeLock() (*oslock.Lock, error) {
	path := os.Getenv(wire.EnvRuntimeLock)
	switch {
	case path == "":
		return nil, fmt.Errorf("缺少环境变量 %s（插件必须由 Core 启动）", wire.EnvRuntimeLock)
	case !filepath.IsAbs(path):
		return nil, fmt.Errorf("%s 必须是绝对路径，实际 %q", wire.EnvRuntimeLock, path)
	}
	l, err := oslock.TryLock(path)
	if errors.Is(err, oslock.ErrLocked) {
		return nil, fmt.Errorf("另一个实例或 Core 仍持有 runtime.lock，拒绝启动: %w", err)
	}
	return l, err
}
