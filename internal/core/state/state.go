// Package state 管理 Core 元数据库 <data_dir>/state.sqlite 与 <data_dir>/core.lock
// （CONTRACTS §6.1、§12.2）。T-02 只交付共同原语：实例锁、状态库、最小表与提交事务；
// 候选资源引用、完整 Journal 与恢复编排由 T-08/T-09 在同一实现上扩展。
package state

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/AlonaBot/alonabot/internal/oslock"
	_ "modernc.org/sqlite" // 注册 "sqlite" 驱动
)

const (
	lockFile  = "core.lock"
	stateFile = "state.sqlite"
	// 连接级设置，驱动对每个物理连接执行；单连接，见 CONTRACTS §6.1。
	pragmas = "_pragma=journal_mode(WAL)&_pragma=synchronous(FULL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
)

var (
	// ErrDataDirLocked 表示同一 data_dir 已有另一个 Core 在运行。
	ErrDataDirLocked = errors.New("data_dir 已被另一个 Core 锁定")
	// ErrStaleGeneration 表示提交时 ActiveGeneration 已不是调用方预期的旧值。
	ErrStaleGeneration = errors.New("ActiveGeneration 已变化，拒绝陈旧候选")
	// ErrNotFound 表示记录不存在。
	ErrNotFound = errors.New("记录不存在")
)

// Store 是持有 core.lock 的状态库；同一进程内只应有一个。
type Store struct {
	db   *sql.DB
	lock *oslock.Lock

	mu      sync.Mutex
	pluginM map[string]*sync.Mutex
}

// Open 先取得 core.lock，再打开（首次启动时创建）state.sqlite 并核对连接设置与 schema。
// 已存在但无法读取、不是 SQLite 或缺少 schema 的状态库一律报错，不会重建成空库。
func Open(ctx context.Context, dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, fmt.Errorf("创建 data_dir 失败: %w", err)
	}
	lock, err := oslock.TryLock(filepath.Join(dataDir, lockFile))
	if errors.Is(err, oslock.ErrLocked) {
		return nil, fmt.Errorf("%w（%s）", ErrDataDirLocked, dataDir)
	} else if err != nil {
		return nil, err
	}
	s, err := open(ctx, filepath.Join(dataDir, stateFile))
	if err != nil {
		_ = lock.Unlock()
		return nil, err
	}
	s.lock = lock
	return s, nil
}

func open(ctx context.Context, path string) (*Store, error) {
	if strings.ContainsAny(path, "?#") {
		return nil, fmt.Errorf("状态库路径 %q 含 ?/#", path)
	}
	_, statErr := os.Stat(path)
	fresh := errors.Is(statErr, os.ErrNotExist)
	if statErr != nil && !fresh {
		return nil, fmt.Errorf("检查状态库失败: %w", statErr)
	}

	if !fresh {
		// 先只读预检，确认是 AlonaBot 状态库后再以可写方式打开：可写 DSN 在连接时就会执行
		// journal_mode(WAL) 等设置并改写文件头，不能在确认之前碰已有文件。
		if err := preflight(ctx, path); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path+"?"+pragmas)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	s := &Store{db: db, pluginM: map[string]*sync.Mutex{}}
	if err := s.verify(ctx, fresh); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close 关闭状态库并释放 core.lock。
func (s *Store) Close() error {
	return errors.Join(s.db.Close(), s.lock.Unlock())
}

// Pragmas 是状态库连接上实际生效的设置。
type Pragmas struct {
	JournalMode string
	Synchronous int // FULL = 2
	ForeignKeys int
	BusyTimeout int
}

// ReadPragmas 回读实际生效的连接设置。
func (s *Store) ReadPragmas(ctx context.Context) (Pragmas, error) {
	var p Pragmas
	err := errors.Join(
		s.db.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&p.JournalMode),
		s.db.QueryRowContext(ctx, "PRAGMA synchronous").Scan(&p.Synchronous),
		s.db.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&p.ForeignKeys),
		s.db.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&p.BusyTimeout),
	)
	return p, err
}

func (s *Store) verify(ctx context.Context, fresh bool) error {
	p, err := s.ReadPragmas(ctx)
	if err != nil {
		return fmt.Errorf("状态库无法读取（可能已损坏，不会重建）: %w", err)
	}
	if want := (Pragmas{"wal", 2, 1, 5000}); p != want {
		return fmt.Errorf("状态库连接设置未生效：实际 %+v，期望 %+v", p, want)
	}
	return s.migrate(ctx, fresh)
}

