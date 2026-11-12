package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/proto"
)

const adapterMediaReadCapability = "media.read"

// adapterRuntime 按实例串行化启停；全局状态锁只保护登记，不跨平台操作持有。
// Session 生命周期沿用 activation；每个实例 Core 客户端独立关闭交付入口。
type adapterRuntime struct {
	activation *activation // Session.Start 前发布，提供出站停止门禁
	Base
	adapter   adapter.Adapter
	mu        sync.Mutex // 保护实例登记、状态与配置
	stopping  bool       // 与停止快照一起在 mu 下发布，阻止已准入但尚未执行的 Start
	instances map[string]*adapterEntry
	session   *wire.Session // Session.Start 前赋值，此后保持不变
}

// adapterEntry 保留实例状态与脱敏配置，生命周期操作由本实例锁串行化。
// busy、config、instance 与 core 由 adapterRuntime.mu 保护。
type adapterEntry struct {
	lifecycle sync.Mutex // 只串行化本实例的 Start 与 Stop，不阻塞其他实例
	busy      bool       // 正在启动或停止，由 adapterRuntime.mu 保护
	config    []byte     // 错误脱敏所需配置，由 adapterRuntime.mu 保护
	instance  adapter.Instance
	core      *adapterCore
}

type adapterCore struct {
	runtime     *adapterRuntime
	instanceID  string
	stopped     atomic.Bool // 停止或报告失败后锁存，阻止新的交付
	commandSync sync.Mutex  // 覆盖本次运行的实际菜单回调，RPC 到期后仍持有。
}

// RunAdapter 作为 Adapter 进程的 main 入口，复用插件的锁、握手、激活与停止状态机。
func RunAdapter(a adapter.Adapter) error {
	setupLogging()
	// 1. 先取得进程锁，再执行 Adapter 描述与启动流程。
	lock, err := acquireRuntimeLock()
	if err != nil {
		return err
	}
	defer lock.Unlock()
	l, err := readLaunch()
	if err != nil {
		return err
	}
	conn, err := dial(l)
	if err != nil {
		return err
	}
	defer conn.Close()
	ack, err := clientHandshake(conn, l)
	if err != nil {
		return err
	}
	setLogLevel(ack.GetBootstrap().GetLogLevel())
	if err := prepareAdapter(conn, l, ack, a); err != nil {
		return err
	}
	// 2. 复用激活与停止回调，仅将业务处理器替换成实例服务。
	runtime := &adapterRuntime{adapter: a, instances: make(map[string]*adapterEntry)}
	active := &activation{l: l, ack: ack, root: runtime, readyCh: make(chan struct{}), failErr: make(chan error, 1), finished: make(chan struct{})}
	active.lifetime, active.stopLifetime = context.WithCancel(context.Background())
	limits := ack.GetBootstrap().GetSessionLimits()
	active.sess = wire.New(conn, wire.Options{Local: pb.InitiatorSide_INITIATOR_SIDE_PLUGIN, SessionID: ack.GetSessionId(), ProtocolMinor: ack.GetChosenMinor(), SourcePluginID: l.pluginID,
		Handler: runtime.handle, Admit: active.admit, Control: active.control, MaxInflight: int(limits.GetMaxInflight()), MaxBytes: int64(limits.GetMaxPayloadBytes()), DefaultDeadline: time.Duration(limits.GetRequestTimeoutMs()) * time.Millisecond})
	runtime.activation = active
	runtime.session = active.sess
	activeSession.Store(active)
	active.sess.Start() // 回调使用的 Session 指针先发布，再启动读循环。
	if err := active.waitReady(); err != nil {
		return err
	}
	return serve(active.sess, active)
}

