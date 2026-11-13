//go:build unix

package alonabot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	coreadapter "github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk"
	sdkadapter "github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/proto"
)

const (
	pipelineModeEnv     = "ALONABOT_PIPELINE_MODE"
	pipelineAuditEnv    = "ALONABOT_PIPELINE_AUDIT"
	pipelineIDEnv       = "ALONABOT_PIPELINE_ID"
	pipelinePriorityEnv = "ALONABOT_PIPELINE_PRIORITY"
	pipelineInstanceID  = "pipeline-instance"
	pipelineAccountID   = "pipeline-account"
	pipelineGeneration  = "pipeline-generation"
	pipelineTimeout     = 10 * time.Second
	pipelineStopDrain   = 50 * time.Millisecond
	pipelineWaitTimeout = 20 * time.Millisecond
)

// TestMain 让同一测试二进制在子进程中运行真实 SDK 插件或回环 Adapter。
func TestMain(m *testing.M) {
	mode := os.Getenv(pipelineModeEnv)
	if mode == "" {
		os.Exit(m.Run())
	}
	var err error
	if mode == "adapter" {
		err = sdk.RunAdapter(&pipelineAdapter{})
	} else {
		err = sdk.Run(pipelinePlugin(mode))
	}
	if err != nil {
		slog.Error("pipeline process failed", "error", err)
		os.Exit(1)
	}
	os.Exit(0)
}

type pipelineRecord struct {
	Stage, Plugin, Route, Text, User, Session, Token, Code  string
	Action, Value, Job, RunID, MessageID, Instance, Account string
	Permissions                                             []string
}

// pipelineAudit 同步写入父进程，记录返回后才允许当前 Handler 继续。
func pipelineAudit(record pipelineRecord) error {
	record.Plugin = os.Getenv(pipelineIDEnv)
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	response, err := http.Post(os.Getenv(pipelineAuditEnv), "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("audit status %d", response.StatusCode)
	}
	return nil
}

func pipelineObserve(stage, route string, c *sdk.Context) error {
	record := pipelineRecord{Stage: stage, Route: route, Text: c.Text(), Session: string(c.Session())}
	if ev := c.Event(); ev != nil {
		record.Instance = ev.AdapterInstanceID
		record.Account = ev.AccountID
		switch payload := ev.Payload.(type) {
		case *message.CreatedPayload:
			record.MessageID = payload.PlatformMessageID
		case *message.ButtonClickedPayload:
			record.Action = payload.ActionID
			record.Value = payload.Value
		case *message.ScheduledPayload:
			record.Job = payload.JobID
			record.RunID = payload.RunID
		}
		if ev.Actor != nil {
			record.User = ev.Actor.ExternalUserID
			record.Permissions = ev.Actor.Permissions
		}
	}
	return pipelineAudit(record)
}

func pipelinePlugin(mode string) *sdk.Plugin {
	p := sdk.NewPlugin()
	switch mode {
	case "order":
		priority, err := strconv.Atoi(os.Getenv(pipelinePriorityEnv))
		if err != nil {
			panic(err)
		}
		// 逆注册顺序同时检验 route_id 的排序。
		for _, id := range []string{"z", "a"} {
			p.On(id, sdk.MessageCreated).Priority(int32(priority)).Handle(func(c *sdk.Context) error { return pipelineObserve("handler", id, c) })
		}
	case "rules":
		for _, name := range []string{"false", "error"} {
			rule := func(c *sdk.Context) (bool, error) {
				if err := pipelineObserve("rule", name, c); err != nil {
					return false, err
				}
				if name == "error" {
					return false, errors.New("rule rejected")
				}
				return false, nil
			}
			p.Command(name).Rule(rule).Handle(func(c *sdk.Context) error { return pipelineObserve("unexpected", name, c) })
			p.Keyword("text."+name, name).Rule(rule).Handle(func(c *sdk.Context) error { return pipelineObserve("unexpected", "text."+name, c) })
			p.On(name, sdk.MessageCreated).Rule(rule).Handle(func(c *sdk.Context) error { return pipelineObserve("unexpected", name, c) })
		}
		p.Raw("raw").Handle(func(c *sdk.Context) error { return pipelineObserve("raw", "raw", c) })
		p.On("observe", sdk.MessageCreated).Rule(sdk.ToMe()).Handle(func(c *sdk.Context) error { return pipelineObserve("handler", "observe", c) })
		p.Command("stop").Handle(func(c *sdk.Context) error {
			c.StopPropagation()
			return pipelineObserve("stopped", "cmd.stop", c)
		})
	case "hooks", "trusted", "buttons", "scheduled", "button-attack", "job-attack":
		pipelineHookPlugin(p, mode)

	case "wait", "control", "policy":
		pipelineConcurrentPlugin(p, mode)
	case "handlers":
		p.Command("auto").Handle(func(c *sdk.Context) (message.Content, error) {
			c.StopPropagation()
			return message.Text("automatic"), nil
		})
		for _, name := range []string{"explicit", "empty"} {
			p.Command(name).Handle(func(c *sdk.Context) (message.Content, error) {
				c.StopPropagation()
				if _, err := c.Send(message.Text("explicit")); err != nil {
					return nil, err
				}
				if name == "empty" {
					return nil, nil
				}
				return message.Text("automatic-after-explicit"), nil
			})
		}
		p.BeforeSend("before").Handle(func(_ *sdk.Context, _ message.Session, chain message.Chain) (message.Chain, error) {
			return nil, pipelineAudit(pipelineRecord{Stage: "before", Text: chain.Text()})
		})
		p.AfterSend("after").Handle(func(_ *sdk.Context, _ message.Session, result *message.SendResult, sendErr error) error {
			if sendErr != nil {
				return sendErr
			}
			return pipelineAudit(pipelineRecord{Stage: "after", MessageID: result.PrimaryPlatformMessageID})
		})
		p.Command("manual").Handle(func(c *sdk.Context) error {
			c.StopPropagation()
			for _, text := range []string{"first", "second"} {
				if _, err := c.Send(message.Text(text)); err != nil {
					return err
				}
			}
			return nil
		})
	}
	return p
}

type pipelineInput struct {
	Text, User, Token, Job  string
	Permissions             []string
	ForgeActor, ForgeTarget bool
}
type pipelineAdapter struct{}
type pipelineLoop struct {
	core     sdkadapter.Core
	mu       sync.Mutex
	sequence int
}

// Descriptor 声明回环平台的权限与测试输入动作。
func (*pipelineAdapter) Descriptor() *sdkadapter.Descriptor {
	return &sdkadapter.Descriptor{AdapterID: "loop", Permissions: []sdkadapter.PermissionDeclaration{{Name: "loop.owner", Implies: []string{"chat_admin"}, LabelKey: "permissions.owner"}}, Actions: []sdkadapter.ActionDeclaration{sdkadapter.Action[pipelineInput, struct{}]("loop.inject", 1, sdkadapter.ActionScopeChat)}}
}

// Start 将真实 Core 客户端绑定到回环账号实例。
func (*pipelineAdapter) Start(_ context.Context, _ sdkadapter.StartRequest, core sdkadapter.Core) (sdkadapter.Instance, error) {
	return &pipelineLoop{core: core}, nil
}

// Account 返回该测试平台验证的固定账号。
func (*pipelineLoop) Account() sdkadapter.Account { return sdkadapter.Account{ID: pipelineAccountID} }