// preflight 以只读、无 PRAGMA 的方式打开已有状态库，确认含 schema_version 表且完整性检查通过。
func preflight(ctx context.Context, path string) error {
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p // Windows 盘符路径需写成 file:///C:/...
	}
	uri := (&url.URL{Scheme: "file", Path: p, RawQuery: "mode=ro"}).String()
	db, err := sql.Open("sqlite", uri)
	if err != nil {
		return err
	}
	defer db.Close()
	var n int
	err = db.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='schema_version'").Scan(&n)
	if err != nil {
		return fmt.Errorf("状态库无法读取（可能已损坏，不会重建）: %w", err)
	}
	if n == 0 {
		return errors.New("state.sqlite 已存在但不含 AlonaBot schema（空文件或其他数据库），拒绝初始化为空库")
	}
	var check string
	if err := db.QueryRowContext(ctx, "PRAGMA quick_check").Scan(&check); err != nil || check != "ok" {
		return fmt.Errorf("状态库完整性检查失败（%q, %v），不会重建", check, err)
	}
	return nil
}

// migrations 按顺序执行，版本号即下标加 1；已发布的条目不得修改，只能追加。
var migrations = []string{
	`CREATE TABLE schema_version (version INTEGER NOT NULL);
	 INSERT INTO schema_version (version) VALUES (0);
	 CREATE TABLE installs (
	   plugin_id    TEXT PRIMARY KEY,
	   install_seq  INTEGER NOT NULL UNIQUE,
	   source       TEXT NOT NULL,
	   entrypoint   TEXT NOT NULL,
	   build_digest BLOB NOT NULL,
	   created_at   INTEGER NOT NULL
	 );
	 CREATE TABLE generations (
	   generation_id   TEXT PRIMARY KEY,
	   plugin_id       TEXT NOT NULL REFERENCES installs(plugin_id),
	   config_snapshot TEXT NOT NULL,
	   config_version  INTEGER NOT NULL,
	   created_at      INTEGER NOT NULL,
	   committed_at    INTEGER
	 );
	 CREATE TABLE plugin_state (
	   plugin_id         TEXT PRIMARY KEY REFERENCES installs(plugin_id),
	   active_generation TEXT REFERENCES generations(generation_id),
	   enabled           INTEGER NOT NULL DEFAULT 0
	 );`,
}

// migrate 只在全新文件上从版本 0 开始；已有文件的 schema_version 表已由 verify 确认存在。
func (s *Store) migrate(ctx context.Context, fresh bool) error {
	current := 0
	if !fresh {
		if err := s.db.QueryRowContext(ctx, "SELECT version FROM schema_version").Scan(&current); err != nil {
			return fmt.Errorf("读取 schema 版本失败: %w", err)
		}
	}
	if current > len(migrations) {
		return fmt.Errorf("状态库 schema 版本 %d 高于本程序支持的 %d", current, len(migrations))
	}
	for v := current; v < len(migrations); v++ {
		err := s.tx(ctx, func(tx *sql.Tx) error {
			if _, err := tx.ExecContext(ctx, migrations[v]); err != nil {
				return err
			}
			_, err := tx.ExecContext(ctx, "UPDATE schema_version SET version = ?", v+1)
			return err
		})
		if err != nil {
			return fmt.Errorf("状态库迁移到版本 %d 失败: %w", v+1, err)
		}
	}
	return nil
}

func (s *Store) tx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// LockPlugin 串行化同一插件的管理操作（install/enable/disable/reload/update/rollback/uninstall），
// 返回解锁函数；不同插件互不阻塞，也不影响普通 Handler。
func (s *Store) LockPlugin(pluginID string) (unlock func()) {
	s.mu.Lock()
	m, ok := s.pluginM[pluginID]
	if !ok {
		m = &sync.Mutex{}
		s.pluginM[pluginID] = m
	}
	s.mu.Unlock()
	m.Lock()
	return m.Unlock
}

// Install 是最小安装记录；来源绑定、pin、卸载保留等由 T-09 扩展。
type Install struct {
	PluginID    string
	Source      string
	Entrypoint  string
	BuildDigest []byte
}

