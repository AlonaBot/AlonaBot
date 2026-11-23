// Package plugintest 在 go test 中装配真实 Core、被测插件和 loopback 测试 Adapter。
//
// [Start] 创建独立临时数据目录，[Harness.Message]、[Harness.Event] 与 [Harness.Click] 注入事件，
// [Harness.Trigger] 与 [Harness.TriggerTask] 经 Scheduler 立即触发任务。
// [Result] 按 trace 汇总 Core 完成或等待通知前的发送与 Handler 结果，后台记录由 [Harness.Background] 读取。
// [Connect] 经管理认证连接已有 Core，复用同一消息与结果类型。
package plugintest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlonaBot/AlonaBot"
	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/internal/generate"
	"github.com/AlonaBot/AlonaBot/internal/generated"
	"github.com/AlonaBot/AlonaBot/internal/inproc"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
	sdkdebug "github.com/AlonaBot/AlonaBot/sdk/debug"
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
	instanceID string
	id         string
	chatType   message.ChatType
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
	eventID                                   string
	AdapterID, InstanceID, AccountID, TraceID string
	Err                                       *errcode.Error
	Media                                     []Media
	Op                                        string // "send" 或 "edit"。
	To                                        string
	Chain                                     message.Chain
	Edited                                    string
	Result                                    *message.SendResult
}

// Text 按顺序拼接消息链的文本片段。
func (s Sent) Text() string { return s.Chain.Text() }

// Outcome 保存实际调用的 Handler 结果，包括 Handler 返回的安全错误。
type Outcome struct {
	TraceID           string
	PluginID, RouteID string
	Err               *errcode.Error // nil 表示成功。
	Waiting           bool
}

const (
	// PhaseAdmission 标识 Handler 调用前的权限、生命周期和路由准入检查。
	PhaseAdmission = sdkdebug.PhaseAdmission
	// PhaseArguments 标识 Handler 调用前的命令参数解析。
	PhaseArguments = sdkdebug.PhaseArguments
)

// Diagnostic 保留 Handler 调用前的阶段、路由、trace 与安全错误。
type Diagnostic = sdkdebug.Diagnostic

// Result 汇总一次注入的发送、实际 Handler 结果与调用前诊断。
// RequireReply 继续只检查 Handler 结果和回复；解析与准入错误通过 Diagnostics 断言。
type Result struct {
	TraceIDs    []string
	Calls       []Call
	Sends       []Sent
	Outcomes    []Outcome
	Diagnostics []Diagnostic
	Waiting     bool
}

// RequireReply 要求全部 Handler 成功，且恰好发送一条文本为 want 的消息。
func (r Result) RequireReply(t testing.TB, want string) {
	t.Helper()
	for _, outcome := range r.Outcomes {
		if outcome.Err != nil {
			t.Fatalf("handler %s/%s: %v", outcome.PluginID, outcome.RouteID, outcome.Err)
		}
	}
	if len(r.Sends) != 1 || r.Sends[0].Text() != want {
		t.Fatalf("sends = %+v, want one reply %q", r.Sends, want)
	}
}

