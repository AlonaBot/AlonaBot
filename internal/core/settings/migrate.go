package settings

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// BackupDir 指定迁移写回前保存 core_config.json 原文件的目录（CONTRACTS §5.5）。
const BackupDir = "backups"

const backupTimeFormat = "20060102T150405.000Z"

type coreMigration struct {
	to    int
	apply func(*Document) error
}

// Core 设置的迁移按目标版本递增排列；当前格式尚无迁移步骤。
var coreMigrations = []coreMigration{}

func positiveVersion(n *node) (int, error) {
	version, err := n.integer()
	if err != nil {
		return 0, err
	}
	if version <= 0 {
		return 0, n.fail(errors.New("version must be a positive integer"))
	}
	return version, nil
}

func readVersion(n *node) error {
	version, err := positiveVersion(n)
	if err != nil {
		return err
	}
	if version > currentVersion {
		return n.fail(fmt.Errorf("settings version %d is newer than supported version %d", version, currentVersion))
	}
	if version < currentVersion {
		return n.fail(fmt.Errorf("settings version %d requires migration to version %d", version, currentVersion))
	}
	return nil
}

func migrateSettings(name string, document *Document) error {
	index := document.root.index("$version")
	if index < 0 {
		return nil // 省略版本时按当前版本读取，留到下一次写回时补上。
	}
	n := document.root.members[index].value
	version, err := positiveVersion(n)
	if err != nil {
		return err
	}
	if version >= currentVersion {
		return readVersion(n)
	}
	oldVersion := version
	// 1. 先在内存完成全部迁移和校验，失败时保留磁盘原文件。
	for _, migration := range coreMigrations {
		if migration.to <= version {
			continue
		}
		if err := migration.apply(document); err != nil {
			return fmt.Errorf("migrate settings from version %d to %d: %w", version, migration.to, err)
		}
		version = migration.to
	}
	if version != currentVersion {
		return n.fail(fmt.Errorf("no migration from settings version %d to supported version %d", version, currentVersion))
	}
	if err := document.Set([]string{"$version"}, currentVersion); err != nil {
		return err
	}
	if _, err := document.settings(); err != nil {
		return err
	}
	content, err := document.formatted()
	if err != nil {
		return err
	}
	// 2. 原文件备份持久化后才允许替换，沿用设置文件的提交方式。
	if err := backupSettings(name, oldVersion); err != nil {
		return err
	}
	if err := replaceFile(name, content); err != nil {
		return err
	}
	if err := syncDirectory(filepath.Dir(name)); err != nil {
		return fmt.Errorf("settings saved but directory sync failed: %w", err)
	}
	return nil
}

func backupSettings(name string, version int) error {
	// 1. 从磁盘复制原文，权限只允许运行账号读取配置中的秘密。
	source, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("open settings backup source: %w", err)
	}
	dir := filepath.Join(filepath.Dir(name), BackupDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return errors.Join(fmt.Errorf("create settings backup directory: %w", err), source.Close())
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return errors.Join(fmt.Errorf("restrict settings backup directory: %w", err), source.Close())
	}
	stamp := time.Now().UTC()
	var backupName string
	var backup *os.File
	for {
		backupName = filepath.Join(dir, fmt.Sprintf("core_config.%d.%s.json", version, stamp.Format(backupTimeFormat)))
		backup, err = os.OpenFile(backupName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if errors.Is(err, fs.ErrExist) {
			stamp = stamp.Add(time.Millisecond)
			continue
		}
		if err != nil {
			return errors.Join(fmt.Errorf("create settings backup: %w", err), source.Close())
		}
		break
	}
	_, copyErr := io.Copy(backup, source)
	if copyErr == nil {
		copyErr = backup.Sync()
	}
	if err := errors.Join(copyErr, source.Close(), backup.Close()); err != nil {
		return errors.Join(fmt.Errorf("copy settings backup: %w", err), os.Remove(backupName))
	}
	// 2. 确认备份目录项落盘，调用方随后才写回设置文件。
	if err := syncDirectory(dir); err != nil {
		return fmt.Errorf("sync settings backup directory: %w", err)
	}
	return nil
}