func prepareAdapter(conn *wire.Conn, l launch, ack *pb.HelloAck, a adapter.Adapter) error {
	boot := ack.GetBootstrap()
	limits := boot.GetSessionLimits()
	// 1. 在信任边界检查 Core 提供的预算，Adapter 没有包级配置或用户 OnStart。
	if boot.GetPhaseLimits().GetOnStartMs() == 0 || boot.GetPreparationDeadlineUnixMs() == 0 || limits.GetMaxInflight() == 0 || limits.GetMaxPayloadBytes() == 0 || limits.GetRequestTimeoutMs() == 0 {
		return sessionFailure(conn, ack.SessionId, ack.ChosenMinor, l.pluginID, pb.SessionPhase_SESSION_PHASE_CONFIG, pb.FrameKind_FRAME_KIND_HELLO_ACK, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "Bootstrap is missing valid phase or session limits", nil)
	}
	stop := withWatchdog(conn, time.Until(time.UnixMilli(boot.GetPreparationDeadlineUnixMs())))
	defer stop()
	descriptor, err := convert.DescriptorToProto((*convert.Descriptor)(a.Descriptor()))
	if err != nil {
		return err
	}
	reg := &pb.Register{Adapter: descriptor, CandidateConfigJson: boot.GetCandidateConfigJson(), CandidateConfigVersion: boot.GetTargetConfigVersion(), Migration: &pb.MigrationState{Mode: pb.MigrationMode_MIGRATION_MODE_NONE}}
	frame, err := wire.ControlEnvelope(ack.SessionId, ack.ChosenMinor, l.pluginID, pb.FrameKind_FRAME_KIND_REGISTER, reg)
	if err != nil {
		return err
	}
	if err := conn.WriteFrame(frame); err != nil {
		if errors.Is(err, wire.ErrFrameTooLarge) {
			return sessionFailure(conn, ack.SessionId, ack.ChosenMinor, l.pluginID, pb.SessionPhase_SESSION_PHASE_REGISTER, pb.FrameKind_FRAME_KIND_REGISTER, pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "registration exceeds 4 MiB", err)
		}
		return err
	}
	// 2. 读取 Core 的准备结果：PREPARED 继续，SessionError 转为错误返回。
	frame, err = conn.ReadFrame()
	if err != nil {
		return err
	}
	if err := wire.CheckPeer(frame, ack.SessionId, ack.ChosenMinor, ""); err != nil {
		return err
	}
	result, err := validate.DecodeControl(frame)
	if err != nil {
		return err
	}
	switch result.(type) {
	case *pb.Prepared:
		return nil
	case *pb.SessionError:
		return &Error{Code: errcode.Code(validate.PublicCode(frame.GetError().GetCode())), Message: frame.GetError().GetSafeMessage()}
	default:
		return fmt.Errorf("unexpected %s during adapter preparation", frame.GetKind())
	}
}

