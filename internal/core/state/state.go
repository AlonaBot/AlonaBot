// Package state 管理 Core 的元数据库与实例锁。
//
// 状态库位于 <data_dir>/state.sqlite，实例锁位于 <data_dir>/core.lock。Open 先取得实例锁，
// 保证同一个 data_dir 只有一个 Core；已有但无法读取的状态库会报错，不会被重建成空库。
// 提交 ActiveGeneration 时核对预期的旧值，拒绝陈旧候选。规则见 spec/CONTRACTS-v0.1.md §6.1 与 §12.2。
//
// Secret 与配置路径随代次保存；正式 Journal 与恢复编排由管理层负责。
// Unix 核对数据库及 WAL/SHM 的 0600 权限，Windows 为 data_dir 设置当前账号专用的可继承 DACL。
package state

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/oslock"
	"github.com/libtnb/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
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
	errEmptyPluginID        = errors.New("plugin_id is required")
	errEmptyConversationKey = errors.New("conversation key is required")

	// ErrDataDirLocked 表示同一 data_dir 已有另一个 Core 在运行。
	ErrDataDirLocked = errors.New("data_dir is locked by another Core")
	// ErrStaleGeneration 表示提交时 ActiveGeneration 已不是调用方预期的旧值。
	ErrStaleGeneration = errors.New("ActiveGeneration changed; stale candidate rejected")
	// ErrNotFound 表示记录不存在。
	ErrNotFound = errors.New("record not found")
)

// Store 持有 core.lock 并访问状态库；同一进程内只应有一个。
type Store struct {
	db   *gorm.DB     // 单连接的状态库句柄
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
	if err := protectStateDirectory(dataDir); err != nil {
		return nil, err
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
		slog.Log(ctx, databaseErrorLevel(err), "state database open or verification failed", "error", err)
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
	} else {
		// 主库先以仅当前账号可读写的权限创建，WAL/SHM 由 SQLite 继承保护。
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			return nil, fmt.Errorf("create state database %q: %w", path, err)
		}
		if err := file.Close(); err != nil {
			return nil, fmt.Errorf("close new state database %q: %w", path, err)
		}
	}
	db, err := gorm.Open(sqlite.Open(path+"?"+pragmas), &gorm.Config{Logger: stateLogger{logger.Default.LogMode(logger.Silent)}})
	if err != nil {
		return nil, err
	}
	connection, err := db.DB()
	if err != nil {
		return nil, err
	}
	connection.SetMaxOpenConns(1)
	connection.SetMaxIdleConns(1)
	s := &Store{db: db, pluginM: map[string]*sync.Mutex{}}
	if err := s.verify(ctx, fresh); err != nil {
		_ = connection.Close()
		return nil, err
	}
	if err := checkStatePermissions(path); err != nil {
		return nil, errors.Join(err, connection.Close())
	}
	return s, nil
}

// Close 关闭状态库并释放 core.lock。
func (s *Store) Close() error {
	connection, err := s.db.DB()
	if err != nil {
		return errors.Join(err, s.lock.Unlock())
	}
	return errors.Join(connection.Close(), s.lock.Unlock())
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
		s.db.WithContext(ctx).Raw("PRAGMA journal_mode").Scan(&p.JournalMode).Error,
		s.db.WithContext(ctx).Raw("PRAGMA synchronous").Scan(&p.Synchronous).Error,
		s.db.WithContext(ctx).Raw("PRAGMA foreign_keys").Scan(&p.ForeignKeys).Error,
		s.db.WithContext(ctx).Raw("PRAGMA busy_timeout").Scan(&p.BusyTimeout).Error,
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

//go:embed schema.sql
var initialSchema string

// migrations 按顺序执行，版本号即下标加 1。首次发布前只有初始这一条，schema 改动直接写在其中，
// 不追加迁移（编码规范「首次发布前不写兼容层」）。
var migrations = []string{initialSchema}

// migrate 只在全新文件上从版本 0 开始；已有文件的 schema_version 表已由 verify 确认存在。
func (s *Store) migrate(ctx context.Context, fresh bool) error {
	current := 0
	if !fresh {
		var version schemaVersion
		if err := s.db.WithContext(ctx).Take(&version).Error; err != nil {
			return fmt.Errorf("read schema version: %w", err)
		}
		current = version.Version
	}
	if current > len(migrations) {
		return fmt.Errorf("state database schema version %d is newer than supported version %d", current, len(migrations))
	}
	for v := current; v < len(migrations); v++ {
		err := s.tx(ctx, func(tx *gorm.DB) error {
			if err := tx.Exec(migrations[v]).Error; err != nil {
				return err
			}
			return tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&schemaVersion{}).Updates(&schemaVersion{Version: v + 1}).Error
		})
		if err != nil {
			return fmt.Errorf("migrate state database to version %d: %w", v+1, err)
		}
	}
	return nil
}

