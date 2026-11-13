//go:build unix

package alonabot

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/conversation"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/tool"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk"
	sdkadapter "github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/proto"
)

const (
	t11Provider    = "tool-provider"
	t11Caller      = "tool-caller"
	t11InputSchema = `{"type":"object","properties":{"value":{"type":"string"},"key":{"type":"string"}},"required":["value","key"]}`
	t11Description = "Echo the supplied value"
)

type t11Arguments struct {
	Value string `json:"value"`
	Key   string `json:"key"`
}
type t11Command struct {
	Key, Action, Name, Arguments, Summary, Conversation string
	Entries                                             []string
	From                                                uint64
}
type t11Stored struct {
	Seq  uint64
	JSON string
}
type t11Reply struct {
	Code, Error, Result string
	Tools               []sdk.ToolInfo
	LastSeq, Next       uint64
	More                bool
	Entries             []t11Stored
}

func t11Process(mode string) error {
	if mode == "t11-permissions" {
		return t11PermissionProcess()
	}
	if strings.HasPrefix(mode, "t11-adapter") {
		return sdk.RunAdapter(&t11Adapter{buttons: mode != "t11-adapter-text"})
	}
	return sdk.Run(t11Plugin(mode))
}
func t11Plugin(mode string) *sdk.Plugin {
	p := sdk.NewPlugin()
	sdk.UseGenerated(p, sdk.Generated{ServiceSchemas: map[string][]byte{
		chainSchema[t11Arguments](): []byte(t11InputSchema),
		chainSchema[t11Command]():   []byte(`{"type":"object"}`),
		chainSchema[t11Reply]():     []byte(`{"type":"object"}`),
	}})
	id := os.Getenv(pipelineIDEnv)
	sdk.Serve(p, sdk.JSON[t11Command, t11Reply](id+"/probe")).Access(sdk.AccessPublic).Handle(t11Execute)
	if mode == "t11-provider" {
		for _, entry := range []struct {
			name      string
			access    sdk.Access
			allowlist []string
			policy    sdk.ToolDefault
		}{
			{"run", sdk.AccessPublic, nil, sdk.ToolConfirm},
			{"direct", sdk.AccessPublic, nil, sdk.ToolAllow},
			{"listed", sdk.AccessAllowlist, []string{t11Caller}, sdk.ToolAllow},
			{"hidden", sdk.AccessAllowlist, []string{"someone-else"}, sdk.ToolAllow},
		} {
			sdk.Serve(p, sdk.JSON[t11Arguments, t11Arguments](id+"/"+entry.name)).Access(entry.access, entry.allowlist...).Tool(t11Description, entry.policy).Handle(func(c *sdk.Context, in t11Arguments) (t11Arguments, error) {
				b, err := json.Marshal(in)
				if err != nil {
					return in, err
				}
				err = pipelineAudit(pipelineRecord{Stage: "execute", Route: entry.name, Text: in.Key, Value: string(b), User: string(c.User()), Session: string(c.Session())})
				return in, err
			})
		}
	} else {
		p.On("invoke", sdk.MessageCreated).Handle(func(c *sdk.Context) error {
			var in t11Command
			if err := json.Unmarshal([]byte(c.Text()), &in); err != nil {
				return err
			}
			out, err := t11Execute(c, in)
			if err != nil {
				return err
			}
			b, err := json.Marshal(out)
			if err != nil {
				return err
			}
			return pipelineAudit(pipelineRecord{Stage: "result", Text: in.Key, Value: string(b), Code: out.Code, User: string(c.User()), Session: string(c.Session())})
		})
		p.BeforeSend("audit").Handle(func(_ *sdk.Context, _ message.Session, chain message.Chain) (message.Chain, error) {
			return chain, pipelineAudit(pipelineRecord{Stage: "before", Text: chain.Text()})
		})
	}
	return p
}
func t11Execute(c *sdk.Context, in t11Command) (t11Reply, error) {
	var out t11Reply
	var err error
	switch in.Action {
	case "list":
		out.Tools, err = sdk.ListTools(c.Context())
	case "invoke":
		var result json.RawMessage
		result, err = sdk.InvokeTool(c.Context(), in.Name, json.RawMessage(in.Arguments))
		out.Result = string(result)
	case "operation":
		err = sdk.ConfirmOperation(c.Context(), in.Name, in.Summary)
		if err == nil {
			err = pipelineAudit(pipelineRecord{Stage: "operation", Text: in.Key, Route: in.Name})
		}
	case "append":
		entries := make([]json.RawMessage, len(in.Entries))
		for i, e := range in.Entries {
			entries[i] = json.RawMessage(e)
		}
		out.LastSeq, err = sdk.AppendConversation(c.Context(), in.Conversation, entries...)
	case "read":
		var entries []sdk.ConversationEntry
		entries, out.Next, out.More, err = sdk.ListConversation(c.Context(), in.Conversation, in.From)
		for _, entry := range entries {
			out.Entries = append(out.Entries, t11Stored{Seq: entry.Seq, JSON: string(entry.Entry)})
		}
	case "clear":
		err = sdk.ClearConversation(c.Context(), in.Conversation)
	default:
		return out, fmt.Errorf("unknown command %q", in.Action)
	}
	out.Code = chainCode(err).String()
	if err != nil {
		out.Error = err.Error()
	}
	return out, nil
}