// Capabilities 声明文本、按钮与编辑支持。
func (*pipelineLoop) Capabilities() []message.Capability {
	return []message.Capability{{Name: "message.text", Version: 1, Supported: true}, {Name: "message.edit", Version: 1, Supported: true}, {Name: "message.buttons", Version: 1, Supported: true}}
}

// Stop 结束无后台接收任务的回环实例。
func (*pipelineLoop) Stop(context.Context) error { return nil }

// Action 只承担测试平台输入；事件仍经 Adapter 的真实 Deliver 进入 Core。
func (i *pipelineLoop) Action(ctx context.Context, to message.Session, name string, _ uint32, params json.RawMessage) (json.RawMessage, *errcode.Error) {
	var input pipelineInput
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, errcode.InvalidArgument.New(err.Error())
	}
	ev := &message.Event{Type: "message.created", AdapterID: "loop", AdapterInstanceID: pipelineInstanceID, AccountID: pipelineAccountID, Session: to, ChatType: message.ChatGroup,
		Actor: &message.Actor{ExternalUserID: input.User, AdapterInstanceID: pipelineInstanceID, Permissions: input.Permissions}, Payload: &message.CreatedPayload{Chain: message.Chain{message.Text(input.Text)}, PlatformMessageID: "source", ToMe: true}}
	if input.Token != "" {
		ev.Type = "button.clicked"
		ev.Payload = &message.ButtonClickedPayload{ActionID: "forged-action", ActionSchemaVersion: 99, Value: input.Token, CallbackQueryID: "callback", OriginPlatformMessageID: "source"}
	}
	if input.Job != "" {
		ev.Type = "scheduled"
		ev.Payload = &message.ScheduledPayload{JobID: input.Job}
	}
	if input.ForgeActor {
		ev.Actor.AdapterInstanceID = "foreign-instance"
	}
	if input.ForgeTarget {
		ev.AdapterInstanceID = "foreign-instance"
	}
	accepted, err := i.core.Deliver(ctx, ev, nil)
	if err != nil {
		var safe *errcode.Error
		if errors.As(err, &safe) {
			return nil, safe
		}
		return nil, errcode.Internal.New(err.Error())
	}
	if !accepted {
		return nil, errcode.ResourceExhausted.New("event was not accepted")
	}
	return json.RawMessage(`{}`), nil
}

// Send 记录平台实际收到的链并返回确认回执。
func (i *pipelineLoop) Send(_ context.Context, to message.Session, chain message.Chain) (*message.SendResult, *errcode.Error) {
	return i.receipt("send", to, "", chain)
}

// Edit 保留消息 ID，记录平台实际收到的编辑链。
func (i *pipelineLoop) Edit(_ context.Context, to message.Session, id string, chain message.Chain) (*message.SendResult, *errcode.Error) {
	return i.receipt("edit", to, id, chain)
}
func (i *pipelineLoop) receipt(stage string, to message.Session, id string, chain message.Chain) (*message.SendResult, *errcode.Error) {
	i.mu.Lock()
	defer i.mu.Unlock() // 平台回执序号与审计次序共同串行化。
	i.sequence++
	if id == "" {
		id = strconv.Itoa(i.sequence)
	}
	record := pipelineRecord{Stage: stage, Text: chain.Text(), Session: string(to), MessageID: id}
	for _, segment := range chain {
		if buttons, ok := segment.(message.ButtonsSegment); ok {
			record.Token = buttons.Rows[0][0].CoreToken
		}
	}
	if err := pipelineAudit(record); err != nil {
		return nil, errcode.Internal.New(err.Error())
	}
	return &message.SendResult{Status: message.SendAccepted, Session: to, AdapterInstanceID: pipelineInstanceID, PrimaryPlatformMessageID: id, PlatformMessageIDs: []string{id}}, nil
}

type pipelineOutcome struct {
	Plugin, Route string
	Request       *pb.HandlerRequest
	Outcome       *pb.HandlerOutcome
	Code          pb.ErrorCode
}

// pipelineObserver 只记录请求与回执，所有调用仍交给真实 Session。
type pipelineObserver struct {
	router.Plugin
	harness *pipelineHarness
}

// Call 记录真实 Session 的 HandlerOutcome 与 PreEvent 返回载荷。
func (p *pipelineObserver) Call(ctx context.Context, request wire.Request) (wire.Response, error) {
	response, err := p.Plugin.Call(ctx, request)
	if request.Payload.SchemaID == validate.SchemaID(&pb.HandlerRequest{}) {
		invocation := &pb.HandlerRequest{}
		outcome := &pb.HandlerOutcome{}
		decodeErr := proto.Unmarshal(request.Payload.Bytes, invocation)
		if len(response.Payload.Bytes) > 0 {
			decodeErr = errors.Join(decodeErr, proto.Unmarshal(response.Payload.Bytes, outcome))
		}
		p.harness.mu.Lock()
		p.harness.outcomes = append(p.harness.outcomes, pipelineOutcome{Plugin: p.PluginID(), Route: request.Target, Request: invocation, Outcome: outcome, Code: response.Err.GetCode()})
		p.harness.decodeErr = errors.Join(p.harness.decodeErr, decodeErr)
		p.harness.mu.Unlock()
		select {
		case p.harness.changed <- struct{}{}:
		default:
		}
	}

	if request.Payload.SchemaID == validate.SchemaID(&pb.PreEventRequest{}) {
		result := &pb.PreEventResult{}
		decodeErr := proto.Unmarshal(response.Payload.Bytes, result)
		record := pipelineRecord{Stage: "pre-response", Plugin: p.PluginID(), Route: request.Target, Code: response.Err.GetCode().String(), Action: result.EventPayload.GetButtonClicked().GetActionId(), Value: result.EventPayload.GetButtonClicked().GetValue(), Job: result.EventPayload.GetScheduled().GetJobId()}
		p.harness.mu.Lock()
		p.harness.records = append(p.harness.records, record)
		p.harness.decodeErr = errors.Join(p.harness.decodeErr, decodeErr)
		p.harness.mu.Unlock()
	}

	return response, err
}

// pipelineHarness 拥有生产装配、子进程与审计端点；清理先停进程，再关闭端点和状态库。
type pipelineHarness struct {
	t          *testing.T
	store      *state.Store
	assembly   *assembly
	values     tune.Values
	dir        string
	server     *httptest.Server
	mu         sync.Mutex // 保护 HTTP 审计与 Session 观察者的共同记录。
	logs       []byte     // Core 日志由 slog 写入，与审计共用锁。
	records    []pipelineRecord
	outcomes   []pipelineOutcome
	decodeErr  error
	changed    chan struct{}
	release    chan struct{}
	components map[string]*runningComponent
}