// Harness 持有一次测试的 Core 与 Adapter；Submit 支持并行事件，结果按 trace 归属。
type Harness struct {
	close       func() error
	debugClient *sdkdebug.Client
	traceOwners map[string]string
	instances   map[string]InstanceConfig
	cores       map[string]adapter.Core
	fixtures    map[string]map[string]Media
	calls       []Call
	changed     chan struct{}
	t           testing.TB
	ctx         context.Context
	busy        atomic.Bool // 串行显式按钮点击与任务触发。
	sequence    uint64      // 注入入口以原子操作分配来源 ID。
	mu          sync.Mutex  // 保护 Adapter 发布、发送记录与通知接收者。
	sends       []Sent
	messageID   uint64
	pending     map[string]chan dispatchResult
	bridge      *inproc.Harness
	plugins     map[*sdk.Plugin]string // 启动前登记，运行期间只读。
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
	o.addLLM()
	// 1. 身份来自生成 manifest，配置按开发组件的目录规则写入。
	dir := t.TempDir()
	pluginID := func(p *sdk.Plugin) string {
		if p == nil {
			t.Fatal("plugin has no generated project manifest")
		}
		resources := scanResources(t, p)
		parsed, err := manifest.ParseProject(resources.Manifest)
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
	admins := slices.Clone(o.platformAdmins)
	for _, userID := range o.admins {
		admins = append(admins, o.descriptor.AdapterID+":"+userID)
	}
	instances := o.instances
	if len(instances) == 0 {
		instances = []InstanceConfig{{AdapterID: o.descriptor.AdapterID, InstanceID: InstanceID, AccountID: AccountID, Descriptor: o.descriptor, Capabilities: o.capabilities, Behavior: o.behavior}}
	}
	groups := map[string]any{}
	descriptors := map[string]adapter.Descriptor{}
	instanceMap := map[string]InstanceConfig{}
	for _, config := range instances {
		if config.AdapterID == "" || config.InstanceID == "" || config.AccountID == "" {
			t.Fatal("instance requires adapter, instance and account IDs")
		}
		if _, exists := instanceMap[config.InstanceID]; exists {
			t.Fatalf("duplicate test instance %q", config.InstanceID)
		}
		if config.Capabilities == nil {
			config.Capabilities = slices.Clone(o.capabilities)
		}
		descriptor := config.Descriptor
		descriptor.AdapterID = config.AdapterID
		if descriptor.AdapterID == o.descriptor.AdapterID && len(descriptor.Capabilities) == 0 && len(descriptor.Permissions) == 0 && len(descriptor.Actions) == 0 && len(descriptor.Events) == 0 && len(descriptor.Extensions) == 0 {
			descriptor = o.descriptor
		}
		for _, name := range config.Capabilities {
			if !slices.ContainsFunc(descriptor.Capabilities, func(c message.Capability) bool { return c.Name == name }) {
				descriptor.Capabilities = append(descriptor.Capabilities, message.Capability{Name: name, Version: 1, Supported: true})
			}
		}
		config.Descriptor = descriptor
		instanceMap[config.InstanceID] = config
		group, ok := groups[config.AdapterID].(map[string]any)
		if !ok {
			group = map[string]any{"package": adapterComponentID + "_" + config.AdapterID, "instances": map[string]any{}, "default_instance": o.defaults[config.AdapterID]}
			groups[config.AdapterID] = group
		}
		group["instances"].(map[string]any)[config.InstanceID] = map[string]any{}
		if existing, ok := descriptors[config.AdapterID]; ok {
			if len(descriptor.Actions) > 0 {
				if len(existing.Actions) > 0 && !reflect.DeepEqual(existing.Actions, descriptor.Actions) {
					t.Fatalf("conflicting Actions declarations for adapter %q", config.AdapterID)
				}
				existing.Actions = descriptor.Actions
			}
			if len(descriptor.Permissions) > 0 {
				if len(existing.Permissions) > 0 && !reflect.DeepEqual(existing.Permissions, descriptor.Permissions) {
					t.Fatalf("conflicting Permissions declarations for adapter %q", config.AdapterID)
				}
				existing.Permissions = descriptor.Permissions
			}
			if len(descriptor.Events) > 0 {
				if len(existing.Events) > 0 && !reflect.DeepEqual(existing.Events, descriptor.Events) {
					t.Fatalf("conflicting Events declarations for adapter %q", config.AdapterID)
				}
				existing.Events = descriptor.Events
			}
			if len(descriptor.Extensions) > 0 {
				if len(existing.Extensions) > 0 && !reflect.DeepEqual(existing.Extensions, descriptor.Extensions) {
					t.Fatalf("conflicting Extensions declarations for adapter %q", config.AdapterID)
				}
				existing.Extensions = descriptor.Extensions
			}
			for _, capability := range descriptor.Capabilities {
				if !slices.ContainsFunc(existing.Capabilities, func(c message.Capability) bool { return c.Name == capability.Name && c.Version == capability.Version }) {
					existing.Capabilities = append(existing.Capabilities, capability)
				}
			}
			descriptors[config.AdapterID] = existing
		} else {
			descriptors[config.AdapterID] = descriptor
		}
	}
	configuration, err := json.Marshal(map[string]any{"admins": admins, "adapters": groups})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, coreConfigFile), configuration, 0o600); err != nil {
		t.Fatal(err)
	}
	// 2. 在组件激活前登记桥接，让 Core 写入发信许可并发送分发通知。
	ctx, cancel := context.WithCancel(context.Background())
	h := &Harness{traceOwners: map[string]string{}, instances: instanceMap, cores: map[string]adapter.Core{}, fixtures: map[string]map[string]Media{}, changed: make(chan struct{}), t: t, ctx: ctx, pending: make(map[string]chan dispatchResult), plugins: make(map[*sdk.Plugin]string)}
	for _, p := range o.plugins {
		h.plugins[p] = pluginID(p)
	}
	h.bridge = &inproc.Harness{AllowedInstances: allowed, Dispatched: h.dispatched, Scheduled: func(runID, traceID string, explicit bool) {
		if explicit {
			h.mu.Lock()
			h.traceOwners[traceID] = scheduledSourcePrefix + runID
			h.mu.Unlock()
		}
	}}
	if o.clock != nil {
		h.bridge.Now = o.clock.Now
		h.bridge.AfterFunc = func(d time.Duration, f func()) inproc.Timer { return o.clock.AfterFunc(d, f) }
		h.bridge.NewTimer = func(d time.Duration) inproc.Timer { return o.clock.NewTimer(d) }
	}
	unregister := inproc.RegisterHarness(dir, h.bridge)
	for _, descriptor := range descriptors {
		test := &testAdapter{h: h, descriptor: descriptor}
		test.bindResources()
		rootOptions = append(rootOptions, alonabot.Adapter(test))
	}

	core, err := alonabot.Start(ctx, rootOptions...)
	if err != nil {
		cancel()
		unregister()
		t.Fatalf("start plugin test harness: %v", err)
	}
	h.debugClient = &sdkdebug.Client{URL: h.bridge.DebugURL, Host: h.bridge.DebugHost, Token: h.bridge.DebugToken}
	// 3. Wait 完成后再注销，保留停止期间的桥接生命周期。
	h.close = sync.OnceValue(func() error { cancel(); err := core.Wait(); unregister(); return err })
	t.Cleanup(func() {
		if err := h.Close(); err != nil {
			t.Errorf("stop plugin test harness: %v", err)
		}
	})
	return h
}