type t11Adapter struct {
	pipelineAdapter
	buttons bool
}
type t11Loop struct {
	pipelineLoop
	buttons bool
}

// Start 绑定本次实例的按钮能力，其他平台输入沿用真实 Action 入口。
func (a *t11Adapter) Start(_ context.Context, _ sdkadapter.StartRequest, core sdkadapter.Core) (sdkadapter.Instance, error) {
	return &t11Loop{pipelineLoop: pipelineLoop{core: core}, buttons: a.buttons}, nil
}

// Capabilities 只对本矩阵的实例缩减按钮能力。
func (i *t11Loop) Capabilities() []message.Capability {
	capabilities := i.pipelineLoop.Capabilities()
	if !i.buttons {
		capabilities = slices.DeleteFunc(capabilities, func(c message.Capability) bool { return c.Name == "message.buttons" })
	}
	return capabilities
}

// Send 记录 Core 发到平台的完整文本和两枚 token。
func (i *t11Loop) Send(_ context.Context, to message.Session, chain message.Chain) (*message.SendResult, *errcode.Error) {
	return i.receipt("send", to, "", chain)
}

// Edit 记录确认完成后的文本与实际剩余按钮。
func (i *t11Loop) Edit(_ context.Context, to message.Session, id string, chain message.Chain) (*message.SendResult, *errcode.Error) {
	return i.receipt("edit", to, id, chain)
}
func (i *t11Loop) receipt(stage string, to message.Session, id string, chain message.Chain) (*message.SendResult, *errcode.Error) {
	i.mu.Lock()
	defer i.mu.Unlock() // 平台消息编号和对应审计保持相同次序。
	i.sequence++
	if id == "" {
		id = strconv.Itoa(i.sequence)
	}
	record := pipelineRecord{Stage: stage, Text: chain.Text(), Session: string(to), MessageID: id}
	for _, segment := range chain {
		if buttons, ok := segment.(message.ButtonsSegment); ok {
			for _, row := range buttons.Rows {
				for _, button := range row {
					record.Tokens = append(record.Tokens, button.CoreToken)
					record.Labels = append(record.Labels, button.Label)
				}
			}
		}
	}
	if err := pipelineAudit(record); err != nil {
		return nil, errcode.Internal.New(err.Error())
	}
	return &message.SendResult{Status: message.SendAccepted, Session: to, AdapterInstanceID: pipelineInstanceID, PrimaryPlatformMessageID: id, PlatformMessageIDs: []string{id}}, nil
}

func t11Harness(t *testing.T, stdio bool, values tune.Values, adapterMode string) *pipelineHarness {
	t.Helper()
	h := newPipelineAt(t, t.TempDir(), values, stdio)
	t11Components(h, adapterMode, true)
	return h
}
func t11ArgumentsFor(key string) string {
	return `{"value":"原文 & value","key":` + strconv.Quote(key) + `}`
}

func t11Begin(h *pipelineHarness, in t11Command, permissions ...string) (string, int) {
	h.t.Helper()
	start := len(h.snapshot())
	in.Key = in.Action + "-" + strconv.Itoa(start)
	if in.Arguments == "" {
		in.Arguments = t11ArgumentsFor(in.Key)
	}
	b, err := json.Marshal(in)
	if err != nil {
		h.t.Fatal(err)
	}
	h.inject(pipelineInput{Text: string(b), User: "alice", Permissions: permissions})
	return in.Key, start
}
func t11Result(h *pipelineHarness, key string) t11Reply {
	h.t.Helper()
	h.await(func(records []pipelineRecord) bool {
		return slices.ContainsFunc(records, func(r pipelineRecord) bool { return r.Stage == "result" && r.Text == key })
	})
	h.idle()
	for _, r := range h.snapshot() {
		if r.Stage == "result" && r.Text == key {
			var out t11Reply
			if err := json.Unmarshal([]byte(r.Value), &out); err != nil {
				h.t.Fatal(err)
			}
			return out
		}
	}
	h.t.Fatal("missing result")
	return t11Reply{}
}
func t11Code(t *testing.T, out t11Reply, want errcode.Code) {
	t.Helper()
	if out.Code != want.String() {
		t.Fatalf("code=%s want=%s error=%s", out.Code, want, out.Error)
	}
}
func t11Root(h *pipelineHarness, id string, in t11Command) t11Reply {
	h.t.Helper()
	b, err := json.Marshal(in)
	if err != nil {
		h.t.Fatal(err)
	}
	response := h.assembly.mux.For(router.Caller{})(h.t.Context(), &wire.Inbound{Target: id + "/probe", Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, SchemaID: chainSchema[t11Command](), SchemaVersion: 1, Bytes: b}})
	if response.Err != nil {
		h.t.Fatalf("probe %s: %v", id, response.Err)
	}
	var out t11Reply
	if err := json.Unmarshal(response.Payload.Bytes, &out); err != nil {
		h.t.Fatal(err)
	}
	return out
}
func t11Count(h *pipelineHarness, stage string) int { return len(pipelineStages(h.snapshot(), stage)) }
func t11Prompt(h *pipelineHarness, start int) pipelineRecord {
	h.t.Helper()
	h.await(func(records []pipelineRecord) bool {
		return slices.ContainsFunc(records[start:], func(r pipelineRecord) bool { return r.Stage == "send" })
	})
	for _, r := range h.snapshot()[start:] {
		if r.Stage == "send" {
			if len(r.Tokens) != 2 || r.Tokens[0] == "" || r.Tokens[1] == "" || r.Tokens[0] == r.Tokens[1] {
				h.t.Fatalf("invalid confirmation tokens: %+v", r)
			}
			return r
		}
	}
	h.t.Fatal("missing prompt")
	return pipelineRecord{}
}
func t11Click(h *pipelineHarness, token, user string) {
	h.t.Helper()
	h.inject(pipelineInput{Token: token, User: user})
}
func t11Policy(h *pipelineHarness, service string, p pb.ToolPolicy) {
	h.t.Helper()
	if err := tool.New(h.store, h.assembly.router).SetPolicy(h.t.Context(), t11Provider, service, p); err != nil {
		h.t.Fatal(err)
	}
}
func t11Directory(h *pipelineHarness, want map[string]string) {
	h.t.Helper()
	out := t11Root(h, t11Caller, t11Command{Action: "list"})
	t11Code(h.t, out, 0)
	got := map[string]string{}
	for _, entry := range out.Tools {
		got[entry.Name] = entry.Policy
		if entry.Description != t11Description || entry.InputSchema == nil {
			h.t.Fatalf("tool metadata=%+v", entry)
		}
		var schema, expected any
		if err := json.Unmarshal(entry.InputSchema, &schema); err != nil {
			h.t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(t11InputSchema), &expected); err != nil {
			h.t.Fatal(err)
		}
		if !reflect.DeepEqual(schema, expected) {
			h.t.Fatalf("schema=%s", entry.InputSchema)
		}
	}
	if !reflect.DeepEqual(got, want) {
		h.t.Fatalf("directory=%v want=%v", got, want)
	}
}