// tx 在一个事务中执行 fn：fn 返回错误时回滚，否则提交。
func (s *Store) tx(ctx context.Context, fn func(*gorm.DB) error, opts ...*sql.TxOptions) error {
	var operationErr error
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		operationErr = fn(tx)
		return operationErr
	}, opts...)
	// SQL 执行错误已由 Trace 记录；这里记录事务开始或提交本身的失败。
	if err != nil && operationErr == nil {
		slog.Log(ctx, databaseErrorLevel(err), "state database transaction failed", "error", err)
	}
	return err
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

// Role 标记可信安装记录中的进程角色。
type Role string

const (
	// RolePlugin 指定插件进程。
	RolePlugin Role = "plugin"
	// RoleAdapter 指定承载平台账号实例的 Adapter 进程。
	RoleAdapter Role = "adapter"
	// RoleAgent 指定全局唯一的 Agent 组件。
	RoleAgent Role = "agent"
	// RoleWebUI 指定全局唯一的静态管理界面组件。
	RoleWebUI Role = "webui"
)

// Install 保存组件身份、来源绑定、版本约束与卸载保留状态。
type Install struct {
	Role           Role   `gorm:"column:role"`                 // 必填的安装角色
	AdapterID      string `gorm:"column:adapter_id"`           // RoleAdapter 必填的平台类型
	PluginID       string `gorm:"column:plugin_id;primaryKey"` // 插件 ID，主键
	Source         string `gorm:"column:source"`               // 安装来源（库模式为 dev）
	Entrypoint     string `gorm:"column:entrypoint"`           // 入口可执行文件的路径
	BuildDigest    []byte `gorm:"column:build_digest"`         // 安装时已验证的归档摘要，经启动上下文交给 SDK 回显
	Version        string `gorm:"column:version"`
	PublisherKeyID string `gorm:"column:publisher_key_id"` // 无签名时为空串。
	Pin            string `gorm:"column:pin"`              // 空串表示允许所有版本。
	Uninstalled    bool   `gorm:"column:uninstalled"`      // 保留数据与来源绑定的已卸载记录。
}

// AddInstall 登记安装记录，install_seq 按登记顺序递增（first-wins 的依据）。
// Role 必须明确指定；Adapter 按 adapter_id 唯一，Agent 与 WebUI 各自全局唯一。
// 同来源、同发布者的已卸载记录可以重装，保留启用状态行与代次并刷新安装顺序。
func (s *Store) AddInstall(ctx context.Context, in Install) error {
	// 1. 校验安装角色与 Adapter 的必填身份。
	if in.Role != RolePlugin && in.Role != RoleAdapter && in.Role != RoleAgent && in.Role != RoleWebUI {
		return fmt.Errorf("invalid installation role %q", in.Role)
	}
	if in.Role == RoleAdapter && in.AdapterID == "" {
		return errors.New("adapter installation requires adapter_id")
	}
	return s.tx(ctx, func(tx *gorm.DB) error {
		// 2. 在单连接事务内核对平台类型唯一性并写入安装与启用状态。
		var rows []installation
		if err := tx.Find(&rows).Error; err != nil {
			return err
		}
		var previous *installation
		for i := range rows {
			row := &rows[i]
			if row.PluginID == in.PluginID {
				if !row.Uninstalled {
					return fmt.Errorf("component %s is already installed", in.PluginID)
				}
				if row.Source != in.Source || row.PublisherKeyID != in.PublisherKeyID || row.Role != in.Role || row.AdapterID != in.AdapterID {
					return fmt.Errorf("component %s installation binding differs", in.PluginID)
				}
				previous = row
			}
			if !row.Uninstalled && row.Role == in.Role && (in.Role == RoleAgent || in.Role == RoleWebUI || (in.Role == RoleAdapter && row.AdapterID == in.AdapterID)) {
				return fmt.Errorf("component %s of role %s is already installed", row.PluginID, in.Role)
			}
		}
		var last installation
		err := tx.Order(clause.OrderByColumn{Column: clause.Column{Name: columnInstallSeq}, Desc: true}).Take(&last).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		sequence := last.InstallSeq + 1
		if previous != nil {
			in.Pin = previous.Pin
			previous.Install, previous.InstallSeq = in, sequence
			return tx.Save(previous).Error
		}
		row := installation{Install: in, InstallSeq: sequence, CreatedAt: time.Now().UnixMilli()}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		if err := tx.Create(&pluginState{PluginID: in.PluginID}).Error; err != nil {
			return err
		}
		if in.Role == RolePlugin || in.Role == RoleAgent {
			return tx.Create(&pluginRuntime{PluginID: in.PluginID, RuntimeStatus: stoppedRuntime("installed")}).Error
		}
		return nil
	})
}

