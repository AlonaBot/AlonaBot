package settings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"

	"github.com/AlonaBot/AlonaBot/internal/core/logging"
)

// WithPublisher 返回共享文档与快照的受管视图，Update/Reload 交由 publish 执行。
// publish 应同步完成 change，再在文件锁外应用已发布设置；原视图供所有者直接使用。
func (f *File) WithPublisher(publish func(change func() error) error) *File {
	return &File{fileState: f.fileState, publish: publish}
}

// Update 在未展开的文档副本上编辑，校验通过后以 0600 临时文件替换原文件。
// Rename 为提交点：此前失败保留文件与快照；此后目录同步失败仍发布新快照，
// 并返回设置已写入但持久化尚未确认的错误。edit 执行期间持有写锁，应只操作传入的 Document。
func (f *File) Update(edit func(doc *Document) error) error {
	if f.publish != nil {
		return f.publish(func() error { return f.update(edit) })
	}
	return f.update(edit)
}

func (f *File) update(edit func(doc *Document) error) error {
	// 1. 持锁读取磁盘原文，与 Reload 串行，保留运行期间的手动修改。
	f.mu.Lock()
	defer f.mu.Unlock()
	content, err := os.ReadFile(f.name)
	if err != nil {
		return fmt.Errorf("read %s: %w", f.name, err)
	}
	draft, err := parseDocument(content)
	if err != nil {
		return fmt.Errorf("%s: %w", f.name, err)
	}
	if _, err := draft.settings(); err != nil {
		return fmt.Errorf("%s: %w", f.name, err)
	}
	if err := edit(draft); err != nil {
		return fmt.Errorf("edit %s: %w", f.name, err)
	}
	// 2. 从最终格式重新解析，诊断中的行列对应将要保存的内容。
	content, err = draft.formatted()
	if err != nil {
		return fmt.Errorf("format %s: %w", f.name, err)
	}
	document, err := parseDocument(content)
	if err != nil {
		return fmt.Errorf("%s: %w", f.name, err)
	}
	current, err := document.settings()
	if err != nil {
		return fmt.Errorf("%s: %w", f.name, err)
	}
	// 3. 重命名成功后才发布成对的文档和设置。
	if err := replaceFile(f.name, content); err != nil {
		return err
	}
	f.document, f.current = document, current
	logCorrections(current.Corrections)
	if err := syncDirectory(filepath.Dir(f.name)); err != nil {
		slog.Log(context.Background(), logging.LevelCritical, "settings directory sync failed", "file", f.name, "error", err)
		return fmt.Errorf("settings saved but directory sync failed: %w", err)
	}
	return nil
}

// Reload 重新读取并校验文件，返回有效值发生变化的顶层键，按字典序排列。
// 失败时保留当前文档与快照；只改变格式、键顺序或显式写出缺省值时返回空列表。
func (f *File) Reload() (changed []string, err error) {
	if f.publish != nil {
		err = f.publish(func() error {
			var reloadErr error
			changed, reloadErr = f.reload()
			return reloadErr
		})
		return changed, err
	}
	return f.reload()
}

func (f *File) reload() (changed []string, err error) {
	f.mu.Lock() // 文件读取与发布位于同一个锁区间，避免覆盖同时完成的 Update。
	defer f.mu.Unlock()
	content, err := os.ReadFile(f.name)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", f.name, err)
	}
	document, err := parseDocument(content)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", f.name, err)
	}
	current, err := document.settings()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", f.name, err)
	}
	before, err := settingsValues(f.current)
	if err != nil {
		return nil, fmt.Errorf("compare current settings: %w", err)
	}
	after, err := settingsValues(current)
	if err != nil {
		return nil, fmt.Errorf("compare reloaded settings: %w", err)
	}
	for key, value := range after {
		if !reflect.DeepEqual(before[key], value) {
			changed = append(changed, key)
		}
	}
	slices.Sort(changed)
	f.document, f.current = document, current
	logCorrections(current.Corrections)
	return changed, nil
}

func replaceFile(name string, content []byte) (err error) {
	defer func() {
		if err != nil {
			slog.Log(context.Background(), logging.LevelCritical, "settings write failed", "file", name, "error", err)
		}
	}()
	temporary, err := os.CreateTemp(filepath.Dir(name), "."+FileName+"-*") // CreateTemp 使用 0600。
	if err != nil {
		return fmt.Errorf("create temporary settings: %w", err)
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, os.Remove(temporary.Name()))
		}
	}()
	_, writeErr := temporary.Write(content)
	if writeErr == nil {
		writeErr = temporary.Sync() // 内容落盘后才允许重命名。
	}
	closeErr := temporary.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return fmt.Errorf("write temporary settings: %w", err)
	}
	if err := os.Rename(temporary.Name(), name); err != nil {
		return fmt.Errorf("replace %s: %w", name, err)
	}
	return nil
}

func settingsValues(s Settings) (map[string]any, error) {
	// 实例配置比较 JSON 对象的值，忽略成员顺序，并保留大整数的精度。
	adapters := make(map[string]any, len(s.Adapters))
	for id, adapter := range s.Adapters {
		instances := make(map[string]any, len(adapter.Instances))
		for instance, raw := range adapter.Instances {
			var value any
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			if err := decoder.Decode(&value); err != nil {
				return nil, err
			}
			instances[instance] = value
		}
		adapters[id] = map[string]any{"package": adapter.Package, "instances": instances, "default_instance": adapter.DefaultInstance}
	}
	return map[string]any{
		"listen": s.Listen, "base_url": s.BaseURL,
		"trusted_proxies": s.TrustedProxies, "proxy": s.Proxy, "admins": s.Admins,
		"github_api_base": s.GitHubAPIBase, "github_download_prefix": s.GitHubDownloadPrefix,
		"command_prefixes": s.CommandPrefixes,
		"reply":            s.Reply,
		"sleep":            s.Sleep,
		"default_locale":   s.DefaultLocale, "adapters": adapters,
		"log": map[string]any{"level": s.LogLevel, "format": s.LogFormat, "file": s.LogFile, "sources": s.LogSources},
	}, nil
}
