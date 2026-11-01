//go:build unix

package remote_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter/remote"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk"
	sdkadapter "github.com/AlonaBot/AlonaBot/sdk/adapter"
	"google.golang.org/protobuf/proto"
)

const helperEnv = "ALONABOT_LOOP_ADAPTER"

// TestMain 将测试二进制切换为经真实 SDK 运行的回环 Adapter 或插件。
func TestMain(m *testing.M) {
	mode := os.Getenv(helperEnv)
	if mode == "" {
		os.Exit(m.Run())
	}
	var err error
	if mode == "plugin" {
		err = sdk.Run(&loopPlugin{})
	} else {
		err = sdk.RunAdapter(&loopAdapter{})
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

type loopPlugin struct{ sdk.Base }

func (*loopPlugin) Events(r *sdk.EventRegistry) {
	r.On("loop", "message.created", func(*sdk.Context) error { return nil })
}

type loopAdapter struct{}

func (*loopAdapter) Descriptor() *pb.AdapterDescriptor {
	return &pb.AdapterDescriptor{AdapterId: "loop"}
}
func (*loopAdapter) Start(_ context.Context, req sdkadapter.StartRequest, core sdkadapter.Core) (sdkadapter.Instance, error) {
	var config struct {
		Account     string `json:"account"`
		Token       string `json:"token"`
		DisplayName string `json:"display_name"`
		Username    string `json:"username"`
		Media       bool   `json:"media"`
	}
	if err := json.Unmarshal(req.ConfigJSON, &config); err != nil {
		return nil, err
	}
	if config.Account == "start-error" {
		return nil, fmt.Errorf("invalid token: %s", config.Token)
	}
	n := 0
	if len(req.Cursor) > 0 {
		var err error
		n, err = strconv.Atoi(string(req.Cursor))
		if err != nil {
			return nil, err
		}
	}
	instance := &loopInstance{id: req.InstanceID, account: config.Account, displayName: config.DisplayName, username: config.Username, core: core, sequence: n, initial: string(req.Cursor), token: config.Token}
	if config.Media {
		return &mediaLoopInstance{loopInstance: instance}, nil
	}
	return instance, nil
}

type mediaLoopInstance struct{ *loopInstance }

func (*mediaLoopInstance) ReadPlatformFile(_ context.Context, fileID string, offset uint64, length uint32) ([]byte, uint64, bool, error) {
	if fileID != "platform-file" {
		return nil, 0, false, &sdk.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "platform file not found"}
	}
	const content = "image bytes"
	if offset > uint64(len(content)) {
		return nil, 0, false, &sdk.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "media offset exceeds file size"}
	}
	end := min(offset+uint64(length), uint64(len(content)))
	return []byte(content[offset:end]), end, end == uint64(len(content)), nil
}

type loopInstance struct {
	token                 string
	stopFailed            bool
	id, account, initial  string
	displayName, username string
	core                  sdkadapter.Core
	mu                    sync.Mutex
	sequence              int
}

func (i *loopInstance) Account() sdkadapter.Account {
	return sdkadapter.Account{ID: i.account, DisplayName: i.displayName, Username: i.username}
}
func (i *loopInstance) Capabilities() []*pb.Capability {
	return []*pb.Capability{{Name: "message.text", Version: 1, Supported: true}}
}
func (i *loopInstance) Stop(context.Context) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.account == "stop-error" && !i.stopFailed {
		i.stopFailed = true
		return errors.New("platform stop failed: " + i.token)
	}
	return nil
}
func (i *loopInstance) Send(ctx context.Context, target *pb.Target, chain *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo) {
	i.mu.Lock()
	defer i.mu.Unlock()
	text := chain.GetSegments()[0].GetText().GetText()
	if text == "fail" {
		if err := i.core.ReportFailure(ctx, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAUTHENTICATED, SafeMessage: "account rejected"}); err != nil {
			return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}
		}
		if _, err := i.core.Deliver(ctx, &pb.Event{}, nil); err == nil {
			return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "failed client accepted event"}
		}
		return &pb.SendResult{Status: pb.SendStatus_SEND_STATUS_ACCEPTED}, nil
	}
	if text == "cursor" {
		return &pb.SendResult{Status: pb.SendStatus_SEND_STATUS_ACCEPTED, PlatformMessageIds: []string{i.initial}}, nil
	}
	i.sequence++
	id := i.id
	if text == "forge" {
		id = "other-instance"
	}
	ev := &pb.Event{EventId: "adapter-forged-id", EventType: "message.created", AdapterId: proto.String("loop"), AdapterInstanceId: proto.String(id), AccountId: proto.String(i.account), Target: target, EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_MessageCreated{MessageCreated: &pb.MessageCreatedPayload{Chain: chain, PlatformMessageId: strconv.Itoa(i.sequence)}}}}
	switch text {
	case "bot.online", "bot.offline", "scheduled":
		ev.EventType = text
	case "mismatched-payload":
		ev.EventType = "message.edited"
	}
	accepted, err := i.core.Deliver(ctx, ev, []byte(strconv.Itoa(i.sequence)))
	if err != nil {
		return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: err.Error()}
	}
	if !accepted {
		return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, SafeMessage: "event queue full"}
	}
	return &pb.SendResult{Status: pb.SendStatus_SEND_STATUS_ACCEPTED}, nil
}

