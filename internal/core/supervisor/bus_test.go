//go:build unix

package supervisor

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/broker"
	"github.com/AlonaBot/AlonaBot/internal/core/bus"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

const (
	busControlSchema = "main.control"
	busReplySchema   = "main.reply"
	busTestTimeout   = 5 * time.Second
)

type busControl struct {
	Operation string
	Event     string
	Number    int
	Version   uint32
}

type busReply struct {
	EventID   string
	Seen      []int
	Versioned []int
	Entered   bool
	Code      pb.ErrorCode
	Error     string
}

// busWriteGate 只阻塞 slow 的写出；其余方法与开闸后的发送都委托真实插件实例。
type busWriteGate struct {
	router.Plugin
	entered chan struct{}
	release chan struct{}
	once    sync.Once // 只报告首条事件已离开 Bus 等待队列
}

func (p *busWriteGate) SendEvent(ctx context.Context, target string, ev *pb.BusEvent) error {
	p.once.Do(func() { close(p.entered) })
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.release:
		return p.Plugin.SendEvent(ctx, target, ev)
	}
}

// TestBusProcesses 验证真实进程的广播顺序、双类慢订阅者隔离、版本匹配与离线不补发。
func TestBusProcesses(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "busplugin")
	if output, err := exec.Command("go", "build", "-race", "-o", binary, "./testdata/busplugin").CombinedOutput(); err != nil {
		t.Fatalf("build bus plugin: %v\n%s", err, output)
	}
	for _, stdio := range []bool{false, true} {
		t.Run(fmt.Sprintf("stdio=%v", stdio), func(t *testing.T) { testBusProcesses(t, binary, stdio) })
	}
}

