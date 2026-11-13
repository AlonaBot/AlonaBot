// Package plugintest 在 go test 中装配真实 Core、被测插件和 loopback 测试 Adapter。
//
// [Start] 创建独立临时数据目录，[Harness.Message] 与 [Harness.Event] 注入事件，
// [Result] 汇总 Core 分发结束或进入多轮等待前的发送与 Handler 结果（CONTRACTS §17）。
package plugintest

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlonaBot/AlonaBot"
	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/internal/inproc"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// AdapterID 指定缺省测试 Adapter 的平台名称。
const AdapterID = "plugintest"

// InstanceID 指定测试实例。
const InstanceID = "test"

// AccountID 指定测试实例的机器人账号。
const AccountID = "bot"

const (
	adapterComponentID = "plugintest_adapter"
	coreConfigFile     = "core_config.json"
	pluginConfigFile   = "config.json"
	loopbackListen     = "127.0.0.1:0"
)

// Chat 保存测试会话的身份与类别。
type Chat struct {
	id       string
	chatType message.ChatType
}

// Private 创建 chat_id 为 private:<userID> 的私聊。
func Private(userID string) Chat { return Chat{id: "private:" + userID, chatType: message.ChatPrivate} }

// Group 创建 chat_id 为 group:<chatID> 的群聊。
func Group(chatID string) Chat { return Chat{id: "group:" + chatID, chatType: message.ChatGroup} }

// User 提供会话中的用户与平台权限。
type User struct {
	ID, Name    string
	Permissions []sdk.Permission
}

// Sent 记录测试 Adapter 接受的一次发送或编辑。
type Sent struct {
	Op     string // "send" 或 "edit"。
	To     message.Session
	Chain  message.Chain
	Edited message.MessageRef
	Result *message.SendResult
}

// Text 按顺序拼接消息链的文本片段。
func (s Sent) Text() string { return s.Chain.Text() }

// Outcome 保存 Handler 的成功或安全错误。
type Outcome struct {
	PluginID, RouteID string
	Err               *errcode.Error // nil 表示成功。
}

// Result 汇总一次注入到分发通知之间的发送与调用结果。
type Result struct {
	Sends    []Sent
	Outcomes []Outcome
	Waiting  bool
}

// Harness 持有一次测试的 Core 与 Adapter；事件注入按顺序调用。
type Harness struct {
	t         testing.TB
	ctx       context.Context
	adapterID string
	busy      atomic.Bool // 立即拒绝同一套件的并发注入。
	sequence  uint64      // 由注入入口独占。
	mu        sync.Mutex  // 保护 Adapter 发布、发送记录与通知接收者。
	core      adapter.Core
	sends     []Sent
	messageID uint64
	pending   map[string]chan dispatchResult
}

