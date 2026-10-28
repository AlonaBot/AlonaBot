// Package adapter 定义独立 Adapter 进程的平台接入接口。
//
// [Adapter] 创建账号实例，[Instance] 提供发送和停止，[Core] 将事件与游标交给宿主。
// 实例配置只在启动请求中传递；生命周期由 sdk.RunAdapter 管理。
package adapter

import (
	"context"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// Account 保存已经平台验证的账号身份。
type Account struct {
	ID          string
	DisplayName string // 空值表示平台没有提供展示名
}

// StartRequest 携带实例配置与 Core 已保存的不透明游标。
type StartRequest struct {
	InstanceID string
	ConfigJSON []byte // 含凭证，只在实例启动时使用
	Cursor     []byte
}

// Core 绑定一个实例，交付事件、推进游标和报告致命故障。
type Core interface {
	Deliver(ctx context.Context, ev *pb.Event, cursor []byte) (accepted bool, err error)
	SaveCursor(ctx context.Context, cursor []byte) error
	// ReportFailure 关闭实例交付入口；调用后 Deliver 返回错误。
	ReportFailure(ctx context.Context, safe *pb.ErrorInfo) error
}

// Instance 提供一个平台账号的身份、能力、发送与停止操作。
type Instance interface {
	Account() Account
	Capabilities() []*pb.Capability
	// Send 同时保留发送结果和安全错误，包括部分成功。
	Send(ctx context.Context, target *pb.Target, chain *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo)
	Stop(ctx context.Context) error
}

// Adapter 描述平台接入，并为每个账号创建独立实例。
type Adapter interface {
	Descriptor() *pb.AdapterDescriptor
	// Start 完成凭证与账号检查、启动接收后返回；Core 在实例停止前一直有效。
	Start(ctx context.Context, req StartRequest, core Core) (Instance, error)
}
