package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
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

// ServiceOption 设置服务的访问规则。
type ServiceOption func(*serviceOptions)

type serviceOptions struct {
	access    Access
	allowlist []string
}

// WithAccess 设置访问范围；AccessAllowlist 的名单使用插件 ID。
func WithAccess(a Access, allowlist ...string) ServiceOption {
	return func(o *serviceOptions) {
		o.access = a
		o.allowlist = append([]string(nil), allowlist...)
	}
}

// ServiceRegistry 收集静态服务；零值可用，Collect 统一返回描述错误。
type ServiceRegistry struct {
	services []serviceRoute
	errs     []error
}

type serviceRoute struct {
	descriptor *pb.RouteDescriptor
	invoke     func(*Context, []byte) ([]byte, error)
}

// JSONService 注册类型化 JSON 服务；schema 标识随包固定，输入和输出使用 version。
func JSONService[In, Out any](r *ServiceRegistry, name, schemaIn, schemaOut string, version uint32, h func(*Context, In) (Out, error), opts ...ServiceOption) {
	if h == nil {
		r.errs = append(r.errs, fmt.Errorf("handler for service %q is nil", name))
		return
	}
	r.add(name, schemaIn, schemaOut, version, pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, func(c *Context, b []byte) ([]byte, error) {
		var input In
		if err := json.Unmarshal(b, &input); err != nil {
			return nil, &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "decode service input: " + err.Error()}
		}
		out, err := h(c, input)
		if err != nil {
			return nil, err
		}
		encoded, err := json.Marshal(out)
		if err != nil {
			return nil, fmt.Errorf("encode service output: %w", err)
		}
		return encoded, nil
	}, opts)
}

// ProtoService 注册 Protobuf 服务，schema 来自消息类型，版本固定为 1。
// In 与 Out 使用生成的 Protobuf 消息指针类型。
func ProtoService[In, Out proto.Message](r *ServiceRegistry, name string, h func(*Context, In) (Out, error), opts ...ServiceOption) {
	if h == nil {
		r.errs = append(r.errs, fmt.Errorf("handler for service %q is nil", name))
		return
	}
	input, output := newServiceMessage[In](), newServiceMessage[Out]()
	r.add(name, validate.SchemaID(input), validate.SchemaID(output), invocationSchemaVersion, pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, func(c *Context, b []byte) ([]byte, error) {
		input := newServiceMessage[In]()
		if err := proto.Unmarshal(b, input); err != nil {
			return nil, &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "decode service input: " + err.Error()}
		}
		out, err := h(c, input)
		if err != nil {
			return nil, err
		}
		encoded, err := proto.Marshal(out)
		if err != nil {
			return nil, fmt.Errorf("encode service output: %w", err)
		}
		return encoded, nil
	}, opts)
}

func newServiceMessage[T proto.Message]() T {
	var message T
	return message.ProtoReflect().Type().New().Interface().(T)
}

func (r *ServiceRegistry) add(name, schemaIn, schemaOut string, version uint32, encoding pb.PayloadEncoding, invoke func(*Context, []byte) ([]byte, error), opts []ServiceOption) {
	var options serviceOptions
	for _, option := range opts {
		option(&options)
	}
	access := pb.ServiceAccess_SERVICE_ACCESS_UNSPECIFIED
	switch options.access {
	case AccessPrivate:
		access = pb.ServiceAccess_SERVICE_ACCESS_PRIVATE
	case AccessAllowlist:
		access = pb.ServiceAccess_SERVICE_ACCESS_ALLOWLIST
	case AccessPublic:
		access = pb.ServiceAccess_SERVICE_ACCESS_PUBLIC
	}
	r.services = append(r.services, serviceRoute{descriptor: &pb.RouteDescriptor{
		RouteId: serviceRoutePrefix + name, Kind: pb.RouteKind_ROUTE_KIND_SERVICE, DefaultEnabled: true, RequiredPermission: pb.Permission_PERMISSION_ANY,
		InputSchemaId: schemaIn, InputSchemaVersion: version, OutputSchemaId: schemaOut, OutputSchemaVersion: version,
		Spec: &pb.RouteDescriptor_Service{Service: &pb.ServiceSpec{ServiceName: name, Access: access, AllowlistPluginIds: options.allowlist, Encoding: encoding}},
	}, invoke: invoke})
}

func collectServices(root Root, reg *Registration) error {
	r := &ServiceRegistry{}
	root.Services(r)
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
	c := &Context{ctx: ctx, inbound: in}
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

// CallJSON 调用 plugin_id/service_name，使用 schemaIn 和 version 标识输入。
// Handler 中传入 Context.Context() 以关联父请求；后台工作使用独立 context。
func CallJSON[Out any](ctx context.Context, target, schemaIn string, version uint32, in any) (Out, error) {
	var out Out
	session, err := currentSession()
	if err != nil {
		return out, err
	}
	encoded, err := json.Marshal(in)
	if err != nil {
		return out, &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "encode service input: " + err.Error()}
	}
	response, err := callService(ctx, session, target, wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, SchemaID: schemaIn, SchemaVersion: version, Bytes: encoded})
	if err != nil {
		return out, err
	}
	if response.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_JSON {
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
