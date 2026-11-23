package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const snapshotFile = "config.json"

// SyncSnapshot 导出当前 Active，仅覆盖仍对应旧 Active 或本次导入内容的编辑快照。
// 缺失文件直接补建；用户在准备期间修改的内容（包括未完成的 JSON）保持原样。
func (s *Service) SyncSnapshot(ctx context.Context, pluginID, previous string, imported []byte) error {
	unlock := s.store.LockPlugin(pluginID)
	defer unlock()
	active, err := s.active(ctx, pluginID)
	if err != nil {
		return err
	}
	content, err := s.store.GenerationConfig(ctx, active)
	if err != nil {
		return err
	}
	path := filepath.Join(s.dataDir, "plugins", pluginID, snapshotFile)
	raw, err := os.ReadFile(path)
	missing := errors.Is(err, os.ErrNotExist)
	if err != nil && !missing {
		return fmt.Errorf("read config snapshot: %w", err)
	}
	if !missing {
		if len(imported) != 0 {
			imported, err = Canonical(imported)
			if err != nil {
				return err
			}
		}
		current, err := Canonical(raw)
		if err != nil {
			return nil // 编辑中的无效 JSON 留给下一次显式导入报告。
		}
		matches := bytes.Equal(current, imported)
		if !matches && previous != "" {
			old, err := s.store.GenerationConfig(ctx, previous)
			if err != nil {
				return err
			}
			matches = bytes.Equal(current, []byte(old))
		}
		if !matches {
			return nil
		}
	}
	return writeSnapshot(path, []byte(content), raw, missing)
}

// writeSnapshot 在发布前再次核对编辑内容，缺失文件用 Link 发布以保留并发创建的文件。
func writeSnapshot(path string, content, expected []byte, missing bool) (err error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".config-*")
	if err != nil {
		return err
	}
	defer func() {
		if removeErr := os.Remove(temporary.Name()); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			err = errors.Join(err, removeErr)
		}
	}()
	_, writeErr := temporary.Write(content)
	if writeErr == nil {
		writeErr = temporary.Sync()
	}
	if err := errors.Join(writeErr, temporary.Close()); err != nil {
		return err
	}
	if missing {
		err := os.Link(temporary.Name(), path)
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		if err != nil {
			return err
		}
		return syncDirectory(filepath.Dir(path))
	}
	current, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil // 用户删除文件也是一次尚未应用的编辑。
	}
	if err != nil {
		return err
	}
	if !bytes.Equal(current, expected) {
		return nil
	}
	if err := os.Rename(temporary.Name(), path); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(path))
}