func newPipeline(t *testing.T) *pipelineHarness {
	t.Helper()
	dir := t.TempDir()
	store, err := state.Open(t.Context(), filepath.Join(dir, "data"))
	if err != nil {
		t.Fatal(err)
	}
	h := &pipelineHarness{t: t, store: store, dir: dir, values: tune.Defaults(), changed: make(chan struct{}, 1), release: make(chan struct{}, 1), components: make(map[string]*runningComponent)}
	h.values.StopDrain = pipelineStopDrain
	h.assembly, err = assemble(store, dir, h.values)
	if err != nil {
		t.Fatal(err)
	}
	h.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/release" {
			select {
			case <-h.release:
				w.WriteHeader(http.StatusNoContent)
			case <-r.Context().Done():
			}
			return
		}

		var record pipelineRecord
		if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.mu.Lock()
		h.records = append(h.records, record)
		h.mu.Unlock()
		select {
		case h.changed <- struct{}{}:
		default:
		} // 多条审计合并唤醒，断言每次读取完整快照。
		w.WriteHeader(http.StatusNoContent)
	}))
	h.assembly.router.Start()
	t.Cleanup(func() {
		select {
		case h.release <- struct{}{}:
		default:
		} // 清理失败路径时唤醒尚在等待的测试 Handler。
		h.assembly.scheduler.Close()
		for _, component := range h.components {
			h.stop(component)
		}
		ctx, cancel := context.WithTimeout(context.Background(), pipelineTimeout)
		defer cancel()
		if err := h.assembly.router.Close(ctx); err != nil {
			t.Error(err)
		}
		h.server.Close()
		h.mu.Lock()
		decodeErr := h.decodeErr
		h.mu.Unlock()
		if decodeErr != nil {
			t.Error(decodeErr)
		}
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	return h
}
func (h *pipelineHarness) install(id, mode string) {
	h.t.Helper()
	role := state.RolePlugin
	adapterID := ""
	if mode == "adapter" {
		role = state.RoleAdapter
		adapterID = "loop"
	}
	if err := h.store.AddInstall(h.t.Context(), state.Install{PluginID: id, Role: role, AdapterID: adapterID, BuildDigest: bytes.Repeat([]byte{1}, 32)}); err != nil {
		h.t.Fatal(err)
	}
	if err := h.store.AddGeneration(h.t.Context(), id, pipelineGeneration+":"+id, "{}", 1); err != nil {
		h.t.Fatal(err)
	}
}
func (h *pipelineHarness) launch(id, mode string, priority int) *runningComponent {
	h.t.Helper()
	// 1. 复用持久化安装身份和代次，准备进程目录。
	install, err := h.store.GetInstall(h.t.Context(), id)
	if err != nil {
		h.t.Fatal(err)
	}
	active, err := h.store.Active(h.t.Context(), id)
	if err != nil {
		h.t.Fatal(err)
	}
	pluginDir := filepath.Join(h.dir, id)
	if err := os.MkdirAll(pluginDir, 0700); err != nil {
		h.t.Fatal(err)
	}
	c := &runningComponent{id: id, install: install, caller: router.NewCaller(id, pipelineGeneration+":"+id)}
	// 2. 通过生产 Handler 绑定启动真实插件或 Adapter。
	handler := h.assembly.handler(c, install)
	c.process, err = supervisor.Launch(h.t.Context(), h.store, supervisor.Spec{PluginID: id, Role: install.Role, AdapterID: install.AdapterID, PluginDir: pluginDir, Executable: os.Args[0], BuildDigest: install.BuildDigest, GenerationID: pipelineGeneration + ":" + id,
		Bootstrap: bootstrap(pipelineGeneration+":"+id, 1), NewCandidate: active == "", Limits: host.NewLimits(h.values), Stop: host.NewStopTiming(h.values), Log: h, Handler: handler, Global: h.assembly.global,
		ExtraEnv: map[string]string{pipelineModeEnv: mode, pipelineIDEnv: id, pipelinePriorityEnv: strconv.Itoa(priority), pipelineAuditEnv: h.server.URL, "GORACE": "atexit_sleep_ms=0"}})
	if err != nil {
		h.t.Fatal(err)
	}
	// 3. 接入生产装配，观察者只记录原 Session 请求和结果。
	c.instance = c.process.Instance
	h.components[id] = c
	if err := h.assembly.attach(h.t.Context(), c); err != nil {
		h.t.Fatal(err)
	}
	observed := &pipelineObserver{Plugin: router.FromInstance(c.instance), harness: h}
	h.assembly.router.Attach(observed, c.caller)
	if mode == "adapter" {
		if err := c.remote.StartInstance(h.t.Context(), pipelineInstanceID, []byte(`{}`)); err != nil {
			h.t.Fatal(err)
		}
		h.idle()
	}
	return c
}
func (h *pipelineHarness) stop(c *runningComponent) {
	h.assembly.detach(c)
	if result := c.process.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE); result.Err != nil {
		h.t.Error(result.Err)
	}
	delete(h.components, c.id)
}

// Line 将子进程日志交给当前测试保留完整诊断。
func (h *pipelineHarness) Line(id, stream string, line []byte) {
	h.t.Logf("%s %s: %s", id, stream, line)
}

// Write 保存 Core 的结构化日志，供失败链断言检查 WARN 与错误码。
func (h *pipelineHarness) Write(data []byte) (int, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.logs = append(h.logs, data...)
	return len(data), nil
}

// SetSession 满足 Supervisor 日志接口；当前记录按插件 ID 区分。
func (h *pipelineHarness) SetSession([]byte) {}
func (h *pipelineHarness) snapshot() []pipelineRecord {
	h.mu.Lock()
	defer h.mu.Unlock()
	return slices.Clone(h.records)
}
func (h *pipelineHarness) await(predicate func([]pipelineRecord) bool) {
	h.t.Helper()
	timer := time.NewTimer(pipelineTimeout)
	defer timer.Stop()
	for {
		if predicate(h.snapshot()) {
			return
		}
		select {
		case <-h.changed:
		case <-timer.C:
			h.t.Fatalf("audit timed out: %+v", h.snapshot())
		}
	}
}
func (h *pipelineHarness) idle() {
	h.t.Helper()
	timer := time.NewTimer(pipelineTimeout)
	defer timer.Stop()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		stats := h.assembly.router.Stats()
		if stats.Running == 0 && stats.Queued == 0 {
			return
		}
		select {
		case <-ticker.C:
		case <-timer.C:
			h.t.Fatalf("router did not drain: %+v", stats)
		}
	}
}
func (h *pipelineHarness) deliver(input pipelineInput) *pb.ErrorInfo {
	h.t.Helper()
	instance, ok := h.assembly.adapters.Get(pipelineInstanceID)
	if !ok {
		h.t.Fatal("missing loop instance")
	}
	params, err := json.Marshal(input)
	if err != nil {
		h.t.Fatal(err)
	}
	// Core 直接操作测试平台入口；插件发送永远走生产 message 服务。
	executor := instance.(coreadapter.ActionExecutor)
	_, safe := executor.Action(h.t.Context(), &pb.Target{AdapterId: "loop", AdapterInstanceId: pipelineInstanceID, AccountId: pipelineAccountID, ChatId: "group", ThreadId: proto.String("topic")}, "loop.inject", 1, string(params))
	return safe
}

