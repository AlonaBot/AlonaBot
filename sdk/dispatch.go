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
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/proto"
)

const invocationSchemaVersion = 1

func businessHandler(reg *Registration) wire.Handler {
	routes := make(map[string]EventRoute, len(reg.Events))
	commands := make(map[string]Command, len(reg.Commands))
	for _, route := range reg.Events {
		routes[route.RouteID] = route
	}
	for _, command := range reg.Commands {
		commands[commandRouteID(command)] = command
	}
	services := serviceHandler(reg)
	return func(ctx context.Context, in *wire.Inbound) wire.Response {
		if response, ok := services(ctx, in); ok {
			return response
		}
		route, eventRoute := routes[in.Target]
		command, commandRoute := commands[in.Target]
		if !eventRoute && !commandRoute {
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
			args, err := commandArguments(command, request.GetCommand())
			if err != nil {
				return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
			}
			var groups []string
			if command.spec.GetMatch() == pb.CommandMatch_COMMAND_MATCH_REGEX {
				groups = request.GetCommand().GetRegexGroups()
			}
			return invokeHandler(ctx, in, request.GetEvent(), command.Mode == pb.HandlerMode_HANDLER_MODE_AUTO_REPLY, groups, func(c *Context) (message.Chain, error) {
				args[0] = reflect.ValueOf(c)
				values := reflect.ValueOf(command.Handler).Call(args)
				var result message.Chain
				if c.autoReply {
					result = values[0].Interface().(message.Chain)
				}
				if value := values[len(values)-1]; !value.IsNil() {
					return result, value.Interface().(error)
				}
				return result, nil
			})
		}
		return invokeHandler(ctx, in, request.GetEvent(), false, nil, func(c *Context) (message.Chain, error) { return nil, route.Handler(c) })
	}
}

func invokeHandler(ctx context.Context, in *wire.Inbound, event *pb.Event, autoReply bool, groups []string, handler func(*Context) (message.Chain, error)) (response wire.Response) {
	// 1. 为 Handler 建立独立有效期，返回或 panic 后取消所有仍在使用 Context 的工作。
	ctx, cancel := context.WithCancel(ctx)
	c := &Context{ctx: ctx, inbound: in, event: event, autoReply: autoReply, match: groups}
	bindServiceContext(c)
	outcome := &pb.HandlerOutcome{Matched: true, Propagation: pb.Propagation_PROPAGATION_CONTINUE}
	c.active.Store(true)
	defer func() {
		c.active.Store(false)
		cancel()
		if p := recover(); p != nil {
			c.Logger().Error("handler panic", "panic", fmt.Sprint(p), "stack", string(debug.Stack()))
			response.Err = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "handler panicked"}
			outcome.Chain = nil
		}
		// 2. 出错仍输出传播标记，保留 Handler 已经设置的 STOP。
		if c.stopped.Load() {
			outcome.Propagation = pb.Propagation_PROPAGATION_STOP
		}
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
	}()
	chain, err := handler(c)
	if err != nil {
		var public *Error
		if errors.As(err, &public) {
			response.Err = &pb.ErrorInfo{Code: public.Code, SafeMessage: public.Message}
		} else {
			c.Logger().Error("handler returned an error", "error", err)
			response.Err = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "handler returned an error"}
		}
	}
	if err == nil && autoReply && len(chain) != 0 {
		outcome.Chain = chain.Proto()
	}

	return response
}

// Match 返回本次 Regex 的捕获组，包含下标 0 的整体匹配；其他路由返回 nil。
func (c *Context) Match() []string { return c.match }

func commandArguments(command Command, invocation *pb.CommandInvocation) ([]reflect.Value, error) {
	// 1. 调用载荷与 Handler 必须对应注册时的同一份参数描述。
	mode, types, err := checkHandler(command.Handler)
	if err != nil {
		return nil, err
	}
	params := command.spec.GetParams()
	if invocation == nil || len(invocation.GetArgs()) != len(params) || len(types) != len(params) || mode != command.Mode {
		return nil, errors.New("command arguments or handler do not match the registered signature")
	}
	args := make([]reflect.Value, len(params)+1)
	handlerType := reflect.TypeOf(command.Handler)
	for i, param := range params {
		arg := invocation.Args[i]
		if arg.GetName() != param.GetName() || types[i] != param.GetType() {
			return nil, fmt.Errorf("parameter %q does not match the registered name or type", param.GetName())
		}
		// 2. 无 value 只表示缺失的无默认值可选参数；其 Go 值为声明类型零值。
		if arg.GetValue() == nil {
			if param.GetRequired() || param.DefaultValue != nil {
				return nil, fmt.Errorf("parameter %q is missing a value", param.GetName())
			}
			args[i+1] = reflect.Zero(handlerType.In(i + 1))
			continue
		}
		switch value := arg.GetValue().(type) {
		case *pb.ArgValue_StringValue:
			args[i+1] = reflect.ValueOf(value.StringValue)
		case *pb.ArgValue_IntValue:
			args[i+1] = reflect.ValueOf(value.IntValue)
		case *pb.ArgValue_FloatValue:
			args[i+1] = reflect.ValueOf(value.FloatValue)
		case *pb.ArgValue_BoolValue:
			args[i+1] = reflect.ValueOf(value.BoolValue)
		}
		if args[i+1].Type() != handlerType.In(i+1) {
			return nil, fmt.Errorf("parameter %q value has an incorrect type", param.GetName())
		}
	}
	return args, nil
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
