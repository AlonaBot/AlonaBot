package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"slices"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/google/jsonschema-go/jsonschema"
	"google.golang.org/protobuf/proto"
)

const serviceRoutePrefix = "svc."

// Access 指定服务允许的直接调用方。
type Access int

const (
	// AccessPrivate 只允许 Core 和本插件调用，作为默认访问规则。
	AccessPrivate Access = iota
	// AccessAllowlist 允许 Core、本插件和明确列出的插件调用。
	AccessAllowlist
	// AccessPublic 允许已启用插件调用。
	AccessPublic
)

// serviceRegistry 收集静态服务；零值可用，Collect 统一返回描述错误。
type serviceRegistry struct {
	services []serviceRoute
	builders []*serviceBuilder
	schemas  map[string][]byte
	errs     []error
}

type serviceRoute struct {
	plugin     *Plugin
	subID      string
	descriptor *pb.RouteDescriptor
	invoke     func(*Context, []byte) ([]byte, error)
}

type serviceBuilder struct {
	plugin      *Plugin
	subID       string
	name        string
	target      string
	encoding    pb.PayloadEncoding
	access      Access
	allowlist   []string
	handles     int
	afterHandle bool
	prepare     func() (serviceRoute, error)
}

// Service 绑定服务目标与输入输出类型；零值无效。
type Service[In, Out any] struct {
	target   string
	encoding pb.PayloadEncoding
}

// JSON 声明完整目标的 JSON 服务，Schema 由生成器提供。
func JSON[In, Out any](target string) Service[In, Out] {
	return Service[In, Out]{target: target, encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON}
}

// Proto 声明完整目标的 Protobuf 服务，Schema 版本保持为 1。
func Proto[In, Out proto.Message](target string) Service[In, Out] {
	return Service[In, Out]{target: target, encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF}
}

// ServiceBuilder 构建声明的访问规则与 Handler；Handle 结束注册。
type ServiceBuilder[In, Out any] struct {
	builder         *serviceBuilder
	handler         func(*Context, In) (Out, error)
	toolDescription *string // nil 表示未声明为工具，空说明留给注册校验拒绝。
	toolDefault     ToolDefault
}

// Serve 开始在插件或子插件上注册声明，Collect 统一检查构建结果。
func Serve[In, Out any](owner Registrar, declaration Service[In, Out]) *ServiceBuilder[In, Out] {
	scope := owner.registrar()
	scope.plugin.checkRegistration()
	_, name, _ := strings.Cut(declaration.target, "/")
	b := &serviceBuilder{plugin: scope.plugin, subID: scope.subID, target: declaration.target, name: name, encoding: declaration.encoding}
	s := &ServiceBuilder[In, Out]{builder: b}
	b.prepare = func() (serviceRoute, error) {
		if b.encoding == pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF {
			if s.toolDescription != nil {
				return serviceRoute{}, fmt.Errorf("Protobuf service cannot be a tool")
			}
			return s.prepareProto()
		}
		return s.prepareJSON()
	}
	scope.plugin.services.builders = append(scope.plugin.services.builders, b)
	return s
}