// Start 在 t.TempDir 装配真实 Core，测试清理时停止全部组件并注销桥接。
func Start(t testing.TB, opts ...Option) *Harness {
	t.Helper()
	o := options{t: t, configs: make(map[*sdk.Plugin]string), allowed: make(map[*sdk.Plugin][]string),
		capabilities: []string{"message.send", "message.markdown", "message.album", "message.reference", "message.buttons", "message.forward", "message.edit"},
		descriptor:   adapter.Descriptor{AdapterID: AdapterID}}
	for _, apply := range opts {
		apply(&o)
	}
	// 1. 身份来自生成 manifest，配置按开发组件的目录规则写入。
	dir := t.TempDir()
	pluginID := func(p *sdk.Plugin) string {
		if p == nil || len(p.Generated().Manifest) == 0 {
			t.Fatal("plugin has no generated project manifest")
		}
		parsed, err := manifest.ParseProject(p.Generated().Manifest)
		if err != nil {
			t.Fatalf("parse plugin manifest: %v", err)
		}
		return parsed.Manifest.ID
	}
	rootOptions := []alonabot.Option{alonabot.DataDir(dir), alonabot.Listen(loopbackListen)}
	for _, p := range o.plugins {
		pluginID(p)
		rootOptions = append(rootOptions, alonabot.Plugin(p))
	}
	for p, config := range o.configs {
		path := filepath.Join(dir, "plugins", pluginID(p))
		if err := os.MkdirAll(path, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, pluginConfigFile), []byte(config), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	allowed := make(map[string][]string)
	for p, instances := range o.allowed {
		allowed[pluginID(p)] = slices.Clone(instances)
	}
	admins := make([]string, 0, len(o.admins))
	for _, userID := range o.admins {
		admins = append(admins, o.descriptor.AdapterID+":"+userID)
	}
	configuration, err := json.Marshal(map[string]any{"$version": 1, "admins": admins,
		"adapters": map[string]any{o.descriptor.AdapterID: map[string]any{"package": adapterComponentID, "instances": map[string]any{InstanceID: map[string]any{}}}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, coreConfigFile), configuration, 0o600); err != nil {
		t.Fatal(err)
	}
	// 2. 在组件激活前登记桥接，让 Core 写入发信许可并发送分发通知。
	ctx, cancel := context.WithCancel(context.Background())
	h := &Harness{t: t, ctx: ctx, adapterID: o.descriptor.AdapterID, pending: make(map[string]chan dispatchResult)}
	unregister := inproc.RegisterHarness(dir, &inproc.Harness{AllowedInstances: allowed, Dispatched: h.dispatched})
	for _, name := range o.capabilities {
		o.descriptor.Capabilities = append(o.descriptor.Capabilities, message.Capability{Name: name, Version: 1, Supported: true})
	}
	rootOptions = append(rootOptions, alonabot.Adapter(&testAdapter{h: h, descriptor: o.descriptor}))
	core, err := alonabot.Start(ctx, rootOptions...)
	if err != nil {
		cancel()
		unregister()
		t.Fatalf("start plugin test harness: %v", err)
	}
	// 3. Wait 完成后再注销，保留停止期间的桥接生命周期。
	t.Cleanup(func() {
		cancel()
		if err := core.Wait(); err != nil {
			t.Errorf("stop plugin test harness: %v", err)
		}
		unregister()
	})
	return h
}

func (h *Harness) dispatched(d inproc.Dispatch) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if receiver := h.pending[d.SourceEventID]; d.InstanceID == InstanceID && receiver != nil {
		receiver <- dispatchResult{dispatch: d, end: len(h.sends)} // 一次通知，缓冲允许 Core 在 Deliver 返回前完成分发。
	}
}

// Message 经测试实例注入一条消息，并等待 Core 的分发或等待通知。
func (h *Harness) Message(chat Chat, from User, content message.Content) Result {
	h.t.Helper()
	if !h.busy.CompareAndSwap(false, true) {
		h.t.Fatal("concurrent event injection on one Harness")
	}
	defer h.busy.Store(false)

	var chain message.Chain
	switch c := content.(type) {
	case message.Chain:
		chain = c
	case message.Segment:
		chain = message.Chain{c}
	default:
		h.t.Fatalf("unsupported message content %T", content)
	}
	toMe := chat.chatType == message.ChatPrivate
	for _, segment := range chain {
		if mention, ok := segment.(message.MentionSegment); ok && mention.ExternalUserID == AccountID {
			toMe = true
		}
	}
	return h.deliver(chat, from, "message.created", &message.CreatedPayload{Chain: chain, ToMe: toMe})
}

// Text 注入一个纯文本消息。
func (h *Harness) Text(chat Chat, from User, text string) Result {
	h.t.Helper()
	return h.Message(chat, from, message.Text(text))
}

// Event 经测试实例注入声明过的扩展事件，载荷由 Core 按 Adapter Schema 检查。
func (h *Harness) Event(chat Chat, from User, event sdk.EventType, payload any) Result {
	h.t.Helper()
	if !h.busy.CompareAndSwap(false, true) {
		h.t.Fatal("concurrent event injection on one Harness")
	}
	defer h.busy.Store(false)

	raw, err := json.Marshal(payload)
	if err != nil {
		h.t.Fatalf("encode adapter event: %v", err)
	}
	return h.deliver(chat, from, event.Name(), &message.AdapterEventPayload{Version: event.Version(), PayloadJSON: string(raw)})
}

func (h *Harness) deliver(chat Chat, from User, eventType string, payload message.Payload) Result {
	h.t.Helper()
	h.sequence++
	id := strconv.FormatUint(h.sequence, 10)
	if created, ok := payload.(*message.CreatedPayload); ok {
		created.PlatformMessageID = id
	}
	var permissions []string
	for _, p := range from.Permissions {
		permissions = append(permissions, p.Names()...)
	}
	ev := &message.Event{Type: eventType, Source: message.SourceAdapter, OccurredAt: time.Now(), AdapterID: h.adapterID,
		AdapterInstanceID: InstanceID, AccountID: AccountID, Session: message.NewSession(h.adapterID, chat.id, ""), ChatType: chat.chatType,
		Actor:         &message.Actor{ExternalUserID: from.ID, AdapterInstanceID: InstanceID, DisplayName: from.Name, Permissions: permissions},
		SourceEventID: id, Payload: payload}
	receiver := make(chan dispatchResult, 1)
	h.mu.Lock()
	start := len(h.sends)
	h.pending[id] = receiver
	core := h.core
	h.mu.Unlock()
	defer func() { h.mu.Lock(); delete(h.pending, id); h.mu.Unlock() }()
	accepted, err := core.Deliver(h.ctx, ev, nil)
	if err != nil || !accepted {
		h.t.Fatalf("deliver event: accepted=%v error=%v", accepted, err)
	}
	// Core 拥有结束判定；这里不使用 sleep、轮询或套件自己的期限。
	notification := <-receiver
	d := notification.dispatch
	result := Result{Waiting: d.Waiting}
	h.mu.Lock()
	result.Sends = slices.Clone(h.sends[start:notification.end])
	h.mu.Unlock()
	for _, o := range d.Outcomes {
		outcome := Outcome{PluginID: o.PluginID, RouteID: o.RouteID}
		if o.ErrorCode != 0 {
			outcome.Err = errcode.Code(o.ErrorCode).New(o.SafeMessage)
		}
		result.Outcomes = append(result.Outcomes, outcome)
	}
	return result
}

// dispatchResult 固定通知时的发送记录终点，后续后台发送不会混入本次结果。
type dispatchResult struct {
	dispatch inproc.Dispatch
	end      int
}
