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

// ServiceRegistry 收集静态服务；零值可用，Collect 统一返回描述错误。
type ServiceRegistry struct {
	services []serviceRoute
	builders []*serviceBuilder
	schemas  map[string][]byte
	errs     []error
}

type serviceRoute struct {
	descriptor *pb.RouteDescriptor
	invoke     func(*Context, []byte) ([]byte, error)
}

type serviceBuilder struct {
	name        string
	access      Access
	allowlist   []string
	handles     int
	afterHandle bool
	prepare     func() (serviceRoute, error)
}

// JSONService 构建具名输入/输出类型的 JSON 服务；Handle 结束声明。
type JSONService[In, Out any] struct {
	builder *serviceBuilder
	handler func(*Context, In) (Out, error)
}

// JSON 开始构建 JSON 服务，Schema 由生成器提供。
func JSON[In, Out any](r *ServiceRegistry, name string) *JSONService[In, Out] {
	b := &serviceBuilder{name: name}
	s := &JSONService[In, Out]{builder: b}
	b.prepare = func() (serviceRoute, error) {
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
		return b.route(inputID, outputID, inputVersion, outputVersion, pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, func(c *Context, data []byte) ([]byte, error) {
			if err := validateServiceJSON(inputSchema, data); err != nil {
				return nil, &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "invalid service input"}
			}
			var input In
			if err := json.Unmarshal(data, &input); err != nil {
				return nil, &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "decode service input: " + err.Error()}
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
				return nil, &Error{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, Message: "invalid service output"}
			}
			return encoded, nil
		}), nil
	}
	r.builders = append(r.builders, b)
	return s
}

// Access 设置直接调用方规则，重复调用采用最后一次。
func (s *JSONService[In, Out]) Access(a Access, allowlist ...string) *JSONService[In, Out] {
	s.builder.setAccess(a, allowlist)
	return s
}

// Handle 结束 JSON 服务声明。
func (s *JSONService[In, Out]) Handle(h func(*Context, In) (Out, error)) {
	s.builder.handles++
	s.handler = h
}

// ProtoService 构建 Protobuf 服务；消息类型提供 Schema 标识。
type ProtoService[In, Out proto.Message] struct {
	builder *serviceBuilder
	handler func(*Context, In) (Out, error)
}

// Proto 开始构建 Protobuf 服务，版本保持为 1。
func Proto[In, Out proto.Message](r *ServiceRegistry, name string) *ProtoService[In, Out] {
	b := &serviceBuilder{name: name}
	s := &ProtoService[In, Out]{builder: b}
	b.prepare = func() (serviceRoute, error) {
		handler := s.handler // 收集后冻结 Handler，保持 Session 注册固定。
		if handler == nil {
			return serviceRoute{}, fmt.Errorf("handler is nil")
		}
		input, output := newServiceMessage[In](), newServiceMessage[Out]()
		return b.route(validate.SchemaID(input), validate.SchemaID(output), invocationSchemaVersion, invocationSchemaVersion, pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, func(c *Context, data []byte) ([]byte, error) {
			input := newServiceMessage[In]()
			if err := proto.Unmarshal(data, input); err != nil {
				return nil, &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "decode service input: " + err.Error()}
			}
			output, err := handler(c, input)
			if err != nil {
				return nil, err
			}
			encoded, err := proto.Marshal(output)
			if err != nil {
				return nil, fmt.Errorf("encode service output: %w", err)
			}
			return encoded, nil
		}), nil
	}
	r.builders = append(r.builders, b)
	return s
}

// Access 设置直接调用方规则。
func (s *ProtoService[In, Out]) Access(a Access, allowlist ...string) *ProtoService[In, Out] {
	s.builder.setAccess(a, allowlist)
	return s
}

// Handle 结束 Protobuf 服务声明。
func (s *ProtoService[In, Out]) Handle(h func(*Context, In) (Out, error)) {
	s.builder.handles++
	s.handler = h
}

func (b *serviceBuilder) setAccess(a Access, allowlist []string) {
	if b.handles != 0 {
		b.afterHandle = true
	}
	b.access, b.allowlist = a, slices.Clone(allowlist)
}

func newServiceMessage[T proto.Message]() T {
	var message T
	return message.ProtoReflect().Type().New().Interface().(T)
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
	return serviceRoute{descriptor: &pb.RouteDescriptor{
		RouteId: serviceRoutePrefix + b.name, Kind: pb.RouteKind_ROUTE_KIND_SERVICE, DefaultEnabled: true,
		InputSchemaId: inputID, InputSchemaVersion: inputVersion, OutputSchemaId: outputID, OutputSchemaVersion: outputVersion,
		Spec: &pb.RouteDescriptor_Service{Service: &pb.ServiceSpec{ServiceName: b.name, Access: access, AllowlistPluginIds: b.allowlist, Encoding: encoding}},
	}, invoke: invoke}
}