func (s *ServiceBuilder[In, Out]) prepareJSON() (serviceRoute, error) {
	b := s.builder
	r := &b.plugin.services

	// 1. 解析输入输出 Schema 并冻结 Handler，调用时复用校验器。
	inputID, inputVersion, err := jsonType[In]()
	if err != nil {
		return serviceRoute{}, err
	}
	outputID, outputVersion, err := jsonType[Out]()
	if err != nil {
		return serviceRoute{}, err
	}
	inputSchema, err := resolveServiceSchema(r.schemas, inputID)
	if err != nil {
		return serviceRoute{}, err
	}
	outputSchema, err := resolveServiceSchema(r.schemas, outputID)
	if err != nil {
		return serviceRoute{}, err
	}
	handler := s.handler // 收集后冻结 Handler，保持 Session 注册固定。
	if handler == nil {
		return serviceRoute{}, fmt.Errorf("handler is nil")
	}
	route := b.route(inputID, outputID, inputVersion, outputVersion, pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, func(c *Context, data []byte) ([]byte, error) {
		if err := validateServiceJSON(inputSchema, data); err != nil {
			return nil, &Error{Code: errcode.InvalidArgument, Message: "invalid service input"}
		}
		var input In
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, &Error{Code: errcode.InvalidArgument, Message: "decode service input: " + err.Error()}
		}
		out, err := handler(c, input)
		if err != nil {
			return nil, err
		}
		encoded, err := json.Marshal(out)
		if err != nil {
			return nil, fmt.Errorf("encode service output: %w", err)
		}
		if err := validateServiceJSON(outputSchema, encoded); err != nil {
			c.Logger().Error("service output schema validation failed", "schema_id", outputID)
			return nil, &Error{Code: errcode.Internal, Message: "invalid service output"}
		}
		return encoded, nil
	})
	// 2. 工具描述携带同一输入 Schema 原文，完整注册校验交给 Collect。
	if s.toolDescription != nil {
		var policy pb.ToolPolicy
		switch s.toolDefault {
		case ToolAllow:
			policy = pb.ToolPolicy_TOOL_POLICY_ALLOW
		case ToolConfirm:
			policy = pb.ToolPolicy_TOOL_POLICY_CONFIRM
		default:
			return serviceRoute{}, fmt.Errorf("unknown tool default policy %d", s.toolDefault)
		}
		route.descriptor.GetService().Tool = &pb.ToolSpec{Description: *s.toolDescription, DefaultPolicy: policy, InputSchemaJson: string(r.schemas[inputID])}
	}
	return route, nil
}

func (s *ServiceBuilder[In, Out]) prepareProto() (serviceRoute, error) {
	b := s.builder

	handler := s.handler // 收集后冻结 Handler，保持 Session 注册固定。
	if handler == nil {
		return serviceRoute{}, fmt.Errorf("handler is nil")
	}
	input, output := newServiceMessage[In](), newServiceMessage[Out]()
	return b.route(validate.SchemaID(input), validate.SchemaID(output), invocationSchemaVersion, invocationSchemaVersion, pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, func(c *Context, data []byte) ([]byte, error) {
		input := newServiceMessage[In]()
		if err := proto.Unmarshal(data, input); err != nil {
			return nil, &Error{Code: errcode.InvalidArgument, Message: "decode service input: " + err.Error()}
		}
		output, err := handler(c, input.(In))
		if err != nil {
			return nil, err
		}
		encoded, err := proto.Marshal(any(output).(proto.Message))
		if err != nil {
			return nil, fmt.Errorf("encode service output: %w", err)
		}
		return encoded, nil
	}), nil
}

// Access 设置直接调用方规则，重复调用采用最后一次。
func (s *ServiceBuilder[In, Out]) Access(a Access, allowlist ...string) *ServiceBuilder[In, Out] {
	s.builder.setAccess(a, allowlist)
	return s
}

// Tool 设置工具说明与缺省策略；Proto 声明或 Handle 后调用由 Collect 报错。
func (s *ServiceBuilder[In, Out]) Tool(description string, defaultPolicy ToolDefault) *ServiceBuilder[In, Out] {
	s.builder.plugin.checkRegistration()
	if s.builder.handles != 0 {
		s.builder.afterHandle = true
	}
	s.toolDescription, s.toolDefault = &description, defaultPolicy
	return s
}

// Handle 结束服务注册，冻结的 Handler 在 Session 内按声明类型调用。
func (s *ServiceBuilder[In, Out]) Handle(h func(*Context, In) (Out, error)) {
	s.builder.plugin.checkRegistration()
	s.builder.handles++
	s.handler = h
}

func (b *serviceBuilder) setAccess(a Access, allowlist []string) {
	b.plugin.checkRegistration()
	if b.handles != 0 {
		b.afterHandle = true
	}
	b.access, b.allowlist = a, slices.Clone(allowlist)
}

