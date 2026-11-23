// Package policy 管理 Core 的管理员身份、全局禁止与实例发信许可。
//
// [Policy] 每次命令判定读取全局禁止与作用域覆盖；每次发送先检查会话禁止，
// 通过后按原始完整 Target 豁免或状态库中的实例许可准入。
// 配置管理员由设置服务发布，运行期管理员由装配处登记；有效身份合并两份内存列表（D-82）。
package policy

import (
	"context"
	"sync"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/profile"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"google.golang.org/protobuf/proto"
)

// Policy 持有内存中的管理员列表，并逐次读取状态库中的命令策略与发信许可。
type Policy struct {
	store           *state.Store
	adminsMu        sync.RWMutex // 保护管理员列表的整体替换与读取
	admins          []Admin
	runtimeAdmins   []Admin                          // 随本次 Core 运行保留，配置发布只替换 admins。
	commandsChanged func()                           // 开始处理命令策略前装配一次，此后只读。
	namesMu         sync.RWMutex                     // 串行化命令名写入并保护快照；通知 Router 前释放。
	names           map[string]map[string][][]string // plugin_id → route_id → 命令名列表
}

// New 创建使用指定状态库的策略，初始管理员列表为空。
func New(store *state.Store) *Policy { return &Policy{store: store} }

// CheckSessionAccess 按目标实际使用的配置文件检查会话访问权限。
func (p *Policy) CheckSessionAccess(ctx context.Context, target *pb.Target) *pb.ErrorInfo {
	denied, err := profile.SessionDenied(ctx, p.store, target)
	if err != nil {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "read profile session access: " + err.Error()}
	}
	if denied {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, SafeMessage: "target session is denied by its profile"}
	}
	return nil
}

// CheckSend 先拒绝被禁止的会话，再按原始完整 Target 或当前实例许可准入。
// target 由发送服务选定；查询失败返回 INTERNAL，缺少许可返回 PERMISSION_DENIED。
func (p *Policy) CheckSend(ctx context.Context, pluginID string, inv *router.Invocation, target *pb.Target) *pb.ErrorInfo {
	// 1. 会话禁止先于所有实例许可与原目标回复豁免。
	denied, err := p.store.HasGlobalDeny(ctx, state.GlobalDeny{Kind: string(DenyChat), AdapterID: target.GetAdapterId(), SubjectID: target.GetChatId()})
	if err != nil {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "read chat deny: " + err.Error()}
	}
	if denied {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, SafeMessage: "target chat is globally denied"}
	}
	if failure := p.CheckSessionAccess(ctx, target); failure != nil {
		return failure
	}
	// 2. 原目标回复豁免实例许可，thread_id 也参与完整目标比较。
	if inv != nil && inv.PluginID == pluginID && proto.Equal(target, inv.DefaultTarget) {
		return nil
	}
	// 3. 每次准入重新读库，使撤销在下一次发送时生效。
	return p.checkInstance(ctx, pluginID, target.GetAdapterInstanceId())
}

// CheckInstance 按同插件可信在途事件的原实例豁免或当前实例许可准入账号动作。
func (p *Policy) CheckInstance(ctx context.Context, pluginID string, inv *router.Invocation, instanceID string) *pb.ErrorInfo {
	if inv != nil && inv.PluginID == pluginID {
		// bot.online 等事件没有 Target，原实例身份仍保存在可信事件上。
		originalInstanceID := inv.Event.GetAdapterInstanceId()
		if originalInstanceID == "" {
			originalInstanceID = inv.DefaultTarget.GetAdapterInstanceId()
		}
		if originalInstanceID != "" && originalInstanceID == instanceID {
			return nil
		}
	}
	return p.checkInstance(ctx, pluginID, instanceID)
}

func (p *Policy) checkInstance(ctx context.Context, pluginID, instanceID string) *pb.ErrorInfo {
	allowed, err := p.store.AdapterInstanceAllowed(ctx, pluginID, instanceID)
	if err != nil {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "read allowed adapter instances: " + err.Error()}
	}
	if !allowed {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, SafeMessage: "adapter instance is not allowed for plugin"}
	}
	return nil
}
