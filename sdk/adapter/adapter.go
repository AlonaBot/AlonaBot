// Package adapter 定义独立 Adapter 进程的平台接入接口。
//
// [Adapter] 创建账号实例，[Instance] 提供发送和停止，[MediaSource] 提供可选的平台文件读取。
// [Editor] 提供可选的消息编辑入口，[CommandSynchronizer] 提供命令菜单同步，
// [MessageChecker] 只读检查出站格式与组合。
// [Core] 将事件与游标交给宿主，并读取出站媒体。
// 实例配置只在启动请求中传递；生命周期由 sdk.RunAdapter 管理。
// 包级 HTTP 路由在 [Descriptor.HTTP] 中声明，Listener 随 Session 托管。
package adapter

import (
	"context"
	"time"

	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// Account 保存已经平台验证的账号身份。
type Account struct {
	ID          string
	DisplayName string // 空值表示平台没有提供展示名
	Username    string // 平台上可寻址的账号标识；空值表示平台没有用户名
}

// StartRequest 携带实例配置与 Core 已保存的不透明游标。
type StartRequest struct {
	InstanceID string
	ConfigJSON []byte // 含凭证，只在实例启动时使用
	Cursor     []byte
}

// Core 绑定一个实例，交付事件、推进游标、报告致命故障并读取出站媒体。
type Core interface {
	// Deliver 交付事件与游标；accepted=false 表示队列满而丢弃，游标已持久化，Adapter 不重放。
	Deliver(ctx context.Context, ev *message.Event, cursor []byte) (accepted bool, err error)
	// SaveCursor 保存没有事件时推进的接收游标。
	SaveCursor(ctx context.Context, cursor []byte) error
	// ReportFailure 关闭实例交付入口；调用后 Deliver 返回错误。
	ReportFailure(ctx context.Context, safe *errcode.Error) error
	// ReadMedia 读取 Core 为当前 Adapter Session 签发的只读媒体令牌。
	ReadMedia(ctx context.Context, token string, offset uint64, length uint32) ([]byte, uint64, bool, error)
	// MediaURL 为本次发送的只读令牌取得临时 URL；baseURL 为空时使用 Core 配置。
	MediaURL(ctx context.Context, token, baseURL string) (url string, expiresAt time.Time, err error)
}

// MediaSource 提供实例可选的平台文件读取能力，length 为 1～1 MiB。
// 未实现此接口的实例在收到 adapter.media.read 时返回 UNSUPPORTED。
type MediaSource interface {
	ReadPlatformFile(ctx context.Context, fileID string, offset uint64, length uint32) ([]byte, uint64, bool, error)
}

// Instance 提供一个平台账号的身份、能力、发送与停止操作。
type Instance interface {
	// Account 返回已经平台验证的账号身份。
	Account() Account
	// Capabilities 返回本实例声明的平台能力。
	Capabilities() []message.Capability
	// Send 同时保留发送结果和安全错误，包括部分成功。
	Send(ctx context.Context, to string, chain message.Chain) (*message.SendResult, *errcode.Error)
	// Stop 返回前停止接收；此后实例的 Deliver、SaveCursor 均返回 FAILED_PRECONDITION。
	Stop(ctx context.Context) error
}

// Adapter 描述平台接入，并为每个账号创建独立实例。
type Adapter interface {
	// Descriptor 返回平台注册描述。
	Descriptor() *Descriptor
	// Start 完成凭证与账号检查、启动接收后返回；Core 在实例停止前一直有效。
	Start(ctx context.Context, req StartRequest, core Core) (Instance, error)
}

// NodesReader 读取合并转发资源或同会话消息引用，返回完整递归节点及部分错误。
type NodesReader interface {
	ReadNodes(ctx context.Context, to string, segment message.ForwardSegment) ([]message.ForwardItem, *errcode.Error)
}

// MessageChecker 检查格式、来源及组合；只读取消息和 Core 授权的元数据。
// 不取媒体、请求平台、签发按钮或创建租约；缺少决定组合的元数据返回 unknown。
type MessageChecker interface {
	CheckMessage(ctx context.Context, to string, chain message.Chain) (message.SupportStatus, *errcode.Error)
}
