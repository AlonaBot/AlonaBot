package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/broker"
	"github.com/AlonaBot/AlonaBot/internal/core/buttons"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/google/jsonschema-go/jsonschema"
	"google.golang.org/protobuf/proto"
)

// InvokeOptions 提供工具转发与 Core 直接确认所需的装配依赖。
type InvokeOptions struct {
	Call           func(ctx context.Context, c router.Caller, in *wire.Inbound) wire.Response
	Adapters       *adapter.Registry
	Buttons        *buttons.Service
	ConfirmTimeout time.Duration
	DefaultLocale  func() string // 每次确认读取当前 Core 设置
}

// RegisterInvoke 注册工具调用与操作确认，并接入确认按钮的点击处理。
func (c *Catalog) RegisterInvoke(m *router.Mux, opt InvokeOptions) {
	i := &invoker{catalog: c, options: opt, pending: make(map[string]*confirmation)}
	opt.Buttons.OnConfirm(i.clicked)
	m.Handle(wire.ServiceToolInvoke, i.invoke)
	m.Handle(wire.ServiceToolConfirm, i.confirm)
}

func (i *invoker) invoke(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	// 1. 父附件只由 Session 从本次会话的在途 Core 请求恢复。
	request := &pb.ToolInvokeRequest{}
	if err := decodeRequest(in, request); err != nil {
		return wire.Response{Err: err}
	}
	parent := in.ParentAttach.(*router.Invocation)
	pluginID, name, _ := strings.Cut(request.GetName(), "/")
	plugin, ok := i.catalog.router.Plugin(pluginID)
	if !ok || !plugin.Activated() {
		return failure(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "tool not found")
	}
	var service *pb.RouteDescriptor
	for _, route := range plugin.Routes() {
		spec := route.GetService()
		if route.GetKind() == pb.RouteKind_ROUTE_KIND_SERVICE && spec.GetServiceName() == name && spec.GetTool() != nil && router.ServiceCallable(spec, caller.PluginID, pluginID) {
			service = route
			break
		}
	}
	if service == nil {
		return failure(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "tool not found")
	}
	policy, err := i.catalog.Effective(ctx, pluginID, name)
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "read tool policy: "+err.Error())
	}
	if policy == pb.ToolPolicy_TOOL_POLICY_DENY {
		return failure(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "tool policy denies invocation")
	}
	// 2. 校验副本供 Schema 使用；转发和确认始终保留实际参数原文。
	arguments := request.GetArgumentsJson()
	var value any
	if !utf8.ValidString(arguments) || json.Unmarshal([]byte(arguments), &value) != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "tool arguments must be a UTF-8 JSON object")
	}
	if _, ok := value.(map[string]any); !ok {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "tool arguments must be a JSON object")
	}
	schema := &jsonschema.Schema{}
	if err := json.Unmarshal([]byte(service.GetService().GetTool().GetInputSchemaJson()), schema); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "decode tool input schema: "+err.Error())
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "resolve tool input schema: "+err.Error())
	}
	if err := resolved.Validate(value); err != nil {
		// 校验器错误包含参数原值，公开错误与日志只保留声明信息。
		slog.DebugContext(ctx, "tool input schema validation failed", "tool", request.GetName(), "schema_id", service.GetInputSchemaId())
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, fmt.Sprintf("invalid tool arguments for %q", request.GetName()))
	}
	permitted := len(service.GetRequiredPermissions()) == 0
	for _, required := range service.GetRequiredPermissions() {
		permitted = permitted || slices.Contains(parent.Event.GetActor().GetPermissions(), required)
	}
	if !permitted {
		return failure(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "tool service permission denied")
	}
	// 3. 策略快照只约束本次调用；确认等待继承父调用的取消与期限。
	if policy == pb.ToolPolicy_TOOL_POLICY_CONFIRM {
		if err := i.ask(ctx, caller, parent, request.GetName(), arguments, true); err != nil {
			return wire.Response{Err: err}
		}
	}
	if err := ctx.Err(); err != nil {
		return contextFailure(ctx)
	}
	forward := *in
	forward.Target = request.GetName()
	forward.Payload = wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, SchemaID: service.GetInputSchemaId(), SchemaVersion: service.GetInputSchemaVersion(), Bytes: []byte(arguments)}
	response := i.options.Call(broker.WithSelfToolCall(ctx), caller, &forward)
	if response.Err == nil || response.Payload.SchemaID != "" {
		encoded := toolResponse(&pb.ToolInvokeResponse{ResultJson: string(response.Payload.Bytes)})
		response.Payload = encoded.Payload
	}
	return response
}

func decodeRequest(in *wire.Inbound, request proto.Message) *pb.ErrorInfo {
	if parent, ok := in.ParentAttach.(*router.Invocation); in.ParentCallID == 0 || !ok || parent == nil {
		return failure(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "tool operation requires an active parent call").Err
	}
	if in.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.Payload.SchemaID != validate.SchemaID(request) || in.Payload.SchemaVersion != schemaVersion {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "unexpected tool request schema").Err
	}
	if err := proto.Unmarshal(in.Payload.Bytes, request); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid tool request protobuf: "+err.Error()).Err
	}
	return nil
}

func toolResponse(message proto.Message) wire.Response {
	encoded, err := proto.Marshal(message)
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "encode tool response: "+err.Error())
	}
	return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(message), SchemaVersion: schemaVersion, Bytes: encoded}}
}

func contextFailure(ctx context.Context) wire.Response {
	code := pb.ErrorCode_ERROR_CODE_CANCELLED
	if ctx.Err() == context.DeadlineExceeded {
		code = pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED
	}
	return failure(code, ctx.Err().Error())
}
