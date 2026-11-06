//go:build unix

package supervisor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter/memory"
	"github.com/AlonaBot/AlonaBot/internal/core/message"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const dispatchAuditEnv = "ALONABOT_TEST_DISPATCH_AUDIT"
const dispatchPriorityEnv = "ALONABOT_TEST_DISPATCH_PRIORITY"
const dispatchWait = 10 * time.Second

type dispatchEntry struct {
	Plugin, Route, Phase string
	Event                json.RawMessage
}
type dispatchHarness struct {
	events   *router.Router
	platform *memory.Adapter
	registry *adapter.Registry
	audit    string
}

func newDispatchHarness(t *testing.T, opt router.Options) *dispatchHarness {
	t.Helper()
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	registry := adapter.NewRegistry()
	platform := memory.New("memory", "dispatch-instance", "dispatch-bot")
	if err := registry.Add(platform); err != nil {
		t.Fatal(err)
	}
	opt.Adapters = registry
	events := router.New(opt)
	events.Start()
	if err := platform.Start(t.Context(), events); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := platform.Stop(context.Background()); err != nil {
			t.Error(err)
		}
		if err := events.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})
	return &dispatchHarness{events: events, platform: platform, registry: registry, audit: path}
}
func (h *dispatchHarness) launch(t *testing.T, binary, id string, priority int, stdio bool) {
	t.Helper()
	store, spec := commandSpec(t, binary, id, "{}")
	spec.Stdio = stdio
	spec.ExtraEnv = map[string]string{dispatchAuditEnv: h.audit, dispatchPriorityEnv: fmt.Sprint(priority)}
	sender := message.New(h.registry, policy.New(store), nil, tune.Defaults().RequestTimeout)
	mux := router.NewMux()
	mux.Handle(wire.ServiceSend, sender.Send)
	caller := router.NewCaller(id, spec.GenerationID)
	spec.Handler = mux.For(caller)
	running, err := Launch(t.Context(), store, spec)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		h.events.Detach(id)
		result := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
		if result.Err != nil || result.Killed {
			t.Errorf("dispatch plugin stop: %+v", result)
		}
		if t.Failed() {
			t.Logf("dispatch plugin %s logs:\n%s", id, spec.Log.(*logBuf).String())
		}
	})
	for _, route := range running.Instance.Routes {
		if route.GetRouteId() == "scheduled.user" {
			route.RequiredPermissions = []string{"user"}
		}
	}
	h.events.Attach(router.FromInstance(running.Instance), caller)
}
func (h *dispatchHarness) read(t *testing.T) []dispatchEntry {
	t.Helper()
	file, err := os.Open(h.audit)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Error(err)
		}
	}()
	decoder := json.NewDecoder(file)
	var entries []dispatchEntry
	for {
		var entry dispatchEntry
		err := decoder.Decode(&entry)
		if errors.Is(err, io.EOF) {
			return entries
		}
		if err != nil {
			t.Fatal(err)
		}
		entries = append(entries, entry)
	}
}
func waitDispatch(t *testing.T, description string, ready func() bool) {
	t.Helper()
	deadline := time.NewTimer(dispatchWait)
	defer deadline.Stop()
	tick := time.NewTicker(5 * time.Millisecond)
	defer tick.Stop()
	for {
		if ready() {
			return
		}
		select {
		case <-tick.C:
		case <-deadline.C:
			t.Fatalf("timed out waiting for %s", description)
		}
	}
}
func (h *dispatchHarness) waitState(t *testing.T, running, queued int) {
	t.Helper()
	waitDispatch(t, fmt.Sprintf("running=%d queued=%d", running, queued), func() bool { s := h.events.Stats(); return s.Running == running && s.Queued == queued })
}
func dispatchInput(id, text string) *pb.Event {
	return &pb.Event{EventId: id, EventType: "message.created", SourceKind: pb.EventSourceKind_EVENT_SOURCE_KIND_ADAPTER,
		AdapterId: proto.String("memory"), AdapterInstanceId: proto.String("dispatch-instance"), AccountId: proto.String("dispatch-bot"),
		Target:       &pb.Target{AdapterId: "memory", AdapterInstanceId: "dispatch-instance", AccountId: "dispatch-bot", ChatId: "dispatch-chat"},
		Actor:        &pb.Actor{ExternalUserId: "user", AdapterInstanceId: "dispatch-instance"},
		EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_MessageCreated{MessageCreated: &pb.MessageCreatedPayload{PlatformMessageId: id, Chain: &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: text}}}}}}}}}
}
func (h *dispatchHarness) inject(t *testing.T, ev *pb.Event) {
	t.Helper()
	if err := h.platform.Inject(t.Context(), ev); err != nil {
		t.Fatal(err)
	}
}