// GetInstall 读取安装记录。
func (s *Store) GetInstall(ctx context.Context, pluginID string) (Install, error) {
	var row installation
	err := s.db.WithContext(ctx).Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Take(&row).Error
	row.Install.PluginID = pluginID
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row.Install, fmt.Errorf("plugin %s: %w", pluginID, ErrNotFound)
	}
	return row.Install, err
}

// AddGeneration 持久化一个候选代次（尚未提交）。配置快照只能含 secret 引用，不含明文。
func (s *Store) AddGeneration(ctx context.Context, pluginID, generationID, configSnapshot string, configVersion uint32) error {
	return s.db.WithContext(ctx).Create(&generation{
		GenerationID: generationID, PluginID: pluginID, ConfigSnapshot: configSnapshot,
		ConfigVersion: configVersion, CreatedAt: time.Now().UnixMilli(),
	}).Error
}

// CommitActive 在一个事务内把 ActiveGeneration 从 expectedOld 切换为 newGen，并写入 COMMITTED 标记。
// expectedOld 为空表示首次激活。实际旧值不同返回 ErrStaleGeneration；事务内只操作元数据，不等待插件 RPC。
func (s *Store) CommitActive(ctx context.Context, pluginID, expectedOld, newGen string) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		// 1. 核对当前 Active 等于调用方预期的旧值。
		var current pluginState
		err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Take(&current).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("plugin %s: %w", pluginID, ErrNotFound)
		} else if err != nil {
			return err
		}
		if current.ActiveGeneration.String != expectedOld {
			return fmt.Errorf("%w: expected %q, got %q", ErrStaleGeneration, expectedOld, current.ActiveGeneration.String)
		}
		// 2. 给属于该插件且尚未提交的新代次写入 COMMITTED 标记。
		var candidate generation
		err = tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: newGen}).Take(&candidate).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) || candidate.PluginID != pluginID || candidate.CommittedAt != nil {
			return fmt.Errorf("generation %s does not exist, does not belong to plugin %s, or was already committed", newGen, pluginID)
		}
		committedAt := time.Now().UnixMilli()
		candidate.CommittedAt = &committedAt
		if err := tx.Save(&candidate).Error; err != nil {
			return err
		}
		// 3. 切换 Active；事务内保留其余状态字段。
		current.ActiveGeneration = sql.NullString{String: newGen, Valid: true}
		return tx.Save(&current).Error
	})
}

// Active 返回插件当前的 ActiveGeneration（未激活时为空）。只有已 COMMITTED 的代次能成为 Active。
func (s *Store) Active(ctx context.Context, pluginID string) (string, error) {
	var row pluginState
	err := s.db.WithContext(ctx).Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("plugin %s: %w", pluginID, ErrNotFound)
	}
	return row.ActiveGeneration.String, err
}

// SetEnabled 持久化启用状态。
func (s *Store) SetEnabled(ctx context.Context, pluginID string, enabled bool) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		var row pluginState
		err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Take(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("plugin %s: %w", pluginID, ErrNotFound)
		}
		if err != nil {
			return err
		}
		row.Enabled = enabled
		return tx.Save(&row).Error
	})
}

// Enabled 读取启用状态。
func (s *Store) Enabled(ctx context.Context, pluginID string) (bool, error) {
	var row pluginState
	err := s.db.WithContext(ctx).Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, fmt.Errorf("plugin %s: %w", pluginID, ErrNotFound)
	}
	return row.Enabled, err
}

// stateLogger 将数据库执行故障交给 Core 日志，避免输出可能包含秘密的 SQL。
type stateLogger struct{ logger.Interface }

func (stateLogger) Trace(ctx context.Context, _ time.Time, _ func() (string, int64), err error) {
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Log(ctx, databaseErrorLevel(err), "state database operation failed", "error", err)
	}
}

func databaseErrorLevel(err error) slog.Level {
	// 请求生命周期结束无需严重告警，按返回的错误分类以保留独立存储故障的级别。
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return slog.LevelDebug
	}
	return logging.LevelCritical
}
