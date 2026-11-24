// Package message 实现 Core 消息发送与编辑服务的目标准入和回执转发。
//
// [Service] 根据可信父调用的原目标或持久化的实例许可准入，先校验载荷、目标和消息链，
// 再执行出站 Hook、准备媒体消费租约并由注册的 Adapter 实例执行发送；结果与错误同时保留。
// 编辑仅接受文本及可选按钮，实例须声明能力并实现 adapter.Editor；
// 撤回逐项返回结果，实例须声明能力并实现 adapter.Deleter。
package message

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const schemaVersion = 1

// Hooks 在每次 Send/Edit 的预检后改写或拒绝消息，并在平台操作后观察完整结果。
type Hooks interface {
	// BeforeSend 每次改写后立即调用 check；nil 表示保留原 Chain，检查失败即终止。
	BeforeSend(ctx context.Context, op pb.SendOperation, edited string, target *pb.Target, chain *pb.MessageChain, parent *router.Invocation, check func(*pb.MessageChain) *pb.ErrorInfo) (*pb.MessageChain, *pb.ErrorInfo)
	// PrepareAfterSend 在平台副作用前预留观察容量；complete 与 release 共享一次释放所有权。
	PrepareAfterSend(ctx context.Context, source router.Caller, op pb.SendOperation, edited string, target *pb.Target, chain *pb.MessageChain, parent *router.Invocation) (complete func(*pb.SendResult, *pb.ErrorInfo), release func(), err error)
}

// Service 按插件发信许可准入目标，再由注册的 Adapter 实例执行发送或编辑。
type Service struct {
	observe         func([]byte, string, *pb.Target, *pb.MessageChain, *pb.SendResult, *pb.ErrorInfo)
	prepareIncoming func(*pb.MessageChain, router.Caller, *pb.Target) (*pb.MessageChain, error)
	replyMu         sync.Mutex // 保护根首条标记与 Session 回执关联，不跨配置读取和平台调用持锁。
	resolveReply    func(context.Context, *pb.Target) (settings.Reply, error)
	replyRoots      map[*router.Invocation]struct{}
	replySessions   map[router.Caller]*replySession
	replyDB         *gorm.DB
	requestTimeout  time.Duration
	adapters        *adapter.Registry
	policy          *policy.Policy
	hooks           Hooks // nil 时跳过出站 Hook
	issue           func(context.Context, router.Caller, *pb.Target, *pb.MessageChain) error
	lease           func(context.Context, router.Caller, string, *pb.Target, *pb.MediaSource, time.Time) (*pb.MediaSource, func(), error)
}

// New 创建消息服务；装配处登记 Send、Edit 与 ReadNodes，并通过 SetIncomingChain 绑定查询媒体授权。
// lease 由媒体 Store 提供，准备平台文件标识或绑定目标 Adapter Session 的只读令牌；
// issue 在媒体租约前签发按钮；nil 仅供无按钮的测试。
// requestTimeout 使用生效配置，在调用没有期限时提供媒体请求预算。
func New(adapters *adapter.Registry, p *policy.Policy, lease func(context.Context, router.Caller, string, *pb.Target, *pb.MediaSource, time.Time) (*pb.MediaSource, func(), error), issue func(context.Context, router.Caller, *pb.Target, *pb.MessageChain) error, requestTimeout time.Duration, hooks Hooks) *Service {
	return &Service{adapters: adapters, policy: p, lease: lease, issue: issue, requestTimeout: requestTimeout, hooks: hooks}
}