func (h *Harness) dispatched(d inproc.Dispatch) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.pending[d.SourceEventID] != nil {
		for _, trace := range d.TraceIDs {
			h.traceOwners[trace] = d.SourceEventID
		}
	}
	if d.Started {
		return
	}
	if receiver := h.pending[d.SourceEventID]; receiver != nil {
		receiver <- dispatchResult{dispatch: d}
	}
}

// Message 经测试实例注入一条消息，并等待 Core 的分发或等待通知。
func (h *Harness) Message(chat Chat, from User, content message.Content) Result {
	h.t.Helper()

	result, err := h.Submit(chat, from, content).Wait(h.ctx)
	if err != nil {
		h.t.Fatal(err)
	}
	return result
}

// Text 注入一个纯文本消息。
func (h *Harness) Text(chat Chat, from User, text string) Result {
	h.t.Helper()
	return h.Message(chat, from, message.Text(text))
}

// Event 经测试实例注入声明过的扩展事件，载荷由 Core 按 Adapter Schema 检查。
func (h *Harness) Event(chat Chat, from User, event sdk.EventType, payload any) Result {
	h.t.Helper()

	raw, err := json.Marshal(payload)
	if err != nil {
		h.t.Fatalf("encode adapter event: %v", err)
	}
	return h.deliver(chat, from, event.Name(), &message.AdapterEventPayload{Version: event.Version(), PayloadJSON: string(raw)})
}

func (h *Harness) deliver(chat Chat, from User, eventType string, payload message.Payload) Result {
	h.t.Helper()
	config := h.selectInstance(chat)
	return h.deliverEvent(from, &message.Event{Type: eventType, AdapterID: config.AdapterID, AdapterInstanceID: config.InstanceID, AccountID: config.AccountID,
		Session: message.SID(config.AdapterID, chat.id, ""), ChatType: chat.chatType, Payload: payload})
}

func (h *Harness) deliverEvent(from User, ev *message.Event) Result {
	result, err := h.submitEvent(from, ev).Wait(h.ctx)
	if err != nil {
		h.t.Fatal(err)
	}
	return result
}

func (h *Harness) submitEvent(from User, ev *message.Event) *Pending {
	h.t.Helper()
	id := strconv.FormatUint(atomic.AddUint64(&h.sequence, 1), 10)
	if created, ok := ev.Payload.(*message.CreatedPayload); ok {
		created.PlatformMessageID = id
	}
	if clicked, ok := ev.Payload.(*message.ButtonClickedPayload); ok {
		clicked.CallbackQueryID = id
	}
	var permissions []string
	for _, p := range from.Permissions {
		permissions = append(permissions, p.Names()...)
	}
	ev.Source, ev.OccurredAt, ev.SourceEventID = message.SourceAdapter, time.Now(), id
	ev.Actor = &message.Actor{ExternalUserID: from.ID, AdapterInstanceID: ev.AdapterInstanceID, DisplayName: from.Name, Permissions: permissions}
	receiver := make(chan dispatchResult, 1)
	h.mu.Lock()

	h.pending[id] = receiver
	core := h.cores[ev.AdapterInstanceID]
	h.mu.Unlock()
	accepted, err := core.Deliver(h.ctx, ev, nil)
	if err != nil || !accepted {
		h.t.Fatalf("deliver event: accepted=%v error=%v", accepted, err)
	}
	pending := &Pending{done: make(chan struct{})}
	go func() {
		defer close(pending.done)
		select {
		case notification := <-receiver:
			pending.result = h.result(notification)
		case <-h.ctx.Done():
			pending.err = h.ctx.Err()
		}
		h.mu.Lock()
		delete(h.pending, id)
		h.mu.Unlock()
	}()
	return pending
}