// Proto 在声明处约束消息类型；统一的 builder 只在编解码边界恢复该消息接口。
func newServiceMessage[T any]() proto.Message {
	var message T
	return any(message).(proto.Message).ProtoReflect().Type().New().Interface()
}

func (b *serviceBuilder) route(inputID, outputID string, inputVersion, outputVersion uint32, encoding pb.PayloadEncoding, invoke func(*Context, []byte) ([]byte, error)) serviceRoute {
	access := pb.ServiceAccess_SERVICE_ACCESS_UNSPECIFIED
	switch b.access {
	case AccessPrivate:
		access = pb.ServiceAccess_SERVICE_ACCESS_PRIVATE
	case AccessAllowlist:
		access = pb.ServiceAccess_SERVICE_ACCESS_ALLOWLIST
	case AccessPublic:
		access = pb.ServiceAccess_SERVICE_ACCESS_PUBLIC
	}
	return serviceRoute{plugin: b.plugin, subID: b.subID, descriptor: &pb.RouteDescriptor{
		RouteId: serviceRoutePrefix + b.name, Kind: pb.RouteKind_ROUTE_KIND_SERVICE, DefaultEnabled: true,
		InputSchemaId: inputID, InputSchemaVersion: inputVersion, OutputSchemaId: outputID, OutputSchemaVersion: outputVersion,
		Spec: &pb.RouteDescriptor_Service{Service: &pb.ServiceSpec{ServiceName: b.name, Access: access, AllowlistPluginIds: b.allowlist, Encoding: encoding}},
	}, invoke: invoke}
}

func jsonType[T any]() (string, uint32, error) {
	typ := reflect.TypeFor[T]()
	if typ.PkgPath() == "" || typ.Name() == "" || strings.Contains(typ.Name(), "[") {
		return "", 0, &Error{Code: errcode.InvalidArgument, Message: fmt.Sprintf("JSON type must be a named non-generic value type: %s", typ)}
	}
	version := uint32(1)
	if value, ok := any(*new(T)).(interface{ SchemaVersion() uint32 }); ok {
		version = value.SchemaVersion()
	}
	return typ.PkgPath() + "." + typ.Name(), version, nil
}

func resolveServiceSchema(schemas map[string][]byte, id string) (*jsonschema.Resolved, error) {
	raw, ok := schemas[id]
	if !ok {
		return nil, fmt.Errorf("service schema for %s is not generated; run alonabot plugin generate", id)
	}
	schema := &jsonschema.Schema{}
	if err := json.Unmarshal(raw, schema); err != nil {
		return nil, fmt.Errorf("decode service schema for %s: %w", id, err)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		return nil, fmt.Errorf("resolve service schema for %s: %w", id, err)
	}
	return resolved, nil
}

// validateServiceJSON 解码校验副本；jsonschema-go 将 json.Number 识别为 string。
func validateServiceJSON(schema *jsonschema.Resolved, data []byte) error {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	return schema.Validate(value)
}

func collectServices(root *Plugin, reg *Registration) error {
	r := &root.services
	r.schemas = generatedFor(root).ServiceSchemas
	r.services = nil
	r.errs = nil
	seen := make(map[string]bool)
	for _, b := range r.builders {
		component, name, ok := strings.Cut(b.target, "/")
		_, componentError := validate.PluginID(component)
		if !ok || !validate.IsServiceName(name) || component != "@agent" && componentError != nil || b.encoding == pb.PayloadEncoding_PAYLOAD_ENCODING_UNSPECIFIED {
			r.errs = append(r.errs, fmt.Errorf("service %q has an invalid declaration", b.target))
			continue
		}
		if seen[b.target] {
			r.errs = append(r.errs, fmt.Errorf("service %q registered twice", b.target))
			continue
		}
		seen[b.target] = true
		if b.handles != 1 || b.afterHandle {
			r.errs = append(r.errs, fmt.Errorf("service %q has invalid Handle sequence", b.target))
			continue
		}
		route, err := b.prepare()
		if err != nil {
			r.errs = append(r.errs, fmt.Errorf("service %q: %w", b.target, err))
			continue
		}
		r.services = append(r.services, route)
	}
	reg.services = r.services
	registration := &pb.Register{CandidateConfigJson: "{}"}
	appendServiceRoutes(reg, registration)
	r.errs = append(r.errs, validate.Register(registration))
	return errors.Join(r.errs...)
}

