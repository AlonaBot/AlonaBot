// Package dbhook 在 SDK 与按需导入的数据库入口之间保存迁移登记及 Session 连接。
//
// SDK 在配置准备完成后调用 [Begin]，在准备失败或停止时调用 [Close]。
// [Connection] 串行化首次打开与关闭，使一个 Session 只有一个受管连接。
// 数据库入口通过 [Register] 登记迁移识别函数，SDK 通过 [Recognize] 收集声明；
// 本包只依赖标准库与 contracts，避免 SDK 根包链接数据库实现。
package dbhook

import (
	"context"
	"database/sql"
	"sync"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

// current 的全部字段由 mu 保护；descriptor 为 nil 表示数据库可用期之外。
var current struct {
	mu         sync.Mutex
	descriptor *pb.DatabaseDescriptor
	db         *sql.DB
	close      func() error
	gorm       any // 缓存按需入口创建的 *gorm.DB，本包保持对 GORM 的静态隔离。
}

// Begin 开始当前 Session 的数据库可用期；SDK 须先关闭上一 Session。
// descriptor 在该可用期内由 SDK 持有且保持不变。
func Begin(descriptor *pb.DatabaseDescriptor) {
	current.mu.Lock()
	defer current.mu.Unlock()
	current.descriptor = descriptor
}

// Connection 返回当前 Session 的连接，首次调用时通过 open 打开并接管关闭职责。
// open 在锁内执行，须只打开连接，不得再次调用本包的连接入口。
func Connection(ctx context.Context, open func(context.Context, *pb.DatabaseDescriptor) (*sql.DB, error)) (*sql.DB, error) {
	current.mu.Lock() // 首次打开与关闭共用这把锁，避免重复打开或停止后重新发布连接。
	defer current.mu.Unlock()
	if current.descriptor == nil {
		return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "plugin database session is not available"}
	}
	if current.db != nil {
		return current.db, nil
	}
	db, err := open(ctx, current.descriptor)
	if err != nil {
		return nil, err
	}
	current.db = db
	current.close = db.Close
	return db, nil
}

// GORM 返回 db 对应的 Session 包装，首次调用时通过 wrap 构造。
// wrap 在连接锁内执行；db 须为 Connection 返回的受管连接。
func GORM(db *sql.DB, wrap func() (any, error)) (any, error) {
	current.mu.Lock()
	defer current.mu.Unlock()
	if current.descriptor == nil || current.db != db {
		return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "plugin database session has stopped"}
	}
	if current.gorm != nil {
		return current.gorm, nil
	}
	wrapped, err := wrap()
	if err != nil {
		return nil, err
	}
	current.gorm = wrapped
	return wrapped, nil
}

// Close 结束数据库可用期并关闭受管连接；重复调用无副作用。
func Close() error {
	current.mu.Lock()
	closeDB := current.close
	current.descriptor = nil
	current.db = nil
	current.close = nil
	current.gorm = nil
	current.mu.Unlock()
	if closeDB != nil {
		return closeDB()
	}
	return nil
}