func jsonType[T any]() (string, uint32, error) {
	typ := reflect.TypeFor[T]()
	if typ.PkgPath() == "" || typ.Name() == "" || strings.Contains(typ.Name(), "[") {
		return "", 0, &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: fmt.Sprintf("JSON type must be a named non-generic value type: %s", typ)}
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

func collectServices(root Root, reg *Registration) error {
	r := &ServiceRegistry{schemas: generatedFor(root).ServiceSchemas}
	root.Services(r)
	for _, b := range r.builders {
		if b.handles != 1 || b.afterHandle {
			r.errs = append(r.errs, fmt.Errorf("service route %q has invalid Handle sequence", serviceRoutePrefix+b.name))
			continue
		}
		route, err := b.prepare()
		if err != nil {
			r.errs = append(r.errs, fmt.Errorf("service route %q: %w", serviceRoutePrefix+b.name, err))
			continue
		}
		r.services = append(r.services, route)
	}
	reg.services = r.services
	registration := &pb.Register{CandidateConfigJson: "{}"}
	appendServiceRoutes(reg, registration)
	// Collect 尚未创建 HTTP Listener，仅校验 Services Hook 的描述。
	r.errs = append(r.errs, validate.Register(registration))
	seen := make(map[string]bool)
	for _, command := range reg.Commands {
		seen[commandRouteID(command)] = true
	}
	for _, event := range reg.Events {
		seen[event.RouteID] = true
	}
	for _, service := range r.services {
		id := service.descriptor.GetRouteId()
		if seen[id] {
			r.errs = append(r.errs, fmt.Errorf("service route %q conflicts with another route", id))
		}
	}
	return errors.Join(r.errs...)
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
	c := &Context{ctx: ctx, inbound: in, actor: request.GetContext().GetActor()}
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
			return dispatchError(public.Code, public.Message)
		}
		c.Logger().Error("service returned an error", "error", err)
		return dispatchError(pb.ErrorCode_ERROR_CODE_INTERNAL, "service returned an error")
	}
	// 3. 成功响应始终带声明的输出 schema，空 Protobuf 消息仍有完整元数据。
	return wire.Response{Payload: wire.Payload{Encoding: descriptor.GetService().GetEncoding(), SchemaID: descriptor.GetOutputSchemaId(), SchemaVersion: descriptor.GetOutputSchemaVersion(), Bytes: encoded}}
}

// CallJSON 按输入输出 Go 类型推导 Schema 标识及版本并调用服务。
// Handler 中传入 Context.Context() 以关联父请求；后台工作使用独立 context。
func CallJSON[In, Out any](ctx context.Context, target string, in In) (Out, error) {
	var out Out
	inputID, inputVersion, err := jsonType[In]()
	if err != nil {
		return out, err
	}
	outputID, outputVersion, err := jsonType[Out]()
	if err != nil {
		return out, err
	}
	session, err := currentSession()
	if err != nil {
		return out, err
	}
	encoded, err := json.Marshal(in)
	if err != nil {
		return out, &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "encode service input: " + err.Error()}
	}
	response, err := callService(ctx, session, target, wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, SchemaID: inputID, SchemaVersion: inputVersion, Bytes: encoded})
	if err != nil {
		return out, err
	}
	if response.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_JSON || response.SchemaID != outputID || response.SchemaVersion != outputVersion {
		return out, &Error{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, Message: "expected JSON service output"}
	}
	if err := json.Unmarshal(response.Bytes, &out); err != nil {
		return out, &Error{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, Message: "decode service output: " + err.Error()}
	}
	return out, nil
}

// CallProto 调用 plugin_id/service_name，按生成的 Out 消息类型解码结果。
// Handler 中传入 Context.Context() 以关联父请求；后台工作使用独立 context。
func CallProto[Out proto.Message](ctx context.Context, target string, in proto.Message) (Out, error) {
	var zero Out
	session, err := currentSession()
	if err != nil {
		return zero, err
	}
	encoded, err := proto.Marshal(in)
	if err != nil {
		return zero, &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "encode service input: " + err.Error()}
	}
	response, err := callService(ctx, session, target, wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(in), SchemaVersion: invocationSchemaVersion, Bytes: encoded})
	if err != nil {
		return zero, err
	}
	out := newServiceMessage[Out]()
	if err := decodeInvocation(response, out); err != nil {
		return zero, &Error{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, Message: "decode service output: " + err.Error()}
	}
	return out, nil
}

func callService(ctx context.Context, session *wire.Session, target string, payload wire.Payload) (wire.Payload, error) {
	request := wire.Request{Target: target, Payload: payload}
	if c, ok := ctx.Value(serviceContextKey{}).(*Context); ok {
		if !c.active.Load() {
			return wire.Payload{}, &Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "handler context has expired"}
		}
		request.ParentCallID = c.inbound.CallID
		request.TraceID = c.inbound.TraceID
	}
	response, err := session.Call(ctx, request)
	if err != nil {
		info := err.(*wire.CallError).ErrorInfo
		return response.Payload, &Error{Code: validate.PublicCode(info.GetCode()), Message: info.GetSafeMessage()}
	}
	return response.Payload, nil
}
