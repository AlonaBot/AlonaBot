//go:build unix

package app

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	coreadapter "github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/core/process"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/generate"
	genresources "github.com/AlonaBot/AlonaBot/internal/generated"
	"github.com/AlonaBot/AlonaBot/internal/pluginkit"
	"google.golang.org/protobuf/proto"
)

// TestIdle 验证自动空闲退出、首命令恢复和设置 never 的即时唤醒，保留原代次和业务数据。
func TestIdle(t *testing.T) {
	const id = "wakeprobe"
	events := make(chan string, 16)
	stopGate, commandGate := make(chan struct{}, 4), make(chan struct{}, 4)
	var commands atomic.Int32
	audit := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		event := strings.TrimPrefix(r.URL.Path, "/")
		if event == "command" {
			commands.Add(1)
		}
		events <- event
		gate := stopGate
		if event == "command" {
			gate = commandGate
		}
		select {
		case <-gate:
		case <-r.Context().Done():
		}
	}))
	source := strings.Replace(wakePluginSource, " return p", ` p.OnStop(func(ctx context.Context)error{
  db,err:=sqlitedb.Open(ctx);if err!=nil{return err}
  var value string
  if err:=db.QueryRowContext(ctx,"SELECT value FROM wake_marker").Scan(&value);err!=nil{return err}
  if value!="original"{return fmt.Errorf("OnStop database value changed: %q",value)}
  return observe("stop")
 })
 return p`, 1)
	h, core := startIdleProbe(t, source, audit.URL)
	t.Cleanup(func() { close(stopGate); close(commandGate); h.stop(); audit.Close() })
	h.action("enable", id, http.StatusOK)
	original, ok := core.assembly.router.Caller(id)
	if !ok {
		t.Fatal("enabled plugin has no Caller")
	}
	var first *runningComponent
	if err := core.operation(t.Context(), func(context.Context) error {
		for _, r := range core.components {
			if r != nil && r.id == id {
				first = r
				return nil
			}
		}
		return errors.New("plugin missing")
	}); err != nil {
		t.Fatal(err)
	}
	initial, err := core.store.RuntimeStatus(t.Context(), id)
	if err != nil || initial.State != state.RuntimeActive {
		t.Fatalf("initial=%+v error=%v", initial, err)
	}
	dbPath := filepath.Join(h.dir, "plugins", id, "generations", original.GenerationID, "db.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	// 1. OnStop 的同步点暴露真实 QUIESCING，返回后检查整棵旧树和 runtime.lock。
	awaitIdleEvent(t, events, "stop")
	status, err := core.store.RuntimeStatus(t.Context(), id)
	if err != nil || status.State != state.RuntimeQuiescing || !status.LastWakeAt.Equal(initial.LastWakeAt) {
		t.Fatalf("quiescing=%+v error=%v", status, err)
	}
	if core.assembly.waker.Activity(id).Ready {
		t.Fatal("quiescing still accepts new work")
	}
	stopGate <- struct{}{}
	waitIdleRuntime(t, core, id, state.RuntimeSuspended)
	status, err = core.store.RuntimeStatus(t.Context(), id)
	if err != nil || status.Reason != pb.StopReason_STOP_REASON_IDLE.String() || status.LastSuspendAt.IsZero() {
		t.Fatalf("suspended=%+v error=%v", status, err)
	}
	if empty, err := first.process.Process.WaitTree(time.Now()); err != nil || !empty {
		t.Fatalf("old tree: empty=%v error=%v", empty, err)
	}
	if err := syscall.Kill(first.process.Process.Pid(), 0); !errors.Is(err, syscall.ESRCH) {
		t.Fatalf("old process still exists: %v", err)
	}
	lock, err := process.LockRuntime(filepath.Join(h.dir, "plugins", id))
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.Unlock(); err != nil {
		t.Fatal(err)
	}

	// 2. 第一条命令经真实回环 Adapter 进入 Waker，只交付一次并使用新的 Session。
	injectIdleCommand(t, core, "/ping")
	awaitIdleEvent(t, events, "command")
	current, ok := core.assembly.router.Caller(id)
	if !ok || current == original || current.GenerationID != original.GenerationID {
		t.Fatalf("wake Caller: original=%+v current=%+v", original, current)
	}
	if a := core.assembly.waker.Activity(id); !a.Ready || a.References == 0 {
		t.Fatalf("command activity=%+v", a)
	}
	commandGate <- struct{}{}
	waitTaskState(t, func() bool {
		return core.assembly.waker.Activity(id).References == 0 && core.assembly.router.Stats().Running == 0
	})
	awaitIdleEvent(t, events, "stop")
	stopGate <- struct{}{}
	waitIdleRuntime(t, core, id, state.RuntimeSuspended)
	if commands.Load() != 1 {
		t.Fatalf("command deliveries=%d", commands.Load())
	}
	generation, err := core.store.Active(t.Context(), id)
	if err != nil || generation != original.GenerationID {
		t.Fatalf("generation=%q error=%v", generation, err)
	}
	config, err := core.store.GenerationConfig(t.Context(), generation)
	if err != nil || config != `{"marker":"keep-original"}` {
		t.Fatalf("config=%q error=%v", config, err)
	}
	var marker string
	if err := db.QueryRowContext(t.Context(), "SELECT value FROM wake_marker").Scan(&marker); err != nil || marker != "original" {
		t.Fatalf("business DB=%q error=%v", marker, err)
	}
	enabled, err := core.store.Enabled(t.Context(), id)
	if err != nil || !enabled {
		t.Fatalf("enabled=%v error=%v", enabled, err)
	}
	snapshot, err := core.store.ActiveRouteSnapshot(t.Context(), id)
	if err != nil || snapshot.GenerationID != original.GenerationID {
		t.Fatalf("route snapshot=%+v error=%v", snapshot, err)
	}

	// 3. 在运行循环保存策略，在循环外应用，唤醒沿用原代次且后续检查保持 ACTIVE。
	never := "never"
	if err := core.operation(t.Context(), func(context.Context) error {
		return core.assembly.settings.Update(func(d *settings.Document) error { return d.SetSleepPolicy(id, &never) })
	}); err != nil {
		t.Fatal(err)
	}
	if err := core.applySleep(t.Context()); err != nil {
		t.Fatal(err)
	}
	waitIdleRuntime(t, core, id, state.RuntimeActive)
	resident, ok := core.assembly.router.Caller(id)
	if !ok || resident == current || resident.GenerationID != original.GenerationID {
		t.Fatalf("resident Caller=%+v", resident)
	}
	select {
	case unexpected := <-events:
		t.Fatalf("never policy delivered unexpected event %q", unexpected)
	case <-time.After(2*idleCheckInterval + 50*time.Millisecond):
	}
	status, err = core.store.RuntimeStatus(t.Context(), id)
	if err != nil || status.State != state.RuntimeActive || core.assembly.settings.Settings().Sleep.Plugins[id] != "never" {
		t.Fatalf("never status=%+v error=%v", status, err)
	}
	if commands.Load() != 1 {
		t.Fatalf("duplicate command after restore: %d", commands.Load())
	}
}

