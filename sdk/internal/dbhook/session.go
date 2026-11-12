// Package dbhook 为 SDK 的配置与数据库准备提供私有状态接入。
//
// [ConfigView] 连接登记时取得的子插件视图与准备后的取值。
// SDK 为每个组件建立 [Session] 并把归属传入生命周期和调用 context。
// [Connection] 串行化首次打开与关闭，使一个 Session 只有一个受管连接。
// 数据库入口通过 [Register] 在插件对象上登记迁移，SDK 通过 [Migrators] 读取声明；
// 本包只依赖标准库与 contracts，避免 SDK 根包链接数据库实现。
package dbhook

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"sync/atomic"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

// Session 保留一个组件的受管连接；descriptor 为 nil 表示数据库可用期之外。
// 全部可变字段由 mu 保护，独立进程入口可以将本 Session 设为默认归属。
type Session struct {
	mu         sync.Mutex
	descriptor *pb.DatabaseDescriptor
	db         *sql.DB
	gorm       any // 缓存按需入口创建的 *gorm.DB，本包保持对 GORM 的静态隔离。
}

type sessionContextKey struct{}

var defaultSession atomic.Pointer[Session]

// Logger 由 SDK 在 init 时登记按组件 ctx 选择日志的函数；独立数据库工具使用默认 Logger。
var Logger = func(context.Context) *slog.Logger { return slog.Default() }

// New 创建一次启动的数据库归属，standalone 指定是否供独立进程的后台调用使用。
func New(standalone bool) *Session {
	s := &Session{}
	if standalone {
		defaultSession.Store(s)
	}
	return s
}

// Context 将本 Session 的归属附在 ctx 上，保留其取消与期限。
func (s *Session) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, s)
}

func sessionFor(ctx context.Context) *Session {
	if s, ok := ctx.Value(sessionContextKey{}).(*Session); ok {
		return s
	}
	return defaultSession.Load()
}

// Begin 开始本 Session 的数据库可用期；descriptor 由 SDK 持有且保持不变。
func (s *Session) Begin(descriptor *pb.DatabaseDescriptor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.descriptor = descriptor
}

// Connection 返回 ctx 所属 Session 的连接，首次调用时通过 open 打开并接管关闭职责。
// open 在锁内执行，须只打开连接，不得再次调用本包的连接入口。
func Connection(ctx context.Context, open func(context.Context, *pb.DatabaseDescriptor) (*sql.DB, error)) (*sql.DB, error) {
	s := sessionFor(ctx)
	if s == nil {
		return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "plugin database session is not available"}
	}
	s.mu.Lock() // 首次打开与关闭共用这把锁，避免重复打开或停止后重新发布连接。
	defer s.mu.Unlock()
	if s.descriptor == nil {
		return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "plugin database session is not available"}
	}
	if s.db != nil {
		return s.db, nil
	}
	db, err := open(ctx, s.descriptor)
	if err != nil {
		return nil, err
	}
	s.db = db
	return db, nil
}

// GORM 返回 ctx 所属 Session 的连接包装，首次调用时通过 wrap 构造。
// wrap 在连接锁内执行；db 须为同一 ctx 的 Connection 返回的受管连接。
func GORM(ctx context.Context, db *sql.DB, wrap func() (any, error)) (any, error) {
	s := sessionFor(ctx)
	if s == nil {
		return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "plugin database session is not available"}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.descriptor == nil || s.db != db {
		return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "plugin database session has stopped"}
	}
	if s.gorm != nil {
		return s.gorm, nil
	}
	wrapped, err := wrap()
	if err != nil {
		return nil, err
	}
	s.gorm = wrapped
	return wrapped, nil
}

// Close 结束本 Session 的数据库可用期并关闭受管连接；重复调用无副作用。
func (s *Session) Close() error {
	s.mu.Lock()
	db := s.db
	s.descriptor = nil
	s.db = nil
	s.gorm = nil
	s.mu.Unlock()
	defaultSession.CompareAndSwap(s, nil)
	if db != nil {
		return db.Close()
	}
	return nil
}
