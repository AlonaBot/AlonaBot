// Package broker 将插件的 Service 调用转发到目标插件当前已激活的 Session。
//
// [Broker.Call] 在转发前校验访问规则、编码、schema 和 JSON 语法，并保留每跳请求的关联。
// SDK 负责按业务类型解码。调用与资源预算复用 wire.Session（CONTRACTS §1.2、§3）。
package broker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/metrics"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

const (
	invocationSchemaVersion = 1
)

// Options 提供路由依赖与 tune.Values 中的请求期限、调用链深度上限。
type Options struct {
	Router         *router.Router
	RequestTimeout time.Duration
	MaxDepth       int                                       // CALL_CHAIN_MAX_DEPTH，包含即将调用的目标插件。
	Metrics        *metrics.Counters                         // nil 时关闭原因计数。
	Agent          func(ctx context.Context) (string, error) // 仅 @agent 目标需要，空 ID 表示未安装。
	EnsureActive   func(ctx context.Context, pluginID string, payloadBytes int64) (release func(), err error)
}

// Broker 保存路由入口和在途转发关系；New 创建，Call 接入 Mux.Forward。
// Session 将新分配的目标 call_id 与 Request.Attach 绑定，calls 以同一附件关联源 Session/call_id。
// 业务函数在锁外执行；目标返回或调用结束时删除关联。
type Broker struct {
	router         *router.Router
	requestTimeout time.Duration
	maxDepth       int
	metrics        *metrics.Counters
	agent          func(ctx context.Context) (string, error)
	ensureActive   func(context.Context, string, int64) (func(), error)
	mu             sync.Mutex // 只保护 calls，转发期间不持锁
	calls          map[*router.Invocation]callOrigin
}

type callOrigin struct {
	caller router.Caller
	callID uint64
	target router.Caller
	cancel context.CancelFunc
}

type selfToolCallKey struct{}

// WithSelfToolCall 标记已通过 core.tool.invoke 策略检查的转发，仅在目标为调用方时放行。
func WithSelfToolCall(ctx context.Context) context.Context {
	return context.WithValue(ctx, selfToolCallKey{}, true)
}

// New 按装配处传入的请求期限和调用链深度上限创建 Broker。
func New(opt Options) *Broker {
	return &Broker{router: opt.Router, requestTimeout: opt.RequestTimeout, maxDepth: opt.MaxDepth, metrics: opt.Metrics, agent: opt.Agent, ensureActive: opt.EnsureActive, calls: make(map[*router.Invocation]callOrigin)}
}

// Call 依次完成定位、访问与输入准入，再以独立编号调用目标 Session。
func (b *Broker) Call(ctx context.Context, c router.Caller, in *wire.Inbound) wire.Response {
	return b.call(ctx, c, in, nil)
}

// CallSession 只向 target 绑定的 Session 转发；目标已替换时返回 UNAVAILABLE。
// 定位后持有原实例，后续替换不会使请求转发到新 Session。
func (b *Broker) CallSession(ctx context.Context, c, target router.Caller, in *wire.Inbound) wire.Response {
	return b.call(ctx, c, in, &target)
}

// Release 取消以该 Session 为调用方或目标的全部在途转发，保留其他 Session 的调用。
func (b *Broker) Release(c router.Caller) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, origin := range b.calls {
		if origin.caller == c || origin.target == c {
			origin.cancel()
		}
	}
}