// TestDispatchProcessesBothTransports 经真实插件、Session 与内存 Adapter 验证有界链及可信 fixture。
func TestDispatchProcessesBothTransports(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "dispatchplugin")
	if out, err := exec.Command("go", "build", "-o", binary, "./testdata/dispatchplugin").CombinedOutput(); err != nil {
		t.Fatalf("build dispatch plugin: %v\n%s", err, out)
	}
	for _, stdio := range []bool{false, true} {
		t.Run(fmt.Sprintf("stdio=%v", stdio), func(t *testing.T) {
			t.Run("stable-order", func(t *testing.T) {
				for _, order := range [][]string{{"b", "c", "a"}, {"a", "c", "b"}} {
					t.Run(strings.Join(order, ""), func(t *testing.T) {
						h := newDispatchHarness(t, router.Options{Concurrency: 2, QueueEvents: 3, QueueBytes: tune.Defaults().DispatchQueueBytes})
						for _, id := range order {
							priority := 10
							if id == "c" {
								priority = -10
							}
							h.launch(t, binary, id, priority, stdio)
						}
						h.inject(t, dispatchInput("order", "0s"))
						h.waitState(t, 0, 0)
						var got []string
						for _, e := range h.read(t) {
							if e.Phase == "entered" {
								got = append(got, e.Plugin+":"+e.Route)
							}
						}
						want := []string{"c:a", "c:z", "a:a", "a:z", "b:a", "b:z"}
						if !reflect.DeepEqual(got, want) {
							t.Fatalf("calls=%v, want %v", got, want)
						}
						if got := len(h.platform.Sent()); got != len(want) {
							t.Fatalf("replies=%d, want %d", got, len(want))
						}
					})
				}
			})
			for _, limit := range []string{"events", "bytes", "global"} {
				t.Run(limit, func(t *testing.T) {
					event := dispatchInput("one", "2s")
					size := int64(proto.Size(event))
					opt := router.Options{Concurrency: 1, QueueEvents: 3, QueueBytes: tune.Defaults().DispatchQueueBytes}
					switch limit {
					case "events":
						opt.Concurrency = 2
						opt.QueueEvents = 1
					case "bytes":
						opt.QueueBytes = size
					case "global":
						opt.Global = wire.NewBudget(size)
					}
					h := newDispatchHarness(t, opt)
					h.launch(t, binary, "limits", 0, stdio)
					for i := range opt.Concurrency {
						ev := proto.CloneOf(event)
						ev.EventId = fmt.Sprintf("r%02d", i)
						h.inject(t, ev)
						h.waitState(t, i+1, 0)
					}
					queued := proto.CloneOf(event)
					h.inject(t, queued)
					h.waitState(t, opt.Concurrency, 1)
					if err := h.platform.Inject(t.Context(), event); !errors.Is(err, adapter.ErrQueueFull) {
						t.Fatalf("over limit=%v", err)
					}
					want := router.Stats{Accepted: uint64(opt.Concurrency + 1), DroppedQueueFull: 1, Running: opt.Concurrency, Queued: 1}
					if limit == "global" {
						want.DroppedQueueFull = 0
						want.DroppedBudget = 1
						if used := opt.Global.Used(); used != size {
							t.Fatalf("budget=%d want %d", used, size)
						}
					}
					if got := h.events.Stats(); got != want {
						t.Fatalf("stats=%+v want %+v", got, want)
					}
					h.waitState(t, 0, 0)
					if opt.Global != nil && opt.Global.Used() != 0 {
						t.Fatalf("drained budget=%d", opt.Global.Used())
					}
					if got := len(h.platform.Sent()); got != 2*(opt.Concurrency+1) {
						t.Fatalf("replies=%d", got)
					}
					called := make(map[string]int)
					for _, entry := range h.read(t) {
						if entry.Phase == "entered" {
							var ev pb.Event
							if err := protojson.Unmarshal(entry.Event, &ev); err != nil {
								t.Fatal(err)
							}
							called[ev.EventId]++
						}
					}
					if len(called) != opt.Concurrency+1 || called[queued.EventId] != 2 {
						t.Fatalf("queued event calls=%v", called)
					}
					for i := range opt.Concurrency {
						if called[fmt.Sprintf("r%02d", i)] != 2 {
							t.Fatalf("running event calls=%v", called)
						}
					}
				})
			}
			t.Run("session-response-during-handler", func(t *testing.T) {
				h := newDispatchHarness(t, router.Options{Concurrency: 1, QueueEvents: 3, QueueBytes: tune.Defaults().DispatchQueueBytes})
				h.launch(t, binary, "probe", 0, stdio)
				h.inject(t, dispatchInput("probe-event", "2s probe"))
				h.waitState(t, 1, 0)
				waitDispatch(t, "independent core.send Response", func() bool {
					for _, e := range h.read(t) {
						if e.Phase == "send.response" {
							return true
						}
					}
					return false
				})
				if s := h.events.Stats(); s.Running != 1 {
					t.Fatalf("slow chain ended before probe Response: %+v", s)
				}
				for _, e := range h.read(t) {
					if e.Phase == "finished" {
						t.Fatal("slow handler finished before independent Response")
					}
				}
				sent := h.platform.Sent()
				if len(sent) != 1 || sent[0].Chain.GetSegments()[0].GetText().GetText() != "probe:probe-event" {
					t.Fatalf("probe sends=%v", sent)
				}
				h.waitState(t, 0, 0)
			})
			t.Run("fixtures", func(t *testing.T) {
				h := newDispatchHarness(t, router.Options{Concurrency: 2, QueueEvents: 3, QueueBytes: tune.Defaults().DispatchQueueBytes})
				h.launch(t, binary, "fixtures", 0, stdio)
				originals := map[string]*pb.Event{}
				for _, name := range []string{"button.clicked", "scheduled"} {
					raw, err := os.ReadFile(filepath.Join("..", "..", "..", "contracts", "testdata", "events", name+".json"))
					if err != nil {
						t.Fatal(err)
					}
					ev := &pb.Event{}
					if err := protojson.Unmarshal(raw, ev); err != nil {
						t.Fatal(err)
					}
					originals[name] = proto.CloneOf(ev)
					if err := h.events.Deliver(t.Context(), ev); err != nil {
						t.Fatal(err)
					}
				}
				h.waitState(t, 0, 0)
				entries := h.read(t)
				if len(entries) != 2 {
					t.Fatalf("fixture calls=%v", entries)
				}
				for _, entry := range entries {
					ev := &pb.Event{}
					if err := protojson.Unmarshal(entry.Event, ev); err != nil {
						t.Fatal(err)
					}
					want := originals[ev.EventType]
					if ev.EventType == "scheduled" {
						if entry.Route != "scheduled.open" || ev.Actor != nil || ev.Target != nil {
							t.Fatalf("scheduled identity/permission=%v %s", ev, entry.Route)
						}
					} else {
						want.Actor.Permissions = []string{"user"}
					}
					if !proto.Equal(ev, want) {
						t.Fatalf("fixture event=%v want %v", ev, want)
					}
				}
				if len(h.platform.Sent()) != 0 {
					t.Fatal("fixture attempted implicit reply")
				}
			})
		})
	}
}
