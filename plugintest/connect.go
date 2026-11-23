package plugintest

import (
	"context"
	"fmt"
	"sync"

	"github.com/AlonaBot/AlonaBot/sdk/debug"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// Connection 连接运行中的 Core；Close 取消本客户端请求并释放观察资源。
// Adapter 的目录与真实操作通过 Client 访问。
type Connection struct {
	Client   *debug.Client
	identity InstanceConfig
	ctx      context.Context
	cancel   context.CancelFunc
	once     sync.Once
}

// Connect 使用已有管理认证连接 Core，并核对指定实例的目录。
func Connect(ctx context.Context, client *debug.Client, identity InstanceConfig) (*Connection, error) {
	if identity.AdapterID == "" || identity.InstanceID == "" || identity.AccountID == "" {
		return nil, fmt.Errorf("connect requires adapter, instance and account IDs")
	}
	if _, err := client.Operations(ctx, identity.InstanceID, ""); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	return &Connection{Client: client, identity: identity, ctx: ctx, cancel: cancel}, nil
}

// Close 结束本客户端及其在途观察，Core 与组件继续运行。
func (c *Connection) Close() error { c.once.Do(c.cancel); return nil }

// Event 经管理入口注入完整标准事件，保留真实路由结果。
func (c *Connection) Event(ctx context.Context, event *message.Event) (Result, error) {
	ctx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(c.ctx, cancel)
	defer func() { stop(); cancel() }()
	if c.ctx.Err() != nil {
		return Result{}, c.ctx.Err()
	}
	response, err := c.Client.Inject(ctx, event)
	result := Result{TraceIDs: response.TraceIDs, Waiting: response.Waiting}
	for _, o := range response.Outcomes {
		result.Outcomes = append(result.Outcomes, Outcome{TraceID: o.TraceID, PluginID: o.PluginID, RouteID: o.RouteID, Waiting: o.Waiting, Err: o.Err})
	}
	result.Diagnostics = response.Diagnostics
	for _, s := range response.Replies {
		result.Sends = append(result.Sends, Sent{AdapterID: event.AdapterID, InstanceID: s.InstanceID, AccountID: s.AccountID, TraceID: s.TraceID, Op: s.Operation, To: s.Session, Chain: s.Chain, Result: s.Result, Err: s.Err})
	}
	return result, err
}

// Message 构造标准入站消息，Core 负责命令匹配、参数解析与权限判断。
func (c *Connection) Message(ctx context.Context, chat Chat, from User, content message.Content) (Result, error) {
	var chain message.Chain
	switch v := content.(type) {
	case message.Chain:
		chain = v
	case message.Segment:
		chain = message.Chain{v}
	default:
		return Result{}, fmt.Errorf("unsupported message content %T", content)
	}
	identity := c.identity
	if chat.instanceID != "" && chat.instanceID != identity.InstanceID {
		return Result{}, fmt.Errorf("chat instance does not match connection")
	}
	toMe := chat.chatType == message.ChatPrivate
	for _, s := range chain {
		if m, ok := s.(message.MentionSegment); ok && m.ExternalUserID == identity.AccountID {
			toMe = true
		}
	}
	event := &message.Event{Type: "message.created", AdapterID: identity.AdapterID, AdapterInstanceID: identity.InstanceID, AccountID: identity.AccountID, ChatType: chat.chatType, Session: message.SID(identity.AdapterID, chat.id, ""), Actor: &message.Actor{ExternalUserID: from.ID, AdapterInstanceID: identity.InstanceID, DisplayName: from.Name}, Payload: &message.CreatedPayload{Chain: chain, ToMe: toMe}}
	for _, permission := range from.Permissions {
		event.Actor.Permissions = append(event.Actor.Permissions, permission.Names()...)
	}
	return c.Event(ctx, event)
}

// Text 注入纯文本标准消息。
func (c *Connection) Text(ctx context.Context, chat Chat, from User, text string) (Result, error) {
	return c.Message(ctx, chat, from, message.Text(text))
}

// Submit 异步提交真实 Core 消息；Wait 可取消等待，Close 结束此客户端的请求。
func (c *Connection) Submit(chat Chat, from User, content message.Content) *Pending {
	pending := &Pending{done: make(chan struct{})}
	go func() { defer close(pending.done); pending.result, pending.err = c.Message(c.ctx, chat, from, content) }()
	return pending
}
