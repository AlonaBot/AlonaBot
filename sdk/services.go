package sdk

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
	"google.golang.org/protobuf/proto"
)

type serviceRegistry = registry.Services[*Context, Permission]
type serviceRoute = registry.Service[*Context]

func collectServices(root *Plugin, reg *registration) error {
	r := &root.services
	r.Schemas = generatedFor(root).ServiceSchemas
	r.Services = nil
	r.Errs = nil
	seen := make(map[string]bool)
	for _, b := range r.Builders {
		component, name, ok := strings.Cut(b.Target, "/")
		_, componentError := validate.PluginID(component)
		invalidDeclaration := !ok || !validate.IsServiceName(name) || component != "@agent" && componentError != nil
		if b.Target == "" && b.Name != "" { // 本组件自有服务只声明名称，归属在注册时确定。
			invalidDeclaration = !validate.IsServiceName(b.Name)
		}
		if invalidDeclaration || b.Encoding == pb.PayloadEncoding_PAYLOAD_ENCODING_UNSPECIFIED {
			r.Errs = append(r.Errs, fmt.Errorf("service %q has an invalid declaration", b.Target))
			continue
		}
		if seen[b.Name] {
			r.Errs = append(r.Errs, fmt.Errorf("service %q registered twice", b.Name))
			continue
		}
		seen[b.Name] = true
		if b.Handles != 1 || b.AfterHandle {
			r.Errs = append(r.Errs, fmt.Errorf("service %q has invalid Handle sequence", b.Target))
			continue
		}
		route, err := b.Prepare()
		if err != nil {
			r.Errs = append(r.Errs, fmt.Errorf("service %q: %w", b.Target, err))
			continue
		}
		r.Services = append(r.Services, route)
	}
	reg.services = r.Services
	registration := &pb.Register{CandidateConfigJson: "{}"}
	appendServiceRoutes(reg, registration)
	r.Errs = append(r.Errs, validate.Register(registration))
	return errors.Join(r.Errs...)
}

// checkServiceTargets 使用启动身份检查归属；@agent 的角色解析留给 Broker。
func checkServiceTargets(root *Plugin, pluginID string) error {
	var errs []error
	for _, b := range root.services.Builders {
		if b.Target == "" && b.Name != "" {
			continue
		}
		component, _, _ := strings.Cut(b.Target, "/")
		if component != pluginID && component != "@agent" {
			errs = append(errs, fmt.Errorf("service %q does not belong to plugin %q", b.Target, pluginID))
		}
	}
	return errors.Join(errs...)
}

func appendServiceRoutes(reg *registration, registration *pb.Register) {
	for _, service := range reg.services {
		registration.Routes = append(registration.Routes, proto.Clone(service.Descriptor).(*pb.RouteDescriptor))
	}
}

type serviceContextKey struct{}

func bindServiceContext(c *Context) { c.ctx = context.WithValue(c.ctx, serviceContextKey{}, c) }

// Caller 返回 Core 下发的服务直接调用方插件 ID；Core 调用时为空。
func (c *Context) Caller() string { return c.caller }

func serviceHandler(reg *registration) func(context.Context, *wire.Inbound) (wire.Response, bool) {
	routes := make(map[string]serviceRoute, len(reg.services))
	for _, service := range reg.services {
		routes[service.Descriptor.GetRouteId()] = service
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
	descriptor := service.Descriptor
	if request.GetRouteId() != in.Target || request.GetEncoding() != descriptor.GetService().GetEncoding() || request.GetSchemaId() != descriptor.GetInputSchemaId() || request.GetSchemaVersion() != descriptor.GetInputSchemaVersion() {
		return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "service invocation does not match registered input")
	}
	// 2. 将父请求绑定到 Handler 有效期；恢复 panic 时记录完整堆栈。
	ctx, cancel := context.WithCancel(ctx)
	c := &Context{plugin: service.Plugin.(*Plugin), subID: service.SubID, ctx: ctx, inbound: in, actor: request.GetContext().GetActor(), defaultTarget: request.GetContext().GetDefaultTarget()}
	c.freezeProfile(request.GetContext())
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
	encoded, err := service.Invoke(c, request.GetPayload())
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
