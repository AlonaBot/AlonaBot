// Package policy 实现 Core 的实例发信许可检查。
//
// [Policy] 对可信调用的原始完整 Target 放行，其他目标逐次读取状态库中的实例许可。
// 许可仅控制发送，不赋予 User/Admin 身份（CONTRACTS §8）。
package policy

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"google.golang.org/protobuf/proto"
)

// Policy 从 Core 状态库读取插件当前的实例发信许可，不缓存查询结果。
type Policy struct{ store *state.Store }

// New 创建使用指定状态库的发送策略。
func New(store *state.Store) *Policy { return &Policy{store: store} }

// CheckSend 放行同一插件对原始完整 Target 的回复，其余目标按当前实例许可准入。
// target 由发送服务选定；查询失败返回 INTERNAL，缺少许可返回 PERMISSION_DENIED。
func (p *Policy) CheckSend(ctx context.Context, pluginID string, inv *router.Invocation, target *pb.Target) *pb.ErrorInfo {
	// 1. 原目标回复豁免实例许可，thread_id 也参与完整目标比较。
	if inv != nil && inv.PluginID == pluginID && proto.Equal(target, inv.DefaultTarget) {
		return nil
	}
	// 2. 每次准入重新读库，使撤销在下一次发送时生效。
	allowed, err := p.store.AdapterInstanceAllowed(ctx, pluginID, target.GetAdapterInstanceId())
	if err != nil {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "read allowed adapter instances: " + err.Error()}
	}
	if !allowed {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, SafeMessage: "adapter instance is not allowed for plugin"}
	}
	return nil
}