// AddInstall 登记安装记录，install_seq 按登记顺序递增（first-wins 的依据）。
func (s *Store) AddInstall(ctx context.Context, in Install) error {
	return s.tx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO installs (plugin_id, install_seq, source, entrypoint, build_digest, created_at)
			 VALUES (?, (SELECT coalesce(max(install_seq), 0) + 1 FROM installs), ?, ?, ?, ?)`,
			in.PluginID, in.Source, in.Entrypoint, in.BuildDigest, time.Now().UnixMilli()); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, "INSERT INTO plugin_state (plugin_id) VALUES (?)", in.PluginID)
		return err
	})
}

// GetInstall 读取安装记录。
func (s *Store) GetInstall(ctx context.Context, pluginID string) (Install, error) {
	in := Install{PluginID: pluginID}
	err := s.db.QueryRowContext(ctx, "SELECT source, entrypoint, build_digest FROM installs WHERE plugin_id = ?", pluginID).
		Scan(&in.Source, &in.Entrypoint, &in.BuildDigest)
	if errors.Is(err, sql.ErrNoRows) {
		return in, fmt.Errorf("插件 %s: %w", pluginID, ErrNotFound)
	}
	return in, err
}

// AddGeneration 持久化一个候选代次（尚未提交）。配置快照只能含 secret 引用，不含明文（T-08）。
func (s *Store) AddGeneration(ctx context.Context, pluginID, generationID, configSnapshot string, configVersion uint32) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO generations (generation_id, plugin_id, config_snapshot, config_version, created_at) VALUES (?, ?, ?, ?, ?)`,
		generationID, pluginID, configSnapshot, configVersion, time.Now().UnixMilli())
	return err
}

// CommitActive 在一个事务内把 ActiveGeneration 从 expectedOld 切换为 newGen，并写入 COMMITTED 标记。
// expectedOld 为空表示首次激活。实际旧值不同返回 ErrStaleGeneration；事务内只操作元数据，不等待插件 RPC。
func (s *Store) CommitActive(ctx context.Context, pluginID, expectedOld, newGen string) error {
	return s.tx(ctx, func(tx *sql.Tx) error {
		var current sql.NullString
		err := tx.QueryRowContext(ctx, "SELECT active_generation FROM plugin_state WHERE plugin_id = ?", pluginID).Scan(&current)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("插件 %s: %w", pluginID, ErrNotFound)
		} else if err != nil {
			return err
		}
		if current.String != expectedOld {
			return fmt.Errorf("%w：预期 %q，实际 %q", ErrStaleGeneration, expectedOld, current.String)
		}
		res, err := tx.ExecContext(ctx,
			"UPDATE generations SET committed_at = ? WHERE generation_id = ? AND plugin_id = ? AND committed_at IS NULL",
			time.Now().UnixMilli(), newGen, pluginID)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n != 1 {
			return fmt.Errorf("代次 %s 不存在、不属于插件 %s 或已提交过", newGen, pluginID)
		}
		_, err = tx.ExecContext(ctx, "UPDATE plugin_state SET active_generation = ? WHERE plugin_id = ?", newGen, pluginID)
		return err
	})
}

// Active 返回插件当前的 ActiveGeneration（未激活时为空）。只有已 COMMITTED 的代次能成为 Active。
func (s *Store) Active(ctx context.Context, pluginID string) (string, error) {
	var g sql.NullString
	err := s.db.QueryRowContext(ctx, "SELECT active_generation FROM plugin_state WHERE plugin_id = ?", pluginID).Scan(&g)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("插件 %s: %w", pluginID, ErrNotFound)
	}
	return g.String, err
}

// SetEnabled 持久化启用状态。
func (s *Store) SetEnabled(ctx context.Context, pluginID string, enabled bool) error {
	res, err := s.db.ExecContext(ctx, "UPDATE plugin_state SET enabled = ? WHERE plugin_id = ?", enabled, pluginID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return fmt.Errorf("插件 %s: %w", pluginID, ErrNotFound)
	}
	return nil
}

// Enabled 读取启用状态。
func (s *Store) Enabled(ctx context.Context, pluginID string) (bool, error) {
	var e bool
	err := s.db.QueryRowContext(ctx, "SELECT enabled FROM plugin_state WHERE plugin_id = ?", pluginID).Scan(&e)
	if errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("插件 %s: %w", pluginID, ErrNotFound)
	}
	return e, err
}
