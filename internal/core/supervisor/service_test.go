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
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter/memory"
	"github.com/AlonaBot/AlonaBot/internal/core/broker"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/message"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

type serviceReply struct {
	Text   string
	Caller string
	Calls  int64
	Code   pb.ErrorCode
	Error  string
}

// TestServiceProcesses 验证真实 a/b/c 进程的访问准入、类型化往返和目标函数调用次数。
func TestServiceProcesses(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "svcplugin")
	if output, err := exec.Command("go", "build", "-o", binary, "./testdata/svcplugin").CombinedOutput(); err != nil {
		t.Fatalf("build service plugin: %v\n%s", err, output)
	}
	for _, stdio := range []bool{false, true} {
		for _, badOutput := range []bool{false, true} {
			t.Run(fmt.Sprintf("stdio=%v/bad_output=%v", stdio, badOutput), func(t *testing.T) { testServiceProcesses(t, binary, stdio, badOutput) })
		}
	}
}

func testServiceProcesses(t *testing.T, binary string, stdio, badOutput bool) {
	t.Helper()
	ctx := context.Background()
	store, err := state.Open(ctx, t.TempDir())
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
			t.Logf("Core and service plugin logs:\n%s", logs.String())
		}
	}()
	values := tune.Defaults()
	registry := adapter.NewRegistry()
	platform := memory.New("memory", "instance", "account")
	if err := registry.Add(platform); err != nil {
		t.Fatal(err)
	}
	replies := make(chan string, 1)
	platform.SetSendFunc(func(target *pb.Target, chain *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo) {
		replies <- chain.Segments[0].GetText().GetText()
		return &pb.SendResult{Status: pb.SendStatus_SEND_STATUS_ACCEPTED, Target: target}, nil
	})
	sender := message.New(registry, policy.New(store), nil)
	events := router.New(output.Logger(), values.DispatchQueueEvents, registry, sender.Send, nil)
	events.Start()
	defer events.Close()
	if err := platform.Start(ctx, events); err != nil {
		t.Fatal(err)
	}
	defer platform.Stop(ctx)
	mux := router.NewMux()
	mux.Forward(broker.New(events, output.Logger(), values.RequestTimeout).Call)
	mux.Handle(wire.ServiceSend, sender.Send)
	// 各进程经过真实握手、REGISTER、激活和独立的 Session；测试只用内存平台替代外部平台。
	var running []*Running
	defer func() {
		for _, instance := range running {
			events.Detach(instance.Instance.PluginID)
			stopped := instance.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
			if stopped.Err != nil || stopped.Killed {
				t.Errorf("service plugin stop: %+v", stopped)
			}
		}
	}()
	for _, id := range []string{"a", "b", "c"} {
		digest := []byte("service integration fixture")
		if err := store.AddInstall(ctx, state.Install{Role: state.RolePlugin, PluginID: id, Source: "test", Entrypoint: binary, BuildDigest: digest}); err != nil {
			t.Fatal(err)
		}
		generation := "generation-" + id
		if err := store.AddGeneration(ctx, id, generation, "{}", 1); err != nil {
			t.Fatal(err)
		}
		dir := filepath.Join(t.TempDir(), id)
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		env := map[string]string{"ALONABOT_TEST_SERVICE_ID": id}
		if id == "a" && badOutput {
			env["ALONABOT_TEST_BAD_SERVICE_OUTPUT"] = "1"
		}
		caller := router.NewCaller(id, generation)
		spec := Spec{PluginID: id, Role: state.RolePlugin, GenerationID: generation, PluginDir: dir, Executable: binary, BuildDigest: digest, NewCandidate: true, Stdio: stdio, ExtraEnv: env,
			Bootstrap: &pb.Bootstrap{GenerationId: generation, CandidateConfigJson: "{}", TargetConfigVersion: 1, DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE}},
			Limits:    host.NewLimits(values), Stop: host.NewStopTiming(values), Log: output.Plugin(id, generation), Handler: mux.For(caller),
		}
		instance, err := Launch(ctx, store, spec)
		if err != nil {
			t.Fatalf("launch %s: %v", id, err)
		}
		running = append(running, instance)
		events.Attach(router.FromInstance(instance.Instance), caller)
	}
	target := &pb.Target{AdapterId: "memory", AdapterInstanceId: "instance", AccountId: "account", ChatId: "chat"}
	command := func(t *testing.T, id, operation, service string) serviceReply {
		t.Helper()
		text := strings.Join([]string{id, operation, service}, " ")
		event := &pb.Event{EventType: "message.created", SourceKind: pb.EventSourceKind_EVENT_SOURCE_KIND_ADAPTER, Target: target, Actor: &pb.Actor{ExternalUserId: "user", AdapterInstanceId: "instance"}, EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_MessageCreated{MessageCreated: &pb.MessageCreatedPayload{Chain: &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: text}}}}}}}}}
		if err := platform.Inject(ctx, event); err != nil {
			t.Fatal(err)
		}
		select {
		case text := <-replies:
			var result serviceReply
			if err := json.Unmarshal([]byte(text), &result); err != nil {
				t.Fatal(err)
			}
			return result
		case <-time.After(5 * time.Second):
			t.Fatalf("service reply missing for %q", text)
			return serviceReply{}
		}
	}
	if badOutput {
		reply := command(t, "b", "json", "a/bad")
		if reply.Code != pb.ErrorCode_ERROR_CODE_INTERNAL {
			t.Fatalf("bad output reply=%+v", reply)
		}
		if text := logs.String(); !strings.Contains(text, "service output does not match") || !strings.Contains(text, `"service_name":"bad"`) || !strings.Contains(text, `"plugin_id":"a"`) {
			t.Fatalf("bad output warning missing: %s", text)
		}
		return
	}
	count := func(t *testing.T) int64 {
		t.Helper()
		reply := command(t, "b", "count", "a/count")
		if reply.Code != 0 {
			t.Fatalf("count failed: %+v", reply)
		}
		return reply.Calls
	}
	var expectedCalls int64
	tests := []struct {
		name, id, operation, target string
		code                        pb.ErrorCode
		executed                    bool
	}{
		{"public", "b", "json", "a/pub", 0, true},
		{"version 2 round trip", "b", "v2", "a/v2", 0, true},
		{"allowlist", "b", "json", "a/allow", 0, true},
		{"allowlist denied", "c", "json", "a/allow", pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, false},
		{"private denied", "c", "json", "a/echo", pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, false},
		{"schema", "b", "schema", "a/pub", pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, false},
		{"encoding", "b", "encoding", "a/pub", pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, false},
		{"invalid JSON rejected by SDK", "b", "invalid_json", "a/pub", pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, false},
		{"unknown service", "b", "json", "a/missing", pb.ErrorCode_ERROR_CODE_NOT_FOUND, false},
		{"unavailable plugin", "b", "json", "offline/pub", pb.ErrorCode_ERROR_CODE_UNAVAILABLE, false},
		{"protobuf", "b", "proto", "a/proto", 0, true},
		{"version", "b", "version", "a/pub", pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, false},
	}
	if got := count(t); got != 0 {
		t.Fatalf("initial calls=%d", got)
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reply := command(t, test.id, test.operation, test.target)
			if reply.Code != test.code {
				t.Fatalf("reply=%+v want=%s", reply, test.code)
			}
			if test.executed {
				expectedCalls++
				if reply.Text != "hello 世界" {
					t.Fatalf("echo text=%q", reply.Text)
				}
				if (test.operation == "json" || test.operation == "v2") && (reply.Caller != test.id || reply.Calls != expectedCalls) {
					t.Fatalf("JSON service result=%+v", reply)
				}
			}
			if got := count(t); got != expectedCalls {
				t.Fatalf("function calls=%d want=%d", got, expectedCalls)
			}
		})
	}
}