// TestPipeline 通过生产装配与真实子进程验证事件流水线。
func TestPipeline(t *testing.T) {
	t.Run("priority", func(t *testing.T) {
		for _, order := range [][]string{{"beta", "alpha", "early"}, {"early", "alpha", "beta"}} {
			t.Run(order[0], func(t *testing.T) {
				h := newPipeline(t)
				for _, id := range order {
					h.install(id, "order")
				}
				h.install("platform", "adapter")
				h.launch("platform", "adapter", 0)
				for i := len(order) - 1; i >= 0; i-- {
					priority := 0
					if order[i] == "early" {
						priority = -1
					}
					h.launch(order[i], "order", priority)
				}
				h.inject(pipelineInput{Text: "order", User: "alice"})
				h.idle()
				var got []string
				for _, record := range h.snapshot() {
					if record.Stage == "handler" {
						got = append(got, record.Plugin+"/"+record.Route)
					}
				}
				want := []string{"early/a", "early/z", "alpha/a", "alpha/z", "beta/a", "beta/z"}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("order=%v want=%v", got, want)
				}
			})
		}
	})
	t.Run("rules", func(t *testing.T) {
		h := newPipeline(t)
		h.install("platform", "adapter")
		h.launch("platform", "adapter", 0)
		h.install("rules", "rules")
		h.launch("rules", "rules", 0)
		for _, text := range []string{"ordinary false", "ordinary error", "/false", "/error"} {
			h.inject(pipelineInput{Text: text, User: "alice"})
			h.idle()
		}
		for _, record := range h.snapshot() {
			if record.Stage == "unexpected" {
				t.Fatalf("Rule invoked Handler: %+v", record)
			}
			if record.Stage == "raw" && record.Text != "ordinary false" && record.Text != "ordinary error" {
				t.Fatalf("consumed command reached Raw: %+v", record)
			}
		}

		if len(pipelineStages(h.snapshot(), "raw")) != 2 {
			t.Fatal("ordinary input did not reach Raw")
		}
		h.mu.Lock()
		outcomes := slices.Clone(h.outcomes)
		h.mu.Unlock()
		counts := map[string]int{}
		for _, call := range outcomes {
			switch call.Route {
			case "false", "error", "cmd.false", "cmd.error", "text.false", "text.error":
				counts[call.Route]++
				wantCode := pb.ErrorCode_ERROR_CODE_UNSPECIFIED
				if call.Route == "error" || call.Route == "cmd.error" || call.Route == "text.error" {
					wantCode = pb.ErrorCode_ERROR_CODE_INTERNAL
				}
				if call.Outcome.GetMatched() || call.Code != wantCode || call.Outcome.GetPropagation() != pb.Propagation_PROPAGATION_CONTINUE {
					t.Fatalf("Rule outcome=%+v", call)
				}
				if (call.Route == "text.false" || call.Route == "text.error") && strings.HasPrefix(call.Request.Event.GetEventPayload().GetMessageCreated().GetChain().GetSegments()[0].GetText().GetText(), "/") {
					t.Fatalf("consumed command reached text matching: %+v", call)
				}
			}
		}
		wantCounts := map[string]int{"false": 4, "error": 4, "cmd.false": 1, "cmd.error": 1, "text.false": 1, "text.error": 1}
		if !reflect.DeepEqual(counts, wantCounts) {
			t.Fatalf("Rule calls=%v want=%v", counts, wantCounts)
		}

		t.Run("consumed_event_route", func(t *testing.T) {
			got := pipelineStages(h.snapshot(), "handler")
			want := []string{"handler:ordinary false", "handler:ordinary error", "handler:/false", "handler:/error"}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("event propagation=%v want=%v", got, want)
			}
			h.inject(pipelineInput{Text: "/stop", User: "alice"})
			h.idle()
			if !reflect.DeepEqual(pipelineStages(h.snapshot(), "handler"), want) || !reflect.DeepEqual(pipelineStages(h.snapshot(), "stopped"), []string{"stopped:/stop"}) || len(pipelineStages(h.snapshot(), "raw")) != 2 {
				t.Fatalf("StopPropagation did not consume event: %+v", h.snapshot())
			}
		})
	})
	t.Run("waiting", pipelineWaitingTest)
	t.Run("control_frames", pipelineControlTest)
	t.Run("permissions", pipelinePolicyTest)
	t.Run("hooks", pipelineHooksTest)
	t.Run("trusted", pipelineTrustedTest)
	t.Run("buttons", pipelineButtonsTest)
	t.Run("scheduled", pipelineScheduledTest)

	t.Run("handlers", func(t *testing.T) {
		h := newPipeline(t)
		h.install("platform", "adapter")
		h.launch("platform", "adapter", 0)
		h.install("handlers", "handlers")
		h.launch("handlers", "handlers", 0)
		for _, text := range []string{"/auto", "/manual"} {
			h.inject(pipelineInput{Text: text, User: "alice"})
			h.idle()
		}
		var got []string
		for _, record := range h.snapshot() {
			if record.Stage == "send" {
				got = append(got, record.Text)
			}
		}
		if !reflect.DeepEqual(got, []string{"automatic", "first", "second"}) {
			t.Fatalf("sends=%v", got)
		}

		h.mu.Lock()
		outcomes := slices.Clone(h.outcomes)
		h.mu.Unlock()
		if len(outcomes) != 2 {
			t.Fatalf("Handler outcomes=%+v", outcomes)
		}
		for _, call := range outcomes {
			if call.Code != pb.ErrorCode_ERROR_CODE_UNSPECIFIED || !call.Outcome.GetMatched() || call.Outcome.GetPropagation() != pb.Propagation_PROPAGATION_STOP {
				t.Fatalf("Handler outcome=%+v", call)
			}
			if (call.Route == "cmd.auto") != (len(call.Outcome.GetChain().GetSegments()) > 0) {
				t.Fatalf("Handler reply chain=%+v", call)
			}
		}
		t.Run("automatic_explicit_send", func(t *testing.T) {
			before := len(h.snapshot())
			for _, text := range []string{"/explicit", "/empty"} {
				h.inject(pipelineInput{Text: text, User: "alice"})
				h.idle()
			}
			records := h.snapshot()[before:]
			got := pipelineStages(records, "before", "send", "after")
			want := []string{"before:explicit", "send:explicit", "after:", "before:automatic-after-explicit", "send:automatic-after-explicit", "after:", "before:explicit", "send:explicit", "after:"}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("explicit and automatic Hook sequence=%v want=%v", got, want)
			}
			var sent, hooked []string
			for _, record := range records {
				if record.Stage == "send" {
					sent = append(sent, record.MessageID)
				}
				if record.Stage == "after" {
					hooked = append(hooked, record.MessageID)
				}
			}
			if !reflect.DeepEqual(sent, hooked) {
				t.Fatalf("AfterSend receipts=%v sends=%v", hooked, sent)
			}
		})
	})
}

func (h *pipelineHarness) inject(input pipelineInput) {
	h.t.Helper()
	if safe := h.deliver(input); safe != nil {
		h.t.Fatal(safe)
	}
}

type pipelineButtonValue struct {
	Choice string `json:"choice"`
}

var pipelineAction = sdk.Action[pipelineButtonValue]("choose")

