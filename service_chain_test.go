//go:build unix

package alonabot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/proto"
)

const (
	chainMode           = "service-chain"
	chainBackgroundMode = "service-chain-background"
	chainBadOutput      = "invalid-output"
	chainDeadline       = pipelineTimeout
	chainBusBytes       = 1 << 20
)

type chainInput struct {
	Path []string `json:"path,omitempty"`
	Mode string   `json:"mode,omitempty"`
	Key  string   `json:"key,omitempty"`
	Size int      `json:"size,omitempty"`
}

type chainOutput struct {
	Value chainText `json:"value"`
}
type chainText string

// MarshalJSON 在目标进程注入合法 JSON、错误业务类型的输出。
func (s chainText) MarshalJSON() ([]byte, error) {
	if s == chainBadOutput {
		return []byte(`false`), nil
	}
	return json.Marshal(string(s))
}

type chainBusPayload struct {
	Key  string `json:"key"`
	Body string `json:"body"`
}

func chainSchema[T any]() string {
	typ := reflect.TypeFor[T]()
	return typ.PkgPath() + "." + typ.Name()
}
func chainService(id string) sdk.Service[chainInput, chainOutput] {
	return sdk.JSON[chainInput, chainOutput](id + "/run")
}

func serviceChainBackgroundPlugin() *sdk.Plugin {
	p := serviceChainPlugin()
	id := os.Getenv(pipelineIDEnv)
	sdk.Serve(p, sdk.JSON[chainInput, chainOutput](id+"/background")).Access(sdk.AccessPublic).Handle(func(c *sdk.Context, in chainInput) (chainOutput, error) {
		// 后台根调用脱离本次 Handler，停止时须由 Core 按 Session 取消。
		go func() {
			_, err := chainService("c").Call(context.Background(), chainInput{Mode: "hang", Key: "background-stop"})
			if err := pipelineAudit(pipelineRecord{Stage: "background-result", Code: chainCode(err).String()}); err != nil {
				panic(err)
			}
		}()
		return chainOutput{Value: chainText(id)}, nil
	})
	p.OnStop(func(ctx context.Context) error {
		if err := pipelineAudit(pipelineRecord{Stage: "on-stop"}); err != nil {
			return err
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, os.Getenv(pipelineAuditEnv)+"/release", nil)
		if err != nil {
			return err
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusNoContent {
			return fmt.Errorf("on-stop barrier status %d", response.StatusCode)
		}
		return nil
	})
	return p
}

func serviceChainPlugin() *sdk.Plugin {
	p := sdk.NewPlugin()
	sdk.UseGenerated(p, sdk.Generated{ServiceSchemas: map[string][]byte{
		chainSchema[chainInput]():  []byte(`{"type":"object","properties":{"path":{"type":"array","items":{"type":"string"}},"mode":{"type":"string"},"key":{"type":"string"},"size":{"type":"integer"}}}`),
		chainSchema[chainOutput](): []byte(`{"type":"object","properties":{"value":{"type":"string"}},"required":["value"]}`),
	}})
	id := os.Getenv(pipelineIDEnv)
	sdk.Serve(p, chainService(id)).Access(sdk.AccessPublic).Handle(func(c *sdk.Context, in chainInput) (chainOutput, error) { return chainHandle(c, id, in) })
	sdk.Serve(p, sdk.JSON[chainInput, chainOutput](id+"/private")).Handle(func(c *sdk.Context, in chainInput) (chainOutput, error) { return chainHandle(c, id, in) })
	if id == "a" {
		p.On("root", sdk.MessageCreated).Handle(func(c *sdk.Context) error {
			var in chainInput
			if err := json.Unmarshal([]byte(c.Text()), &in); err != nil {
				return err
			}
			if in.Mode == "hook" {
				_, err := c.Send(message.Text("original"))
				return err
			}
			if err := chainAudit("root", c, in); err != nil {
				return err
			}
			target := in.Path[0]
			in.Path = in.Path[1:]
			out, err := chainService(target).Call(c.Context(), in)
			return pipelineAudit(pipelineRecord{Stage: "root-result", Text: in.Key, Value: string(out.Value), Code: chainCode(err).String()})
		})
		p.BeforeSend("service").Handle(func(c *sdk.Context, _ message.Session, _ message.Chain) (message.Chain, error) {
			_, err := chainService("b").Call(c.Context(), chainInput{Mode: "send", Key: "hook"})
			return nil, pipelineAudit(pipelineRecord{Stage: "hook-result", Code: chainCode(err).String()})
		})
	}
	// 两个独立慢队列同时达到条数与字节上限，a 继续接收两类事件。
	for _, name := range []string{"small", "large"} {
		if id == "b" && name != "small" || id == "c" && name != "large" {
			continue
		}
		sdk.Subscribe(p, sdk.BusEvent[chainBusPayload]("a/"+name)).Handle(func(c *sdk.Context, in chainBusPayload) error {
			return pipelineAudit(pipelineRecord{Stage: "bus", Text: in.Key, User: string(c.User()), Session: string(c.Session())})
		})
	}
	return p
}

func chainCode(err error) errcode.Code {
	if err == nil {
		return 0
	}
	var public *errcode.Error
	if errors.As(err, &public) {
		return public.Code
	}
	return errcode.Internal
}
func chainAudit(stage string, c *sdk.Context, in chainInput) error {
	deadline, _ := c.Context().Deadline()
	return pipelineAudit(pipelineRecord{Stage: stage, Text: in.Key, Route: in.Mode, User: string(c.User()), Session: string(c.Session()), Value: strconv.FormatInt(deadline.UnixMilli(), 10)})
}
func chainHandle(c *sdk.Context, id string, in chainInput) (chainOutput, error) {
	if err := chainAudit("service", c, in); err != nil {
		return chainOutput{}, err
	}
	if len(in.Path) > 0 {
		next := in.Path[0]
		in.Path = in.Path[1:]
		out, err := chainService(next).Call(c.Context(), in)
		if err != nil {
			return chainOutput{}, err
		}
		return chainOutput{Value: chainText(id + ">" + string(out.Value))}, nil
	}
	switch in.Mode {
	case "hang":
		<-c.Context().Done()
		if err := chainAudit("cancelled", c, in); err != nil {
			return chainOutput{}, err
		}
		return chainOutput{}, errcode.Cancelled.New("service cancelled")
	case "crash":
		os.Exit(23)
	case "send":
		_, err := c.Send(message.Text("forbidden"))
		return chainOutput{}, err
	case "bad-output":
		return chainOutput{Value: chainBadOutput}, nil
	case "barrier":
		request, err := http.NewRequestWithContext(c.Context(), http.MethodGet, os.Getenv(pipelineAuditEnv)+"/release", nil)
		if err != nil {
			return chainOutput{}, err
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return chainOutput{}, err
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusNoContent {
			return chainOutput{}, fmt.Errorf("barrier status %d", response.StatusCode)
		}
	case "error":
		return chainOutput{}, errors.New("/private/plugin.go:42 stack: synthetic diagnostic")
	case "small", "large":
		_, err := sdk.BusEvent[chainBusPayload]("a/"+in.Mode).Publish(c.Context(), chainBusPayload{Key: in.Key, Body: strings.Repeat("x", in.Size)})
		return chainOutput{Value: chainText(id)}, err
	}
	return chainOutput{Value: chainText(id)}, nil
}

type chainHop struct {
	Plugin   string
	Request  *pb.ServiceInvocation
	Trace    []byte
	Deadline time.Time
}

type chainObserver struct {
	router.Plugin
	mu      sync.Mutex
	hops    []chainHop
	gate    <-chan struct{} // 只推迟真实 Session 写出，保留原取消路径。
	entered chan struct{}
	once    sync.Once
}

// Call 记录每一跳后仍调用原生产 Session。
func (p *chainObserver) Call(ctx context.Context, req wire.Request) (wire.Response, error) {
	if req.Payload.SchemaID == validate.SchemaID(&pb.ServiceInvocation{}) {
		in := &pb.ServiceInvocation{}
		if err := proto.Unmarshal(req.Payload.Bytes, in); err != nil {
			return wire.Response{}, err
		}
		p.mu.Lock()
		p.hops = append(p.hops, chainHop{Plugin: p.PluginID(), Request: in, Trace: bytes.Clone(req.TraceID), Deadline: req.Deadline})
		p.mu.Unlock()
	}
	return p.Plugin.Call(ctx, req)
}

// SendEvent 把慢写出稳定保持在真实调用之前；恢复与取消都交给 Session。
func (p *chainObserver) SendEvent(ctx context.Context, target string, ev *pb.BusEvent) error {
	if p.gate != nil {
		p.once.Do(func() { close(p.entered) })
		select {
		case <-p.gate:
		case <-ctx.Done():
		}
	}
	return p.Plugin.SendEvent(ctx, target, ev)
}
func (p *chainObserver) snapshot() []chainHop {
	p.mu.Lock()
	defer p.mu.Unlock()
	return slices.Clone(p.hops)
}

func newServiceChain(t *testing.T, stdio bool) (*pipelineHarness, map[string]*chainObserver) {
	t.Helper()
	h := newPipeline(t, stdio)
	observers := make(map[string]*chainObserver)
	for _, id := range []string{"a", "b", "c"} {
		h.install(id, chainMode)
		c := h.launch(id, chainMode, 0)
		observer := &chainObserver{Plugin: router.FromInstance(c.instance)}
		h.assembly.router.Attach(observer, c.caller)
		observers[id] = observer
	}
	t.Logf("command: %s %q; transport=%s", os.Args[0], os.Args[1:], chainTransport(stdio))
	return h, observers
}
func chainTransport(stdio bool) string {
	if stdio {
		return "stdio"
	}
	return "tcp"
}
func chainRequest(t *testing.T, target string, in chainInput) *wire.Inbound {
	t.Helper()
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	return &wire.Inbound{Target: target, TraceID: bytes.Repeat([]byte{7}, 16), Deadline: time.Now().Add(chainDeadline), Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, SchemaID: chainSchema[chainInput](), SchemaVersion: 1, Bytes: data}}
}
func chainCall(h *pipelineHarness, ctx context.Context, target string, in chainInput) wire.Response {
	req := chainRequest(h.t, target, in)
	return h.assembly.mux.For(router.Caller{})(ctx, req)
}
func chainAssertCode(t *testing.T, response wire.Response, want pb.ErrorCode) {
	t.Helper()
	if response.Err.GetCode() != want {
		t.Fatalf("response code=%s want=%s payload=%s error=%v", response.Err.GetCode(), want, response.Payload.Bytes, response.Err)
	}
}
func chainAssertValue(t *testing.T, response wire.Response, want string) {
	t.Helper()
	chainAssertCode(t, response, 0)
	var out chainOutput
	if err := json.Unmarshal(response.Payload.Bytes, &out); err != nil {
		t.Fatal(err)
	}
	if string(out.Value) != want {
		t.Fatalf("value=%s want=%s", out.Value, want)
	}
}
func chainCount(h *pipelineHarness, stage, plugin, key string) int {
	n := 0
	for _, r := range h.snapshot() {
		if r.Stage == stage && r.Plugin == plugin && (key == "" || r.Text == key) {
			n++
		}
	}
	return n
}
func chainAwait(h *pipelineHarness, stage, plugin, key string, n int) {
	h.t.Helper()
	h.await(func(records []pipelineRecord) bool {
		count := 0
		for _, r := range records {
			if r.Stage == stage && r.Plugin == plugin && (key == "" || r.Text == key) {
				count++
			}
		}
		return count >= n
	})
}
func chainEventually(t *testing.T, predicate func() bool) {
	t.Helper()
	timer := time.NewTimer(pipelineTimeout)
	defer timer.Stop()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for !predicate() {
		select {
		case <-ticker.C:
		case <-timer.C:
			t.Fatal("condition timed out")
		}
	}
}

func chainReceive(t *testing.T, results <-chan wire.Response) wire.Response {
	t.Helper()
	timer := time.NewTimer(pipelineTimeout)
	defer timer.Stop()
	select {
	case response := <-results:
		return response
	case <-timer.C:
		t.Fatal("service call did not finish")
		return wire.Response{}
	}
}

// TestServiceChain 覆盖 Service/Hook/Bus 的真实进程与双传输验收矩阵。
func TestServiceChain(t *testing.T) {
	for _, stdio := range []bool{false, true} {
		t.Run(chainTransport(stdio), func(t *testing.T) {
			t.Run("nested-context-and-hook", func(t *testing.T) {
				h, observers := newServiceChain(t, stdio)
				h.install("platform", "adapter")
				h.launch("platform", "adapter", 0)
				data, _ := json.Marshal(chainInput{Path: []string{"b", "c"}, Key: "trusted"})
				if failure := h.deliver(pipelineInput{Text: string(data), User: "alice", Permissions: []string{"loop.owner"}}); failure != nil {
					t.Fatal(failure)
				}
				chainAwait(h, "root-result", "a", "trusted", 1)
				h.idle()
				var root pipelineRecord
				for _, r := range h.snapshot() {
					if r.Stage == "root" {
						root = r
					}
					if r.Stage == "root-result" && r.Value != "b>c" {
						t.Fatal(r)
					}
				}
				var trace []byte
				var target *pb.Target
				for _, id := range []string{"b", "c"} {
					hops := observers[id].snapshot()
					if len(hops) != 1 {
						t.Fatalf("%s hops=%d", id, len(hops))
					}
					hop := hops[0]
					if hop.Request.Context.GetActor().GetExternalUserId() != "alice" || !slices.Contains(hop.Request.Context.Actor.Permissions, "chat_admin") || hop.Request.Context.GetDefaultTarget().GetThreadId() != "topic" {
						t.Fatalf("untrusted context: %v", hop.Request.Context)
					}
					if trace == nil {
						trace = hop.Trace
					} else if !bytes.Equal(trace, hop.Trace) {
						t.Fatal("trace changed across hops")
					}
					if len(trace) != 16 {
						t.Fatalf("trace length %d", len(trace))
					}
					if target == nil {
						target = hop.Request.Context.DefaultTarget
					} else if !proto.Equal(target, hop.Request.Context.DefaultTarget) {
						t.Fatal("target changed across hops")
					}
					wantChain := []string{"a"}
					if id == "c" {
						wantChain = append(wantChain, "b")
					}
					if !slices.Equal(hop.Request.Context.PluginChain, wantChain) {
						t.Fatalf("plugin chain=%v want=%v", hop.Request.Context.PluginChain, wantChain)
					}
					if strconv.FormatInt(hop.Deadline.UnixMilli(), 10) != root.Value {
						t.Fatalf("deadline %s root=%s", hop.Deadline, root.Value)
					}
					chainAwait(h, "service", id, "trusted", 1)
					for _, r := range h.snapshot() {
						if r.Stage == "service" && r.Plugin == id && (r.User != root.User || r.Session != root.Session || r.Value != root.Value) {
							t.Fatalf("hop binding changed: root=%+v hop=%+v", root, r)
						}
					}
				}
				data, _ = json.Marshal(chainInput{Mode: "hook"})
				if failure := h.deliver(pipelineInput{Text: string(data), User: "alice"}); failure != nil {
					t.Fatal(failure)
				}
				chainAwait(h, "hook-result", "a", "", 1)
				h.idle()
				for _, r := range h.snapshot() {
					if r.Stage == "hook-result" && r.Code != errcode.FailedPrecondition.String() {
						t.Fatal(r)
					}
				}
				if chainCount(h, "send", "platform", "") != 1 {
					t.Fatalf("original Send count: %+v", h.snapshot())
				}
				for _, r := range h.snapshot() {
					if r.Stage == "send" && r.Text != "original" {
						t.Fatal(r)
					}
				}
			})
			t.Run("background-root-stopping", func(t *testing.T) {
				h := newPipeline(t, stdio)
				h.install("c", chainMode)
				h.launch("c", chainMode, 0)
				h.install("b", chainBackgroundMode)
				b := h.launch("b", chainBackgroundMode, 0)
				initial := h.assembly.global.Used()
				chainAssertValue(t, chainCall(h, t.Context(), "b/background", chainInput{}), "b")
				chainAwait(h, "service", "c", "background-stop", 1)
				stopped := make(chan host.StopResult, 1)
				go func() { stopped <- b.process.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE) }()
				chainAwait(h, "on-stop", "b", "", 1)
				chainAwait(h, "cancelled", "c", "background-stop", 1)
				chainAwait(h, "background-result", "b", "", 1)
				chainEventually(t, func() bool { return h.assembly.global.Used() == initial })
				if b.instance.State() != host.StateStopping {
					t.Fatalf("background cancellation state=%s", b.instance.State())
				}
				select {
				case <-b.instance.Session.Done():
					t.Fatal("background root waited for session close")
				default:
				} // OnStop 保持连接开放，确认取消发生在 STOPPING。
				for _, record := range h.snapshot() {
					if record.Stage == "background-result" && record.Code != errcode.Cancelled.String() && record.Code != errcode.Unavailable.String() {
						t.Fatalf("background result=%+v", record)
					}
				}
				if chainCount(h, "service", "c", "background-stop") != 1 || h.assembly.metrics.Snapshot()["c"]["service"]["cancelled"] != 1 {
					t.Fatal("background root repeated or cancellation count missing")
				}
				t.Logf("caller=b pid=%d state=%s transport=%s; background target=c cancelled; budget=%d", b.process.Process.Pid(), b.instance.State(), chainTransport(stdio), h.assembly.global.Used())
				h.release <- struct{}{}
				if result := <-stopped; result.Err != nil {
					t.Fatal(result.Err)
				}
				delete(h.components, "b")
			})
			t.Run("cycles-depth-concurrent", func(t *testing.T) {
				h, observers := newServiceChain(t, stdio)
				cycle := chainCall(h, t.Context(), "a/run", chainInput{Path: []string{"b", "a"}, Key: "cycle"})
				chainAssertCode(t, cycle, pb.ErrorCode_ERROR_CODE_CALL_CYCLE)
				if chainCount(h, "service", "a", "cycle") != 1 || chainCount(h, "service", "b", "cycle") != 1 {
					t.Fatal("cycle executed a target twice")
				}
				if h.assembly.metrics.Snapshot()["a"]["service"]["cycle"] != 1 {
					t.Fatal(h.assembly.metrics.Snapshot())
				}
				// 边界使用 Core 保留的祖先附件；实际尾部仍经 a→b→c 三进程。
				for _, ancestors := range []int{12, 13} {
					req := chainRequest(t, "a/run", chainInput{Path: []string{"b", "c"}, Key: fmt.Sprint(ancestors)})
					chain := make([]string, ancestors)
					for i := range chain {
						chain[i] = fmt.Sprintf("ancestor%d", i)
					}
					req.ParentAttach = &router.Invocation{Depth: ancestors + 1, Chain: chain, TraceID: req.TraceID}
					response := h.assembly.mux.For(router.Caller{})(t.Context(), req)
					if ancestors == 12 {
						chainAssertValue(t, response, "a>b>c")
					} else {
						chainAssertCode(t, response, pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION)
					}
				}
				if h.assembly.metrics.Snapshot()["c"]["service"]["depth"] != 1 {
					t.Fatal(h.assembly.metrics.Snapshot())
				}
				var wg sync.WaitGroup
				results := make(chan wire.Response, 2)
				for _, key := range []string{"one", "two"} {
					wg.Go(func() {
						request := chainRequest(t, "a/run", chainInput{Path: []string{"b", "c"}, Mode: "barrier", Key: key})
						request.ParentAttach = &router.Invocation{Depth: 1, Event: &pb.Event{Actor: &pb.Actor{ExternalUserId: key}}, DefaultTarget: &pb.Target{AdapterId: "loop", ChatId: key}, Scopes: []string{key}, TraceID: bytes.Repeat([]byte{key[0]}, 16)}
						results <- h.assembly.mux.For(router.Caller{})(t.Context(), request)
					})
				}
				for _, key := range []string{"one", "two"} {
					chainAwait(h, "service", "c", key, 1)
				}
				// 两条链同时停在 c，放行后分别取得自己的结果。
				h.release <- struct{}{}
				h.release <- struct{}{}
				wg.Wait()
				close(results)
				for response := range results {
					chainAssertValue(t, response, "a>b>c")
				}
				for _, key := range []string{"one", "two"} {
					for _, id := range []string{"a", "b", "c"} {
						if chainCount(h, "service", id, key) != 1 {
							t.Fatalf("duplicate/missing %s %s", id, key)
						}
					}
				}
				for _, observer := range observers {
					for _, hop := range observer.snapshot() {
						var input chainInput
						if err := json.Unmarshal(hop.Request.Payload, &input); err != nil {
							t.Fatal(err)
						}
						if input.Key != "one" && input.Key != "two" {
							continue
						}
						trusted := hop.Request.Context
						if trusted.Actor.GetExternalUserId() != input.Key || trusted.DefaultTarget.GetChatId() != input.Key || !slices.Equal(trusted.Scopes, []string{input.Key}) || !bytes.Equal(hop.Trace, bytes.Repeat([]byte{input.Key[0]}, 16)) {
							t.Fatalf("concurrent chain context changed: %+v", hop)
						}
					}
				}
				if len(observers["c"].snapshot()) != 3 {
					t.Fatal("depth rejection executed target")
				}
			})
			t.Run("cancellation-exit-replacement", func(t *testing.T) {
				h, _ := newServiceChain(t, stdio)
				for _, leaf := range []string{"b", "c"} {
					ctx, cancel := context.WithCancel(t.Context())
					done := make(chan wire.Response, 1)
					path := []string{"b"}
					if leaf == "c" {
						path = append(path, "c")
					}
					go func() { done <- chainCall(h, ctx, "a/run", chainInput{Path: path, Mode: "hang", Key: leaf}) }()
					chainAwait(h, "service", leaf, leaf, 1)
					cancel()
					response := chainReceive(t, done)
					if code := response.Err.GetCode(); code != pb.ErrorCode_ERROR_CODE_CANCELLED && code != pb.ErrorCode_ERROR_CODE_UNAVAILABLE {
						t.Fatal(response.Err)
					}
					chainAwait(h, "cancelled", leaf, leaf, 1)
				}
				done := make(chan wire.Response, 1)
				go func() {
					done <- chainCall(h, t.Context(), "a/run", chainInput{Path: []string{"b", "c"}, Mode: "hang", Key: "exit"})
				}()
				chainAwait(h, "service", "c", "exit", 1)
				old := h.components["b"]
				if err := old.process.Process.KillTree(); err != nil {
					t.Fatal(err)
				}
				<-old.instance.Session.Done()
				h.assembly.detach(old)
				if _, _, err := old.process.WaitExit(); err != nil {
					t.Fatal(err)
				}
				delete(h.components, "b")
				response := chainReceive(t, done)
				if code := response.Err.GetCode(); code != pb.ErrorCode_ERROR_CODE_CANCELLED && code != pb.ErrorCode_ERROR_CODE_UNAVAILABLE {
					t.Fatal(response.Err)
				}
				chainAwait(h, "cancelled", "c", "exit", 1)
				h.launch("b", chainMode, 0)
				chainAssertValue(t, chainCall(h, t.Context(), "a/run", chainInput{Path: []string{"b", "c"}, Key: "new"}), "a>b>c")
				if chainCount(h, "service", "b", "exit") != 1 || chainCount(h, "service", "b", "new") != 1 {
					t.Fatal("replacement replayed old request")
				}
				if h.assembly.metrics.Snapshot()["b"]["service"]["cancelled"] == 0 {
					t.Fatal("missing cancellation count")
				}
			})

			t.Run("deadline-draining-and-target-exit", func(t *testing.T) {
				h, observers := newServiceChain(t, stdio)
				ctx, cancel := context.WithTimeout(t.Context(), time.Second)
				defer cancel()
				done := make(chan wire.Response, 1)
				go func() {
					done <- chainCall(h, ctx, "a/run", chainInput{Path: []string{"b", "c"}, Mode: "hang", Key: "deadline"})
				}()
				chainAwait(h, "service", "c", "deadline", 1)
				deadline, _ := ctx.Deadline()
				for _, id := range []string{"a", "b", "c"} {
					hops := observers[id].snapshot()
					if len(hops) != 1 || hops[0].Deadline.UnixMilli() != deadline.UnixMilli() {
						t.Fatalf("deadline changed at %s: %+v", id, hops)
					}
				}
				chainAssertCode(t, chainReceive(t, done), pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED)
				chainAwait(h, "cancelled", "c", "deadline", 1)
				go func() {
					done <- chainCall(h, t.Context(), "a/run", chainInput{Path: []string{"b", "c"}, Mode: "barrier", Key: "drain"})
				}()
				chainAwait(h, "service", "c", "drain", 1)
				stopped := make(chan host.StopResult, 1)
				b := h.components["b"]
				go func() { stopped <- b.process.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE) }()
				chainEventually(t, func() bool { return b.instance.State() == host.StateDraining })
				chainAssertCode(t, chainCall(h, t.Context(), "b/run", chainInput{Key: "frozen"}), pb.ErrorCode_ERROR_CODE_UNAVAILABLE)
				if chainCount(h, "service", "b", "frozen") != 0 || h.assembly.metrics.Snapshot()["b"]["service"]["offline"] != 1 {
					t.Fatal("draining admitted a new call")
				}
				h.release <- struct{}{}
				chainAssertValue(t, chainReceive(t, done), "a>b>c")
				if result := <-stopped; result.Err != nil {
					t.Fatal(result.Err)
				}
				h.assembly.detach(b)
				delete(h.components, "b")
				h.launch("b", chainMode, 0)
				chainAssertCode(t, chainCall(h, t.Context(), "a/run", chainInput{Path: []string{"b", "c"}, Mode: "crash", Key: "crash-c"}), pb.ErrorCode_ERROR_CODE_UNAVAILABLE)
				c := h.components["c"]
				<-c.instance.Session.Done()
				h.assembly.detach(c)
				if _, _, err := c.process.WaitExit(); err != nil {
					t.Fatal(err)
				}
				delete(h.components, "c")
				h.launch("c", chainMode, 0)
				chainAssertValue(t, chainCall(h, t.Context(), "a/run", chainInput{Path: []string{"b", "c"}, Key: "after-c"}), "a>b>c")
				if chainCount(h, "service", "c", "crash-c") != 1 {
					t.Fatal("crashed request replayed")
				}
			})
			t.Run("admission-and-malformed", func(t *testing.T) {
				h, _ := newServiceChain(t, stdio)
				caller := h.components["a"].caller
				req := chainRequest(t, "b/private", chainInput{Key: "private"})
				chainAssertCode(t, h.assembly.mux.For(caller)(t.Context(), req), pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED)
				for _, mutate := range []func(*wire.Inbound){func(r *wire.Inbound) { r.Payload.SchemaVersion++ }, func(r *wire.Inbound) { r.Payload.Encoding = pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF }, func(r *wire.Inbound) { r.Payload.Bytes = []byte(`{`) }, func(r *wire.Inbound) { r.Payload.Bytes = []byte(`{"path":123}`) }} {
					req := chainRequest(t, "b/run", chainInput{Key: "invalid"})
					mutate(req)
					chainAssertCode(t, h.assembly.mux.For(caller)(t.Context(), req), pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT)
				}
				if chainCount(h, "service", "b", "invalid") != 0 || chainCount(h, "service", "b", "private") != 0 {
					t.Fatal("rejected input ran target")
				}
				chainAssertCode(t, chainCall(h, t.Context(), "b/run", chainInput{Mode: "bad-output", Key: "output"}), pb.ErrorCode_ERROR_CODE_INTERNAL)
				if chainCount(h, "service", "b", "output") != 1 {
					t.Fatal("invalid output repeated business effect")
				}
				// 占满目标方向槽位，普通响应和 Cancel 仍走优先通道。
				max := h.values.SessionMaxInflight
				pending := make(chan wire.Response, max)
				ctx, cancel := context.WithCancel(t.Context())
				for i := range max {
					go func() {
						pending <- chainCall(h, ctx, "b/run", chainInput{Mode: "hang", Key: fmt.Sprintf("full-%d", i)})
					}()
				}
				chainAwait(h, "service", "b", "", max+1)
				chainAssertCode(t, chainCall(h, t.Context(), "b/run", chainInput{Key: "overflow"}), pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED)
				if chainCount(h, "service", "b", "overflow") != 0 {
					t.Fatal("full target executed overflow")
				}
				targetSession := h.components["b"].instance.Session
				lateResponses := targetSession.Stats().LateResponses
				cancel()
				for range max {
					chainAssertCode(t, chainReceive(t, pending), pb.ErrorCode_ERROR_CODE_CANCELLED)
				}
				// 等这批取消调用的迟到响应全部到达，确认插件已释放接收侧槽位。
				chainEventually(t, func() bool {
					return h.assembly.global.Used() == 0 && targetSession.Stats().LateResponses == lateResponses+int64(max)
				})
				if !h.assembly.global.Acquire(h.values.CoreMaxPayload) {
					t.Fatal("global budget not empty")
				}
				response := chainCall(h, t.Context(), "b/run", chainInput{Key: "bytes-full"})
				h.assembly.global.Release(h.values.CoreMaxPayload)
				chainAssertCode(t, response, pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED)
				if chainCount(h, "service", "b", "bytes-full") != 0 {
					t.Fatal("budget rejection ran target")
				}

				ctx, cancel = context.WithCancel(t.Context())
				cancel()
				chainAssertCode(t, chainCall(h, ctx, "b/run", chainInput{Key: "pre-cancel"}), pb.ErrorCode_ERROR_CODE_CANCELLED)
				if chainCount(h, "service", "b", "pre-cancel") != 0 {
					t.Fatal("pre-cancelled request executed")
				}
				chainAssertValue(t, chainCall(h, t.Context(), "b/run", chainInput{Key: "after"}), "b")
				counts := h.assembly.metrics.Snapshot()["b"]["service"]
				if counts["denied"] != 1 || counts["dropped_budget"] == 0 {
					t.Fatal(counts)
				}
			})
			t.Run("bus-bounds-and-disconnect", func(t *testing.T) {
				h, observers := newServiceChain(t, stdio)
				gates := make(map[string]chan struct{})
				for _, id := range []string{"b", "c"} {
					observer := observers[id]
					gates[id] = make(chan struct{})
					observer.gate = gates[id]
					observer.entered = make(chan struct{})
					h.assembly.bus.Attach(observer)
				}
				publish := func(name, key string, size int) {
					t.Helper()
					chainAssertValue(t, chainCall(h, t.Context(), "a/run", chainInput{Mode: name, Key: key, Size: size}), "a")
					chainAwait(h, "bus", "a", key, 1)
				}
				publish("small", "small-first", 0)
				<-observers["b"].entered
				publish("large", "large-first", chainBusBytes)
				<-observers["c"].entered
				for i := range h.values.BusQueueEvents {
					publish("small", fmt.Sprintf("small-%d", i), 0)
				}
				for i := range 3 {
					publish("large", fmt.Sprintf("large-%d", i), chainBusBytes)
				}
				publish("small", "small-overflow", 0)
				publish("large", "large-overflow", chainBusBytes)
				counts := h.assembly.metrics.Snapshot()
				if counts["b"]["bus"]["dropped_queue"] != 1 || counts["c"]["bus"]["dropped_budget"] != 1 {
					t.Fatal(counts)
				}
				close(gates["b"])
				chainAwait(h, "bus", "b", fmt.Sprintf("small-%d", h.values.BusQueueEvents-1), 1)
				if chainCount(h, "bus", "b", "") != h.values.BusQueueEvents+1 || chainCount(h, "bus", "b", "small-overflow") != 0 {
					t.Fatal("queue did not drop newest")
				}
				// 再挂起一次真实写出，保留旧队列以核对断线清空与替换隔离。
				blocked := &chainObserver{Plugin: router.FromInstance(h.components["b"].instance), gate: make(chan struct{}), entered: make(chan struct{})}
				h.assembly.bus.Attach(blocked)
				publish("small", "offline-first", 0)
				<-blocked.entered
				publish("small", "offline-queued", 0)
				beforeReplacement := chainCount(h, "bus", "b", "")
				h.stop(h.components["b"])
				old := h.components["c"]
				if err := old.process.Process.KillTree(); err != nil {
					t.Fatal(err)
				}
				<-old.instance.Session.Done()
				h.assembly.detach(old)
				if _, _, err := old.process.WaitExit(); err != nil {
					t.Fatal(err)
				}
				delete(h.components, "c")
				for _, id := range []string{"b", "c"} {
					if h.assembly.metrics.Snapshot()[id]["bus"]["offline"] == 0 {
						t.Fatalf("no offline queue drops for %s", id)
					}
					h.launch(id, chainMode, 0)
				}
				publish("small", "small-new", 0)
				publish("large", "large-new", 0)
				chainAwait(h, "bus", "b", "small-new", 1)
				chainAwait(h, "bus", "c", "large-new", 1)
				if chainCount(h, "bus", "b", "") != beforeReplacement+1 || chainCount(h, "bus", "c", "") != 1 {
					t.Fatal("bus replayed old queue")
				}
				chainEventually(t, func() bool {
					return h.assembly.metrics.Snapshot()["b"]["bus"]["cancelled"] > 0 && h.assembly.metrics.Snapshot()["c"]["bus"]["cancelled"] > 0
				})
				for _, r := range h.snapshot() {
					if r.Stage == "bus" && (r.User != "" || r.Session != "") {
						t.Fatal("bus inherited user context")
					}
				}
			})
		})
	}
}
