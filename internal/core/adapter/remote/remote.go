// Package remote 将独立 Adapter 进程中的账号包装为 Core 的 adapter.Instance。
//
// [Process] 用 Bind 更换 Session、恢复已启用实例，用 Unbind 撤销运行登记。
// 启停串行执行，入站交付独立处理；平台请求期间释放状态锁，允许同步回调 Core。
package remote

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

const (
	schemaVersion = 1
	eventIDBytes  = 16
)

// Options 绑定一个安装包的进程身份、状态库与事件接纳入口。
type Options struct {
	AdapterID        string // 平台类型
	PluginID         string // Adapter 包的安装 ID，用于核对调用者与写游标
	Store            *state.Store
	Registry         *adapter.Registry
	Sink             adapter.Sink
	RegisterIncoming func(instanceID, fileID string, size *uint64, mime *string) string // 装配处必须提供媒体 Store 的入站登记入口
	StartCommandSync func(adapter.Instance)                                             // 登记成功后在状态锁外通知。
	StopCommandSync  func(adapter.Instance)                                             // 停止或 Session 撤销后在状态锁外通知。
}

// Process 并发安全地管理一个 Adapter 包进程及其账号实例。
type Process struct {
	opt          Options
	operations   sync.Mutex // 串行化 Bind、Unbind 与实例启停
	mu           sync.Mutex // 保护 session、instances 与实例状态；RPC 期间释放
	session      *host.Instance
	caller       router.Caller // 与当前 Session 的 Mux.For 使用相同身份，由 mu 保护
	instances    map[string]*instance
	declarations declarations // 当前绑定 Session 的声明缓存，由 mu 保护
}

// instance 保存远端账号身份与启停状态；可变字段统一由 Process.mu 保护。
type instance struct {
	syncMu       sync.Mutex // 串行化本次运行的菜单启停通知；不与 Process.mu 同时等待。
	needsStop    bool       // SDK 中仍有待回收的平台实例
	process      *Process
	id           string
	config       []byte // 只保留在内存中，恢复时重新下发
	enabled      bool   // 管理员要求运行，Bind 时恢复
	running      bool   // 已登记到 Registry，绑定当前 Session
	account      string
	displayName  *string
	username     string
	capabilities []*pb.Capability
	dropped      uint64        // 事件队列满时累计丢弃数
	session      *wire.Session // 本次运行所属 Session，旧 Registry 引用不得切到新 Session
	sessionKey   string        // 实例创建时绑定，Session 更换后旧实例仍保留旧值
}

// New 创建尚未绑定 Session 的 Adapter 包进程管理器。
func New(opt Options) *Process { return &Process{opt: opt, instances: make(map[string]*instance)} }

// Bind 绑定已激活的新 Session，并按保存的配置恢复全部已启用实例。
// caller 必须与为该 Session 装配 Mux.For 时使用的身份相同。
func (p *Process) Bind(ctx context.Context, in *host.Instance, caller router.Caller) error {
	p.operations.Lock()
	defer p.operations.Unlock()
	declarations, err := compileDeclarations(in.Adapter)
	if err != nil {
		return err
	}
	p.mu.Lock()
	previous := p.session
	p.mu.Unlock()
	if previous != nil && previous != in {
		p.unbind()
	}
	p.mu.Lock()
	p.session = in
	p.caller = caller
	p.declarations = declarations
	var enabled []*instance
	for _, entry := range p.instances {
		if entry.enabled {
			enabled = append(enabled, entry)
		}
	}
	p.mu.Unlock()
	if previous != in {
		go p.watchSession(in)
	}
	var errs []error
	for _, entry := range enabled {
		errs = append(errs, p.start(ctx, entry.id, entry.config))
	}
	return errors.Join(errs...)
}

// Unbind 在进程退出后撤销运行登记，保持启用状态供下次 Bind 恢复。
func (p *Process) Unbind() {
	p.operations.Lock()
	defer p.operations.Unlock()
	p.unbind()
}

// watchSession 只撤销自己仍持有的绑定；迟到的旧 Session 退出保留新实例。
func (p *Process) watchSession(in *host.Instance) {
	<-in.Session.Done()
	p.operations.Lock()
	defer p.operations.Unlock()
	p.mu.Lock()
	current := p.session == in
	p.mu.Unlock()
	if current {
		p.unbind()
	}
}

// unbind 由持有 operations 的调用方撤销运行状态，保持启用配置。
func (p *Process) unbind() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.session = nil
	p.declarations = declarations{}
	for _, entry := range p.instances {
		if entry.running {
			if err := p.offline(context.Background(), entry, pb.BotOfflineReason_BOT_OFFLINE_REASON_ADAPTER_EXITED, nil); err != nil {
				slog.Error("deliver adapter offline event failed", "plugin_id", p.opt.PluginID, "instance_id", entry.id, "error", err)
			}
		}
	}
}

