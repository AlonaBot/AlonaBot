package journal

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// 固定目录布局：packages/<id>/<version-digest>；plugins/<id>/generations/<generation>。
const (
	packageParentLevels     = 3
	databaseDirectoryLevels = 5
)

// persistPackage 同步包内文件及目录，从叶子向上持久化 MoveTo 发布的路径。
func persistPackage(dir string) error {
	var directories []string
	if err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			directories = append(directories, path)
			return nil
		}
		return syncFile(path)
	}); err != nil {
		return err
	}
	slices.Reverse(directories)
	parent := dir
	for range packageParentLevels {
		parent = filepath.Dir(parent)
		directories = append(directories, parent)
	}
	for _, path := range directories {
		if err := syncDirectory(path); err != nil {
			return err
		}
	}
	return nil
}

// persistDatabase 同步准备中仍打开的 WAL 数据库；SDK 的每次提交已使用 FULL。
// 首装文件由 SDK 按需创建，尚未打开时只持久化目录。
func persistDatabase(path string, firstInstall bool) error {
	files := []string{path, path + "-wal"}
	if firstInstall {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			files = nil
		} else if err != nil {
			return err
		}
	}
	for _, name := range files {
		err := syncFile(name)
		if name != path && errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
	}
	// 新代次和首次插件目录的父目录也须落盘，避免只留下数据库内容而丢失路径。
	directory := filepath.Dir(path)
	for range databaseDirectoryLevels {
		if err := syncDirectory(directory); err != nil {
			return err
		}
		directory = filepath.Dir(directory)
	}
	return nil
}

func syncFile(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	return errors.Join(f.Sync(), f.Close())
}