// Send 按载荷、可信目标、消息链和实例绑定依次准入，不自动重试平台发送。
// 拒绝时同时返回 REJECTED 回执与错误，发送后保留 Adapter 返回的二者。
func (s *Service) Send(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	ctx, cancel := s.requestContext(ctx, in)
	defer cancel()
	ctx = wire.WithTraceID(ctx, in.TraceID)
	// 1. 取得已知的可信目标供拒绝回执使用，再校验协议载荷。
	invocation, _ := in.ParentAttach.(*router.Invocation)
	var defaultTarget *pb.Target
	if invocation != nil {
		defaultTarget = invocation.DefaultTarget
		if invocation.SendForbidden {
			return reject(defaultTarget, pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "sending is forbidden in outbound hooks")
		}
	}
	request := &pb.SendRequest{}
	if in.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.Payload.SchemaID != validate.SchemaID(request) || in.Payload.SchemaVersion != schemaVersion {
		return reject(defaultTarget, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "expected SendRequest protobuf schema version 1")
	}
	if err := proto.Unmarshal(in.Payload.Bytes, request); err != nil {
		return reject(defaultTarget, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid SendRequest protobuf: "+err.Error())
	}
	// 2. 显式目标先补齐实例与账号，否则沿用可信父调用的默认目标，再检查发送许可。
	target := request.To
	if target == nil {
		target = defaultTarget
	} else {
		completed, completionError := s.adapters.Complete(target, defaultTarget)
		if completionError != nil {
			return response(&pb.SendResult{Status: pb.SendStatus_SEND_STATUS_REJECTED, Target: target}, completionError)
		}
		target = completed
	}
	if target == nil {
		return reject(nil, pb.ErrorCode_ERROR_CODE_NO_DEFAULT_TARGET, "no default target")
	}
	if sendError := s.policy.CheckSend(ctx, caller.PluginID, invocation, target); sendError != nil {
		// 准入拒绝只回传调用方已知的目标，避免带出补齐的实例身份。
		rejectedTarget := request.To
		if rejectedTarget == nil {
			rejectedTarget = defaultTarget
		}
		return response(&pb.SendResult{Status: pb.SendStatus_SEND_STATUS_REJECTED, Target: rejectedTarget}, sendError)
	}
	// 3. 复用结构校验，空链和仅元数据均在产生副作用前拒绝。
	if safe := s.checkChain(request.Chain, target); safe != nil {
		return response(&pb.SendResult{Status: pb.SendStatus_SEND_STATUS_REJECTED, Target: target}, safe)
	}
	// 4. 校验最终实例绑定，媒体来源准备也使用这一实例的 Session。
	instance, ok := s.adapters.Resolve(ctx, target.GetAdapterInstanceId())
	if !ok {
		return reject(target, pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is unavailable")
	}
	if target.GetAdapterId() != instance.AdapterID() || target.GetAccountId() != instance.AccountID() {
		return reject(target, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "target adapter or account does not match instance")
	}
	chain, decoration, session, err := s.prepareReply(ctx, caller, invocation, target, instance, request.Chain)
	if err != nil {
		return rejectError(target, err)
	}
	request.Chain = chain
	if safe := s.checkChain(request.Chain, target); safe != nil {
		return response(&pb.SendResult{Status: pb.SendStatus_SEND_STATUS_REJECTED, Target: target}, safe)
	}
	// 5. 整链能力与扩展校验先于按钮签发、媒体租约和平台副作用。
	if safe := checkExtensions(instance, request.Chain); safe != nil {
		return response(&pb.SendResult{Status: pb.SendStatus_SEND_STATUS_REJECTED, Target: target}, safe)
	}
	// 6. Hook 只能改写消息链；复检通过后才允许按钮签发和媒体租约。
	if s.hooks != nil {
		check := func(chain *pb.MessageChain) *pb.ErrorInfo {
			if safe := s.checkChain(chain, target); safe != nil {
				return safe
			}
			return checkExtensions(instance, chain)
		}
		chain, safe := s.hooks.BeforeSend(ctx, pb.SendOperation_SEND_OPERATION_SEND, "", target, request.Chain, invocation, check)
		if safe != nil {
			return response(&pb.SendResult{Status: pb.SendStatus_SEND_STATUS_REJECTED, Target: target}, safe)
		}
		if chain != nil {
			request.Chain = chain
		}
	}
	if s.issue != nil {
		if err := s.issue(ctx, caller, target, request.Chain); err != nil {
			return rejectError(target, err)
		}
	}
	// 7. 准备引用来源；任一失败或发送结束均释放本次已得租约。
	sources := mediaSources(request.Chain)
	if len(sources) > 0 {
		reader, ok := instance.(adapter.MediaSession)
		if !ok {
			return reject(target, pb.ErrorCode_ERROR_CODE_UNSUPPORTED, "adapter instance does not support media tokens")
		}
		deadline := in.Deadline
		if parent, ok := ctx.Deadline(); ok && (deadline.IsZero() || parent.Before(deadline)) {
			deadline = parent
		}
		if deadline.IsZero() {
			// Core 自动回复直接调用 Send；没有 RPC 期限时使用契约的普通请求预算。
			deadline = time.Now().Add(s.requestTimeout)
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, deadline)
		defer cancel()
		for _, source := range sources {
			prepared, release, err := s.lease(ctx, caller, reader.MediaSessionKey(), target, source, deadline)
			if err != nil {
				return rejectError(target, err)
			}
			defer release()
			source.Source = prepared.Source
			source.Mime = prepared.Mime
		}
		if err := ctx.Err(); err != nil {
			return rejectError(target, err)
		}
	}
	// 8. 平台操作前预留容量，等待失败时保持零平台副作用。
	var complete func(*pb.SendResult, *pb.ErrorInfo)
	if s.hooks != nil {
		var release func()
		var err error
		complete, release, err = s.hooks.PrepareAfterSend(ctx, caller, pb.SendOperation_SEND_OPERATION_SEND, "", target, request.Chain, invocation)
		if err != nil {
			return rejectError(target, err)
		}
		defer release()
	}
	if err := ctx.Err(); err != nil {
		return rejectError(target, err)
	}
	// 9. 唯一一次发送完成后，观察任务独立处理，回执立即交给调用者。
	result, sendError := instance.Send(ctx, target, request.Chain)
	if complete != nil {
		complete(result, sendError)
	}
	if s.observe != nil {
		s.observe(in.TraceID, "send", target, request.Chain, result, sendError)
	}
	if err := s.rememberReply(caller, session, target, request.Chain, result, decoration); err != nil {
		slog.Error("save reply decoration failed after platform send", "error", err)
		if sendError == nil {
			sendError = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "platform send completed; save reply decoration: " + err.Error()}
		}
	}
	return response(result, sendError)
}

func (s *Service) checkChain(chain *pb.MessageChain, target *pb.Target) *pb.ErrorInfo {
	if safe := s.normalizeMarkdown(chain, target); safe != nil {
		return safe
	}
	if safe := s.completeReferences(chain, target); safe != nil {
		return safe
	}
	if err := validate.Chain(chain); err != nil {
		var validation *validate.Error
		if errors.As(err, &validation) {
			return &pb.ErrorInfo{Code: validation.Code, SafeMessage: validation.Message}
		}
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: err.Error()}
	}
	return nil
}

