// Package buttons 签发持久化按钮 token，将点击分流到插件动作或 Core 确认记录。
//
// Service 按插件当前 Active 代次与完整 Target 校验，插件动作 token 可在同代次重启后使用。
// 确认 token 由 Core 按父调用和 Actor 校验并作废；插件动作的点击权限由 Router 逐次检查（MESSAGE-MODEL §6.3）。
package buttons

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log/slog"
	"slices"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// MESSAGE-MODEL §6.3 规定随机 token 大小，按钮值字节上限在 Core 签发边界执行。
const (
	tokenBytes     = 32
	maxValueBytes  = 1024
	pluginPurpose  = "plugin"
	confirmPurpose = "confirm"
	allowAction    = "allow"
	denyAction     = "deny"
)

// Service 使用状态库保留映射，以 current 读取当前 Active 代次与路由快照。
type Service struct {
	store   *state.Store
	now     func() time.Time
	ttl     time.Duration
	current func(string) (string, []*pb.RouteDescriptor, bool)
	confirm func(context.Context, string, bool, *pb.Event) // 装配时注册，运行期间只读
}

// New 创建按钮签发与解析服务；时钟、有效期与当前路由由装配处显式提供。
func New(store *state.Store, now func() time.Time, ttl time.Duration, current func(pluginID string) (generationID string, routes []*pb.RouteDescriptor, ok bool)) *Service {
	return &Service{store: store, now: now, ttl: ttl, current: current}
}

// Issue 清除插件提供的 token，验证动作并持久化整条链的映射，成功后才写回 token。
func (s *Service) Issue(ctx context.Context, caller router.Caller, target *pb.Target, chain *pb.MessageChain) error {
	// 1. 所有按钮先清除不可信 token，只有动作按钮需要路由与持久化映射。
	buttons := chainButtons(chain)
	actionsOnly := buttons[:0]
	for _, button := range buttons {
		button.CoreToken = nil
		if button.Url == nil {
			actionsOnly = append(actionsOnly, button)
		}
	}
	buttons = actionsOnly
	if len(buttons) == 0 {
		return nil
	}
	generation, routes, ok := s.current(caller.PluginID)
	if !ok || generation != caller.GenerationID {
		return invalid("button sender generation is no longer active")
	}
	actions := make(map[string]*pb.RouteDescriptor)
	for _, route := range routes {
		if route.GetKind() == pb.RouteKind_ROUTE_KIND_EVENT && slices.Contains(route.GetEvent().GetEventTypeFilter(), "button.clicked") {
			actions[route.GetEvent().GetActionId()] = route
		}
	}
	// 2. 先验证全部动作与值，随机 token 不携带角色和业务数据。
	records := make([]state.ButtonToken, 0, len(buttons))
	expires := s.now().Add(s.ttl).UnixMilli()
	for _, button := range buttons {
		route, ok := actions[button.GetActionId()]
		if !ok {
			return invalid("button action is not registered by the sending plugin")
		}
		if len(button.GetValue()) > maxValueBytes {
			return invalid("button value exceeds 1024 bytes")
		}
		random := make([]byte, tokenBytes)
		rand.Read(random)
		records = append(records, state.ButtonToken{
			Token: base64.RawURLEncoding.EncodeToString(random), Purpose: pluginPurpose, PluginID: caller.PluginID, RouteID: route.GetRouteId(),
			ActionID: button.GetActionId(), ActionSchemaVersion: route.GetEvent().GetActionSchemaVersion(), GenerationID: generation,
			AdapterID: target.GetAdapterId(), InstanceID: target.GetAdapterInstanceId(), AccountID: target.GetAccountId(), ChatID: target.GetChatId(), ThreadID: target.GetThreadId(),
			Value: button.GetValue(), ExpiresAt: expires,
		})
	}
	// 3. 提交全部映射后才将 token 交给 Adapter，错误完整保留为调用错误。
	if err := s.store.PutButtonTokens(ctx, records); err != nil {
		return &wire.CallError{ErrorInfo: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}}
	}
	for i, button := range buttons {
		button.CoreToken = proto.String(records[i].Token)
	}
	return nil
}

// InterceptConfirm 在 Router 入队前消费确认点击；插件按钮继续排队派发。
func (s *Service) InterceptConfirm(ctx context.Context, ev *pb.Event) bool {
	if ev.GetEventType() != "button.clicked" || ev.GetEventPayload().GetButtonClicked() == nil {
		return false
	}
	record, err := s.store.GetButtonToken(ctx, ev.GetEventPayload().GetButtonClicked().GetValue())
	if err != nil || record.Purpose != confirmPurpose {
		return false // 交回原解析路径处理未知 token 或存储错误。
	}
	if _, _, err := s.Resolve(ctx, ev); err != nil {
		slog.Info("discard invalid confirmation click", "error", err)
	}
	return true
}