// StartInstance 启动实例并登记账号归属；成功后交付 bot.online。
func (p *Process) StartInstance(ctx context.Context, instanceID string, configJSON []byte) error {
	p.operations.Lock()
	defer p.operations.Unlock()
	return p.start(ctx, instanceID, configJSON)
}

func (p *Process) start(ctx context.Context, id string, config []byte) error {
	// 1. 核对本地运行状态并读取 Core 持久化的游标。
	p.mu.Lock()
	bound := p.session
	if entry := p.instances[id]; entry != nil && entry.running {
		p.mu.Unlock()
		return failure(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "adapter instance already running")
	}
	previous := p.instances[id]
	cleanup := previous != nil && previous.needsStop && bound != nil && previous.session == bound.Session
	p.mu.Unlock()
	if bound == nil {
		return failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter session is unavailable")
	}
	if cleanup {
		if err := call(ctx, bound.Session, wire.ServiceAdapterInstanceStop, &pb.AdapterInstanceStopRequest{InstanceId: id}, &pb.AdapterInstanceStopResponse{}); err != nil {
			return err
		}
		p.mu.Lock()
		previous.needsStop = false
		p.mu.Unlock()
	}
	cursor, err := p.opt.Store.AdapterCursor(ctx, id)
	if err != nil {
		return err
	}
	response := &pb.AdapterInstanceStartResponse{}
	err = call(ctx, bound.Session, wire.ServiceAdapterInstanceStart, &pb.AdapterInstanceStartRequest{InstanceId: id, ConfigJson: string(config), Cursor: cursor}, response)
	if err != nil {
		return err
	}
	// 2. 账号唯一性由 Core 核对，冲突时停止刚启动的实例。
	p.mu.Lock()
	var conflict bool
	for _, entry := range p.instances {
		if entry.running && entry.account == response.AccountId {
			conflict = true
		}
	}
	p.mu.Unlock()
	if response.AccountId == "" || conflict {
		stopErr := call(ctx, bound.Session, wire.ServiceAdapterInstanceStop, &pb.AdapterInstanceStopRequest{InstanceId: id}, &pb.AdapterInstanceStopResponse{})
		code := pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		message := "adapter returned an empty account_id"
		if conflict {
			code = pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS
			message = "adapter account already running"
		}
		return errors.Join(failure(code, message), stopErr)
	}
	// 3. 运行登记与在线事件在状态锁内完成，服务只接纳已登记实例。
	p.mu.Lock()
	entry := &instance{needsStop: true, process: p, id: id, config: append([]byte(nil), config...), enabled: true, running: true, account: response.AccountId, displayName: response.AccountDisplayName, username: response.GetAccountUsername(), capabilities: response.Capabilities, session: bound.Session}
	entry.sessionKey = p.caller.SessionKey
	var registered adapter.Instance = entry
	for _, capability := range entry.capabilities {
		if capability.GetName() == mediaReadCapability && capability.GetVersion() == 1 && capability.GetSupported() {
			registered = &mediaInstance{instance: entry}
			break
		}
	}
	if err := p.opt.Registry.Add(registered); err != nil {
		p.mu.Unlock()
		stopErr := call(ctx, bound.Session, wire.ServiceAdapterInstanceStop, &pb.AdapterInstanceStopRequest{InstanceId: id}, &pb.AdapterInstanceStopResponse{})
		return errors.Join(err, stopErr)
	}
	p.instances[id] = entry
	ev := p.event(entry, "bot.online")
	ev.EventPayload = &pb.EventPayload{Kind: &pb.EventPayload_BotOnline{BotOnline: &pb.BotOnlinePayload{AdapterId: p.opt.AdapterID, AdapterInstanceId: id, AccountId: entry.account, AccountDisplayName: entry.displayName}}}
	err = p.deliver(ctx, entry, ev)
	p.mu.Unlock()
	p.startCommandSync(entry)
	if errors.Is(err, adapter.ErrQueueFull) {
		return nil
	}
	return err
}