// TestToolsAndConversations 验证真实插件、Core 与回环平台间的工具和会话契约。
func TestToolsAndConversations(t *testing.T) { t11Matrix(t, false) }
func t11Matrix(t *testing.T, stdio bool) {
	t.Run("policies", func(t *testing.T) {
		h := t11Harness(t, stdio, tune.Defaults(), "t11-adapter")
		t11Policy(h, "direct", pb.ToolPolicy_TOOL_POLICY_DENY)
		t11Directory(h, map[string]string{t11Provider + "/run": "confirm", t11Provider + "/listed": "allow"})
		for _, step := range []struct {
			policy     pb.ToolPolicy
			click      int
			code       errcode.Code
			executions int
		}{
			{pb.ToolPolicy_TOOL_POLICY_ALLOW, -1, 0, 1}, {pb.ToolPolicy_TOOL_POLICY_CONFIRM, 0, 0, 2}, {pb.ToolPolicy_TOOL_POLICY_DENY, -1, errcode.PermissionDenied, 2}, {pb.ToolPolicy_TOOL_POLICY_CONFIRM, 1, errcode.PermissionDenied, 2},
		} {
			t11Policy(h, "run", step.policy)
			key, start := t11Begin(h, t11Command{Action: "invoke", Name: t11Provider + "/run", Arguments: t11ArgumentsFor("policy")})
			if step.click >= 0 {
				prompt := t11Prompt(h, start)
				if t11Count(h, "execute") != step.executions-1 && step.click == 0 {
					t.Fatal("tool ran before confirmation")
				}
				t11Click(h, prompt.Tokens[step.click], "alice")
			}
			out := t11Result(h, key)
			t11Code(t, out, step.code)
			if step.code == 0 {
				var got t11Arguments
				if err := json.Unmarshal([]byte(out.Result), &got); err != nil {
					t.Fatal(err)
				}
				if got.Key != "policy" || got.Value != "原文 & value" {
					t.Fatalf("tool output=%+v", got)
				}
			}
			if n := t11Count(h, "execute"); n != step.executions {
				t.Fatalf("executions=%d want=%d", n, step.executions)
			}
		}
	})
	t.Run("confirmations", func(t *testing.T) {
		for _, scenario := range []string{"allow", "deny", "timeout", "other_actor"} {
			t.Run(scenario, func(t *testing.T) {
				values := tune.Defaults()
				if scenario == "timeout" {
					var err error
					values, _, err = tune.Parse([]string{"TOOL_CONFIRM_TIMEOUT=2s"})
					if err != nil {
						t.Fatal(err)
					}
				}
				h := t11Harness(t, stdio, values, "t11-adapter")
				key, start := t11Begin(h, t11Command{Action: "invoke", Name: t11Provider + "/run"})
				prompt := t11Prompt(h, start)
				t11AssertPrompt(t, prompt, t11Provider+"/run", t11ArgumentsFor(key), true)
				if t11Count(h, "execute") != 0 || t11Count(h, "result") != 0 {
					t.Fatal("confirmation did not wait")
				}
				if scenario == "other_actor" {
					t11Click(h, prompt.Tokens[0], "bob")
					// 同步服务调用充当进程审计屏障，确认仍须等待原 Actor。
					t11Code(t, t11Root(h, t11Caller, t11Command{Action: "list"}), 0)
					if t11Count(h, "execute") != 0 || t11Count(h, "result") != 0 {
						t.Fatal("another actor completed confirmation")
					}
				}
				want := errcode.Code(0)
				executions := 1
				switch scenario {
				case "deny":
					want = errcode.PermissionDenied
					executions = 0
					t11Click(h, prompt.Tokens[1], "alice")
				case "timeout":
					want = errcode.DeadlineExceeded
					executions = 0
				default:
					t11Click(h, prompt.Tokens[0], "alice")
				}
				t11Code(t, t11Result(h, key), want)
				if scenario != "timeout" {
					t11AssertEdit(h, prompt, t11Provider+"/run", scenario != "deny")
				}
				for _, token := range prompt.Tokens {
					t11Click(h, token, "alice")
				}
				t11Code(t, t11Root(h, t11Caller, t11Command{Action: "list"}), 0)
				h.idle()
				if n := t11Count(h, "execute"); n != executions {
					t.Fatalf("replayed tokens changed executions: %d want %d", n, executions)
				}
				if t11Count(h, "before") != 0 {
					t.Fatal("confirmation reached caller BeforeSend")
				}
			})
		}
	})
	t.Run("missing_context", func(t *testing.T) {
		t.Run("actor", func(t *testing.T) {
			h := t11Harness(t, stdio, tune.Defaults(), "t11-adapter")
			out := t11Root(h, t11Caller, t11Command{Action: "invoke", Name: t11Provider + "/run", Arguments: t11ArgumentsFor("no-actor")})
			t11Code(t, out, errcode.PermissionDenied)
			h.idle()
			if t11Count(h, "send") != 0 || t11Count(h, "execute") != 0 {
				t.Fatal("actorless root service confirmed tool")
			}
		})
		t.Run("buttons", func(t *testing.T) {
			h := t11Harness(t, stdio, tune.Defaults(), "t11-adapter-text")
			key, _ := t11Begin(h, t11Command{Action: "invoke", Name: t11Provider + "/run"})
			t11Code(t, t11Result(h, key), errcode.Unsupported)
			if t11Count(h, "send") != 0 || t11Count(h, "execute") != 0 {
				t.Fatal("buttonless target confirmed tool")
			}
		})
	})

	t.Run("operations", func(t *testing.T) {
		h := t11Harness(t, stdio, tune.Defaults(), "t11-adapter")
		const operation = "mcp/x/y"
		const summary = "Write {\"path\":\"示例.txt\",\"value\":7}"
		key, start := t11Begin(h, t11Command{Action: "operation", Name: operation, Summary: summary})
		prompt := t11Prompt(h, start)
		t11AssertPrompt(t, prompt, operation, summary, false)
		if t11Count(h, "operation") != 0 || t11Count(h, "result") != 0 {
			t.Fatal("operation ran before confirmation")
		}
		for _, token := range prompt.Tokens {
			binding, err := h.store.GetButtonToken(t.Context(), token)
			if err != nil || binding.PluginID != t11Caller || binding.Purpose != "confirm" {
				t.Fatalf("operation token binding=%+v error=%v", binding, err)
			}
		}
		t11Click(h, prompt.Tokens[0], "alice")
		t11Code(t, t11Result(h, key), 0)
		t11AssertEdit(h, prompt, operation, true)
		if t11Count(h, "operation") != 1 || t11Count(h, "execute") != 0 || t11Count(h, "before") != 0 {
			t.Fatal("operation audit mismatch")
		}
	})
	t.Run("runtime_policy", func(t *testing.T) {
		h := t11Harness(t, stdio, tune.Defaults(), "t11-adapter")
		provider, caller := h.components[t11Provider], h.components[t11Caller]
		providerPID, callerPID := provider.process.Process.Pid(), caller.process.Process.Pid()
		providerSession, callerSession := provider.instance.Session, caller.instance.Session
		catalog := tool.New(h.store, h.assembly.router)
		executions := 0
		for _, policy := range []pb.ToolPolicy{pb.ToolPolicy_TOOL_POLICY_ALLOW, pb.ToolPolicy_TOOL_POLICY_CONFIRM, pb.ToolPolicy_TOOL_POLICY_DENY, pb.ToolPolicy_TOOL_POLICY_UNSPECIFIED} {
			t11Policy(h, "run", policy)
			effective := policy
			if effective == pb.ToolPolicy_TOOL_POLICY_UNSPECIFIED {
				effective = pb.ToolPolicy_TOOL_POLICY_CONFIRM
			}
			actual, err := catalog.Effective(t.Context(), t11Provider, "run")
			if err != nil || actual != effective {
				t.Fatalf("effective=%v want=%v error=%v", actual, effective, err)
			}
			entries, err := catalog.Policies(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			if !slices.ContainsFunc(entries, func(e tool.ToolPolicyEntry) bool {
				return e.PluginID == t11Provider && e.Service == "run" && e.Default == pb.ToolPolicy_TOOL_POLICY_CONFIRM && e.Override == policy
			}) {
				t.Fatalf("saved policies=%+v", entries)
			}
			want := map[string]string{t11Provider + "/direct": "allow", t11Provider + "/listed": "allow"}
			if effective != pb.ToolPolicy_TOOL_POLICY_DENY {
				want[t11Provider+"/run"] = strings.ToLower(strings.TrimPrefix(effective.String(), "TOOL_POLICY_"))
			}
			t11Directory(h, want)
			key, start := t11Begin(h, t11Command{Action: "invoke", Name: t11Provider + "/run"})
			if effective == pb.ToolPolicy_TOOL_POLICY_CONFIRM {
				prompt := t11Prompt(h, start)
				t11Click(h, prompt.Tokens[0], "alice")
			}
			code := errcode.Code(0)
			if effective == pb.ToolPolicy_TOOL_POLICY_DENY {
				code = errcode.PermissionDenied
			} else {
				executions++
			}
			t11Code(t, t11Result(h, key), code)
			if t11Count(h, "execute") != executions {
				t.Fatal("policy execution count mismatch")
			}
			if providerPID != provider.process.Process.Pid() || callerPID != caller.process.Process.Pid() || providerSession != provider.instance.Session || callerSession != caller.instance.Session || provider != h.components[t11Provider] || caller != h.components[t11Caller] {
				t.Fatal("policy update replaced a process or session")
			}
		}
	})

	t.Run("service_access", func(t *testing.T) {
		h := t11Harness(t, stdio, tune.Defaults(), "t11-adapter")
		t11Directory(h, map[string]string{t11Provider + "/run": "confirm", t11Provider + "/direct": "allow", t11Provider + "/listed": "allow"})
		key, _ := t11Begin(h, t11Command{Action: "invoke", Name: t11Provider + "/hidden"})
		t11Code(t, t11Result(h, key), errcode.NotFound)
		if t11Count(h, "execute") != 0 {
			t.Fatal("inaccessible service executed")
		}
		key, _ = t11Begin(h, t11Command{Action: "invoke", Name: t11Provider + "/listed"}, "loop.owner")
		t11Code(t, t11Result(h, key), 0)
		records := h.snapshot()
		if !slices.ContainsFunc(records, func(r pipelineRecord) bool {
			return r.Stage == "execute" && r.Route == "listed" && r.Text == key && r.User == "loop:alice" && r.Session == "loop:group#topic"
		}) {
			t.Fatal("service lost actor or default target")
		}

		// 权限入口尚由协议声明；换成原始 wire provider，调用方与平台仍是 SDK 进程。
		h.stop(h.components[t11Provider])
		h.launch(t11Provider, "t11-permissions", 0)
		executions := t11Count(h, "execute")
		for _, policy := range []pb.ToolPolicy{pb.ToolPolicy_TOOL_POLICY_ALLOW, pb.ToolPolicy_TOOL_POLICY_CONFIRM} {
			t11Policy(h, "run", policy)
			key, start := t11Begin(h, t11Command{Action: "invoke", Name: t11Provider + "/run"})
			t11Code(t, t11Result(h, key), errcode.PermissionDenied)
			if t11Count(h, "execute") != executions || slices.ContainsFunc(h.snapshot()[start:], func(r pipelineRecord) bool { return r.Stage == "send" }) {
				t.Fatal("required permissions were bypassed")
			}
			key, start = t11Begin(h, t11Command{Action: "invoke", Name: t11Provider + "/run"}, "loop.owner")
			if policy == pb.ToolPolicy_TOOL_POLICY_CONFIRM {
				prompt := t11Prompt(h, start)
				if t11Count(h, "execute") != executions {
					t.Fatal("permissioned tool ran before confirmation")
				}
				t11Click(h, prompt.Tokens[0], "alice")
			}
			t11Code(t, t11Result(h, key), 0)
			executions++
			if t11Count(h, "execute") != executions {
				t.Fatal("permissioned execution count mismatch")
			}
			if !slices.ContainsFunc(h.snapshot(), func(r pipelineRecord) bool {
				return r.Stage == "execute" && r.Text == key && r.User == "alice" && r.Session == "group" && r.Instance == pipelineInstanceID && r.Account == pipelineAccountID && r.Value == "topic" && slices.Contains(r.Permissions, "user") && slices.Contains(r.Permissions, "loop.owner") && slices.Contains(r.Permissions, "chat_admin") && !slices.Contains(r.Permissions, "admin")
			}) {
				t.Fatal("wire service lost inherited actor, permissions or default target")
			}
		}
	})

	t.Run("conversations", func(t *testing.T) {
		dir := t.TempDir()
		const key = "shared"
		var providerWant, callerWant []t11Stored
		t.Run("write_and_close", func(t *testing.T) {
			h := newPipelineAt(t, dir, tune.Defaults(), stdio)
			t11Components(h, "t11-adapter", true)
			first := []string{`{ "index": 1 }`, `{"index":2}`}
			out := t11Root(h, t11Provider, t11Command{Action: "append", Conversation: key, Entries: first})
			t11Code(t, out, 0)
			if out.LastSeq != 2 {
				t.Fatalf("batch last sequence=%d", out.LastSeq)
			}
			third := `{ "index" : 3 }`
			out = t11Root(h, t11Provider, t11Command{Action: "append", Conversation: key, Entries: []string{third}})
			t11Code(t, out, 0)
			if out.LastSeq != 3 {
				t.Fatalf("next sequence=%d", out.LastSeq)
			}
			providerWant = []t11Stored{{1, first[0]}, {2, first[1]}, {3, third}}
			out = t11Root(h, t11Provider, t11Command{Action: "append", Conversation: key, Entries: []string{`{"rollback":true}`, `{"broken":`}})
			t11Code(t, out, errcode.InvalidArgument)
			actual, _ := t11ReadAll(h, t11Provider, key)
			if !reflect.DeepEqual(actual, providerWant) {
				t.Fatal("failed batch changed stored entries")
			}
			fourth := `{"index":4,"after":"failed batch"}`
			out = t11Root(h, t11Provider, t11Command{Action: "append", Conversation: key, Entries: []string{fourth}})
			t11Code(t, out, 0)
			if out.LastSeq != 4 {
				t.Fatalf("failed batch consumed a sequence: %d", out.LastSeq)
			}
			providerWant = append(providerWant, t11Stored{4, fourth})
			total := 0
			for i := 0; i < 6; i++ {
				entry := fmt.Sprintf(`{ "index": %d, "text": "%s" }`, i, strings.Repeat(string(rune('a'+i)), 1<<20))
				total += len(entry)
				out = t11Root(h, t11Provider, t11Command{Action: "append", Conversation: key, Entries: []string{entry}})
				t11Code(t, out, 0)
				wantSeq := uint64(len(providerWant) + 1)
				if out.LastSeq != wantSeq {
					t.Fatalf("large append sequence=%d want=%d", out.LastSeq, wantSeq)
				}
				providerWant = append(providerWant, t11Stored{wantSeq, entry})
			}
			actual, pages := t11ReadAll(h, t11Provider, key)
			if total <= wire.MaxFrame || pages < 2 || !reflect.DeepEqual(actual, providerWant) {
				t.Fatalf("paged contents changed: bytes=%d pages=%d entries=%d", total, pages, len(actual))
			}
			callerEntry := `{ "owner": "caller" }`
			out = t11Root(h, t11Caller, t11Command{Action: "append", Conversation: key, Entries: []string{callerEntry}})
			t11Code(t, out, 0)
			callerWant = []t11Stored{{1, callerEntry}}
			actual, _ = t11ReadAll(h, t11Caller, key)
			if !reflect.DeepEqual(actual, callerWant) {
				t.Fatal("same-key conversation crossed plugin identity")
			}
			out = t11Root(h, t11Caller, t11Command{Action: "clear", Conversation: key})
			t11Code(t, out, 0)
			actual, _ = t11ReadAll(h, t11Caller, key)
			if len(actual) != 0 {
				t.Fatal("SDK clear retained caller entries")
			}
			actual, _ = t11ReadAll(h, t11Provider, key)
			if !reflect.DeepEqual(actual, providerWant) {
				t.Fatal("caller clear changed provider conversation")
			}
			out = t11Root(h, t11Caller, t11Command{Action: "append", Conversation: key, Entries: []string{callerEntry}})
			t11Code(t, out, 0)
			callerWant = []t11Stored{{out.LastSeq, callerEntry}}
			service := conversation.New(h.store)
			t11Summary(h, service, t11Provider, key, len(providerWant), providerWant[len(providerWant)-1].Seq)
			t11Summary(h, service, t11Caller, key, len(callerWant), callerWant[0].Seq)
			for _, name := range []string{"managed", "extra"} {
				t11Code(t, t11Root(h, t11Provider, t11Command{Action: "append", Conversation: name, Entries: []string{`{"saved":true}`}}), 0)
			}
			if err := service.Clear(t.Context(), t11Provider, "managed"); err != nil {
				t.Fatal(err)
			}
			actual, _ = t11ReadAll(h, t11Provider, "managed")
			if len(actual) != 0 {
				t.Fatal("management clear retained selected conversation")
			}
			actual, _ = t11ReadAll(h, t11Provider, key)
			if !reflect.DeepEqual(actual, providerWant) {
				t.Fatal("management clear changed another conversation")
			}
			summaries, err := service.List(t.Context(), t11Provider)
			if err != nil {
				t.Fatal(err)
			}
			if len(summaries) != 2 {
				t.Fatalf("remaining summaries=%+v", summaries)
			}
			// 子测试结束时沿公共清理顺序停止所有进程、Scheduler、Router，再关闭 Store。
		})
		t.Run("reopen_and_purge", func(t *testing.T) {
			h := newPipelineAt(t, dir, tune.Defaults(), stdio)
			t11Components(h, "t11-adapter", false)
			for _, entry := range []struct {
				id   string
				want []t11Stored
			}{{t11Provider, providerWant}, {t11Caller, callerWant}} {
				actual, _ := t11ReadAll(h, entry.id, key)
				if !reflect.DeepEqual(actual, entry.want) {
					t.Fatalf("conversation changed after Core restart: plugin=%s entries=%d want=%d", entry.id, len(actual), len(entry.want))
				}
			}
			service := conversation.New(h.store)
			if err := service.Purge(t.Context(), t11Provider); err != nil {
				t.Fatal(err)
			}
			summaries, err := service.List(t.Context(), t11Provider)
			if err != nil || len(summaries) != 0 {
				t.Fatalf("purged summaries=%+v error=%v", summaries, err)
			}
			for _, name := range []string{key, "extra", "managed"} {
				actual, _ := t11ReadAll(h, t11Provider, name)
				if len(actual) != 0 {
					t.Fatalf("purge retained conversation %s", name)
				}
			}
			actual, _ := t11ReadAll(h, t11Caller, key)
			if !reflect.DeepEqual(actual, callerWant) {
				t.Fatal("provider purge changed caller conversation")
			}
			t11Summary(h, service, t11Caller, key, len(callerWant), callerWant[0].Seq)
		})
	})
}

func t11AssertPrompt(t *testing.T, prompt pipelineRecord, name, detail string, arguments bool) {
	t.Helper()
	prefix := "Allow " + name + "?\n"
	if !strings.HasPrefix(prompt.Text, prefix) || prompt.Session != "loop:group#topic" || !reflect.DeepEqual(prompt.Labels, []string{"Allow", "Deny"}) {
		t.Fatalf("confirmation=%+v", prompt)
	}
	got := strings.TrimPrefix(prompt.Text, prefix)
	if arguments {
		var actual, expected any
		if err := json.Unmarshal([]byte(got), &actual); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(detail), &expected); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("displayed arguments=%s want=%s", got, detail)
		}
	} else if got != detail {
		t.Fatalf("summary=%q want=%q", got, detail)
	}
}
func t11AssertEdit(h *pipelineHarness, prompt pipelineRecord, name string, allowed bool) {
	h.t.Helper()
	prefix := "Denied: "
	if allowed {
		prefix = "Allowed: "
	}
	var edits []pipelineRecord
	for _, r := range h.snapshot() {
		if r.Stage == "edit" && r.MessageID == prompt.MessageID {
			edits = append(edits, r)
		}
	}
	if len(edits) != 1 || edits[0].Text != prefix+name || len(edits[0].Tokens) != 0 || edits[0].Session != prompt.Session {
		h.t.Fatalf("confirmation edits=%+v", edits)
	}
}