func (h *Harness) result(notification dispatchResult) Result {
	d := notification.dispatch
	result := Result{Waiting: d.Waiting}
	h.mu.Lock()
	result.TraceIDs = slices.Clone(d.TraceIDs)
	eventID := d.SourceEventID
	if d.RunID != "" {
		eventID = scheduledSourcePrefix + d.RunID
		for trace, owner := range h.traceOwners {
			if owner == eventID {
				result.TraceIDs = append(result.TraceIDs, trace)
			}
		}
	}
	remaining := h.sends[:0]
	for _, sent := range h.sends {
		if sent.eventID == eventID {
			result.Sends = append(result.Sends, sent)
		} else {
			remaining = append(remaining, sent)
		}
	}
	h.sends = remaining

	for _, call := range h.calls {
		if call.eventID == eventID {
			result.Calls = append(result.Calls, call)
		}
	}
	for _, trace := range result.TraceIDs {
		owner := h.traceOwners[trace]
		if owner == eventID {
			delete(h.traceOwners, trace)
		}
	}
	h.mu.Unlock()
	for _, o := range d.Outcomes {
		outcome := Outcome{TraceID: o.TraceID, PluginID: o.PluginID, RouteID: o.RouteID, Waiting: o.Waiting}
		if d.RunID != "" && len(result.TraceIDs) == 1 {
			outcome.TraceID = result.TraceIDs[0]
		}
		if o.ErrorCode != 0 {
			outcome.Err = errcode.Code(o.ErrorCode).New(o.SafeMessage)
		}
		result.Outcomes = append(result.Outcomes, outcome)
	}
	for _, diagnostic := range d.Diagnostics {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{Phase: diagnostic.Phase, TraceID: diagnostic.TraceID, PluginID: diagnostic.PluginID, RouteID: diagnostic.RouteID, Err: errcode.Code(diagnostic.ErrorCode).New(diagnostic.SafeMessage)})
	}
	return result
}

// dispatchResult 保存事件结束或首次等待时的分发快照。
type dispatchResult struct {
	dispatch inproc.Dispatch
}

// resourceScanMu 保证并行 Harness 只扫描并登记同一入口包一次。
var resourceScanMu sync.Mutex

func scanResources(t testing.TB, p *sdk.Plugin) generated.Resources {
	t.Helper()
	pkgPath, _ := generated.Entry(p)
	if resources, ok := generated.Lookup(p, pkgPath); ok {
		return resources
	}
	resourceScanMu.Lock()
	defer resourceScanMu.Unlock()
	if resources, ok := generated.Lookup(p, pkgPath); ok {
		return resources
	}
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	options := generate.Options{Dir: dir, Package: pkgPath}
	if build, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range build.Settings {
			switch setting.Key {
			case "GOOS":
				options.GOOS = setting.Value
			case "GOARCH":
				options.GOARCH = setting.Value
			case "-tags":
				options.Tags = strings.FieldsFunc(setting.Value, func(c rune) bool { return c == ',' || c == ' ' })
			}
		}
	}
	result, err := generate.Scan(context.Background(), options)
	if err != nil {
		t.Fatalf("scan entry package %q: %v", pkgPath, err)
	}
	if len(result.Diagnostics) != 0 {
		var diagnostics []string
		for _, d := range result.Diagnostics {
			diagnostics = append(diagnostics, fmt.Sprintf("%s: %s", d.Pos, d.Message))
		}
		t.Fatalf("scan entry package %q:\n%s", pkgPath, strings.Join(diagnostics, "\n"))
	}
	generated.Register(pkgPath, result.Generated)
	return result.Generated
}

func (h *Harness) selectInstance(chat Chat) InstanceConfig {
	id := chat.instanceID
	if id == "" {
		id = InstanceID
	}
	config, ok := h.instances[id]
	if !ok {
		h.t.Fatalf("unknown inbound instance %q; select one with Chat.OnInstance", id)
	}
	return config
}

// DebugClient 返回已认证的本地管理客户端，供同一 Core 的 Connect 联调。
func (h *Harness) DebugClient() *sdkdebug.Client { return h.debugClient }

// Close 停止本地测试 Core 并等待组件退出，可与测试清理重复调用。
func (h *Harness) Close() error { return h.close() }
