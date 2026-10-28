package sdk

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

const invocationSchemaVersion = 1

func businessHandler(reg *Registration) wire.Handler {
	routes := make(map[string]EventRoute, len(reg.Events))
	commands := make(map[string]bool, len(reg.Commands))
	for _, route := range reg.Events {
		routes[route.RouteID] = route
	}
	for _, command := range reg.Commands {
		commands[commandRouteID(command)] = true
	}
	return func(ctx context.Context, in *wire.Inbound) wire.Response {
		route, ok := routes[in.Target]
		if !ok {
			if commands[in.Target] {
				return dispatchError(pb.ErrorCode_ERROR_CODE_UNSUPPORTED, "command dispatch is not implemented")
			}
			return dispatchError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "event route not found")
		}
		request := &pb.HandlerRequest{}
		if err := decodeInvocation(in.Payload, request); err != nil {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
		}
		if request.GetRouteId() != in.Target {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "handler route_id does not match target")
		}
		return invokeEvent(ctx, in, request.GetEvent(), route.Handler)
	}
}

func invokeEvent(ctx context.Context, in *wire.Inbound, event *pb.Event, handler func(*Context) error) (response wire.Response) {
	// 1. 为 Handler 建立独立有效期，返回或 panic 后取消所有仍在使用 Context 的工作。
	ctx, cancel := context.WithCancel(ctx)
	c := &Context{ctx: ctx, inbound: in, event: event}
	c.active.Store(true)
	defer func() {
		c.active.Store(false)
		cancel()
		if p := recover(); p != nil {
			c.Logger().Error("event handler panic", "panic", fmt.Sprint(p), "stack", string(debug.Stack()))
			response.Err = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "handler panicked"}
		}
		// 2. 出错仍输出传播标记，保留 Handler 已经设置的 STOP。
		outcome := &pb.HandlerOutcome{Matched: true, Propagation: pb.Propagation_PROPAGATION_CONTINUE}
		if c.stopped.Load() {
			outcome.Propagation = pb.Propagation_PROPAGATION_STOP
		}
		encoded, err := proto.Marshal(outcome)
		if err != nil {
			response = dispatchError(pb.ErrorCode_ERROR_CODE_INTERNAL, "encode HandlerOutcome: "+err.Error())
			return
		}
		response.Payload = wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(outcome), SchemaVersion: invocationSchemaVersion, Bytes: encoded}
	}()
	if err := handler(c); err != nil {
		var public *Error
		if errors.As(err, &public) {
			response.Err = &pb.ErrorInfo{Code: public.Code, SafeMessage: public.Message}
		} else {
			c.Logger().Error("handler returned an error", "error", err)
			response.Err = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "handler returned an error"}
		}
	}
	return response
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