func pipelineHookPlugin(p *sdk.Plugin, mode string) {
	if mode == "scheduled" || mode == "job-attack" {
		p.Every("pulse", time.Second).Handle(func(c *sdk.Context) error {
			return pipelineObserve("tick", "pulse", c)
		})
		if mode == "job-attack" {
			p.PreEvent("pre.job", sdk.Scheduled).Handle(func(c *sdk.Context, payload *message.ScheduledPayload) (*message.ScheduledPayload, error) {
				if err := pipelineObserve("pre", "pre.job", c); err != nil {
					return nil, err
				}
				payload.JobID = "forged-job"
				payload.RunID = "forged-run"
				return payload, nil
			})
		}
		p.On("scheduled", sdk.Scheduled).Handle(func(c *sdk.Context) error { return pipelineObserve("scheduled", "scheduled", c) })
		return
	}
	if mode == "buttons" || mode == "button-attack" {
		p.Command("keyboard").Handle(func(c *sdk.Context) error {
			c.StopPropagation()
			_, err := c.Send(message.Chain{message.Text("keyboard"), message.Buttons([]message.Button{pipelineAction.Button("Choose", pipelineButtonValue{Choice: "kept"})})})
			return err
		})
		if mode == "button-attack" {
			p.PreEvent("pre.button", sdk.ButtonClicked).Handle(func(c *sdk.Context, payload *message.ButtonClickedPayload) (*message.ButtonClickedPayload, error) {
				if err := pipelineObserve("pre", "pre.button", c); err != nil {
					return nil, err
				}
				payload.ActionID = "other"
				payload.ActionSchemaVersion = 99
				payload.Value = `{"choice":"forged"}`
				return payload, nil
			})
		}
		sdk.OnButton(p, pipelineAction).Permission(sdk.User).Handle(func(c *sdk.Context, value pipelineButtonValue) error {
			if value.Choice != "kept" {
				return errors.New("button binding changed")
			}
			return pipelineObserve("button", "button.choose", c)
		})
		p.Raw("raw").Handle(func(c *sdk.Context) error { return pipelineObserve("raw", "raw", c) })
		return
	}
	p.PreEvent("pre.message", sdk.MessageCreated).Handle(func(c *sdk.Context, payload *message.CreatedPayload) (*message.CreatedPayload, error) {
		if err := pipelineObserve("pre", "pre.message", c); err != nil {
			return nil, err
		}
		if mode == "hooks" && payload.Chain.Text() == "rewrite" {
			payload.Chain = message.Chain{message.Text("/echo rewritten")}
		}
		if mode == "trusted" {
			ev := c.Event()
			ev.Actor.ExternalUserID = "mallory"
			ev.Actor.Permissions = []string{"admin"}
			ev.Session = "loop:foreign"
			payload.PlatformMessageID = "forged-source"
		}
		return payload, nil
	})
	p.On("observe", sdk.MessageCreated).Handle(func(c *sdk.Context) error { return pipelineObserve("handler", "observe", c) })
	if mode == "trusted" {
		return
	}
	p.Command("echo").Args("value").Handle(func(c *sdk.Context, value string) error {
		c.StopPropagation()
		if err := pipelineObserve("handler", "cmd.echo", c); err != nil {
			return err
		}
		if value == "reject" {
			_, err := c.Send(message.Text(value))
			if errcode.Of(err) != errcode.PermissionDenied {
				return fmt.Errorf("rejection error: %v", err)
			}
			return pipelineAudit(pipelineRecord{Stage: "rejected", Code: errcode.PermissionDenied.String()})
		}
		result, err := c.Send(message.Text(value))
		if err != nil {
			return err
		}
		if value == "edit" {
			_, err = c.Edit(result.Ref(), message.Text("edited"))
			return err
		}
		_, err = c.Send(message.Text(value + "-again"))
		return err
	})
	if mode == "hooks" {
		p.Command("automatic").Handle(func(c *sdk.Context) (message.Content, error) {
			c.StopPropagation()
			return message.Text("automatic"), nil
		})
		p.BeforeSend("before").Handle(func(c *sdk.Context, to message.Session, chain message.Chain) (message.Chain, error) {
			if err := pipelineAudit(pipelineRecord{Stage: "before", Text: chain.Text(), Session: string(to)}); err != nil {
				return nil, err
			}
			// 出站 Hook 的递归 Send 必须由 Core 拒绝，使用同一 Session 的实际请求取证。
			if _, err := c.Send(message.Text("recursive")); errcode.Of(err) != errcode.FailedPrecondition {
				return nil, fmt.Errorf("recursive send error: %v", err)
			}
			if chain.Text() == "reject" {
				return nil, errcode.PermissionDenied.New("send rejected by hook")
			}
			return message.Chain{message.Text("before:" + chain.Text())}, nil
		})
		p.AfterSend("after").Handle(func(_ *sdk.Context, to message.Session, result *message.SendResult, sendErr error) error {
			record := pipelineRecord{Stage: "after", Session: string(to)}
			if sendErr != nil {
				record.Code = sendErr.Error()
			}
			if result != nil {
				record.MessageID = result.PrimaryPlatformMessageID
			}
			if err := pipelineAudit(record); err != nil {
				return err
			}
			return errors.New("after-send diagnostic")
		})
	}
}

func pipelineWithPlugin(t *testing.T, mode string) *pipelineHarness {
	t.Helper()
	h := newPipeline(t)
	h.install("platform", "adapter")
	h.launch("platform", "adapter", 0)
	h.install("pipeline", mode)
	h.launch("pipeline", mode, 0)
	return h
}
func pipelineStages(records []pipelineRecord, stages ...string) []string {
	var result []string
	for _, record := range records {
		if slices.Contains(stages, record.Stage) {
			result = append(result, record.Stage+":"+record.Text)
		}
	}
	return result
}
func pipelineHooksTest(t *testing.T) {
	h := pipelineWithPlugin(t, "hooks")
	for _, text := range []string{"rewrite", "/echo edit", "/echo reject", "/automatic"} {
		h.inject(pipelineInput{Text: text, User: "alice"})
		h.idle()
	}
	got := pipelineStages(h.snapshot(), "before", "send", "edit", "after", "rejected")
	want := []string{"before:rewritten", "send:before:rewritten", "after:", "before:rewritten-again", "send:before:rewritten-again", "after:", "before:edit", "send:before:edit", "after:", "before:edited", "edit:before:edited", "after:", "before:reject", "rejected:", "before:automatic", "send:before:automatic", "after:"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("hook sequence=%v want=%v", got, want)
	}
	var rewritten bool
	for _, record := range h.snapshot() {
		if record.Stage == "handler" && record.Text == "/echo rewritten" {
			rewritten = true
		}
	}
	if !rewritten {
		t.Fatal("PreEvent rewrite did not reach command Handler")
	}
}
func pipelineTrustedTest(t *testing.T) {
	t.Run("button_and_job_pre_event", pipelineBoundPayloadTest)
	h := pipelineWithPlugin(t, "trusted")
	h.inject(pipelineInput{Text: "trusted", User: "alice", Permissions: []string{"admin", "user"}})
	h.idle()
	var observed bool
	for _, record := range h.snapshot() {
		if record.Stage == "handler" {
			observed = true
			if record.User != "alice" || record.Session != "loop:group#topic" || record.Instance != pipelineInstanceID || record.Account != pipelineAccountID || record.MessageID != "source" || !reflect.DeepEqual(record.Permissions, []string{"user"}) {
				t.Fatalf("trusted fields changed: %+v", record)
			}
		}
	}
	if !observed {
		t.Fatal("missing trusted Handler")
	}
	for _, input := range []pipelineInput{{Text: "actor", User: "alice", ForgeActor: true}, {Text: "target", User: "alice", ForgeTarget: true}} {
		if safe := h.deliver(input); safe.GetCode() != pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT {
			t.Fatalf("forged identity error=%v", safe)
		}
	}
	h.idle()
	if len(pipelineStages(h.snapshot(), "handler")) != 1 {
		t.Fatal("forged identity reached Handler")
	}
}