func testBusProcesses(t *testing.T, binary string, stdio bool) {
	t.Helper()
	// 1. 装配真实路由、Broker 与 Bus；控制 Service 与广播共用各插件的 Session。
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
	setDefaultLogger(t, output.Logger())
	defer output.Close()
	defer func() {
		if t.Failed() {
			t.Logf("Core and bus plugin logs:\n%s", logs.String())
		}
	}()
	values := tune.Defaults()
	routes := router.New(values.DispatchQueueEvents, nil, nil, nil, nil)
	defer routes.Close()
	services := broker.New(routes, values.RequestTimeout)
	mux := router.NewMux()
	mux.Forward(services.Call)
	broadcast := bus.New(bus.Options{QueueEvents: 3, QueueBytes: values.BusQueueBytes})
	broadcast.Register(mux)
	budget := wire.NewBudget(values.CoreMaxPayload)
	running := make(map[string]*Running)
	specs := make(map[string]Spec)
	defer func() {
		for id, instance := range running {
			routes.Detach(id)
			broadcast.Detach(id)
			stopped := instance.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
			if stopped.Err != nil || stopped.Killed {
				t.Errorf("stop bus plugin %s: %+v", id, stopped)
			}
		}
	}()
	gate := &busWriteGate{entered: make(chan struct{}), release: make(chan struct{})}
	for _, id := range []string{"pub", "fast", "slow", "stuck"} {
		digest := []byte("bus integration fixture")
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
		caller := router.NewCaller(id, generation)
		spec := Spec{PluginID: id, Role: state.RolePlugin, GenerationID: generation, PluginDir: dir, Executable: binary, BuildDigest: digest,
			NewCandidate: true, Stdio: stdio, ExtraEnv: map[string]string{"ALONABOT_TEST_BUS_ID": id},
			Bootstrap: &pb.Bootstrap{GenerationId: generation, CandidateConfigJson: "{}", TargetConfigVersion: 1, DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_NONE}},
			Limits:    host.NewLimits(values), Stop: host.NewStopTiming(values), Log: output.Plugin(id, generation), Global: budget, Handler: mux.For(caller),
		}
		instance, err := Launch(ctx, store, spec)
		if err != nil {
			t.Fatalf("launch %s: %v", id, err)
		}
		running[id], specs[id] = instance, spec
		plugin := router.FromInstance(instance.Instance)
		if id == "slow" {
			gate.Plugin = plugin
			plugin = gate
		}
		routes.Attach(plugin, caller)
		broadcast.Attach(plugin)
	}
	control := func(id string, request busControl) busReply {
		t.Helper()
		encoded, err := json.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		callCtx, cancel := context.WithTimeout(ctx, busTestTimeout)
		defer cancel()
		response := services.Call(callCtx, router.Caller{}, &wire.Inbound{Target: id + "/control", Payload: wire.Payload{
			Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, SchemaID: busControlSchema, SchemaVersion: 1, Bytes: encoded,
		}})
		if response.Err != nil {
			t.Fatalf("control %s/%s: %v", id, request.Operation, response.Err)
		}
		if response.Payload.SchemaID != busReplySchema {
			t.Fatalf("control response schema: %s", response.Payload.SchemaID)
		}
		var result busReply
		if err := json.Unmarshal(response.Payload.Bytes, &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	snapshot := func(id string) busReply { return control(id, busControl{Operation: "snapshot"}) }
	wait := func(description string, condition func() bool) {
		t.Helper()
		deadline := time.Now().Add(busTestTimeout)
		for !condition() {
			if time.Now().After(deadline) {
				t.Fatalf("timeout waiting for %s; stats: fast=%+v slow=%+v stuck=%+v", description, broadcast.Stats("fast"), broadcast.Stats("slow"), broadcast.Stats("stuck"))
			}
			time.Sleep(time.Millisecond)
		}
	}
	ids := make(map[string]bool)
	publish := func(name string, version uint32, number int) {
		t.Helper()
		result := control("pub", busControl{Operation: "publish", Event: name, Version: version, Number: number})
		raw, err := hex.DecodeString(result.EventID)
		if result.Code != 0 || err != nil || len(raw) != 16 || ids[result.EventID] {
			t.Fatalf("publish %d: %+v; event ID error: %v", number, result, err)
		}
		ids[result.EventID] = true
	}
	want := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	// 2. 首条离队后再填满 slow；每次确认 fast 已处理，避免用调度偶然性制造溢出。
	publish("x", 1, 1)
	select {
	case <-gate.entered:
	case <-time.After(busTestTimeout):
		t.Fatal("slow did not enter SendEvent")
	}
	wait("stuck callback entry", func() bool { return snapshot("stuck").Entered })
	for number := 2; number <= len(want); number++ {
		wait("fast processing and stuck write", func() bool {
			return slices.Equal(snapshot("fast").Seen, want[:number-1]) && broadcast.Stats("stuck").Delivered == uint64(number-1)
		})
		publish("x", 1, number)
	}
	wait("fast full sequence and stuck writes", func() bool {
		return slices.Equal(snapshot("fast").Seen, want) && broadcast.Stats("fast").Delivered == uint64(len(want)) && broadcast.Stats("stuck").Delivered == uint64(len(want))
	})
	if got := broadcast.Stats("slow"); got.Delivered != 0 || got.Dropped != 6 {
		t.Fatalf("blocked slow stats: %+v", got)
	}
	if got := snapshot("stuck"); !got.Entered || len(got.Seen) != 0 || broadcast.Stats("stuck").Dropped != 0 {
		t.Fatalf("stuck callback did not remain blocked: %+v", got)
	}
	if got := broadcast.Stats("pub").Published; got != 10 {
		t.Fatalf("published=%d, want 10", got)
	}
	if result := control("pub", busControl{Operation: "mismatched_version"}); result.Code != 0 || result.EventID == "" {
		t.Fatalf("version probe rejected: %+v", result)
	}
	close(gate.release)
	control("stuck", busControl{Operation: "release"})
	wait("slow first four and stuck full sequence", func() bool {
		return slices.Equal(snapshot("slow").Seen, want[:4]) && broadcast.Stats("slow").Delivered == 4 && slices.Equal(snapshot("stuck").Seen, want)
	})
	if broadcast.Stats("slow").Dropped != 6 || broadcast.Stats("fast").Dropped != 0 {
		t.Fatalf("unexpected drops: fast=%+v slow=%+v", broadcast.Stats("fast"), broadcast.Stats("slow"))
	}
	// 3. 越权发布不准入；不同版本只到匹配订阅，且不增加丢弃数。
	beforePublished := broadcast.Stats("pub").Published
	if result := control("pub", busControl{Operation: "foreign"}); result.Code != pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED || result.EventID != "" {
		t.Fatalf("foreign namespace accepted: %+v", result)
	}
	if broadcast.Stats("pub").Published != beforePublished {
		t.Fatal("rejected publication was counted")
	}
	fastBefore := broadcast.Stats("fast")
	slowBefore := broadcast.Stats("slow")
	publish("version", 1, 20)
	wait("version 1 delivery", func() bool {
		return slices.Equal(snapshot("slow").Versioned, []int{20}) && broadcast.Stats("slow").Delivered == slowBefore.Delivered+1
	})
	if got := snapshot("fast").Versioned; len(got) != 0 || broadcast.Stats("fast") != fastBefore || broadcast.Stats("slow").Dropped != slowBefore.Dropped {
		t.Fatalf("version 1 reached version 2 subscription or changed drops: %v", got)
	}
	slowBefore = broadcast.Stats("slow")
	publish("version", 2, 21)
	wait("version 2 delivery", func() bool {
		return slices.Equal(snapshot("fast").Versioned, []int{21}) && broadcast.Stats("fast").Delivered == fastBefore.Delivered+1
	})
	if got := snapshot("slow"); !slices.Equal(got.Versioned, []int{20}) || !slices.Equal(got.Seen, want[:4]) || broadcast.Stats("slow") != slowBefore || broadcast.Stats("fast").Dropped != fastBefore.Dropped {
		t.Fatalf("version 2 reached version 1 subscription, changed drops or replayed overflow: %+v", got)
	}
	// 4. 停止前撤销两个装配入口；离线发布由其他插件确认，新 Session 用新事件作顺序屏障。
	routes.Detach("slow")
	broadcast.Detach("slow")
	stopped := running["slow"].Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
	if stopped.Err != nil || stopped.Killed {
		t.Fatalf("stop slow: %+v", stopped)
	}
	delete(running, "slow")
	slowBefore = broadcast.Stats("slow")
	publish("x", 1, 11)
	wait("offline publication reaches fast", func() bool { return slices.Equal(snapshot("fast").Seen, append(slices.Clone(want), 11)) })
	if got := broadcast.Stats("slow"); got != slowBefore {
		t.Fatalf("offline subscriber retained publication: %+v, before %+v", got, slowBefore)
	}
	spec := specs["slow"]
	spec.NewCandidate = false
	spec.Log = output.Plugin(spec.PluginID, spec.GenerationID)
	caller := router.NewCaller(spec.PluginID, spec.GenerationID)
	spec.Handler = mux.For(caller)
	restarted, err := Launch(ctx, store, spec)
	if err != nil {
		t.Fatalf("restart slow: %v", err)
	}
	running["slow"] = restarted
	plugin := router.FromInstance(restarted.Instance)
	routes.Attach(plugin, caller)
	broadcast.Attach(plugin)
	publish("x", 1, 12)
	wait("restarted subscriber receives fresh marker", func() bool {
		return len(snapshot("slow").Seen) > 0 && broadcast.Stats("slow").Delivered == slowBefore.Delivered+1
	})
	if got := snapshot("slow"); !slices.Equal(got.Seen, []int{12}) || len(got.Versioned) != 0 || broadcast.Stats("slow").Dropped != slowBefore.Dropped {
		t.Fatalf("restarted subscriber replayed old events: %+v; stats %+v", got, broadcast.Stats("slow"))
	}
}
