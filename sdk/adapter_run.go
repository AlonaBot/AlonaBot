package sdk

import (
	"context"
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
	"google.golang.org/protobuf/proto"
)

// adapterRuntime 串行化实例创建与停止；发送在实例锁外运行，避免同步交付造成死锁。
// Session 生命周期沿用 activation；每个实例 Core 客户端独立关闭交付入口。
type adapterRuntime struct {
	Base
	adapter   adapter.Adapter
	mu        sync.Mutex // 保护 instances 与实例启停
	instances map[string]*adapterEntry
	session   *wire.Session // Session.Start 前赋值，此后保持不变
}

type adapterEntry struct {
	instance adapter.Instance
	core     *adapterCore
}

type adapterCore struct {
	runtime    *adapterRuntime
	instanceID string
	stopped    atomic.Bool // 停止或报告失败后锁存，阻止新的交付
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
	runtime.session = active.sess
	activeSession.Store(active)
	active.sess.Start() // 回调使用的 Session 指针先发布，再启动读循环。
	select {
	case <-active.readyCh:
		return serve(active.sess, active)
	case err := <-active.failErr:
		active.stopLifetime()
		return err
	case <-active.sess.Done():
		active.stopLifetime()
		return fmt.Errorf("adapter session closed before activation: %w", active.sess.Err())
	}
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
	reg := &pb.Register{Adapter: a.Descriptor(), CandidateConfigJson: boot.GetCandidateConfigJson(), CandidateConfigVersion: boot.GetTargetConfigVersion(), Migration: &pb.MigrationState{Mode: pb.MigrationMode_MIGRATION_MODE_NONE}}
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
	// 2. 准备结果仍由 Core 原子校验，角色来源保留在安装记录。
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
		return &Error{Code: validate.PublicCode(frame.GetError().GetCode()), Message: frame.GetError().GetSafeMessage()}
	default:
		return fmt.Errorf("unexpected %s during adapter preparation", frame.GetKind())
	}
}

// OnStop 依次停止全部实例，在 activation 提供的停止预算内回收平台资源。
func (r *adapterRuntime) OnStop(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var errs []error
	for id, entry := range r.instances {
		entry.core.stopped.Store(true)
		errs = append(errs, entry.instance.Stop(ctx))
		delete(r.instances, id)
	}
	return errors.Join(errs...)
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
		defer r.mu.Unlock()
		if _, ok := r.instances[request.InstanceId]; ok {
			return dispatchError(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "adapter instance already started")
		}
		core := &adapterCore{runtime: r, instanceID: request.InstanceId}
		instance, err := r.adapter.Start(ctx, adapter.StartRequest{InstanceID: request.InstanceId, ConfigJSON: []byte(request.ConfigJson), Cursor: request.Cursor}, core)
		if err != nil {
			core.stopped.Store(true)
			// 实例配置含凭证，启动错误中的配置原文也随错误一起脱敏。
			safe := err.Error()
			if request.ConfigJson != "" {
				safe = strings.ReplaceAll(safe, request.ConfigJson, "[redacted]")
			}
			code := pb.ErrorCode_ERROR_CODE_INTERNAL
			var public *Error
			if errors.As(err, &public) {
				code = public.Code
			}
			return dispatchError(code, safe)
		}
		r.instances[request.InstanceId] = &adapterEntry{instance: instance, core: core}
		account := instance.Account()
		response := &pb.AdapterInstanceStartResponse{AccountId: account.ID, Capabilities: instance.Capabilities()}
		if account.DisplayName != "" {
			response.AccountDisplayName = &account.DisplayName
		}
		return adapterResponse(response, nil)
	case wire.ServiceAdapterInstanceStop:
		request := &pb.AdapterInstanceStopRequest{}
		if err := decodeInvocation(in.Payload, request); err != nil {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
		}
		r.mu.Lock()
		defer r.mu.Unlock()
		entry, ok := r.instances[request.InstanceId]
		if !ok {
			return dispatchError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "adapter instance not found")
		}
		entry.core.stopped.Store(true)
		if err := entry.instance.Stop(ctx); err != nil {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INTERNAL, err.Error())
		}
		delete(r.instances, request.InstanceId)
		return adapterResponse(&pb.AdapterInstanceStopResponse{}, nil)
	case wire.ServiceAdapterSend:
		request := &pb.AdapterSendRequest{}
		if err := decodeInvocation(in.Payload, request); err != nil {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
		}
		r.mu.Lock()
		entry, ok := r.instances[request.InstanceId]
		r.mu.Unlock()
		if !ok || entry.core.stopped.Load() {
			return dispatchError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "adapter instance not running")
		}
		result, safe := entry.instance.Send(ctx, request.Target, request.Chain)
		if result == nil {
			return wire.Response{Err: safe}
		}
		return adapterResponse(result, safe)
	default:
		return dispatchError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "adapter service not found")
	}
}

func (c *adapterCore) call(ctx context.Context, target string, request, response proto.Message) error {
	payload, err := adapterPayload(request)
	if err != nil {
		return err
	}
	result, err := c.runtime.session.Call(ctx, wire.Request{Target: target, Payload: payload})
	if result.Err != nil {
		return &Error{Code: validate.PublicCode(result.Err.GetCode()), Message: result.Err.GetSafeMessage()}
	}
	if err != nil {
		return err
	}
	return decodeInvocation(result.Payload, response)
}

func (c *adapterCore) Deliver(ctx context.Context, ev *pb.Event, cursor []byte) (bool, error) {
	if c.stopped.Load() {
		return false, &Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "adapter instance has stopped"}
	}
	response := &pb.AdapterDeliverResponse{}
	err := c.call(ctx, wire.ServiceCoreAdapterDeliver, &pb.AdapterDeliverRequest{InstanceId: c.instanceID, Event: ev, Cursor: cursor}, response)
	return response.Accepted, err
}

func (c *adapterCore) SaveCursor(ctx context.Context, cursor []byte) error {
	if c.stopped.Load() {
		return &Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "adapter instance has stopped"}
	}
	return c.call(ctx, wire.ServiceCoreAdapterCursor, &pb.AdapterCursorRequest{InstanceId: c.instanceID, Cursor: cursor}, &pb.AdapterCursorResponse{})
}

func (c *adapterCore) ReportFailure(ctx context.Context, safe *pb.ErrorInfo) error {
	c.stopped.Store(true)
	return c.call(ctx, wire.ServiceCoreAdapterStatus, &pb.AdapterStatusRequest{InstanceId: c.instanceID, State: pb.AdapterInstanceState_ADAPTER_INSTANCE_STATE_FAILED, SafeError: safe}, &pb.AdapterStatusResponse{})
}