// pipelineBoundPayloadTest 检查无业务载荷事件的 Hook 拒绝与移除 Hook 后的恢复。
func pipelineBoundPayloadTest(t *testing.T) {
	for _, mode := range []string{"button-attack", "job-attack"} {
		t.Run(mode, func(t *testing.T) {
			h := pipelineWithPlugin(t, mode)
			previous, writer, flags := slog.Default(), log.Writer(), log.Flags()
			slog.SetDefault(slog.New(slog.NewTextHandler(h, &slog.HandlerOptions{Level: slog.LevelWarn})))
			defer func() { slog.SetDefault(previous); log.SetOutput(writer); log.SetFlags(flags) }()
			var token, hook, stage, recoveredMode string
			var event *pb.Event
			if mode == "button-attack" {
				h.inject(pipelineInput{Text: "/keyboard", User: "alice"})
				h.idle()
				for _, record := range h.snapshot() {
					if record.Stage == "send" {
						token = record.Token
					}
				}
				if token == "" {
					t.Fatal("missing button token")
				}
				hook, stage, recoveredMode = "pre.button", "button", "buttons"
			} else {
				c := h.components["pipeline"]
				h.assembly.scheduler.Attach(&pipelineObserver{Plugin: router.FromInstance(c.instance), harness: h}, time.Now())
				h.await(func(_ []pipelineRecord) bool { h.mu.Lock(); defer h.mu.Unlock(); return len(h.outcomes) > 0 })
				h.assembly.scheduler.Drain(c.id)
				h.mu.Lock()
				event = proto.CloneOf(h.outcomes[0].Request.Event)
				h.mu.Unlock()
				hook, stage, recoveredMode = "pre.job", "scheduled", "scheduled"
			}
			h.mu.Lock()
			before := len(h.outcomes)
			h.mu.Unlock()
			if event == nil {
				h.inject(pipelineInput{User: "alice", Token: token})
			} else {
				// Scheduler 直接调用任务；将真实 tick 的事件重交 Router，单独验证入站 PreEvent。
				if err := h.assembly.router.Deliver(t.Context(), event); err != nil {
					t.Fatal(err)
				}
			}
			h.idle()
			h.mu.Lock()
			after, logs := len(h.outcomes), string(h.logs)
			h.mu.Unlock()
			if after != before {
				t.Fatalf("rejected payload reached Handler: before=%d after=%d", before, after)
			}
			if len(pipelineStages(h.snapshot(), stage)) != 0 {
				t.Fatalf("forged binding reached plugin: %+v", h.snapshot())
			}
			returned := slices.ContainsFunc(h.snapshot(), func(record pipelineRecord) bool {
				return record.Stage == "pre-response" && record.Route == hook && record.Code == pb.ErrorCode_ERROR_CODE_UNSPECIFIED.String() && ((record.Action == "other" && record.Value == `{"choice":"forged"}`) || record.Job == "forged-job")
			})
			warned := false
			for _, line := range strings.Split(logs, "\n") {
				if strings.Contains(line, "level=WARN") && strings.Contains(line, "pre-event hook failed") && strings.Contains(line, "route_id="+hook) && strings.Contains(line, "INVALID_ARGUMENT") {
					warned = true
				}
			}
			if !returned || !warned {
				t.Fatalf("missing malicious payload or Core WARN: records=%+v logs=%s", h.snapshot(), logs)
			}

			// 同一插件以无恶意 Hook 的注册重启，保留原按钮 token 与任务 ID。
			h.stop(h.components["pipeline"])
			c := h.launch("pipeline", recoveredMode, 0)
			if event == nil {
				h.inject(pipelineInput{User: "alice", Token: token})
				h.idle()
			} else {
				h.assembly.scheduler.Attach(&pipelineObserver{Plugin: router.FromInstance(c.instance), harness: h}, time.Now())
				h.await(func(_ []pipelineRecord) bool { h.mu.Lock(); defer h.mu.Unlock(); return len(h.outcomes) > before })
				h.assembly.scheduler.Drain(c.id)
			}
			h.mu.Lock()
			calls := slices.Clone(h.outcomes[before:])
			h.mu.Unlock()
			if len(calls) != 1 || calls[0].Plugin != "pipeline" || calls[0].Code != pb.ErrorCode_ERROR_CODE_UNSPECIFIED || !calls[0].Outcome.GetMatched() {
				t.Fatalf("recovered Handler=%+v", calls)
			}
			payload := calls[0].Request.Event.GetEventPayload()
			if event == nil {
				button := payload.GetButtonClicked()
				if calls[0].Route != "button.choose" || button.GetActionId() != "choose" || button.GetValue() != `{"choice":"kept"}` || !reflect.DeepEqual(pipelineStages(h.snapshot(), "button"), []string{"button:"}) {
					t.Fatalf("recovered button=%+v", calls[0])
				}
			} else {
				job := payload.GetScheduled()
				if calls[0].Route != "job.pulse" || job.GetJobId() != "pulse" || job.GetRunId() == "" || job.GetRunId() == event.GetEventPayload().GetScheduled().GetRunId() {
					t.Fatalf("recovered tick=%+v", calls[0])
				}
			}
		})
	}
}

