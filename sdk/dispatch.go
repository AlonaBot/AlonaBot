package sdk

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/proto"
)

const invocationSchemaVersion = 1

func businessHandler(reg *registration) wire.Handler {
	handlers := make(map[string]func(*Context) error, len(reg.Events)+len(reg.Schedules))
	rules := make(map[string][]Rule, len(reg.Events))
	commands := make(map[string]command, len(reg.Commands))
	deferredRoutes := make(map[string]bool, len(reg.Events))
	functions := &continuationFunctions{functions: make(map[uint64]continuationFunction)}
	for _, route := range reg.Events {
		route := route
		handlers[route.RouteID] = func(c *Context) error { return invokeEvent(c, route) }
		rules[route.RouteID] = route.rules
		deferredRoutes[route.RouteID] = route.continuation
	}
	// 定时任务与事件共用手动 Handler 路径，取消后仍等待用户函数实际返回。
	for _, job := range reg.Schedules {
		handlers[scheduleRoutePrefix+job.JobID] = job.Handler
	}
	for _, task := range reg.tasks {
		handlers[taskRoutePrefix+task.Name] = task.Invoke
	}
	for _, command := range reg.Commands {
		commands[commandRouteID(command)] = command
	}
	services := serviceHandler(reg)
	hooks := hookHandler(reg)
	return func(ctx context.Context, in *wire.Inbound) wire.Response {
		if in.Target == wire.ServiceProfileSetOverrides {
			return reg.plugin.changeProfileOverrides(in)
		}
		ctx = context.WithValue(ctx, pluginContextKey{}, pluginScope{plugin: reg.plugin, subID: reg.subPlugins[in.Target]})
		if in.Target == "sdk.service.stream.delta" && registry.StreamReceiver != nil {
			return registry.StreamReceiver(ctx, in)
		}
		if response, ok := services(ctx, in); ok {
			return response
		}
		if response, ok := hooks(ctx, in); ok {
			return response
		}
		handler, manualRoute := handlers[in.Target]
		command, commandRoute := commands[in.Target]
		if !manualRoute && !commandRoute {
			return dispatchError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "handler route not found")
		}
		request := &pb.HandlerRequest{}
		if err := decodeInvocation(in.Payload, request); err != nil {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
		}
		if request.GetRouteId() != in.Target {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "handler route_id does not match target")
		}
		if commandRoute {
			args, err := commandArguments(command, request.GetCommand(), request.GetEvent().GetTarget())
			if err != nil {
				return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
			}
			var groups []string
			if command.spec.GetMatch() == pb.CommandMatch_COMMAND_MATCH_REGEX {
				groups = request.GetCommand().GetRegexGroups()
			}
			return invokeHandler(ctx, in, request, command.mode == pb.HandlerMode_HANDLER_MODE_AUTO_REPLY, groups, command.rules, functions, command.continuation, func(c *Context) (message.Content, error) {
				args[0] = reflect.ValueOf(c)
				values := reflect.ValueOf(command.Handler).Call(args)
				var result message.Content
				if c.autoReply {
					if !values[0].IsNil() {
						result = values[0].Interface().(message.Content)
					}
				}
				if value := values[len(values)-1]; !value.IsNil() {
					return result, value.Interface().(error)
				}
				return result, nil
			})
		}
		return invokeHandler(ctx, in, request, false, nil, rules[in.Target], functions, deferredRoutes[in.Target], func(c *Context) (message.Content, error) { return nil, handler(c) })
	}
}