type eventSink struct {
	mu     sync.Mutex
	events []*pb.Event
}

func (s *eventSink) Deliver(_ context.Context, ev *pb.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, proto.Clone(ev).(*pb.Event))
	return nil
}
func (s *eventSink) snapshot() []*pb.Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]*pb.Event(nil), s.events...)
}

type processLog struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *processLog) Line(_, _ string, b []byte) { l.mu.Lock(); defer l.mu.Unlock(); l.b.Write(b) }
func (*processLog) SetSession([]byte)            {}

func (l *processLog) Write(b []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(b)
}

func setup(t *testing.T) (*state.Store, supervisor.Spec, *processLog) {
	t.Helper()
	ctx := context.Background()
	dir := t.TempDir()
	store, err := state.Open(ctx, filepath.Join(dir, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	digest := bytes.Repeat([]byte{1}, 32)
	if err := store.AddInstall(ctx, state.Install{PluginID: "loop-package", Role: state.RoleAdapter, AdapterID: "loop", BuildDigest: digest}); err != nil {
		t.Fatal(err)
	}
	if err := store.AddGeneration(ctx, "loop-package", "gen", "{}", 1); err != nil {
		t.Fatal(err)
	}
	pluginDir := filepath.Join(dir, "plugin")
	if err := os.MkdirAll(pluginDir, 0o700); err != nil {
		t.Fatal(err)
	}
	values := tune.Defaults()
	limits := host.NewLimits(values)
	limits.Handshake = 3 * time.Second
	limits.OnReady = 3 * time.Second
	logs := &processLog{}
	return store, supervisor.Spec{PluginID: "loop-package", Role: state.RoleAdapter, AdapterID: "loop", PluginDir: pluginDir, Executable: os.Args[0], BuildDigest: digest, GenerationID: "gen", Bootstrap: &pb.Bootstrap{GenerationId: "gen", CandidateConfigJson: "{}", TargetConfigVersion: 1}, NewCandidate: true, Limits: limits, Stop: host.NewStopTiming(values), Log: logs, ExtraEnv: map[string]string{helperEnv: "adapter", "GORACE": "atexit_sleep_ms=0"}}, logs
}

// TestRemoteLifecycle 覆盖真实进程的多实例收发、账号唯一性、伪造拒绝、故障与游标恢复。
func TestRemoteLifecycle(t *testing.T) {
	ctx := context.Background()
	store, spec, logs := setup(t)
	registry := adapter.NewRegistry()
	sink := &eventSink{}
	p := remote.New(remote.Options{AdapterID: "loop", PluginID: spec.PluginID, Store: store, Registry: registry, Sink: sink, Log: slog.New(slog.NewTextHandler(logs, nil))})
	launch := func() *supervisor.Running {
		t.Helper()
		mux := router.NewMux()
		p.Services(mux)
		caller := router.NewCaller(spec.PluginID, spec.GenerationID)
		spec.Handler = mux.For(caller)
		running, err := supervisor.Launch(ctx, store, spec)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if res := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE); res.Err != nil {
				t.Error(res.Err)
			}
		})
		if err := p.Bind(ctx, running.Instance, caller); err != nil {
			t.Fatal(err)
		}
		return running
	}
	running := launch()
	pid := running.Process.Pid()
	start := func(id, account string) {
		t.Helper()
		if err := p.StartInstance(ctx, id, []byte(`{"account":"`+account+`","display_name":"Loop display","token":"credential-canary"}`)); err != nil {
			t.Fatal(err)
		}
	}
	send := func(id, text string) (*pb.SendResult, *pb.ErrorInfo) {
		t.Helper()
		in, ok := registry.Get(id)
		if !ok {
			t.Fatalf("missing instance %s", id)
		}
		return in.Send(ctx, &pb.Target{AdapterId: "loop", AdapterInstanceId: id, AccountId: in.AccountID(), ChatId: "chat"}, &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: text}}}}})
	}
	mustSend := func(id, text string) *pb.SendResult {
		t.Helper()
		out, safe := send(id, text)
		if safe != nil {
			t.Fatal(safe)
		}
		return out
	}
	count := func(eventType, id string, reason pb.BotOfflineReason) int {
		n := 0
		for _, ev := range sink.snapshot() {
			if ev.EventType == eventType && ev.GetAdapterInstanceId() == id && (eventType != "bot.offline" || ev.GetEventPayload().GetBotOffline().GetReason() == reason) {
				n++
			}
		}
		return n
	}
	if err := p.StartInstance(ctx, "one", []byte(`{"account":"a","display_name":"Loop display","username":"loop_bot","token":"credential-canary"}`)); err != nil {
		t.Fatal(err)
	}
	start("two", "b")
	for _, c := range []struct{ id, username string }{{"one", "loop_bot"}, {"two", ""}} {
		in, ok := registry.Get(c.id)
		if !ok {
			t.Fatalf("missing instance %s", c.id)
		}
		named, ok := in.(adapter.Named)
		if !ok || named.AccountName() != c.username {
			t.Fatalf("instance %s username = %v, want %q", c.id, in, c.username)
		}
	}
	for _, ev := range sink.snapshot() {
		if got := ev.GetEventPayload().GetBotOnline().GetAccountDisplayName(); got != "Loop display" {
			t.Fatalf("online display name = %q, want Loop display", got)
		}
	}
	if count("bot.online", "one", 0) != 1 || count("bot.online", "two", 0) != 1 {
		t.Fatal("missing online events")
	}
	mustSend("one", "hello")
	mustSend("two", "world")
	ev := sink.snapshot()[2]
	if ev.GetEventPayload().GetMessageCreated().GetChain().GetSegments()[0].GetText().GetText() != "hello" || ev.EventId == "adapter-forged-id" || len(ev.EventId) != 32 || ev.SourceKind != pb.EventSourceKind_EVENT_SOURCE_KIND_ADAPTER || len(ev.Capabilities) != 1 {
		t.Fatalf("event was not normalized: %v", ev)
	}
	before, err := store.AdapterCursor(ctx, "two")
	if !bytes.Equal(before, []byte("1")) {
		t.Fatalf("delivered cursor = %q, want 1", before)
	}
	if err != nil {
		t.Fatal(err)
	}
	eventsBefore := len(sink.snapshot())
	if _, safe := send("two", "forge"); safe == nil {
		t.Fatal("forged instance accepted")
	}
	after, err := store.AdapterCursor(ctx, "two")
	if err != nil || !bytes.Equal(before, after) || len(sink.snapshot()) != eventsBefore {
		t.Fatal("forgery advanced cursor or delivered event")
	}
	for _, text := range []string{"bot.online", "bot.offline", "scheduled", "mismatched-payload"} {
		if _, safe := send("two", text); safe == nil {
			t.Fatalf("invalid event %s accepted", text)
		}
		current, err := store.AdapterCursor(ctx, "two")
		if err != nil || !bytes.Equal(before, current) || len(sink.snapshot()) != eventsBefore {
			t.Fatalf("invalid event %s changed cursor or Sink", text)
		}
	}
	if err := p.StartInstance(ctx, "duplicate", []byte(`{"account":"a"}`)); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS {
		t.Fatalf("duplicate account: %v", err)
	}
	if _, ok := registry.Get("duplicate"); ok {
		t.Fatal("duplicate account registered")
	}
	if err := p.StopInstance(ctx, "one"); err != nil {
		t.Fatal(err)
	}
	if count("bot.offline", "one", pb.BotOfflineReason_BOT_OFFLINE_REASON_STOPPED) != 1 || running.Process.Pid() != pid {
		t.Fatal("instance stop changed process or missed offline")
	}
	mustSend("two", "still alive")
	start("failed", "c")
	mustSend("failed", "fail")
	if _, ok := registry.Get("failed"); ok {
		t.Fatal("failed instance remained registered")
	}
	if count("bot.offline", "failed", pb.BotOfflineReason_BOT_OFFLINE_REASON_INSTANCE_FAILED) != 1 {
		t.Fatal("missing failure event")
	}
	for _, ev := range sink.snapshot() {
		if ev.EventType == "bot.offline" && ev.GetAdapterInstanceId() == "failed" && ev.EventPayload.GetBotOffline().GetSafeError().GetSafeMessage() != "account rejected" {
			t.Fatal("safe error lost")
		}
	}
	// 故障实例可在同一个 Adapter 进程内重新启用。
	start("failed", "c")
	mustSend("failed", "recovered")
	if err := p.StopInstance(ctx, "failed"); err != nil {
		t.Fatal(err)
	}
	// 停止 RPC 报错仍关闭 Core 运行登记，且保留原始错误。
	start("stop-error", "stop-error")
	if err := p.StopInstance(ctx, "stop-error"); err == nil || !strings.Contains(err.Error(), "platform stop failed") {
		t.Fatalf("stop error lost: %v", err)
	}
	if _, ok := registry.Get("stop-error"); ok {
		t.Fatal("failed stop remained registered")
	}
	if count("bot.offline", "stop-error", pb.BotOfflineReason_BOT_OFFLINE_REASON_STOPPED) != 1 {
		t.Fatal("failed stop missed offline event")
	}
	// 未清理的 SDK 实例在显式重新启用前先完成清理。
	start("stop-error", "stop-error")
	if err := p.StopInstance(ctx, "stop-error"); err == nil {
		t.Fatal("expected new instance stop failure")
	}
	err = p.StartInstance(ctx, "start-error", []byte(`{"account":"start-error","token":"credential-canary"}`))
	if err == nil || strings.Contains(err.Error(), "credential-canary") {
		t.Fatalf("startup credential leaked: %v", err)
	}
	start("three", "d")
	mustSend("three", "before crash")
	saved := map[string][]byte{}
	for _, id := range []string{"two", "three"} {
		saved[id], err = store.AdapterCursor(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
	}
	oldInstance, _ := registry.Get("two")
	oldSessionKey := oldInstance.(adapter.MediaSession).MediaSessionKey()
	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	if _, _, err := running.WaitExit(); err != nil {
		t.Fatal(err)
	}
	p.Unbind()
	p.Unbind()
	for _, id := range []string{"two", "three"} {
		if count("bot.offline", id, pb.BotOfflineReason_BOT_OFFLINE_REASON_ADAPTER_EXITED) != 1 {
			t.Fatalf("wrong exit event count for %s", id)
		}
	}
	spec.NewCandidate = false
	running = launch()
	restored, _ := registry.Get("two")
	if restored.(adapter.MediaSession).MediaSessionKey() == oldSessionKey || oldInstance.(adapter.MediaSession).MediaSessionKey() != oldSessionKey {
		t.Fatal("restored instance reused its old media session or changed the old instance identity")
	}
	for _, id := range []string{"one", "failed", "duplicate"} {
		if _, ok := registry.Get(id); ok {
			t.Fatalf("disabled instance %s restored", id)
		}
	}
	for _, id := range []string{"two", "three"} {
		out := mustSend(id, "cursor")
		if len(out.PlatformMessageIds) != 1 || out.PlatformMessageIds[0] != string(saved[id]) {
			t.Fatalf("cursor not restored for %s: %v", id, out)
		}
	}
	for _, ev := range sink.snapshot() {
		if ev.EventType == "bot.online" || ev.EventType == "bot.offline" {
			if ev.Actor != nil || ev.Target != nil || ev.Message != nil {
				t.Fatal("lifecycle event contains platform context")
			}
		}
	}
	logs.mu.Lock()
	defer logs.mu.Unlock()
	if bytes.Contains(logs.b.Bytes(), []byte("credential-canary")) {
		t.Fatal("configuration leaked into logs")
	}
}

