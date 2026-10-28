//go:build unix

package supervisor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter/memory"
	"github.com/AlonaBot/AlonaBot/internal/core/conversation"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/message"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

type conversationReply struct {
	Last    uint64
	Entries []struct {
		Seq  uint64
		JSON string
	}
	Next  uint64
	More  bool
	Code  pb.ErrorCode
	Error string
}

// TestConversationProcesses 覆盖真实插件经 SDK 的原子追加、分页、隔离、重启与 purge（V-56）。
func TestConversationProcesses(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "convplugin")
	if out, err := exec.Command("go", "build", "-o", binary, "./testdata/convplugin").CombinedOutput(); err != nil {
		t.Fatalf("build conversation plugin: %v\n%s", err, out)
	}
	ctx := context.Background()
	dataDir := t.TempDir()
	store, err := state.Open(ctx, dataDir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	}()
	logs := &echoLog{}
	output, err := logging.Open(logging.Options{Format: logging.FormatJSON, Level: slog.LevelDebug, Stderr: logs})
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	defer func() {
		if t.Failed() {
			t.Logf("Core and plugin logs:\n%s", logs.String())
		}
	}()
	platform := memory.New("memory", "instance", "account")
	registry := adapter.NewRegistry()
	if err := registry.Add(platform); err != nil {
		t.Fatal(err)
	}
	// 每次命令等待真实发送回执，避免轮询复制大消息的发送历史。
	replies := make(chan string, 1)
	platform.SetSendFunc(func(target *pb.Target, chain *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo) {
		replies <- chain.GetSegments()[0].GetText().GetText()
		return &pb.SendResult{Status: pb.SendStatus_SEND_STATUS_ACCEPTED, Target: target}, nil
	})
	target := &pb.Target{AdapterId: "memory", AdapterInstanceId: "instance", AccountId: "account", ChatId: "chat"}
	values := tune.Defaults()
	specs := make([]Spec, 2)
	for i, id := range []string{"a", "b"} {
		digest := []byte("conversation integration plugin")
		if err := store.AddInstall(ctx, state.Install{PluginID: id, Source: "test", Entrypoint: binary, BuildDigest: digest}); err != nil {
			t.Fatal(err)
		}
		gen := "generation-" + id
		if err := store.AddGeneration(ctx, id, gen, "{}", 1); err != nil {
			t.Fatal(err)
		}
		dir := filepath.Join(t.TempDir(), id)
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		specs[i] = Spec{PluginID: id, GenerationID: gen, PluginDir: dir, Executable: binary, BuildDigest: digest, NewCandidate: true,
			Bootstrap: &pb.Bootstrap{GenerationId: gen, CandidateConfigJson: "{}", TargetConfigVersion: 1, DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE}},
			Limits:    host.NewLimits(values), Stop: fastStop, Log: output.Plugin(id, gen), ExtraEnv: map[string]string{"CONV_PLUGIN_ID": id}}
	}
	var events *router.Router
	var service *conversation.Service
	var running []*Running
	start := func() {
		t.Helper()
		service = conversation.New(store)
		mux := router.NewMux()
		service.Register(mux)
		mux.Handle(wire.ServiceSend, message.New(registry).Send)
		events = router.New(output.Logger(), values.DispatchQueueEvents)
		events.Start()
		if err := platform.Start(ctx, events); err != nil {
			t.Fatal(err)
		}
		for i := range specs {
			specs[i].Handler = mux.For(router.NewCaller(specs[i].PluginID, specs[i].GenerationID))
			instance, err := Launch(ctx, store, specs[i])
			if err != nil {
				t.Fatal(err)
			}
			running = append(running, instance)
			events.Attach(router.FromInstance(instance.Instance))
			specs[i].NewCandidate = false
		}
	}
	stop := func() {
		t.Helper()
		for _, instance := range running {
			events.Detach(instance.Instance.PluginID)
			result := instance.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
			if result.Err != nil || result.Killed {
				t.Errorf("conversation plugin stop: %+v", result)
			}
		}
		running = nil
		if events != nil {
			events.Close()
			events = nil
		}
		if err := platform.Stop(ctx); err != nil {
			t.Error(err)
		}
	}
	defer stop()
	start()
	command := func(id, op, key, args string) conversationReply {
		t.Helper()
		text := strings.Join([]string{id, op, key, args}, " ")
		event := &pb.Event{EventType: "message.created", SourceKind: pb.EventSourceKind_EVENT_SOURCE_KIND_ADAPTER, Target: target,
			EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_MessageCreated{MessageCreated: &pb.MessageCreatedPayload{Chain: &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: text}}}}}}}}}
		if err := platform.Inject(ctx, event); err != nil {
			t.Fatal(err)
		}
		select {
		case text := <-replies:
			var reply conversationReply
			if err := json.Unmarshal([]byte(text), &reply); err != nil {
				t.Fatal(err)
			}
			return reply
		case <-time.After(5 * time.Second):
			t.Fatalf("conversation %s reply not received", op)
			return conversationReply{}
		}
	}
	appendEntries := func(id, key string, entries ...string) conversationReply {
		t.Helper()
		body, err := json.Marshal(entries)
		if err != nil {
			t.Fatal(err)
		}
		return command(id, "append", key, string(body))
	}
	list := func(id, key string, from uint64) conversationReply {
		t.Helper()
		reply := command(id, "list", key, strconv.FormatUint(from, 10))
		if reply.Code != 0 {
			t.Fatalf("list failed: %s", reply.Error)
		}
		return reply
	}
	first := []string{"{ \"n\": 1, \"secret\": \"conversation-sensitive-entry\" }", "{\"n\":2}", "{\"n\":3}"}
	if reply := appendEntries("a", "same", first...); reply.Code != 0 || reply.Last != 3 {
		t.Fatalf("first append: %+v", reply)
	}
	first = append(first, "{\"n\":4}")
	if reply := appendEntries("a", "same", first[3]); reply.Code != 0 || reply.Last != 4 {
		t.Fatalf("second append: %+v", reply)
	}
	assertEntries := func(reply conversationReply, want []string) {
		t.Helper()
		var got []string
		for i, entry := range reply.Entries {
			if entry.Seq != uint64(i+1) {
				t.Fatalf("seq=%d want=%d", entry.Seq, i+1)
			}
			got = append(got, entry.JSON)
		}
		if !reflect.DeepEqual(got, want) || reply.More {
			t.Fatalf("entries count=%d want=%d more=%v", len(got), len(want), reply.More)
		}
	}
	assertEntries(list("a", "same", 0), first)
	if reply := appendEntries("a", "same", "{}", "[]"); reply.Code != pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT {
		t.Fatalf("invalid entry reply=%+v", reply)
	}
	assertEntries(list("a", "same", 0), first)
	other := []string{"{\"owner\":\"b\"}"}
	if reply := appendEntries("b", "same", other...); reply.Code != 0 || reply.Last != 1 {
		t.Fatalf("plugin b append: %+v", reply)
	}
	assertEntries(list("b", "same", 0), other)
	assertEntries(list("a", "same", 0), first)

	// 每次追加和事件载荷均低于帧上限；总条目大小超过一帧。
	var large []string
	const largeEntries = 6
	for i := range largeEntries {
		entry := fmt.Sprintf("{\"index\":%d,\"text\":\"%s\"}", i, strings.Repeat("x", wire.MaxFrame/4))
		large = append(large, entry)
		if reply := appendEntries("a", "pages", entry); reply.Code != 0 || reply.Last != uint64(i+1) {
			t.Fatalf("large append: last=%d code=%s", reply.Last, reply.Code)
		}
	}
	var got []string
	var from uint64
	pages := 0
	for {
		page := list("a", "pages", from)
		pages++
		for _, entry := range page.Entries {
			if entry.Seq != uint64(len(got)+1) {
				t.Fatalf("page seq=%d want=%d", entry.Seq, len(got)+1)
			}
			got = append(got, entry.JSON)
		}
		if !page.More {
			break
		}
		if page.Next != uint64(len(got)+1) {
			t.Fatalf("next=%d want=%d", page.Next, len(got)+1)
		}
		from = page.Next
	}
	if pages < 2 || !reflect.DeepEqual(got, large) {
		t.Fatalf("pagination pages=%d entries=%d", pages, len(got))
	}
	if reply := command("a", "clear", "pages", ""); reply.Code != 0 {
		t.Fatalf("clear: %+v", reply)
	}
	if page := list("a", "pages", 0); len(page.Entries) != 0 || page.More {
		t.Fatal("clear left entries")
	}

	stop()
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err = state.Open(ctx, dataDir)
	if err != nil {
		t.Fatal(err)
	}
	start()
	assertEntries(list("a", "same", 0), first)
	assertEntries(list("b", "same", 0), other)
	if err := service.Purge(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if page := list("a", "same", 0); len(page.Entries) != 0 || page.More {
		t.Fatal("purge left entries")
	}
	assertEntries(list("b", "same", 0), other)
	if strings.Contains(logs.String(), "conversation-sensitive-entry") {
		t.Fatal("conversation entry leaked into logs")
	}
}