// t11PermissionProcess 以协议声明 Service 权限，业务调用仍经生产 Broker 与 Session。
func t11PermissionProcess() error {
	// 1. 使用与 SDK 相同的 TCP/stdio 和认证握手。
	var conn *wire.Conn
	if os.Getenv(wire.EnvTransport) == wire.TransportStdio {
		conn = wire.NewConn(os.Stdin, os.Stdout, os.Stdin)
	} else {
		socket, err := net.Dial("tcp", os.Getenv(wire.EnvCoreAddr))
		if err != nil {
			return err
		}
		conn = wire.NewStreamConn(socket)
	}
	defer conn.Close()
	id := os.Getenv(wire.EnvPluginID)
	secret, err := hex.DecodeString(os.Getenv(wire.EnvSecret))
	if err != nil {
		return err
	}
	digest, err := hex.DecodeString(os.Getenv(wire.EnvBuildDigest))
	if err != nil {
		return err
	}
	send := func(sid []byte, kind pb.FrameKind, value proto.Message) error {
		env, err := wire.ControlEnvelope(sid, 0, id, kind, value)
		if err != nil {
			return err
		}
		return conn.WriteFrame(env)
	}
	if err := send(nil, pb.FrameKind_FRAME_KIND_HELLO, &pb.Hello{PluginId: id, Secret: secret, BuildDigest: digest, SupportedMajor: 1}); err != nil {
		return err
	}
	env, err := conn.ReadFrame()
	if err != nil {
		return err
	}
	if env.Kind != pb.FrameKind_FRAME_KIND_HELLO_ACK {
		return fmt.Errorf("expected HELLO_ACK: %s", env.Kind)
	}
	ack := &pb.HelloAck{}
	if err := proto.Unmarshal(env.Payload, ack); err != nil {
		return err
	}
	// 2. 权限要求是任一满足；平台 owner 通过 Core 展开的 chat_admin 取得准入。
	registration := &pb.Register{CandidateConfigJson: "{}", Migration: &pb.MigrationState{Mode: pb.MigrationMode_MIGRATION_MODE_NONE}, Routes: []*pb.RouteDescriptor{{
		RouteId: "svc.run", Kind: pb.RouteKind_ROUTE_KIND_SERVICE, DefaultEnabled: true, RequiredPermissions: []string{"admin", "chat_admin"},
		InputSchemaId: chainSchema[t11Arguments](), InputSchemaVersion: 1, OutputSchemaId: chainSchema[t11Arguments](), OutputSchemaVersion: 1,
		Spec: &pb.RouteDescriptor_Service{Service: &pb.ServiceSpec{ServiceName: "run", Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, Access: pb.ServiceAccess_SERVICE_ACCESS_PUBLIC, Tool: &pb.ToolSpec{Description: t11Description, DefaultPolicy: pb.ToolPolicy_TOOL_POLICY_ALLOW, InputSchemaJson: t11InputSchema}}},
	}}}
	if err := send(ack.SessionId, pb.FrameKind_FRAME_KIND_REGISTER, registration); err != nil {
		return err
	}
	env, err = conn.ReadFrame()
	if err != nil {
		return err
	}
	if env.Kind != pb.FrameKind_FRAME_KIND_PREPARED {
		return fmt.Errorf("expected PREPARED: %s", env.Kind)
	}
	env, err = conn.ReadFrame()
	if err != nil {
		return err
	}
	if env.Kind != pb.FrameKind_FRAME_KIND_ACTIVATE {
		return fmt.Errorf("expected ACTIVATE: %s", env.Kind)
	}
	activate := &pb.Activate{}
	if err := proto.Unmarshal(env.Payload, activate); err != nil {
		return err
	}
	if err := send(ack.SessionId, pb.FrameKind_FRAME_KIND_ACTIVATED, &pb.Activated{GenerationId: activate.GenerationId, Ready: true}); err != nil {
		return err
	}
	// 3. 独立读循环维持取消与停止控制，处理函数只审计和回显实际请求。
	stopping := make(chan *pb.Stopping, 1)
	session := wire.New(conn, wire.Options{Local: pb.InitiatorSide_INITIATOR_SIDE_PLUGIN, SessionID: ack.SessionId, SourcePluginID: id,
		Control: func(env *pb.Envelope) error {
			if env.Kind != pb.FrameKind_FRAME_KIND_STOPPING {
				return fmt.Errorf("unexpected control %s", env.Kind)
			}
			stop := &pb.Stopping{}
			if err := proto.Unmarshal(env.Payload, stop); err != nil {
				return err
			}
			stopping <- stop
			return nil
		},
		Handler: func(_ context.Context, in *wire.Inbound) wire.Response {
			invocation := &pb.ServiceInvocation{}
			if err := proto.Unmarshal(in.Payload.Bytes, invocation); err != nil {
				return taskProcessError(err)
			}
			var args t11Arguments
			if err := json.Unmarshal(invocation.Payload, &args); err != nil {
				return taskProcessError(err)
			}
			actor, target := invocation.Context.GetActor(), invocation.Context.GetDefaultTarget()
			if err := pipelineAudit(pipelineRecord{Stage: "execute", Route: invocation.RouteId, Text: args.Key, User: actor.GetExternalUserId(), Permissions: actor.GetPermissions(), Session: target.GetChatId(), Value: target.GetThreadId(), Instance: target.GetAdapterInstanceId(), Account: target.GetAccountId()}); err != nil {
				return taskProcessError(err)
			}
			return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, SchemaID: chainSchema[t11Arguments](), SchemaVersion: 1, Bytes: invocation.Payload}}
		},
	})
	session.Start()
	select { // STOPPING 先等已接纳请求结束，断链则保留协议错误。
	case stop := <-stopping:
		ctx, cancel := context.WithDeadline(context.Background(), time.UnixMilli(stop.DeadlineUnixMs))
		defer cancel()
		if err := session.WaitIdle(ctx); err != nil {
			return err
		}
		return session.Close()
	case <-session.Done():
		return session.Err()
	}
}