// OnStop 依次停止全部实例，在 activation 提供的停止预算内回收平台资源。
func (r *adapterRuntime) OnStop(ctx context.Context) error {
	r.mu.Lock()
	r.stopping = true
	entries := make([]*adapterEntry, 0, len(r.instances))
	for _, entry := range r.instances {
		if entry.core != nil {
			entry.core.stopped.Store(true)
		}
		entries = append(entries, entry)
	}
	r.mu.Unlock()
	var errs []error
	for _, entry := range entries {
		if err := r.stopEntry(ctx, entry); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// stopEntry 先关闭交付入口，再在本实例的生命周期锁内停止平台接收。
func (r *adapterRuntime) stopEntry(ctx context.Context, entry *adapterEntry) error {
	entry.lifecycle.Lock()
	defer entry.lifecycle.Unlock()
	r.mu.Lock()
	instance := entry.instance
	if instance == nil {
		r.mu.Unlock()
		return nil
	}
	entry.busy = true
	entry.core.stopped.Store(true)
	config := entry.config
	r.mu.Unlock()
	err := instance.Stop(ctx)
	r.mu.Lock()
	entry.busy = false
	if err == nil {
		entry.instance = nil
		entry.config = nil
	}
	r.mu.Unlock()
	if err != nil {
		return errors.New(adapterSafeError(err, config))
	}
	return nil
}

func adapterPayload(message proto.Message) (wire.Payload, error) {
	encoded, err := proto.Marshal(message)
	return wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(message), SchemaVersion: invocationSchemaVersion, Bytes: encoded}, err
}

func adapterResponse(message proto.Message, safe *pb.ErrorInfo) wire.Response {
	payload, err := adapterPayload(message)
	if err != nil {
		return dispatchError(pb.ErrorCode_ERROR_CODE_INTERNAL, err.Error())
	}
	return wire.Response{Payload: payload, Err: safe}
}

func (r *adapterRuntime) handle(ctx context.Context, in *wire.Inbound) wire.Response {
	switch in.Target {
	case wire.ServiceAdapterInstanceStart:
		request := &pb.AdapterInstanceStartRequest{}
		if err := decodeInvocation(in.Payload, request); err != nil {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
		}
		r.mu.Lock()
		if r.stopping {
			r.mu.Unlock()
			return dispatchError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter runtime is stopping")
		}
		entry := r.instances[request.InstanceId]
		if entry == nil {
			entry = &adapterEntry{}
			r.instances[request.InstanceId] = entry
		}
		r.mu.Unlock()
		entry.lifecycle.Lock()
		defer entry.lifecycle.Unlock()
		r.mu.Lock()
		if r.stopping {
			r.mu.Unlock()
			return dispatchError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter runtime is stopping")
		}
		if entry.instance != nil {
			r.mu.Unlock()
			return dispatchError(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "adapter instance already started")
		}
		core := &adapterCore{runtime: r, instanceID: request.InstanceId}
		entry.busy = true
		entry.core = core
		entry.config = []byte(request.ConfigJson)
		r.mu.Unlock()
		instance, err := r.adapter.Start(ctx, adapter.StartRequest{InstanceID: request.InstanceId, ConfigJSON: []byte(request.ConfigJson), Cursor: request.Cursor}, core)
		r.mu.Lock()
		entry.busy = false
		if err != nil {
			core.stopped.Store(true)
			entry.config = nil
			r.mu.Unlock()
			code := pb.ErrorCode_ERROR_CODE_INTERNAL
			var public *Error
			if errors.As(err, &public) {
				code = pb.ErrorCode(public.Code)
			}
			return dispatchError(code, adapterSafeError(err, []byte(request.ConfigJson)))
		}
		entry.instance = instance
		r.mu.Unlock()
		account := instance.Account()
		// 读取能力由实例接口决定，避免平台自报能力与实际 RPC 入口不一致。
		var capabilities []*pb.Capability
		for _, capability := range convert.CapabilitiesToProto(instance.Capabilities()) {
			if capability.GetName() != adapterMediaReadCapability {
				capabilities = append(capabilities, capability)
			}
		}
		if _, ok := instance.(adapter.MediaSource); ok {
			capabilities = append(capabilities, &pb.Capability{Name: adapterMediaReadCapability, Version: 1, Supported: true})
		}
		response := &pb.AdapterInstanceStartResponse{AccountId: account.ID, Capabilities: capabilities}
		if account.DisplayName != "" {
			response.AccountDisplayName = &account.DisplayName
		}
		if account.Username != "" {
			response.AccountUsername = &account.Username
		}
		return adapterResponse(response, nil)
	case wire.ServiceAdapterInstanceStop:
		request := &pb.AdapterInstanceStopRequest{}
		if err := decodeInvocation(in.Payload, request); err != nil {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
		}
		r.mu.Lock()
		entry, ok := r.instances[request.InstanceId]
		r.mu.Unlock()
		if !ok {
			return dispatchError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "adapter instance not found")
		}
		if err := r.stopEntry(ctx, entry); err != nil {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INTERNAL, err.Error())
		}
		return adapterResponse(&pb.AdapterInstanceStopResponse{}, nil)
	case wire.ServiceAdapterMediaRead:
		request := &pb.AdapterMediaReadRequest{}
		if err := decodeInvocation(in.Payload, request); err != nil {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
		}
		if request.Length == 0 || request.Length > wire.MaxMediaChunk {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid media read length")
		}
		r.mu.Lock()
		entry, ok := r.instances[request.InstanceId]
		if !ok {
			r.mu.Unlock()
			return dispatchError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "adapter instance not found")
		}
		if entry.busy || entry.instance == nil || entry.core.stopped.Load() {
			r.mu.Unlock()
			return dispatchError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is not running")
		}
		source, supported := entry.instance.(adapter.MediaSource)
		config := entry.config
		r.mu.Unlock()
		if !supported {
			return dispatchError(pb.ErrorCode_ERROR_CODE_UNSUPPORTED, "adapter instance does not support media reads")
		}
		// 平台请求在锁外执行，停止和其他实例的操作可继续推进。
		data, next, eof, err := source.ReadPlatformFile(ctx, request.FileId, request.Offset, request.Length)
		if err != nil {
			code := pb.ErrorCode_ERROR_CODE_UNAVAILABLE
			message := err.Error()
			var public *Error
			if errors.As(err, &public) {
				code = validate.PublicCode(pb.ErrorCode(public.Code))
				message = public.Message
			}
			return dispatchError(code, adapterSafeError(errors.New(message), config))
		}
		return adapterResponse(&pb.AdapterMediaReadResponse{Data: data, NextOffset: next, Eof: eof}, nil)
	case wire.ServiceAdapterCommandsSync:
		request := &pb.AdapterCommandsSyncRequest{}
		if err := decodeInvocation(in.Payload, request); err != nil {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
		}
		r.mu.Lock()
		entry, ok := r.instances[request.InstanceId]
		if !ok {
			r.mu.Unlock()
			return dispatchError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "adapter instance not found")
		}
		if entry.busy || entry.instance == nil || entry.core.stopped.Load() {
			r.mu.Unlock()
			return dispatchError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is not running")
		}
		core := entry.core
		r.mu.Unlock()
		// RPC 取消后平台操作可能仍在执行，等待实际回调返回再处理下一份菜单。
		core.commandSync.Lock()
		defer core.commandSync.Unlock()
		if err := ctx.Err(); err != nil {
			code := pb.ErrorCode_ERROR_CODE_CANCELLED
			if errors.Is(err, context.DeadlineExceeded) {
				code = pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED
			}
			return dispatchError(code, err.Error())
		}
		r.mu.Lock()
		if entry.core != core || entry.busy || entry.instance == nil || core.stopped.Load() {
			r.mu.Unlock()
			return dispatchError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is not running")
		}
		synchronizer, supported := entry.instance.(adapter.CommandSynchronizer)
		r.mu.Unlock()
		if !supported {
			return dispatchError(pb.ErrorCode_ERROR_CODE_UNSUPPORTED, "adapter instance does not support command synchronization")
		}
		// 与其他固定服务共用请求 context，平台调用在状态锁外执行。
		safe := synchronizer.SyncCommands(ctx, convert.CommandMenuFromProto(request.Commands))
		return adapterResponse(&pb.AdapterCommandsSyncResponse{}, convert.ErrorToProto(safe))
	case wire.ServiceAdapterEdit:
		request := &pb.AdapterEditRequest{}
		if err := decodeInvocation(in.Payload, request); err != nil {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
		}
		r.mu.Lock()
		entry, ok := r.instances[request.InstanceId]
		if !ok {
			r.mu.Unlock()
			return dispatchError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "adapter instance not found")
		}
		if entry.busy || entry.instance == nil || entry.core.stopped.Load() {
			r.mu.Unlock()
			return dispatchError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is not running")
		}
		editor, supported := entry.instance.(adapter.Editor)
		r.mu.Unlock()
		if !supported {
			return dispatchError(pb.ErrorCode_ERROR_CODE_UNSUPPORTED, "adapter instance does not support message editing")
		}
		// 平台请求在锁外执行，停止和其他实例的操作可继续推进。
		result, safe := editor.Edit(ctx, convert.SessionFromProto(request.Target), request.PlatformMessageId, convert.ChainFromProto(request.Chain, request.Target))
		if result == nil {
			return wire.Response{Err: convert.ErrorToProto(safe)}
		}
		encoded, err := convert.SendResultToProto(result, request.Target)
		if err != nil {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
		}
		return adapterResponse(encoded, convert.ErrorToProto(safe))
	case wire.ServiceAdapterSend:
		request := &pb.AdapterSendRequest{}
		if err := decodeInvocation(in.Payload, request); err != nil {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
		}
		r.mu.Lock()
		entry, ok := r.instances[request.InstanceId]
		if !ok {
			r.mu.Unlock()
			return dispatchError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "adapter instance not found")
		}
		if entry.busy || entry.instance == nil || entry.core.stopped.Load() {
			r.mu.Unlock()
			return dispatchError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is not running")
		}
		instance := entry.instance
		r.mu.Unlock()
		result, safe := instance.Send(ctx, convert.SessionFromProto(request.Target), convert.ChainFromProto(request.Chain, request.Target))
		if result == nil {
			return wire.Response{Err: convert.ErrorToProto(safe)}
		}
		encoded, err := convert.SendResultToProto(result, request.Target)
		if err != nil {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
		}
		return adapterResponse(encoded, convert.ErrorToProto(safe))
	default:
		return dispatchError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "adapter service not found")
	}
}

