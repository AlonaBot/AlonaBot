// Package state 管理 Core 的元数据库与实例锁。
//
// 状态库位于 <data_dir>/state.sqlite，实例锁位于 <data_dir>/core.lock。Open 先取得实例锁，
// 保证同一个 data_dir 只有一个 Core；已有但无法读取的状态库会报错，不会被重建成空库。
// 提交 ActiveGeneration 时核对预期的旧值，拒绝陈旧候选。规则见 spec/CONTRACTS-v0.1.md §6.1 与 §12.2。
//
// 尚不含候选资源引用、完整的 Journal 与恢复编排。
package state

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/oslock"
	_ "modernc.org/sqlite" // 注册 "sqlite" 驱动
)

const (
	lockFile  = "core.lock"    // data_dir 下的 Core 实例锁
	stateFile = "state.sqlite" // data_dir 下的状态库
)

// 状态库连接设置（CONTRACTS §6.1），以回读 PRAGMA 时得到的值表示。
const (
	journalModeWAL    = "wal"
	synchronousFull   = 2
	foreignKeysOn     = 1
	busyTimeoutMillis = 5000
)

// pragmas 写在 DSN 中，驱动对每个新物理连接都会执行（CONTRACTS §6.1）。
var pragmas = "_pragma=journal_mode(WAL)&_pragma=synchronous(FULL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(" + strconv.Itoa(busyTimeoutMillis) + ")"

var (
	// ErrDataDirLocked 表示同一 data_dir 已有另一个 Core 在运行。
	ErrDataDirLocked = errors.New("data_dir is locked by another Core")
	// ErrStaleGeneration 表示提交时 ActiveGeneration 已不是调用方预期的旧值。
	ErrStaleGeneration = errors.New("ActiveGeneration changed; stale candidate rejected")
	// ErrNotFound 表示记录不存在。
	ErrNotFound = errors.New("record not found")
)

// Store 持有 core.lock 并访问状态库；同一进程内只应有一个。
type Store struct {
	db   *sql.DB      // 单连接的状态库句柄
	lock *oslock.Lock // core.lock，Close 时释放

	// mu 保护 pluginM。
	mu      sync.Mutex
	pluginM map[string]*sync.Mutex // 逐插件管理锁，按插件 ID 惰性创建，不回收
}

// Open 先取得 core.lock，再打开（首次启动时创建）state.sqlite 并核对连接设置与 schema。
// 已存在但无法读取、不是 SQLite 或缺少 schema 的状态库一律报错，不会重建成空库。
func Open(ctx context.Context, dataDir string) (*Store, error) {
	// 1. 统一为绝对路径：预检使用的 file URI 只能表达绝对路径。
	dataDir, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, fmt.Errorf("resolve data_dir: %w", err)
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, fmt.Errorf("create data_dir: %w", err)
	}
	// 2. 先取 core.lock，确保同一 data_dir 只有一个 Core 打开状态库。
	lock, err := oslock.TryLock(filepath.Join(dataDir, lockFile))
	if errors.Is(err, oslock.ErrLocked) {
		return nil, fmt.Errorf("%w（%s）", ErrDataDirLocked, dataDir)
	} else if err != nil {
		return nil, err
	}
	// 3. 打开并核对状态库；失败时释放锁。
	s, err := open(ctx, filepath.Join(dataDir, stateFile))
	if err != nil {
		_ = lock.Unlock()
		return nil, err
	}
	s.lock = lock
	return s, nil
}