func startIdleProbe(t *testing.T, source, auditURL string, args ...string) (*webUIHarness, *Core) {
	t.Helper()
	project, out := t.TempDir(), t.TempDir()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	module, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	module = bytes.Replace(module, []byte("module github.com/AlonaBot/AlonaBot"), []byte("module example.invalid/idleprobe"), 1)
	module = append(module, []byte("\nrequire github.com/AlonaBot/AlonaBot v0.0.0\nreplace github.com/AlonaBot/AlonaBot => "+root+"\n")...)
	sum, err := os.ReadFile(filepath.Join(root, "go.sum"))
	if err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string][]byte{
		"go.mod": module, "go.sum": sum,
		"plugin.yaml":        []byte("id: wakeprobe\nname: Idle probe\nversion: 0.1.0\nauthor: AlonaBot\nsource: github.com/AlonaBot/wakeprobe\nrole: plugin\nsleep_capable: true\n"),
		"config_schema.json": []byte(`{"marker":{"type":"string","default":"keep-original"}}`),
		"main.go":            []byte(strings.ReplaceAll(source, "AUDIT_URL", auditURL)),
	} {
		if err := os.WriteFile(filepath.Join(project, name), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	archive, err := pluginkit.Pack(t.Context(), generate.Options{Dir: project}, "https://example.invalid/idleprobe.tar.gz", "", out)
	if err != nil {
		t.Fatal(err)
	}
	releaseJSON, err := os.ReadFile(filepath.Join(out, "release.json"))
	if err != nil {
		t.Fatal(err)
	}
	h := newWebUIHarness(t)
	h.start(nil)
	manager := manage.New(manage.Options{Store: h.core.store, Settings: h.core.assembly.settings, Packages: filepath.Join(h.dir, "packages"), Limits: packages.NewLimits(h.core.values), Adapters: h.core.assembly.adapters, State: h.core.assembly.componentState})
	if _, err := manager.Install(t.Context(), manage.InstallRequest{Archive: archive, ReleaseJSON: string(releaseJSON), Confirm: true}); err != nil {
		t.Fatal(err)
	}
	h.stop()
	settingsJSON := `{"listen":"127.0.0.1:0","base_url":"http://127.0.0.1","sleep":{"mode":"component_idle","idle_timeout":"150ms","plugins":{"wakeprobe":"150ms"}},"adapters":{"loop":{"instances":{"pipeline-instance":{}}}}}`
	if err := os.WriteFile(filepath.Join(h.dir, settings.FileName), []byte(settingsJSON), 0600); err != nil {
		t.Fatal(err)
	}
	loop := &pipelineAdapter{}
	genresources.Bind(loop, genresources.Resources{Manifest: []byte("id: idle-loop\nname: Idle loop\nversion: 0.1.0\nsource: test/idle-loop\nrole: adapter\nadapter_id: loop\n"), AdapterConfigSchema: []byte(`{}`)})
	ctx, cancel := context.WithCancel(t.Context())
	args = append([]string{"SLEEP_MIN_UPTIME=50ms,SLEEP_COLD_START_WINDOW=1h,SLEEP_COLD_START_MAX=100"}, args...)
	core, err := Start(ctx, Options{DataDir: h.dir, Components: []Component{AdapterComponent(loop)}, Args: args})
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	h.core, h.cancel = core, cancel
	return h, core
}

func injectIdleCommand(t *testing.T, core *Core, text string) {
	t.Helper()
	instance, ok := core.assembly.adapters.Get(pipelineInstanceID)
	if !ok {
		t.Fatal("loop Adapter instance missing")
	}
	data, err := json.Marshal(pipelineInput{Text: text, User: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	_, failure := instance.(coreadapter.ActionExecutor).Action(t.Context(), &pb.Target{AdapterId: "loop", AdapterInstanceId: pipelineInstanceID, AccountId: pipelineAccountID, ChatId: "group", ThreadId: proto.String("topic")}, "loop.inject", 1, string(data))
	if failure != nil {
		t.Fatal(failure.String())
	}
}

func awaitIdleEvent(t *testing.T, events <-chan string, want string) {
	t.Helper()
	select {
	case got := <-events:
		if got != want {
			t.Fatalf("event=%q want=%q", got, want)
		}
	case <-time.After(pipelineTimeout):
		t.Fatalf("waiting for event %q timed out", want)
	}
}

func waitIdleRuntime(t *testing.T, core *Core, id string, want state.RuntimeState) {
	t.Helper()
	waitTaskState(t, func() bool {
		status, err := core.store.RuntimeStatus(t.Context(), id)
		if err != nil {
			t.Fatal(err)
		}
		return status.State == want
	})
}