// checkServiceTargets 使用启动身份检查归属；@agent 的角色解析留给 Broker。
func checkServiceTargets(root *Plugin, pluginID string) error {
	var errs []error
	for _, b := range root.services.builders {
		component, _, _ := strings.Cut(b.target, "/")
		if component != pluginID && component != "@agent" {
			errs = append(errs, fmt.Errorf("service %q does not belong to plugin %q", b.target, pluginID))
		}
	}
	return errors.Join(errs...)
}

func appendServiceRoutes(reg *Registration, registration *pb.Register) {
	for _, service := range reg.services {
		registration.Routes = append(registration.Routes, proto.Clone(service.descriptor).(*pb.RouteDescriptor))
	}
}

type serviceContextKey struct{}

func bindServiceContext(c *Context) { c.ctx = context.WithValue(c.ctx, serviceContextKey{}, c) }

// Caller 返回 Core 下发的服务直接调用方插件 ID；Core 调用时为空。
func (c *Context) Caller() string { return c.caller }

func serviceHandler(reg *Registration) func(context.Context, *wire.Inbound) (wire.Response, bool) {
	routes := make(map[string]serviceRoute, len(reg.services))
	for _, service := range reg.services {
		routes[service.descriptor.GetRouteId()] = service
	}
	return func(ctx context.Context, in *wire.Inbound) (wire.Response, bool) {
		service, ok := routes[in.Target]
		if !ok {
			return wire.Response{}, false
		}
		return invokeService(ctx, in, service), true
	}
}

func invokeService(ctx context.Context, in *wire.Inbound, service serviceRoute) (response wire.Response) {
	// 1. 校验 Core 的调用封装和输入声明，解码成功后才进入用户函数。
	request := &pb.ServiceInvocation{}
	if err := decodeInvocation(in.Payload, request); err != nil {
		return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	descriptor := service.descriptor
	if request.GetRouteId() != in.Target || request.GetEncoding() != descriptor.GetService().GetEncoding() || request.GetSchemaId() != descriptor.GetInputSchemaId() || request.GetSchemaVersion() != descriptor.GetInputSchemaVersion() {
		return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "service invocation does not match registered input")
	}
	// 2. 将父请求绑定到 Handler 有效期；恢复 panic 时记录完整堆栈。
	ctx, cancel := context.WithCancel(ctx)
	c := &Context{plugin: service.plugin, subID: service.subID, ctx: ctx, inbound: in, actor: request.GetContext().GetActor(), defaultTarget: request.GetContext().GetDefaultTarget()}
	chain := request.GetContext().GetPluginChain()
	if len(chain) > 0 {
		c.caller = chain[len(chain)-1]
	}
	bindServiceContext(c)
	c.active.Store(true)
	defer func() {
		c.active.Store(false)
		cancel()
		if p := recover(); p != nil {
			c.Logger().Error("service handler panic", "panic", fmt.Sprint(p), "stack", string(debug.Stack()))
			response = dispatchError(pb.ErrorCode_ERROR_CODE_INTERNAL, "service panicked")
		}
	}()
	encoded, err := service.invoke(c, request.GetPayload())
	if err != nil {
		var public *Error
		if errors.As(err, &public) {
			return dispatchError(pb.ErrorCode(public.Code), public.Message)
		}
		c.Logger().Error("service returned an error", "error", err)
		return dispatchError(pb.ErrorCode_ERROR_CODE_INTERNAL, "service returned an error")
	}
	// 3. 成功响应始终带声明的输出 schema，空 Protobuf 消息仍有完整元数据。
	return wire.Response{Payload: wire.Payload{Encoding: descriptor.GetService().GetEncoding(), SchemaID: descriptor.GetOutputSchemaId(), SchemaVersion: descriptor.GetOutputSchemaVersion(), Bytes: encoded}}
}