func invokeHandler(ctx context.Context, in *wire.Inbound, request *pb.HandlerRequest, autoReply bool, groups []string, rules []Rule, functions *continuationFunctions, deferred bool, handler func(*Context) (message.Content, error)) (response wire.Response) {
	// 1. 为 Handler 建立独立有效期，返回或 panic 后取消所有仍在使用 Context 的工作。
	parentContext := ctx
	ctx, cancel := context.WithCancel(ctx)
	event := request.GetEvent()
	scope, _ := ctx.Value(pluginContextKey{}).(pluginScope)
	c := &Context{plugin: scope.plugin, subID: scope.subID, ctx: ctx, inbound: in, event: event, autoReply: autoReply, match: groups, continuations: functions, continuation: request.ContinuationOf != nil, commandConsumed: request.GetContext().GetCommandConsumed(), chatType: request.GetContext().GetChatType()}
	c.freezeProfile(request.GetContext())
	bindServiceContext(c)
	outcome := &pb.HandlerOutcome{Propagation: pb.Propagation_PROPAGATION_CONTINUE}
	c.active.Store(true)
	finish := func() {
		// 与 StopPropagation 共用锁，使失效与传播结果的快照成为同一步。
		c.propagationMu.Lock()
		c.active.Store(false)
		if c.stopped && !deferred && !c.continuation {
			outcome.Propagation = pb.Propagation_PROPAGATION_STOP
		}
		c.propagationMu.Unlock()
		cancel()
	}
	defer func() {
		finish()
		if p := recover(); p != nil {
			c.Logger().Error("handler panic", "panic", fmt.Sprint(p), "stack", string(debug.Stack()))
			response.Err = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "handler panicked"}
			outcome.Chain = nil
		}
		// 出错仍输出传播标记，保留 Handler 已经设置的 STOP。
		encoded, err := proto.Marshal(outcome)
		if err != nil {
			response.Err = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "encode HandlerOutcome: " + err.Error()}
			// 非法回复链编码失败时仍返回独立的传播标记，保留 Handler 的 STOP。
			outcome.Chain = nil
			encoded, err = proto.Marshal(outcome)
			if err != nil {
				response = dispatchError(pb.ErrorCode_ERROR_CODE_INTERNAL, "encode propagation outcome: "+err.Error())
				return
			}
		}
		response.Payload = wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(outcome), SchemaVersion: invocationSchemaVersion, Bytes: encoded}
		if response.Err != nil || parentContext.Err() != nil {
			c.discardContinuation()
		}
	}()
	// 2. Rule 与目标函数共享恢复、错误映射和生命周期；全部通过才标记匹配。
	var chain message.Content
	var err error
	matched := true
	if c.continuation {
		rules = nil // 原调用已执行全部 Rule，重放只执行 Handler 或登记的函数。
		if !deferred {
			functions.mu.Lock()
			fn, ok := functions.functions[request.GetContinuationOf()]
			if ok && fn.routeID == in.Target {
				delete(functions.functions, request.GetContinuationOf())
			}
			functions.mu.Unlock()
			if !ok || fn.routeID != in.Target {
				return dispatchError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "continuation function not found")
			}
			handler = func(c *Context) (message.Content, error) { return nil, fn.fn(c) }
		}
	}
	for _, rule := range rules {
		matched, err = rule(c)
		if err != nil || !matched {
			break
		}
	}
	if matched && err == nil {
		outcome.Matched = true
		c.inHandler = true
		if deferred && !c.continuation {
			err = c.callContinuation(wire.ServiceContinuationStart, &pb.ContinuationStartRequest{}, &pb.ContinuationStartResponse{})
		} else {
			chain, err = handler(c)
		}
	}
	if err == nil && c.continuation && autoReply && len(convert.ContentChain(chain)) != 0 {
		_, err = c.Send(chain)
		chain = nil // 延续调用的回复在有效父调用内发送，Core 不再处理该返回值。
	}
	finish()
	if err != nil {
		var public *Error
		if errors.As(err, &public) {
			response.Err = &pb.ErrorInfo{Code: pb.ErrorCode(public.Code), SafeMessage: public.Message}
		} else {
			c.Logger().Error("handler returned an error", "error", err)
			response.Err = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "handler returned an error"}
		}
	}
	if err == nil && autoReply && len(convert.ContentChain(chain)) != 0 {
		encoded, err := convert.ContentToProto(chain, event.GetTarget())
		if err != nil {
			failure := err.(*Error)
			response.Err = convert.ErrorToProto(failure)
		} else {
			outcome.Chain = encoded
		}
	}

	return response
}

// Match 返回本次 Regex 的捕获组，包含下标 0 的整体匹配；其他路由返回 nil。
func (c *Context) Match() []string { return c.match }