// Resolve 将插件点击还原为动作与值，将确认点击交给回调并返回空 pluginID 结束派发。
// 无效点击返回错误，由 Router 记录并丢弃。
func (s *Service) Resolve(ctx context.Context, ev *pb.Event) (pluginID, routeID string, err error) {
	// 1. 点击来自真实平台 Actor，载荷中的 value 此时只携 callback_data。
	payload := ev.GetEventPayload().GetButtonClicked()
	if ev.GetEventType() != "button.clicked" || payload == nil || ev.GetTarget() == nil || ev.GetActor() == nil || ev.GetActor().GetExternalUserId() == "" || ev.GetActor().GetAdapterInstanceId() != ev.GetTarget().GetAdapterInstanceId() {
		return "", "", errors.New("invalid button click event")
	}
	record, err := s.store.GetButtonToken(ctx, payload.GetValue())
	if err != nil {
		return "", "", err
	}
	if !s.now().Before(time.UnixMilli(record.ExpiresAt)) {
		return "", "", errors.New("button token has expired")
	}
	// 2. 代次与目标都从签发记录检查，不接受点击载荷覆盖这些绑定。
	generation, routes, ok := s.current(record.PluginID)
	if !ok || generation != record.GenerationID {
		return "", "", errors.New("button token generation is no longer active")
	}
	target := ev.GetTarget()
	if target.GetAdapterId() != record.AdapterID || target.GetAdapterInstanceId() != record.InstanceID || target.GetAccountId() != record.AccountID || target.GetChatId() != record.ChatID || target.GetThreadId() != record.ThreadID {
		return "", "", errors.New("button token target does not match")
	}
	if record.Purpose == confirmPurpose {
		s.confirm(ctx, record.Value, record.ActionID == allowAction, ev)
		return "", "", nil
	}
	for _, route := range routes {
		event := route.GetEvent()
		if route.GetRouteId() == record.RouteID && slices.Contains(event.GetEventTypeFilter(), "button.clicked") && event.GetActionId() == record.ActionID && event.GetActionSchemaVersion() == record.ActionSchemaVersion {
			payload.ActionId, payload.ActionSchemaVersion, payload.Value = record.ActionID, record.ActionSchemaVersion, record.Value
			return record.PluginID, record.RouteID, nil
		}
	}
	return "", "", errors.New("button token route is no longer registered")
}

// IssueConfirm 签发绑定确认记录、调用插件代次和完整 Target 的允许与拒绝 token。
func (s *Service) IssueConfirm(ctx context.Context, caller router.Caller, target *pb.Target, recordID string) (allow, deny string, err error) {
	generation, _, ok := s.current(caller.PluginID)
	if !ok || generation != caller.GenerationID {
		return "", "", invalid("confirmation sender generation is no longer active")
	}
	expires := s.now().Add(s.ttl)
	if deadline, ok := ctx.Deadline(); ok {
		expires = deadline
	}
	records := make([]state.ButtonToken, 0, 2)
	for _, action := range []string{allowAction, denyAction} {
		random := make([]byte, tokenBytes)
		rand.Read(random)
		records = append(records, state.ButtonToken{
			Token: base64.RawURLEncoding.EncodeToString(random), Purpose: confirmPurpose,
			PluginID: caller.PluginID, GenerationID: generation, ActionID: action, Value: recordID,
			AdapterID: target.GetAdapterId(), InstanceID: target.GetAdapterInstanceId(), AccountID: target.GetAccountId(),
			ChatID: target.GetChatId(), ThreadID: target.GetThreadId(), ExpiresAt: expires.UnixMilli(),
		})
	}
	if err := s.store.PutButtonTokens(ctx, records); err != nil {
		return "", "", &wire.CallError{ErrorInfo: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}}
	}
	return records[0].Token, records[1].Token, nil
}

// RevokeConfirm 作废确认记录的全部按钮 token。
func (s *Service) RevokeConfirm(ctx context.Context, recordID string) error {
	return s.store.RevokeConfirmTokens(ctx, recordID)
}

// OnConfirm 在装配阶段设置确认点击的处理函数；运行期间不得替换。
func (s *Service) OnConfirm(fn func(ctx context.Context, recordID string, allow bool, ev *pb.Event)) {
	s.confirm = fn
}

// Purge 在卸载插件时删除其全部按钮映射。
func (s *Service) Purge(ctx context.Context, pluginID string) error {
	return s.store.PurgeButtonTokens(ctx, pluginID)
}

func chainButtons(chain *pb.MessageChain) []*pb.Button {
	var buttons []*pb.Button
	for _, segment := range chain.GetSegments() {
		for _, row := range segment.GetButtons().GetRows() {
			buttons = append(buttons, row.GetButtons()...)
		}
		for _, item := range segment.GetForward().GetItems() {
			buttons = append(buttons, chainButtons(item.GetChain())...)
		}
	}
	return buttons
}

func invalid(message string) error {
	return &wire.CallError{ErrorInfo: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: message}}
}