// Call 调用声明的服务，JSON 输入输出按生成的 Schema 校验。
// Handler 中传入 Context.Context() 以关联父请求；后台工作使用独立 context。
func (s Service[In, Out]) Call(ctx context.Context, in In) (Out, error) {
	if s.target == "" || s.encoding == pb.PayloadEncoding_PAYLOAD_ENCODING_UNSPECIFIED {
		var out Out
		return out, errcode.InvalidArgument.Errorf("invalid service declaration %q", s.target)
	}
	if s.encoding == pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF {
		return s.callProto(ctx, in)
	}
	return s.callJSON(ctx, in)
}

func (s Service[In, Out]) callJSON(ctx context.Context, in In) (Out, error) {
	var out Out
	inputID, inputVersion, err := jsonType[In]()
	if err != nil {
		return out, err
	}
	outputID, outputVersion, err := jsonType[Out]()
	if err != nil {
		return out, err
	}
	session, err := currentSession(ctx)
	if err != nil {
		return out, err
	}
	schemas := componentFor(ctx).active.Load().l.serviceSchemas
	inputSchema, err := resolveServiceSchema(schemas, inputID)
	if err != nil {
		return out, err
	}
	outputSchema, err := resolveServiceSchema(schemas, outputID)
	if err != nil {
		return out, err
	}
	encoded, err := json.Marshal(in)
	if err != nil {
		return out, &Error{Code: errcode.InvalidArgument, Message: "encode service input: " + err.Error()}
	}
	if err := validateServiceJSON(inputSchema, encoded); err != nil {
		return out, errcode.InvalidArgument.Errorf("invalid service input for %q: %v", s.target, err)
	}
	response, err := callService(ctx, session, s.target, wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, SchemaID: inputID, SchemaVersion: inputVersion, Bytes: encoded})
	if err != nil {
		return out, err
	}
	if response.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_JSON || response.SchemaID != outputID || response.SchemaVersion != outputVersion {
		return out, &Error{Code: errcode.Internal, Message: "expected JSON service output"}
	}
	if err := validateServiceJSON(outputSchema, response.Bytes); err != nil {
		return out, errcode.Internal.Errorf("invalid service output for %q: %v", s.target, err)
	}
	if err := json.Unmarshal(response.Bytes, &out); err != nil {
		return out, &Error{Code: errcode.Internal, Message: "decode service output: " + err.Error()}
	}
	return out, nil
}

func (s Service[In, Out]) callProto(ctx context.Context, in In) (Out, error) {
	var zero Out
	session, err := currentSession(ctx)
	if err != nil {
		return zero, err
	}
	input := any(in).(proto.Message)
	encoded, err := proto.Marshal(input)
	if err != nil {
		return zero, &Error{Code: errcode.InvalidArgument, Message: "encode service input: " + err.Error()}
	}
	response, err := callService(ctx, session, s.target, wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(input), SchemaVersion: invocationSchemaVersion, Bytes: encoded})
	if err != nil {
		return zero, err
	}
	out := newServiceMessage[Out]()
	if err := decodeInvocation(response, out); err != nil {
		return zero, &Error{Code: errcode.Internal, Message: "decode service output: " + err.Error()}
	}
	return out.(Out), nil
}

func callService(ctx context.Context, session *wire.Session, target string, payload wire.Payload) (wire.Payload, error) {
	request := wire.Request{Target: target, Payload: payload}
	if c, ok := ctx.Value(serviceContextKey{}).(*Context); ok {
		if !c.active.Load() {
			return wire.Payload{}, &Error{Code: errcode.FailedPrecondition, Message: "handler context has expired"}
		}
		request.ParentCallID = c.inbound.CallID
		request.TraceID = c.inbound.TraceID
	}
	response, err := session.Call(ctx, request)
	if err != nil {
		info := err.(*wire.CallError).ErrorInfo
		return response.Payload, &Error{Code: errcode.Code(validate.PublicCode(info.GetCode())), Message: info.GetSafeMessage()}
	}
	return response.Payload, nil
}