func t11Components(h *pipelineHarness, adapterMode string, fresh bool) {
	h.t.Helper()
	for _, entry := range []struct{ id, mode string }{{"platform", adapterMode}, {t11Provider, "t11-provider"}, {t11Caller, "t11-caller"}} {
		if fresh {
			h.install(entry.id, entry.mode)
		}
		h.launch(entry.id, entry.mode, 0)
	}
}
func t11ReadAll(h *pipelineHarness, id, key string) ([]t11Stored, int) {
	h.t.Helper()
	var all []t11Stored
	var from uint64
	pages := 0
	for {
		out := t11Root(h, id, t11Command{Action: "read", Conversation: key, From: from})
		t11Code(h.t, out, 0)
		pages++
		for _, entry := range out.Entries {
			if len(all) > 0 && entry.Seq != all[len(all)-1].Seq+1 {
				h.t.Fatalf("noncontiguous page sequence=%d", entry.Seq)
			}
			all = append(all, entry)
		}
		if !out.More {
			if out.Next != 0 {
				h.t.Fatalf("terminal next sequence=%d", out.Next)
			}
			return all, pages
		}
		if len(out.Entries) == 0 || out.Next != out.Entries[len(out.Entries)-1].Seq+1 || out.Next <= from {
			h.t.Fatalf("invalid next sequence=%d from=%d", out.Next, from)
		}
		from = out.Next
	}
}
func t11Summary(h *pipelineHarness, service *conversation.Service, id, key string, count int, last uint64) {
	h.t.Helper()
	summaries, err := service.List(h.t.Context(), id)
	if err != nil {
		h.t.Fatal(err)
	}
	if !slices.ContainsFunc(summaries, func(summary state.ConversationSummary) bool {
		return summary.Key == key && summary.Entries == count && summary.LastSeq == last && !summary.UpdatedAt.IsZero()
	}) {
		h.t.Fatalf("conversation summaries=%+v", summaries)
	}
}