// TestRemoteRoleMismatch 在进程启动边界确认两种角色伪装都在 REGISTER 阶段失败。
func TestRemoteRoleMismatch(t *testing.T) {
	for _, mode := range []string{"adapter", "plugin"} {
		t.Run(mode, func(t *testing.T) {
			store, spec, _ := setup(t)
			spec.ExtraEnv[helperEnv] = mode
			if mode == "adapter" {
				spec.Role = state.RolePlugin
			}
			running, err := supervisor.Launch(context.Background(), store, spec)
			if running != nil {
				running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
				t.Fatal("role mismatch activated")
			}
			var phase *host.PhaseError
			if !errors.As(err, &phase) || phase.Phase != pb.SessionPhase_SESSION_PHASE_REGISTER {
				t.Fatalf("role mismatch failed outside REGISTER: %v", err)
			}
		})
	}
}

// TestRemoteMediaRead 经真实 Adapter 进程分块取件，保持 Session 身份与平台错误。
func TestRemoteMediaRead(t *testing.T) {
	ctx := t.Context()
	store, spec, logs := setup(t)
	registry := adapter.NewRegistry()
	p := remote.New(remote.Options{AdapterID: "loop", PluginID: spec.PluginID, Store: store, Registry: registry, Sink: &eventSink{}, Log: slog.New(slog.NewTextHandler(logs, nil))})
	mux := router.NewMux()
	p.Services(mux)
	caller := router.NewCaller(spec.PluginID, spec.GenerationID)
	spec.Handler = mux.For(caller)
	running, err := supervisor.Launch(ctx, store, spec)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if result := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE); result.Err != nil {
			t.Error(result.Err)
		}
	})
	if err := p.Bind(ctx, running.Instance, caller); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ id, config string }{
		{"readable", `{"account":"media-reader-account","media":true}`},
		{"unsupported", `{"account":"unsupported-account"}`},
	} {
		if err := p.StartInstance(ctx, c.id, []byte(c.config)); err != nil {
			t.Fatal(err)
		}
	}
	in, _ := registry.Get("readable")
	if got := in.(adapter.MediaSession).MediaSessionKey(); got != caller.SessionKey {
		t.Fatalf("media session = %q, want %q", got, caller.SessionKey)
	}
	reader := in.(adapter.MediaReader)
	var content []byte
	var offset uint64
	for {
		data, next, eof, err := reader.ReadMedia(ctx, "platform-file", offset, 4)
		if err != nil {
			t.Fatal(err)
		}
		content = append(content, data...)
		if next != offset+uint64(len(data)) {
			t.Fatalf("next offset = %d after %d bytes", next, len(data))
		}
		offset = next
		if eof {
			break
		}
	}
	if string(content) != "image bytes" {
		t.Fatalf("content = %q", content)
	}
	if data, next, eof, err := reader.ReadMedia(ctx, "platform-file", offset, 4); err != nil || len(data) != 0 || next != offset || !eof {
		t.Fatalf("read at EOF = %q %d %v %v", data, next, eof, err)
	}
	if _, _, _, err := reader.ReadMedia(ctx, "missing", 0, 4); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_NOT_FOUND {
		t.Fatalf("missing file error = %v", err)
	} else {
		var call *wire.CallError
		if !errors.As(err, &call) || call.ErrorInfo.SafeMessage != "platform file not found" {
			t.Fatalf("platform safe message changed: %v", err)
		}
	}
	for _, length := range []uint32{0, wire.MaxMediaChunk + 1} {
		if _, _, _, err := reader.ReadMedia(ctx, "platform-file", 0, length); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT {
			t.Fatalf("length %d error = %v", length, err)
		}
	}
	unsupported, _ := registry.Get("unsupported")
	if _, _, _, err := unsupported.(adapter.MediaReader).ReadMedia(ctx, "platform-file", 0, 4); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_UNSUPPORTED {
		t.Fatalf("unsupported instance error = %v", err)
	}
	if err := p.StopInstance(ctx, "readable"); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := reader.ReadMedia(ctx, "platform-file", 0, 4); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_UNAVAILABLE {
		t.Fatalf("stopped instance error = %v", err)
	}
}
