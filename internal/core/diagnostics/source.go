// Package diagnostics 收集 Core 日志、设置、Active 配置和组件环境，写出诊断 zip。
//
// [Running] 读取 Core 已持有的状态，[Directory] 只读打开数据目录；两者通过 [Write]
// 共用分组、时间筛选与脱敏规则。归档只包含明确选择的文件。
package diagnostics

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/libtnb/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// Source 提供运行中 Core 或只读数据目录的诊断输入。
type Source interface {
	load(context.Context, Options) (snapshot, error)
}

// Directory 从 DataDir 只读收集诊断输入，Version 为当前 CLI 的构建版本。
// 收集期间独占已有的 core.lock；锁被占用或尚未创建时返回错误。
type Directory struct {
	DataDir string
	Version string
}

// Running 复用 Core 的状态库和运行登记；Schema 返回组件或 Adapter 实例的配置声明。
type Running struct {
	Directory
	Store  *state.Store
	State  func(string) string
	Schema func(ctx context.Context, id string, instance bool) ([]byte, error)
}

type component struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Version   string `json:"version"`
	Source    string `json:"source"`
	Signed    bool   `json:"signed"`
	Enabled   bool   `json:"enabled"`
	State     string `json:"state"`
	LastError string `json:"last_error"`

	install state.Install
	active  string
	config  []byte
	paths   []string
}

type snapshot struct {
	Directory
	components []component
	schema     func(context.Context, component, bool) ([]byte, error)
	secret     func(context.Context, string, string) (state.Secret, error)
	close      func() error
}

func (s Running) load(ctx context.Context, opts Options) (snapshot, error) {
	result := snapshot{Directory: s.Directory, components: []component{}, secret: s.Store.GetSecret, close: func() error { return nil }}
	result.schema = func(ctx context.Context, c component, instance bool) ([]byte, error) {
		return s.Schema(ctx, c.ID, instance)
	}
	if !opts.Settings && !opts.Components {
		return result, nil
	}
	installs, err := s.Store.Installs(ctx)
	if err != nil {
		return result, err
	}
	for _, install := range installs {
		if install.Uninstalled {
			continue
		}
		c := describe(install)
		c.Enabled, err = s.Store.Enabled(ctx, c.ID)
		if err != nil {
			return result, err
		}
		c.active, err = s.Store.Active(ctx, c.ID)
		if err != nil {
			return result, err
		}
		if current := s.State(c.ID); current != "" {
			c.State = current
		} else if c.Enabled && c.Role == "webui" {
			c.State = "Activated"
		} else if c.active == "" {
			c.State = "Installed"
		} else if c.Enabled {
			c.State = "Degraded"
		}
		if opts.Settings && c.active != "" && install.Role != state.RoleWebUI {
			raw, err := s.Store.GenerationConfig(ctx, c.active)
			if err != nil {
				return result, err
			}
			c.config = []byte(raw)
			c.paths, err = s.Store.GenerationSecretPaths(ctx, c.active)
			if err != nil {
				return result, err
			}
		}
		result.components = append(result.components, c)
	}
	return result, nil
}

func describe(install state.Install) component {
	return component{ID: install.PluginID, Role: string(install.Role), Version: install.Version,
		Source: install.Source, Signed: install.PublisherKeyID != "", State: "Stopped", install: install}
}