func (s *Service) completeReferences(chain *pb.MessageChain, target *pb.Target) *pb.ErrorInfo {
	for _, segment := range chain.GetSegments() {
		if reference := segment.GetReference(); reference.GetTarget() != nil {
			completed, completionError := s.adapters.Complete(reference.Target, target)
			if completionError != nil {
				return completionError
			}
			reference.Target = completed
		}
		for _, item := range segment.GetForward().GetItems() {
			if completionError := s.completeReferences(item.GetChain(), target); completionError != nil {
				return completionError
			}
		}
	}
	return nil
}

// mediaSources 按消息顺序收集所有需租约的来源，HTTPS 与内联数据原样交给 Adapter。
func mediaSources(chain *pb.MessageChain) []*pb.MediaSource {
	var sources []*pb.MediaSource
	for _, segment := range chain.GetSegments() {
		var source *pb.MediaSource
		switch value := segment.GetData().(type) {
		case *pb.Segment_Image:
			source = value.Image.GetSource()
		case *pb.Segment_Video:
			source = value.Video.GetSource()
		case *pb.Segment_Voice:
			source = value.Voice.GetSource()
		case *pb.Segment_File:
			source = value.File.GetSource()
		}
		switch source.GetSource().(type) {
		case *pb.MediaSource_MediaRef, *pb.MediaSource_IncomingRef:
			sources = append(sources, source)
		}
		for _, item := range segment.GetForward().GetItems() {
			sources = append(sources, mediaSources(item.GetChain())...)
		}
	}
	return sources
}

func rejectError(target *pb.Target, err error) wire.Response {
	var call *wire.CallError
	if errors.As(err, &call) {
		return response(&pb.SendResult{Status: pb.SendStatus_SEND_STATUS_REJECTED, Target: target}, call.ErrorInfo)
	}
	code := pb.ErrorCode_ERROR_CODE_INTERNAL
	if errors.Is(err, context.Canceled) {
		code = pb.ErrorCode_ERROR_CODE_CANCELLED
	} else if errors.Is(err, context.DeadlineExceeded) {
		code = pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED
	}
	return reject(target, code, err.Error())
}

func reject(target *pb.Target, code pb.ErrorCode, message string) wire.Response {
	return response(&pb.SendResult{Status: pb.SendStatus_SEND_STATUS_REJECTED, Target: target}, &pb.ErrorInfo{Code: code, SafeMessage: message})
}

func response(result *pb.SendResult, sendError *pb.ErrorInfo) wire.Response {
	out := wire.Response{Err: sendError}
	if result == nil {
		return out
	}
	encoded, err := proto.Marshal(result)
	if err != nil {
		return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "encode SendResult: " + err.Error()}}
	}
	out.Payload = wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(result), SchemaVersion: schemaVersion, Bytes: encoded}
	return out
}

// SetObserver 在开始接纳请求前绑定平台回执观察器。
func (s *Service) SetObserver(observe func([]byte, string, *pb.Target, *pb.MessageChain, *pb.SendResult, *pb.ErrorInfo)) {
	s.observe = observe
}

// requestContext 统一媒体与普通 Send/Edit 的请求预算，显式期限取最早值。
func (s *Service) requestContext(ctx context.Context, in *wire.Inbound) (context.Context, context.CancelFunc) {
	deadline := in.Deadline
	if parent, ok := ctx.Deadline(); ok && (deadline.IsZero() || parent.Before(deadline)) {
		deadline = parent
	}
	if deadline.IsZero() {
		deadline = time.Now().Add(s.requestTimeout)
	}
	return context.WithDeadline(ctx, deadline)
}