func (c *adapterCore) call(ctx context.Context, target string, request, response proto.Message) error {
	a := c.runtime.activation
	if !a.ready.Load() || a.stopping.Load() {
		return &Error{Code: errcode.FailedPrecondition, Message: "adapter session is not active"}
	}

	payload, err := adapterPayload(request)
	if err != nil {
		return err
	}
	result, err := c.runtime.session.Call(ctx, wire.Request{Target: target, Payload: payload})
	if result.Err != nil {
		return &Error{Code: errcode.Code(validate.PublicCode(result.Err.GetCode())), Message: result.Err.GetSafeMessage()}
	}
	if err != nil {
		return err
	}
	return decodeInvocation(result.Payload, response)
}

func (c *adapterCore) Deliver(ctx context.Context, ev *message.Event, cursor []byte) (bool, error) {
	if c.stopped.Load() {
		return false, &Error{Code: errcode.FailedPrecondition, Message: "adapter instance has stopped"}
	}
	event, err := convert.EventToProto(ev)
	if err != nil {
		return false, err
	}
	response := &pb.AdapterDeliverResponse{}
	err = c.call(ctx, wire.ServiceCoreAdapterDeliver, &pb.AdapterDeliverRequest{InstanceId: c.instanceID, Event: event, Cursor: cursor}, response)
	return response.Accepted, err
}