// StopInstance 停用单个实例；实例保留为未启用，下次 Bind 跳过它。
func (p *Process) StopInstance(ctx context.Context, instanceID string) error {
	p.operations.Lock()
	defer p.operations.Unlock()
	p.mu.Lock()
	entry, ok := p.instances[instanceID]
	bound := p.session
	cleanup := ok && entry.needsStop && bound != nil && entry.session == bound.Session
	p.mu.Unlock()
	if !ok {
		return failure(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "adapter instance not found")
	}
	var stopErr error
	if cleanup {
		stopErr = call(ctx, bound.Session, wire.ServiceAdapterInstanceStop, &pb.AdapterInstanceStopRequest{InstanceId: instanceID}, &pb.AdapterInstanceStopResponse{})
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	entry.enabled = false
	if cleanup && stopErr == nil {
		entry.needsStop = false
	}
	if entry.running {
		return errors.Join(stopErr, p.offline(ctx, entry, pb.BotOfflineReason_BOT_OFFLINE_REASON_STOPPED, nil))
	}
	return stopErr
}

func (p *Process) offline(ctx context.Context, entry *instance, reason pb.BotOfflineReason, safe *pb.ErrorInfo) error {
	entry.running = false
	p.opt.Registry.Remove(entry.id)
	if p.opt.StopCommandSync != nil {
		// 回调可重新读取实例状态，必须释放 Process.mu；停止只携带本次运行身份。
		p.mu.Unlock()
		entry.syncMu.Lock()
		p.opt.StopCommandSync(entry)
		entry.syncMu.Unlock()
		p.mu.Lock()
	}
	ev := p.event(entry, "bot.offline")
	ev.EventPayload = &pb.EventPayload{Kind: &pb.EventPayload_BotOffline{BotOffline: &pb.BotOfflinePayload{AdapterId: p.opt.AdapterID, AdapterInstanceId: entry.id, AccountId: entry.account, Reason: reason, SafeError: safe}}}
	err := p.deliver(ctx, entry, ev)
	if errors.Is(err, adapter.ErrQueueFull) {
		return nil
	}
	return err
}

func (p *Process) event(entry *instance, eventType string) *pb.Event {
	id := make([]byte, eventIDBytes)
	rand.Read(id)
	return &pb.Event{EventId: hex.EncodeToString(id), EventType: eventType, SourceKind: pb.EventSourceKind_EVENT_SOURCE_KIND_ADAPTER, OccurredAtUnixMs: time.Now().UnixMilli(), AdapterId: proto.String(p.opt.AdapterID), AdapterInstanceId: proto.String(entry.id), AccountId: proto.String(entry.account)}
}

// deliver 在调用方持有状态锁时投递，队列满只记丢弃；其他错误保留给调用方。
func (p *Process) deliver(ctx context.Context, entry *instance, ev *pb.Event) error {
	p.registerIncoming(entry.id, ev)
	err := p.opt.Sink.Deliver(ctx, ev)
	if errors.Is(err, adapter.ErrQueueFull) {
		entry.dropped++
		slog.Warn("adapter event dropped", "plugin_id", p.opt.PluginID, "instance_id", entry.id, "dropped", entry.dropped)
		return adapter.ErrQueueFull
	}
	return err
}

func (entry *instance) AdapterID() string  { return entry.process.opt.AdapterID }
func (entry *instance) InstanceID() string { return entry.id }
func (entry *instance) AccountID() string  { return entry.account }

// AccountName 返回启动响应中的 account_username；平台没有用户名时返回空串。
func (entry *instance) AccountName() string {
	return entry.username
}
func (entry *instance) Capabilities() []*pb.Capability { return entry.capabilities }

// Start 沿用保存的配置重启实例，事件统一经 Options.Sink 交付。
func (entry *instance) Start(ctx context.Context, _ adapter.Sink) error {
	return entry.process.StartInstance(ctx, entry.id, entry.config)
}
func (entry *instance) Stop(ctx context.Context) error {
	return entry.process.StopInstance(ctx, entry.id)
}
func (entry *instance) Send(ctx context.Context, target *pb.Target, chain *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo) {
	p := entry.process
	p.mu.Lock()
	running := entry.running
	p.mu.Unlock()
	if !running {
		return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "adapter instance is unavailable"}
	}
	request, err := payload(&pb.AdapterSendRequest{InstanceId: entry.id, Target: target, Chain: chain})
	if err != nil {
		return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}
	}
	response, err := entry.session.Call(ctx, wire.Request{Target: wire.ServiceAdapterSend, Payload: request})
	var result *pb.SendResult
	if response.Payload.SchemaID != "" {
		result = &pb.SendResult{}
		if err := decode(response.Payload, result); err != nil {
			return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}
		}
	}
	if response.Err != nil {
		return result, response.Err
	}
	if err != nil {
		return result, &pb.ErrorInfo{Code: wire.CodeOf(err), SafeMessage: err.Error()}
	}
	return result, nil
}

func failure(code pb.ErrorCode, message string) error {
	return &wire.CallError{ErrorInfo: &pb.ErrorInfo{Code: code, SafeMessage: message}}
}
func payload(message proto.Message) (wire.Payload, error) {
	encoded, err := proto.Marshal(message)
	return wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(message), SchemaVersion: schemaVersion, Bytes: encoded}, err
}
func decode(in wire.Payload, message proto.Message) error {
	if in.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.SchemaID != validate.SchemaID(message) || in.SchemaVersion != schemaVersion {
		return fmt.Errorf("expected %s protobuf schema version %d", message.ProtoReflect().Descriptor().Name(), schemaVersion)
	}
	return proto.Unmarshal(in.Bytes, message)
}
func call(ctx context.Context, session *wire.Session, target string, request, response proto.Message) error {
	encoded, err := payload(request)
	if err != nil {
		return err
	}
	out, err := session.Call(ctx, wire.Request{Target: target, Payload: encoded})
	if out.Payload.SchemaID != "" {
		if decodeErr := decode(out.Payload, response); decodeErr != nil {
			return decodeErr
		}
	}
	return err
}
