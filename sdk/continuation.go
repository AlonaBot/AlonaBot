package sdk

import (
	"context"
	"sync"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/proto"
)

type continuationFunction struct {
	routeID string
	fn      func(*Context) error
}
type continuationFunctions struct {
	mu        sync.Mutex
	functions map[uint64]continuationFunction // call_id 只在本次 Session 内有效。
}

// Continue 为当前 Handler 请求一次延续调用，fn 在原调用成功结束后执行。
func (c *Context) Continue(fn func(*Context) error) error {
	c.continueMu.Lock()
	defer c.continueMu.Unlock()
	if !c.active.Load() || !c.inHandler || c.continuation || c.continued || c.continuations == nil {
		return errcode.FailedPrecondition.New("continuation requires an active original Handler and may only be requested once")
	}
	if fn == nil {
		return errcode.InvalidArgument.New("continuation function is nil")
	}
	c.continued = true
	c.continuations.mu.Lock()
	c.continuations.functions[c.inbound.CallID] = continuationFunction{routeID: c.inbound.Target, fn: fn}
	c.continuations.mu.Unlock()
	err := c.callContinuation(wire.ServiceContinuationStart, &pb.ContinuationStartRequest{}, &pb.ContinuationStartResponse{})
	if err != nil {
		c.discardContinuation()
	}
	return err
}

func (c *Context) discardContinuation() {
	if c.continuations == nil {
		return
	}
	c.continuations.mu.Lock()
	delete(c.continuations.functions, c.inbound.CallID)
	c.continuations.mu.Unlock()
}

// Wait 等待同用户在原会话中的下一条消息；返回的 Context 与当前延续调用共同失效。
func (c *Context) Wait(timeout time.Duration) (*Context, error) {
	if !c.active.Load() || !c.continuation {
		return nil, errcode.FailedPrecondition.New("wait requires an active continuation Handler")
	}
	millis := timeout / time.Millisecond
	if millis < 1 || millis > time.Duration(^uint32(0)) {
		return nil, errcode.InvalidArgument.New("wait timeout is out of range")
	}
	result := &pb.WaitResponse{}
	if err := c.callContinuation(wire.ServiceWait, &pb.WaitRequest{TimeoutMs: uint32(millis)}, result); err != nil {
		return nil, err
	}
	next := &Context{ctx: c.ctx, inbound: c.inbound, event: result.Event, continuation: true}
	next.active.Store(true)
	// 命中消息仅替换可读事件；所有子请求继续使用当前延续调用作为父请求。
	context.AfterFunc(c.ctx, func() { next.active.Store(false) })
	return next, nil
}

// Text 返回 message.created 的消息纯文本，其他事件返回空字符串。
func (c *Context) Text() string {
	event := c.Event()
	if event != nil && event.Type == "message.created" {
		if payload, ok := event.Payload.(*message.CreatedPayload); ok {
			return payload.Chain.Text()
		}
	}
	return ""
}

func (c *Context) callContinuation(target string, request, result proto.Message) error {
	session, err := currentSession()
	if err != nil {
		return err
	}
	encoded, err := proto.Marshal(request)
	if err != nil {
		return errcode.InvalidArgument.New("encode continuation request: " + err.Error())
	}
	// core.wait 显式使用父延续的剩余期限，避免普通请求缺省期限提前结束等待。
	deadline, _ := c.ctx.Deadline()
	response, err := session.Call(c.ctx, wire.Request{Target: target, ParentCallID: c.inbound.CallID, TraceID: c.inbound.TraceID, Deadline: deadline,
		Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: invocationSchemaVersion, Bytes: encoded}})
	if err != nil {
		failure := err.(*wire.CallError).ErrorInfo
		return &Error{Code: errcode.Code(validate.PublicCode(failure.GetCode())), Message: failure.GetSafeMessage()}
	}
	if err := decodeInvocation(response.Payload, result); err != nil {
		return errcode.Internal.New(err.Error())
	}
	return nil
}