// open 打开 path 处的状态库：已有文件先只读预检，再以单连接、固定 PRAGMA 打开并核对设置与 schema。
func open(ctx context.Context, path string) (*Store, error) {
	if strings.ContainsAny(path, "?#") {
		return nil, fmt.Errorf("state database path %q contains ? or #", path)
	}
	_, statErr := os.Stat(path)
	fresh := errors.Is(statErr, os.ErrNotExist)
	if statErr != nil && !fresh {
		return nil, fmt.Errorf("inspect state database: %w", statErr)
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

// Pragmas 记录状态库连接上实际生效的设置。
type Pragmas struct {
	// JournalMode 应为 wal。
	JournalMode string
	// Synchronous 应为 2（FULL）。
	Synchronous int
	// ForeignKeys 应为 1（开启）。
	ForeignKeys int
	BusyTimeout int // 忙等待毫秒数，应为 5000
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

// verify 核对连接设置确实生效，然后执行 schema 迁移；fresh 表示文件是本次新建的。
func (s *Store) verify(ctx context.Context, fresh bool) error {
	p, err := s.ReadPragmas(ctx)
	if err != nil {
		return fmt.Errorf("state database is unreadable (possibly corrupted, will not be recreated): %w", err)
	}
	if want := (Pragmas{journalModeWAL, synchronousFull, foreignKeysOn, busyTimeoutMillis}); p != want {
		return fmt.Errorf("state database connection settings not applied: got %+v, want %+v", p, want)
	}
	return s.migrate(ctx, fresh)
}

// preflight 以只读、无 PRAGMA 的方式打开已有状态库，确认含 schema_version 表且完整性检查通过。
func preflight(ctx context.Context, path string) error {
	p := filepath.ToSlash(path) // path 已由 Open 转为绝对路径
	if !strings.HasPrefix(p, "/") {
		p = "/" + p // Windows 盘符路径 C:/... 需写成 file:///C:/...
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
		return fmt.Errorf("state database is unreadable (possibly corrupted, will not be recreated): %w", err)
	}
	if n == 0 {
		return errors.New("state.sqlite exists but has no AlonaBot schema (empty file or another database); refusing to initialize it as empty")
	}
	var check string
	if err := db.QueryRowContext(ctx, "PRAGMA quick_check").Scan(&check); err != nil || check != "ok" {
		return fmt.Errorf("state database integrity check failed (%q, %v); will not be recreated", check, err)
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
	`CREATE TABLE pending_configs (
	   plugin_id       TEXT PRIMARY KEY REFERENCES installs(plugin_id),
	   base_generation TEXT NOT NULL,
	   config_json     TEXT NOT NULL,
	   updated_at      INTEGER NOT NULL
	 );`,
}

// migrate 只在全新文件上从版本 0 开始；已有文件的 schema_version 表已由 verify 确认存在。
func (s *Store) migrate(ctx context.Context, fresh bool) error {
	current := 0
	if !fresh {
		if err := s.db.QueryRowContext(ctx, "SELECT version FROM schema_version").Scan(&current); err != nil {
			return fmt.Errorf("read schema version: %w", err)
		}
	}
	if current > len(migrations) {
		return fmt.Errorf("state database schema version %d is newer than supported version %d", current, len(migrations))
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
			return fmt.Errorf("migrate state database to version %d: %w", v+1, err)
		}
	}
	return nil
}

// tx 在一个事务中执行 fn：fn 返回错误时回滚，否则提交。
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

// Install 记录插件的最小安装信息；尚不含来源绑定、pin、卸载保留等内容。
type Install struct {
	PluginID    string // 插件 ID，主键
	Source      string // 安装来源（开发模式为 dev-local）
	Entrypoint  string // 入口可执行文件的路径
	BuildDigest []byte // 安装时已验证的归档摘要，经启动上下文交给 SDK 回显
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
		return in, fmt.Errorf("plugin %s: %w", pluginID, ErrNotFound)
	}
	return in, err
}

// AddGeneration 持久化一个候选代次（尚未提交）。配置快照只能含 secret 引用，不含明文。
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
		// 1. 核对当前 Active 等于调用方预期的旧值。
		var current sql.NullString
		err := tx.QueryRowContext(ctx, "SELECT active_generation FROM plugin_state WHERE plugin_id = ?", pluginID).Scan(&current)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("plugin %s: %w", pluginID, ErrNotFound)
		} else if err != nil {
			return err
		}
		if current.String != expectedOld {
			return fmt.Errorf("%w: expected %q, got %q", ErrStaleGeneration, expectedOld, current.String)
		}
		// 2. 给新代次写入 COMMITTED 标记；它必须存在、属于该插件且从未提交过。
		res, err := tx.ExecContext(ctx,
			"UPDATE generations SET committed_at = ? WHERE generation_id = ? AND plugin_id = ? AND committed_at IS NULL",
			time.Now().UnixMilli(), newGen, pluginID)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n != 1 {
			return fmt.Errorf("generation %s does not exist, does not belong to plugin %s, or was already committed", newGen, pluginID)
		}
		// 3. 切换 Active。
		_, err = tx.ExecContext(ctx, "UPDATE plugin_state SET active_generation = ? WHERE plugin_id = ?", newGen, pluginID)
		return err
	})
}

// Active 返回插件当前的 ActiveGeneration（未激活时为空）。只有已 COMMITTED 的代次能成为 Active。
func (s *Store) Active(ctx context.Context, pluginID string) (string, error) {
	var g sql.NullString
	err := s.db.QueryRowContext(ctx, "SELECT active_generation FROM plugin_state WHERE plugin_id = ?", pluginID).Scan(&g)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("plugin %s: %w", pluginID, ErrNotFound)
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
		return fmt.Errorf("plugin %s: %w", pluginID, ErrNotFound)
	}
	return nil
}

// Enabled 读取启用状态。
func (s *Store) Enabled(ctx context.Context, pluginID string) (bool, error) {
	var e bool
	err := s.db.QueryRowContext(ctx, "SELECT enabled FROM plugin_state WHERE plugin_id = ?", pluginID).Scan(&e)
	if errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("plugin %s: %w", pluginID, ErrNotFound)
	}
	return e, err
}