func pipelineButtonsTest(t *testing.T) {
	h := pipelineWithPlugin(t, "buttons")
	h.inject(pipelineInput{Text: "/keyboard", User: "alice"})
	h.idle()
	var token string
	for _, record := range h.snapshot() {
		if record.Stage == "send" {
			token = record.Token
		}
	}
	if len(token) != 43 {
		t.Fatalf("button token=%q", token)
	}
	h.install("z-other", "buttons")
	h.launch("z-other", "buttons", 0)
	for _, user := range []string{"alice", "bob"} {
		h.inject(pipelineInput{User: user, Token: token})
		h.idle()
	}
	var users []string
	for _, record := range h.snapshot() {
		if record.Stage == "button" {
			users = append(users, record.User)
			if record.Plugin != "pipeline" || record.Action != "choose" || record.Value != `{"choice":"kept"}` {
				t.Fatalf("button binding=%+v", record)
			}
		}
	}
	if !reflect.DeepEqual(users, []string{"alice", "bob"}) {
		t.Fatalf("button users=%v", users)
	}
	// 同一个 token 使用当前点击者的权限，再次点击时立即应用管理覆盖。
	p := policy.New(h.store)
	if err := p.SetOverride(t.Context(), "pipeline", policy.Scope{Kind: policy.ScopeRoute, RouteID: "button.choose"}, policy.Override{Permissions: []string{"chat_admin"}}); err != nil {
		t.Fatal(err)
	}
	h.inject(pipelineInput{User: "bob", Token: token})
	h.idle()
	if len(pipelineStages(h.snapshot(), "button")) != 2 {
		t.Fatal("button reused sender permissions")
	}
	h.inject(pipelineInput{User: "alice", Token: token, Permissions: []string{"loop.owner"}})
	h.idle()
	if len(pipelineStages(h.snapshot(), "button")) != 3 {
		t.Fatal("current clicker permission was not applied")
	}
	record, err := h.store.GetButtonToken(t.Context(), token)
	if err != nil {
		t.Fatal(err)
	}
	record.Token = strings.Repeat("x", 43)
	record.ExpiresAt = time.Now().Add(-time.Second).UnixMilli()
	if err := h.store.PutButtonTokens(t.Context(), []state.ButtonToken{record}); err != nil {
		t.Fatal(err)
	}
	h.inject(pipelineInput{User: "alice", Token: record.Token, Permissions: []string{"loop.owner"}})
	h.idle()
	if len(pipelineStages(h.snapshot(), "button")) != 3 || len(pipelineStages(h.snapshot(), "raw")) != 0 {
		t.Fatalf("rejected button fell back: %+v", h.snapshot())
	}
}
func pipelineScheduledTest(t *testing.T) {
	h := pipelineWithPlugin(t, "scheduled")
	if safe := h.deliver(pipelineInput{User: "alice", Job: "forged-job"}); safe.GetCode() != pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT {
		t.Fatalf("forged scheduled error=%v", safe)
	}
	c := h.components["pipeline"]
	h.assembly.scheduler.Attach(&pipelineObserver{Plugin: router.FromInstance(c.instance), harness: h}, time.Now())
	h.await(func(records []pipelineRecord) bool { return len(pipelineStages(records, "tick")) > 0 })
	h.assembly.scheduler.Drain(c.id)
	h.await(func(_ []pipelineRecord) bool { h.mu.Lock(); defer h.mu.Unlock(); return len(h.outcomes) > 0 })
	h.mu.Lock()
	outcomes := slices.Clone(h.outcomes)
	h.mu.Unlock()
	for _, call := range outcomes {
		ev := call.Request.Event
		if ev.GetEventPayload().GetScheduled().GetJobId() != "pulse" || ev.GetSourceKind() != pb.EventSourceKind_EVENT_SOURCE_KIND_SCHEDULER || ev.Actor != nil || ev.Target != nil || call.Code != pb.ErrorCode_ERROR_CODE_UNSPECIFIED || !call.Outcome.GetMatched() {
			t.Fatalf("Scheduler invocation=%+v", call)
		}
	}
	for _, record := range h.snapshot() {
		if record.Stage == "tick" && (record.Job != "pulse" || record.RunID == "" || record.User != "" || record.Session != "") {
			t.Fatalf("Scheduler identity=%+v", record)
		}
	}
}

const pipelineSlowDuration = 2 * time.Second

func pipelineConcurrentPlugin(p *sdk.Plugin, mode string) {
	switch mode {
	case "wait":
		p.Command("wait").Handle(func(c *sdk.Context) error {
			c.StopPropagation()
			return c.Continue(func(c *sdk.Context) error {
				type waited struct {
					next *sdk.Context
					err  error
				}
				results := make(chan waited, 2)
				// 两个同键请求中只有一个登记成功；ALREADY_EXISTS 的回执证明另一个已进入等待表。
				for range 2 {
					go func() { next, err := c.Wait(pipelineTimeout); results <- waited{next, err} }()
				}
				first := <-results
				if errcode.Of(first.err) != errcode.AlreadyExists {
					return fmt.Errorf("wait readiness: %v", first.err)
				}
				if err := pipelineObserve("ready", "cmd.wait", c); err != nil {
					return err
				}
				result := <-results
				if result.err != nil {
					return pipelineAudit(pipelineRecord{Stage: "wait-error", User: c.Event().Actor.ExternalUserID, Code: errcode.Of(result.err).String()})
				}
				return pipelineObserve("wait", "cmd.wait", result.next)
			})
		})
		p.Command("timeout").Handle(func(c *sdk.Context) error {
			c.StopPropagation()
			return c.Continue(func(c *sdk.Context) error {
				_, err := c.Wait(pipelineWaitTimeout)
				if errcode.Of(err) != errcode.DeadlineExceeded {
					return fmt.Errorf("wait timeout: %v", err)
				}
				return pipelineAudit(pipelineRecord{Stage: "timeout", Code: errcode.Of(err).String()})
			})
		})
		p.On("ordinary", sdk.MessageCreated).Handle(func(c *sdk.Context) error { return pipelineObserve("ordinary", "ordinary", c) })
	case "control":
		p.On("slow", sdk.MessageCreated).Rule(func(c *sdk.Context) (bool, error) { return c.Text() == "slow", nil }).Handle(func(c *sdk.Context) error {
			if err := pipelineObserve("slow-start", "slow", c); err != nil {
				return err
			}
			timer := time.NewTimer(pipelineSlowDuration)
			defer timer.Stop()
			select {
			case <-timer.C:
				return pipelineObserve("slow-end", "slow", c)
			case <-c.Context().Done():
				return c.Context().Err()
			}
		})
		p.On("fast", sdk.MessageCreated).Rule(sdk.ToMe(), func(c *sdk.Context) (bool, error) { return c.Text() == "fast", nil }).Handle(func(c *sdk.Context) error {
			if _, err := c.Send(message.Text("fast-response")); err != nil {
				return err
			}
			return pipelineObserve("fast", "fast", c)
		})
		p.On("cancel", sdk.MessageCreated).Rule(func(c *sdk.Context) (bool, error) { return c.Text() == "cancel", nil }).Handle(func(c *sdk.Context) error {
			if err := pipelineObserve("cancel-start", "cancel", c); err != nil {
				return err
			}
			<-c.Context().Done()
			return pipelineAudit(pipelineRecord{Stage: "cancelled", Code: c.Context().Err().Error()})
		})
	case "policy":
		p.Command("admin").Permission(sdk.Admin).Handle(func(c *sdk.Context) error { c.StopPropagation(); return pipelineObserve("admin", "cmd.admin", c) })
		group := p.Group("secure").Permission(sdk.Admin)
		group.Command("leaf").Handle(func(c *sdk.Context) error {
			c.StopPropagation()
			return pipelineObserve("allowed", "cmd.secure.leaf", c)
		})
		group.Command("hold").Permission(sdk.User).Handle(func(c *sdk.Context) error {
			c.StopPropagation()
			if err := pipelineObserve("policy-start", "cmd.secure.hold", c); err != nil {
				return err
			}
			request, err := http.NewRequestWithContext(c.Context(), http.MethodGet, os.Getenv(pipelineAuditEnv)+"/release", nil)
			if err != nil {
				return err
			}
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				return err
			}
			if err := response.Body.Close(); err != nil {
				return err
			}
			if response.StatusCode != http.StatusNoContent {
				return fmt.Errorf("release status %d", response.StatusCode)
			}
			if _, err := c.Send(message.Text("policy-complete")); err != nil {
				return err
			}
			return pipelineObserve("policy-end", "cmd.secure.hold", c)
		})
		p.On("identity", sdk.MessageCreated).Handle(func(c *sdk.Context) error { return pipelineObserve("identity", "identity", c) })
	}
}

