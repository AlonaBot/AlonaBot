// Package broker 将插件的 Service 调用转发到目标插件当前已激活的 Session。
//
// [Broker.Call] 在转发前校验访问规则、编码、schema 和 JSON 语法，并保留每跳请求的关联。
// SDK 负责按业务类型解码。调用与资源预算复用 wire.Session（CONTRACTS §1.2、§3）。
package broker

import (
	"context"
	"encoding/json"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
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
	MaxDepth       int // CALL_CHAIN_MAX_DEPTH，包含即将调用的目标插件。
}

// Broker 保存路由入口和在途转发关系；New 创建，Call 接入 Mux.Forward。
// Session 将新分配的目标 call_id 与 Request.Attach 绑定，calls 以同一附件关联源 Session/call_id。
// 业务函数在锁外执行；目标返回或调用结束时删除关联。
type Broker struct {
	router         *router.Router
	requestTimeout time.Duration
	maxDepth       int
	mu             sync.Mutex // 只保护 calls，转发期间不持锁
	calls          map[*router.Invocation]callOrigin
}

type callOrigin struct {
	caller router.Caller
	callID uint64
}

// New 按装配处传入的请求期限和调用链深度上限创建 Broker。
func New(opt Options) *Broker {
	return &Broker{router: opt.Router, requestTimeout: opt.RequestTimeout, maxDepth: opt.MaxDepth, calls: make(map[*router.Invocation]callOrigin)}
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

func (b *Broker) call(ctx context.Context, c router.Caller, in *wire.Inbound, target *router.Caller) wire.Response {
	// 1. 目标语法必须先于插件与服务查找检查。
	pluginID, name, ok := strings.Cut(in.Target, "/")
	if _, err := validate.PluginID(pluginID); err != nil || !ok || !validate.IsServiceName(name) {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid service target")
	}
	var plugin router.Plugin
	if target != nil {
		if target.PluginID != pluginID {
			return failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "target session does not match service target")
		}
		plugin, ok = b.router.PluginFor(*target)
	} else {
		plugin, ok = b.router.Plugin(pluginID)
	}
	if !ok || !plugin.Activated() {
		return failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "target plugin is not active")
	}
	var service *pb.RouteDescriptor
	for _, route := range plugin.Routes() {
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
		return failure(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "service access denied")
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
	invocation := &router.Invocation{PluginID: pluginID, TraceID: in.TraceID}
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
	}
	invocation.Chain = append(invocation.Chain, c.PluginID)
	if slices.Contains(invocation.Chain, pluginID) {
		return failure(pb.ErrorCode_ERROR_CODE_CALL_CYCLE, "service call chain contains target plugin")
	}
	if len(invocation.Chain)+1 > b.maxDepth {
		return failure(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "service call chain exceeds maximum depth")
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
	deadline := time.Now().Add(b.requestTimeout)
	if !in.Deadline.IsZero() && in.Deadline.Before(deadline) {
		deadline = in.Deadline
	}
	if parentDeadline, ok := ctx.Deadline(); ok && parentDeadline.Before(deadline) {
		deadline = parentDeadline
	}
	// 4. wire.Session 为本跳分配 call_id，并将它与此附件绑定；不回填源请求编号。
	b.mu.Lock()
	b.calls[invocation] = callOrigin{caller: c, callID: in.CallID}
	b.mu.Unlock()
	defer func() {
		b.mu.Lock()
		delete(b.calls, invocation)
		b.mu.Unlock()
	}()
	response, err := plugin.Call(ctx, wire.Request{Target: service.GetRouteId(), TraceID: invocation.TraceID, Deadline: deadline, Attach: invocation,
		Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: invocationSchemaVersion, Bytes: encoded}})
	// 5. 保留目标的业务错误与有效载荷；成功结果必须符合声明的输出契约。
	if response.Err != nil {
		response.Err = proto.Clone(response.Err).(*pb.ErrorInfo)
		response.Err.Code = validate.PublicCode(response.Err.GetCode())
		return response
	}
	if err != nil {
		return failure(validate.PublicCode(wire.CodeOf(err)), err.Error())
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