func commandArguments(command command, invocation *pb.CommandInvocation, target *pb.Target) ([]reflect.Value, error) {
	// 1. 在协议边界核对载荷与注册描述；Handler 签名已在 collect 时检查。
	params := command.spec.GetParams()
	if invocation == nil || len(invocation.GetArgs()) != len(params) {
		return nil, errors.New("command arguments or handler do not match the registered signature")
	}
	handlerType := reflect.TypeOf(command.Handler)
	args := make([]reflect.Value, handlerType.NumIn())
	if command.structure != nil {
		args[1] = reflect.New(command.structure).Elem()
	}
	for i, param := range params {
		arg := invocation.Args[i]
		if arg.GetName() != param.GetName() {
			return nil, fmt.Errorf("parameter %q does not match the registered name or type", param.GetName())
		}
		var typ reflect.Type
		if command.structure != nil {
			typ = command.structure.Field(i).Type
		} else {
			typ = handlerType.In(i + 1)
		}
		// 2. 只赋入 Core 已解析的值；缺失的可选指针保持 nil。
		value, err := commandArgument(arg, param, typ, target)
		if err != nil {
			return nil, err
		}
		if command.structure != nil {
			args[1].Field(i).Set(value)
		} else {
			args[i+1] = value
		}
	}
	return args, nil
}

func commandArgument(arg *pb.ArgValue, param *pb.CommandParam, typ reflect.Type, target *pb.Target) (reflect.Value, error) {
	if arg.GetValue() == nil {
		if param.GetRequired() || param.DefaultValue != nil {
			return reflect.Value{}, fmt.Errorf("parameter %q is missing a value", param.GetName())
		}
		return reflect.Zero(typ), nil
	}
	var value reflect.Value
	var paramType pb.ParamType
	switch v := arg.GetValue().(type) {
	case *pb.ArgValue_StringValue:
		value, paramType = reflect.ValueOf(v.StringValue), pb.ParamType_PARAM_TYPE_STRING
	case *pb.ArgValue_IntValue:
		value, paramType = reflect.ValueOf(v.IntValue), pb.ParamType_PARAM_TYPE_INT
	case *pb.ArgValue_FloatValue:
		value, paramType = reflect.ValueOf(v.FloatValue), pb.ParamType_PARAM_TYPE_FLOAT
	case *pb.ArgValue_BoolValue:
		value, paramType = reflect.ValueOf(v.BoolValue), pb.ParamType_PARAM_TYPE_BOOL
	case *pb.ArgValue_UserId:
		value, paramType = reflect.ValueOf(UserID(v.UserId)), pb.ParamType_PARAM_TYPE_USER
	case *pb.ArgValue_Image:
		value, paramType = reflect.ValueOf(convert.ImageFromProto(v.Image)), pb.ParamType_PARAM_TYPE_IMAGE
	case *pb.ArgValue_Message:
		value, paramType = reflect.ValueOf(convert.ChainFromProto(v.Message, target)), pb.ParamType_PARAM_TYPE_MESSAGE
	}
	if paramType != param.GetType() {
		return reflect.Value{}, fmt.Errorf("parameter %q value has an incorrect type", param.GetName())
	}
	if typ.Kind() == reflect.Pointer && value.Type() != typ {
		pointer := reflect.New(typ.Elem())
		pointer.Elem().Set(value.Convert(typ.Elem()))
		return pointer.Convert(typ), nil
	}
	return value.Convert(typ), nil
}

func decodeInvocation(payload wire.Payload, message proto.Message) error {
	if payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || payload.SchemaID != validate.SchemaID(message) || payload.SchemaVersion != invocationSchemaVersion {
		return fmt.Errorf("expected %s protobuf schema version 1", message.ProtoReflect().Descriptor().Name())
	}
	if err := proto.Unmarshal(payload.Bytes, message); err != nil {
		return fmt.Errorf("decode %s: %w", message.ProtoReflect().Descriptor().Name(), err)
	}
	return nil
}

func dispatchError(code pb.ErrorCode, message string) wire.Response {
	return wire.Response{Err: &pb.ErrorInfo{Code: code, SafeMessage: message}}
}