func (s Directory) load(ctx context.Context, opts Options) (result snapshot, err error) {
	result = snapshot{Directory: s, components: []component{}, close: func() error { return nil }}
	result.schema = func(_ context.Context, c component, instance bool) ([]byte, error) {
		if instance && c.Source == "dev" {
			return nil, nil // 开发声明由运行 Core 持有；离线由 maskAdapter 完整遮盖实例值。
		}
		name := "config_schema.json"
		if instance {
			name = "adapter_config_schema.json"
		}
		digest := hex.EncodeToString(c.install.BuildDigest)
		const digestPrefix = 12 // 与不可变程序目录的命名一致。
		if len(digest) < digestPrefix {
			return nil, fmt.Errorf("invalid package digest for %s", c.ID)
		}
		path := filepath.Join(s.DataDir, "packages", c.ID, c.Version+"-"+digest[:digestPrefix], name)
		raw, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil // 没有配置声明的组件仍可能有 Active 配置。
		}
		return raw, err
	}
	// 1. 对只读句柄取得与 Core 相同的 OS 锁，确认并持续保持停机状态。
	lock, err := os.Open(filepath.Join(s.DataDir, "core.lock"))
	if err != nil {
		return result, fmt.Errorf("open existing Core instance lock: %w", err)
	}
	if err := lockDirectory(lock); err != nil {
		return result, errors.Join(fmt.Errorf("stop Core before offline diagnosis: %w", err), lock.Close())
	}
	result.close = lock.Close
	defer func() {
		if err != nil {
			err = errors.Join(err, result.close())
		}
	}()
	if !opts.Settings && !opts.Components {
		return result, nil
	}
	name, err := filepath.Abs(filepath.Join(s.DataDir, "state.sqlite"))
	if err != nil {
		return result, err
	}
	if _, err := os.Stat(name); errors.Is(err, os.ErrNotExist) {
		return result, nil // Core 首次启动失败时可能尚未创建状态库。
	} else if err != nil {
		return result, err
	}
	// 2. SQLite 的 mode=ro 仍会创建 WAL/SHM；在临时目录读取主库和 WAL。
	temporary, err := os.MkdirTemp("", "alonabot-diagnostics-state-*")
	if err != nil {
		return result, err
	}
	result.close = func() error { return errors.Join(os.RemoveAll(temporary), lock.Close()) }
	for _, suffix := range []string{"", "-wal"} {
		source, err := os.Open(name + suffix)
		if suffix == "-wal" && errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return result, err
		}
		if err := copyState(ctx, source, filepath.Join(temporary, "state.sqlite"+suffix)); err != nil {
			return result, err
		}
	}
	path := filepath.ToSlash(filepath.Join(temporary, "state.sqlite"))
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	uri := (&url.URL{Scheme: "file", Path: path, RawQuery: "mode=ro"}).String()
	db, err := gorm.Open(sqlite.Open(uri), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return result, fmt.Errorf("open diagnostic state: %w", err)
	}
	connection, err := db.DB()
	if err != nil {
		return result, err
	}
	result.close = func() error { return errors.Join(connection.Close(), os.RemoveAll(temporary), lock.Close()) }
	result.secret = func(ctx context.Context, id, ref string) (state.Secret, error) {
		var secret state.Secret
		err := db.WithContext(ctx).Where(clause.Eq{Column: "plugin_id", Value: id}).
			Where(clause.Eq{Column: "ref_id", Value: ref}).Take(&secret).Error
		return secret, err
	}
	var installs []state.Install
	if err := db.WithContext(ctx).Table("install").Order("install_seq").Find(&installs).Error; err != nil {
		return result, err
	}
	for _, install := range installs {
		if install.Uninstalled {
			continue
		}
		c := describe(install)
		var current struct {
			ActiveGeneration sql.NullString
			Enabled          bool
		}
		if err := db.WithContext(ctx).Table("plugin_state").Where(clause.Eq{Column: "plugin_id", Value: c.ID}).Take(&current).Error; err != nil {
			return result, err
		}
		c.Enabled, c.active = current.Enabled, current.ActiveGeneration.String
		if opts.Settings && c.active != "" && install.Role != state.RoleWebUI {
			var generation struct{ ConfigSnapshot string }
			if err := db.WithContext(ctx).Table("generation").Where(clause.Eq{Column: "generation_id", Value: c.active}).Take(&generation).Error; err != nil {
				return result, err
			}
			c.config = []byte(generation.ConfigSnapshot)
			if err := db.WithContext(ctx).Table("generation_secret_path").Where(clause.Eq{Column: "generation_id", Value: c.active}).Pluck("path", &c.paths).Error; err != nil {
				return result, err
			}
		}
		result.components = append(result.components, c)
	}
	return result, nil
}

func copyState(ctx context.Context, source *os.File, name string) (err error) {
	defer func() { err = errors.Join(err, source.Close()) }()
	target, err := os.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, target.Close()) }()
	const copyBufferBytes = 32 << 10
	buffer := make([]byte, copyBufferBytes)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, err := source.Read(buffer)
		if n > 0 {
			if _, err := target.Write(buffer[:n]); err != nil {
				return err
			}
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func decodeJSON(raw []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber() // 设置中的大整数和配置数字保持原文。
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, errors.New("expected one JSON value")
		}
		return nil, err
	}
	return value, nil
}