func (b *Broker) call(ctx context.Context, c router.Caller, in *wire.Inbound, target *router.Caller) (result wire.Response) {
	// 冷启动也使用本次调用开始时固定的期限。
	deadline := in.Deadline
	if deadline.IsZero() {
		deadline = time.Now().Add(b.requestTimeout)
	}
	ctx, stopDeadline := context.WithDeadline(ctx, deadline)
	defer stopDeadline()
	deadline, _ = ctx.Deadline()
	trace := in.TraceID
	if parent, ok := in.ParentAttach.(*router.Invocation); ok {
		trace = parent.TraceID
	}
	defer func() {
		if result.Err != nil {
			result.Err.TraceId = bytes.Clone(trace)
		}
	}()
	// 1. 目标语法必须先于插件与服务查找检查。
	pluginID, name, ok := strings.Cut(in.Target, "/")
	if !ok || !validate.IsServiceName(name) {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid service target")
	}
	if pluginID == "@agent" {
		var err *pb.ErrorInfo
		pluginID, err = b.agentID(ctx)
		if err != nil {
			return wire.Response{Err: err}
		}
	} else if _, err := validate.PluginID(pluginID); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid service target")
	}
	var plugin router.Plugin
	var destination router.Caller
	var routes []*pb.RouteDescriptor
	var generation string
	fromSnapshot := false
	if target != nil {
		if target.PluginID != pluginID {
			b.metrics.Add(c.PluginID, "service", "offline")
			return failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "target session does not match service target")
		}
		destination = *target
	} else {
		destination, ok = b.router.Caller(pluginID)
	}
	if ok {
		plugin, ok = b.router.PluginFor(destination)
	}
	if ok && plugin.Activated() {
		routes = plugin.Routes()
		generation = destination.GenerationID
	} else if target == nil && b.ensureActive != nil {
		generation, routes, ok = b.router.RouteSnapshot(pluginID)
		fromSnapshot = ok
	}
	if !ok {
		owner := c.PluginID
		if _, known := b.router.Plugin(pluginID); known {
			owner = pluginID
		}
		b.metrics.Add(owner, "service", "offline")
		return failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "target plugin is not active")
	}
	if plugin != nil && !plugin.Activated() && b.ensureActive == nil {
		b.metrics.Add(pluginID, "service", "offline")
		return failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "target plugin is not active")
	}
	var service *pb.RouteDescriptor
	for _, route := range routes {
		if route.GetKind() == pb.RouteKind_ROUTE_KIND_SERVICE && route.GetService().GetServiceName() == name {
			service = route
			break
		}
	}
	if service == nil {
		return failure(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "service not found")
	}
	// 2. Caller 来自认证 Session；Core 与本插件始终满足缺省访问规则。
	spec := service.GetService()
	if !router.ServiceCallable(spec, c.PluginID, pluginID) {
		b.metrics.Add(pluginID, "service", "denied")
		return failure(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "service access denied")
	}
	permitted := len(service.GetRequiredPermissions()) == 0
	if parent, ok := in.ParentAttach.(*router.Invocation); ok {
		for _, required := range service.GetRequiredPermissions() {
			permitted = permitted || slices.Contains(parent.Event.GetActor().GetPermissions(), required)
		}
	}
	if !permitted {
		b.metrics.Add(pluginID, "service", "denied")
		return failure(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "service permission denied")
	}
	payload := in.Payload
	if payload.Encoding != spec.GetEncoding() || payload.SchemaID != service.GetInputSchemaId() || payload.SchemaVersion != service.GetInputSchemaVersion() {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "service input does not match registered encoding or schema")
	}
	// Service 准入按约定复查 UTF-8 JSON；json.Valid 本身接受无效 UTF-8 字符串。
	if payload.Encoding == pb.PayloadEncoding_PAYLOAD_ENCODING_JSON && !validJSON(payload.Bytes) {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "service input is not valid UTF-8 JSON")
	}
	// 3. 父附件由 Core 保留；每次复制祖先链，防止独立并发调用修改同一底层切片。
	invocation := &router.Invocation{ProfileID: "default", PluginID: pluginID, TraceID: in.TraceID, Depth: 2}
	trusted := &pb.InvocationContext{}
	if parent, ok := in.ParentAttach.(*router.Invocation); ok {
		trusted.Actor = parent.Event.GetActor()
		trusted.DefaultTarget = parent.DefaultTarget
		invocation.Chain = slices.Clone(parent.Chain)
		invocation.Scopes = slices.Clone(parent.Scopes)
		invocation.SendForbidden = parent.SendForbidden
		invocation.Event = parent.Event
		invocation.DefaultTarget = parent.DefaultTarget
		invocation.TraceID = parent.TraceID
		invocation.Depth = parent.Depth + 1
		invocation.ProfileID, invocation.Root = parent.ProfileID, parent.Root
	}
	if self, _ := ctx.Value(selfToolCallKey{}).(bool); !self || pluginID != c.PluginID {
		invocation.Chain = append(invocation.Chain, c.PluginID)
	}
	if slices.Contains(invocation.Chain, pluginID) {
		b.metrics.Add(pluginID, "service", "cycle")
		return failure(pb.ErrorCode_ERROR_CODE_CALL_CYCLE, "service call chain contains target plugin")
	}
	if invocation.Depth > b.maxDepth {
		b.metrics.Add(pluginID, "service", "depth")
		return failure(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "service call chain exceeds maximum depth")
	}
	// 准入全部通过后才唤醒；等待不持转发表锁，也不占目标 Session 的业务槽。
	if b.ensureActive != nil && target == nil {
		release, err := b.ensureActive(ctx, pluginID, int64(len(payload.Bytes)))
		if err != nil {
			code := wire.CodeOf(err)
			if code == pb.ErrorCode_ERROR_CODE_UNSPECIFIED {
				code = pb.ErrorCode_ERROR_CODE_UNAVAILABLE
			}
			return failure(code, err.Error())
		}
		defer release()
		destination, ok = b.router.Caller(pluginID)
		if !ok || destination.GenerationID != generation {
			return failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "service generation changed during wake")
		}
		plugin, ok = b.router.PluginFor(destination)
		if !ok || !plugin.Activated() || !slices.ContainsFunc(plugin.Routes(), func(current *pb.RouteDescriptor) bool { return proto.Equal(current, service) }) {
			return failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "service route changed during wake")
		}
		if fromSnapshot {
			activeGeneration, committed, exists := b.router.RouteSnapshot(pluginID)
			if !exists || activeGeneration != generation || !slices.ContainsFunc(committed, func(current *pb.RouteDescriptor) bool { return proto.Equal(current, service) }) {
				return failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "committed service route changed during wake")
			}
		}
		if !router.ServiceCallable(service.GetService(), c.PluginID, pluginID) {
			return failure(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "service access denied after wake")
		}
	}
	if invocation.ProfileID != "default" {
		trusted.ProfileId = invocation.ProfileID
	}
	trusted.PluginChain = invocation.Chain
	trusted.Scopes = invocation.Scopes
	trusted.SendForbidden = invocation.SendForbidden
	request := &pb.ServiceInvocation{RouteId: service.GetRouteId(), Context: trusted, Encoding: payload.Encoding, SchemaId: payload.SchemaID, SchemaVersion: payload.SchemaVersion, Payload: payload.Bytes}
	encoded, err := proto.Marshal(request)
	if err != nil {
		slog.Error("encode service invocation failed", "plugin_id", pluginID, "service_name", name, "error", err)
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "encode service invocation failed")
	}
	// 4. wire.Session 为本跳分配 call_id，并将它与此附件绑定；不回填源请求编号。
	if ctx.Err() != nil {
		b.metrics.Add(pluginID, "service", "cancelled")
		if ctx.Err() == context.DeadlineExceeded {
			return failure(pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, "service deadline expired before dispatch")
		}
		return failure(pb.ErrorCode_ERROR_CODE_CANCELLED, "service call cancelled before dispatch")
	}
	b.mu.Lock()
	// 与 Release 串行复核绑定，关闭已准入请求在撤销后才登记的窗口。
	_, current := b.router.PluginFor(destination)
	if !current || !plugin.Activated() {
		b.mu.Unlock()
		b.metrics.Add(pluginID, "service", "offline")
		return failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "target session is no longer active")
	}
	if c.PluginID != "" {
		if _, current := b.router.PluginFor(c); !current {
			b.mu.Unlock()
			b.metrics.Add(pluginID, "service", "cancelled")
			return failure(pb.ErrorCode_ERROR_CODE_CANCELLED, "caller session is no longer active")
		}
	}
	ctx, cancel := context.WithCancel(ctx)
	b.calls[invocation] = callOrigin{caller: c, callID: in.CallID, target: destination, cancel: cancel}
	b.mu.Unlock()
	defer func() {
		cancel()
		b.mu.Lock()
		delete(b.calls, invocation)
		b.mu.Unlock()
	}()
	defer invocation.Begin()()
	response, err := plugin.Call(ctx, wire.Request{Target: service.GetRouteId(), TraceID: invocation.TraceID, Deadline: deadline, Attach: invocation,
		Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: invocationSchemaVersion, Bytes: encoded}})
	// 5. 保留目标的业务错误与有效载荷；成功结果必须符合声明的输出契约。
	if response.Err != nil || err != nil {
		code := validate.PublicCode(wire.CodeOf(err))
		message := "service call failed"
		if response.Err != nil {
			code = validate.PublicCode(response.Err.GetCode())
			message = response.Err.GetSafeMessage()
		}
		switch code {
		case pb.ErrorCode_ERROR_CODE_CANCELLED, pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, pb.ErrorCode_ERROR_CODE_UNAVAILABLE:
			b.metrics.Add(pluginID, "service", "cancelled")
		case pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED:
			b.metrics.Add(pluginID, "service", "dropped_budget")
		}
		if code == pb.ErrorCode_ERROR_CODE_INTERNAL {
			slog.Error("service call failed", "plugin_id", pluginID, "service_name", name, "trace_id", fmt.Sprintf("%x", trace), "error", err, "remote_error", message)
			message = "service call failed"
		}
		// 重新构造公开错误，避免透传远端未知字段；业务载荷仍随错误保留。
		response.Err = &pb.ErrorInfo{Code: code, SafeMessage: message}
		return response
	}
	out := response.Payload
	if out.Encoding != spec.GetEncoding() || out.SchemaID != service.GetOutputSchemaId() || out.SchemaVersion != service.GetOutputSchemaVersion() || out.Encoding == pb.PayloadEncoding_PAYLOAD_ENCODING_JSON && !validJSON(out.Bytes) {
		slog.Warn("service output does not match registered encoding or schema", "plugin_id", pluginID, "service_name", name)
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "service output does not match registered encoding or schema")
	}
	return response
}

func failure(code pb.ErrorCode, message string) wire.Response {
	return wire.Response{Err: &pb.ErrorInfo{Code: code, SafeMessage: message}}
}

func validJSON(b []byte) bool { return utf8.Valid(b) && json.Valid(b) }