func (c *adapterCore) SaveCursor(ctx context.Context, cursor []byte) error {
	if c.stopped.Load() {
		return &Error{Code: errcode.FailedPrecondition, Message: "adapter instance has stopped"}
	}
	return c.call(ctx, wire.ServiceCoreAdapterCursor, &pb.AdapterCursorRequest{InstanceId: c.instanceID, Cursor: cursor}, &pb.AdapterCursorResponse{})
}

func (c *adapterCore) ReadMedia(ctx context.Context, token string, offset uint64, length uint32) ([]byte, uint64, bool, error) {
	if c.stopped.Load() {
		return nil, 0, false, &Error{Code: errcode.FailedPrecondition, Message: "adapter instance has stopped"}
	}
	response := &pb.ReadResponse{}
	if err := c.call(ctx, wire.ServiceMediaRead, &pb.ReadRequest{Token: token, Offset: offset, Length: length}, response); err != nil {
		return nil, 0, false, err
	}
	return response.Data, response.NextOffset, response.Eof, nil
}

func (c *adapterCore) ReportFailure(ctx context.Context, safe *errcode.Error) error {
	c.stopped.Store(true)
	return c.call(ctx, wire.ServiceCoreAdapterStatus, &pb.AdapterStatusRequest{InstanceId: c.instanceID, State: pb.AdapterInstanceState_ADAPTER_INSTANCE_STATE_FAILED, SafeError: convert.ErrorToProto(safe)}, &pb.AdapterStatusResponse{})
}

// adapterSafeError 保留错误上下文，移除实例配置全文与其中的字符串值。
func adapterSafeError(err error, config []byte) string {
	safe := err.Error()
	if len(config) > 0 {
		safe = strings.ReplaceAll(safe, string(config), "[redacted]")
	}
	var value any
	if json.Unmarshal(config, &value) == nil {
		var redact func(any)
		redact = func(value any) {
			switch v := value.(type) {
			case string:
				if v != "" {
					safe = strings.ReplaceAll(safe, v, "[redacted]")
				}
			case map[string]any:
				for _, item := range v {
					redact(item)
				}
			case []any:
				for _, item := range v {
					redact(item)
				}
			}
		}
		redact(value)
	}
	return safe
}