func pipelineWaitingTest(t *testing.T) {
	// 1. 两个用户登记等待，再按相反次序交付消息，检查消费与身份。
	h := pipelineWithPlugin(t, "wait")
	for _, user := range []string{"alice", "bob"} {
		h.inject(pipelineInput{Text: "/wait", User: user})
	}
	h.await(func(records []pipelineRecord) bool { return len(pipelineStages(records, "ready")) == 2 })
	for _, user := range []string{"bob", "alice"} {
		h.inject(pipelineInput{Text: "reply-" + user, User: user})
	}
	h.idle()
	h.await(func(records []pipelineRecord) bool { return len(pipelineStages(records, "wait")) == 2 })
	got := map[string]string{}
	for _, record := range h.snapshot() {
		if record.Stage == "wait" {
			got[record.User] = record.Text
		}
		if record.Stage == "ordinary" {
			t.Fatalf("waited message entered ordinary dispatch: %+v", record)
		}
	}
	if !reflect.DeepEqual(got, map[string]string{"alice": "reply-alice", "bob": "reply-bob"}) {
		t.Fatalf("wait results=%v", got)
	}
	// 2. 超时只结束等待，停用释放 Session 内的等待，重启后只处理新事件。
	h.inject(pipelineInput{Text: "/timeout", User: "timeout-user"})
	h.await(func(records []pipelineRecord) bool { return len(pipelineStages(records, "timeout")) == 1 })
	h.inject(pipelineInput{Text: "/wait", User: "stopped-user"})
	h.await(func(records []pipelineRecord) bool { return len(pipelineStages(records, "ready")) == 3 })
	h.stop(h.components["pipeline"])
	h.inject(pipelineInput{Text: "offline-reply", User: "stopped-user"})
	h.idle()
	h.launch("pipeline", "wait", 0)
	h.inject(pipelineInput{Text: "fresh-message", User: "stopped-user"})
	h.idle()
	records := h.snapshot()
	if len(pipelineStages(records, "wait")) != 2 {
		t.Fatalf("stopped wait was replayed: %+v", records)
	}
	ordinary := pipelineStages(records, "ordinary")
	if !reflect.DeepEqual(ordinary, []string{"ordinary:fresh-message"}) {
		t.Fatalf("restart dispatch=%v", ordinary)
	}
}

func pipelineControlTest(t *testing.T) {
	// 1. 慢 Handler 执行期间，另一事件仍能完成 Send 及其 Response。
	h := pipelineWithPlugin(t, "control")
	h.inject(pipelineInput{Text: "slow", User: "alice"})
	h.await(func(records []pipelineRecord) bool { return len(pipelineStages(records, "slow-start")) == 1 })
	h.inject(pipelineInput{Text: "fast", User: "bob"})
	h.await(func(records []pipelineRecord) bool { return len(pipelineStages(records, "fast")) == 1 })
	// 2. 通过同一真实 Session 发第二个 Handler REQUEST，再发 Cancel，核对对端实际观察到取消。
	request := &pb.HandlerRequest{RouteId: "cancel", Context: &pb.InvocationContext{}, Event: &pb.Event{EventType: "message.created", EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_MessageCreated{MessageCreated: &pb.MessageCreatedPayload{Chain: &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "cancel"}}}}}}}}}}
	encoded, err := proto.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	finished := make(chan error, 1)
	go func() {
		_, err := h.components["pipeline"].instance.Session.Call(ctx, wire.Request{Target: "cancel", Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: 1, Bytes: encoded}})
		finished <- err
	}()
	h.await(func(records []pipelineRecord) bool { return len(pipelineStages(records, "cancel-start")) == 1 })
	cancel()
	select {
	case err := <-finished:
		if wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_CANCELLED {
			t.Fatalf("cancel error=%v", err)
		}
	case <-time.After(pipelineTimeout):
		t.Fatal("Cancel did not complete")
	}
	h.await(func(records []pipelineRecord) bool { return len(pipelineStages(records, "cancelled")) == 1 })
	records := h.snapshot()
	if len(pipelineStages(records, "slow-end")) != 0 {
		t.Fatalf("control frames waited for slow Handler: %+v", records)
	}
	if !reflect.DeepEqual(pipelineStages(records, "send"), []string{"send:fast-response"}) {
		t.Fatalf("fast Response=%+v", records)
	}
	h.idle()
	if len(pipelineStages(h.snapshot(), "slow-end")) != 1 {
		t.Fatal("slow Handler did not finish")
	}
}

func pipelinePolicyTest(t *testing.T) {
	h := pipelineWithPlugin(t, "policy")
	p := policy.New(h.store)
	owner := []string{"loop.owner", "admin"}
	h.inject(pipelineInput{Text: "/admin", User: "alice", Permissions: owner})
	h.idle()
	var identity bool
	for _, record := range h.snapshot() {
		if record.Stage == "admin" {
			t.Fatalf("chat_admin acquired Core Admin: %+v", record)
		}
		if record.Stage == "identity" {
			identity = true
			if !slices.Contains(record.Permissions, "chat_admin") || slices.Contains(record.Permissions, "admin") {
				t.Fatalf("permissions=%v", record.Permissions)
			}
		}
	}
	if !identity {
		t.Fatal("missing chat_admin identity evidence")
	}
	group := policy.Scope{Kind: policy.ScopeGroup, Group: []string{"secure"}}
	leaf := policy.Scope{Kind: policy.ScopeRoute, RouteID: "cmd.secure.leaf"}
	enabled, disabled := true, false
	set := func(scope policy.Scope, override policy.Override) {
		t.Helper()
		if err := p.SetOverride(t.Context(), "pipeline", scope, override); err != nil {
			t.Fatal(err)
		}
	}
	set(group, policy.Override{Enabled: &disabled})
	set(leaf, policy.Override{Enabled: &enabled, Permissions: []string{"user"}})
	h.inject(pipelineInput{Text: "/secure leaf", User: "alice"})
	h.idle()
	if len(pipelineStages(h.snapshot(), "allowed")) != 0 {
		t.Fatal("child enable bypassed disabled group")
	}
	set(group, policy.Override{Permissions: []string{"admin"}})
	set(leaf, policy.Override{})
	h.inject(pipelineInput{Text: "/secure leaf", User: "alice"})
	h.idle()
	if len(pipelineStages(h.snapshot(), "allowed")) != 0 {
		t.Fatal("leaf failed to inherit parent permission")
	}
	set(leaf, policy.Override{Permissions: []string{"user"}})
	h.inject(pipelineInput{Text: "/secure leaf", User: "alice"})
	h.idle()
	if len(pipelineStages(h.snapshot(), "allowed")) != 1 {
		t.Fatal("nearest explicit permission did not override parent")
	}
	set(policy.Scope{Kind: policy.ScopeRoute, RouteID: "cmd.secure.hold"}, policy.Override{Permissions: []string{"user"}})
	h.inject(pipelineInput{Text: "/secure hold", User: "alice"})
	h.await(func(records []pipelineRecord) bool { return len(pipelineStages(records, "policy-start")) == 1 })
	set(policy.Scope{Kind: policy.ScopeRoute, RouteID: "cmd.secure.hold"}, policy.Override{Permissions: []string{"admin"}})
	h.inject(pipelineInput{Text: "/secure hold", User: "alice"})
	// 第二条事件结束后，仍在进行的第一条保持唯一运行链。
	h.await(func(records []pipelineRecord) bool {
		for _, record := range records {
			if record.Stage == "identity" && record.Text == "/secure hold" {
				return true
			}
		}
		return false
	})
	h.release <- struct{}{}
	h.idle()
	if len(pipelineStages(h.snapshot(), "policy-start")) != 1 || len(pipelineStages(h.snapshot(), "policy-end")) != 1 || !reflect.DeepEqual(pipelineStages(h.snapshot(), "send"), []string{"send:policy-complete"}) {
		t.Fatalf("runtime policy=%+v", h.snapshot())
	}
}
