//go:build unix

package app

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/metrics"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"google.golang.org/protobuf/proto"
)

// TestSleepMetrics 验证真实休眠、百请求共享恢复和常驻恢复的计数、阶段日志与持久数据。
func TestSleepMetrics(t *testing.T) {
	const id, requests = "wakeprobe", 100
	readyBlocked, readyGate, deliveryGate := make(chan struct{}), make(chan struct{}), make(chan struct{})
	armed := make(chan struct{}, 1)
	var readyOpen, deliveryOpen sync.Once
	var migrations, starts, readies, stops atomic.Int32
	var mu sync.Mutex
	deliveries := make(map[string]int)
	audit := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var gate <-chan struct{}
		switch strings.TrimPrefix(r.URL.Path, "/") {
		case "migration":
			migrations.Add(1)
		case "start":
			starts.Add(1)
		case "ready":
			if readies.Add(1) == 2 {
				close(readyBlocked)
				gate = readyGate
			}
		case "stop":
			stops.Add(1)
		case "armed":
			armed <- struct{}{}
		case "http":
			mu.Lock()
			deliveries[r.URL.Query().Get("index")]++
			mu.Unlock()
			gate = deliveryGate
		default:
			http.Error(w, "unexpected audit event", http.StatusBadRequest)
			return
		}
		if gate != nil {
			select {
			case <-gate:
			case <-r.Context().Done():
			}
		}
	}))
	t.Cleanup(audit.Close)
	// 夹具明确使用 150ms 空闲、50ms 最短运行及一小时内最多 100 次冷启动。
	h, core := startIdleProbe(t, sleepMetricsPluginSource, audit.URL)
	t.Cleanup(func() {
		readyOpen.Do(func() { close(readyGate) })
		deliveryOpen.Do(func() { close(deliveryGate) })
		h.stop()
	})
	h.expect("PUT", "/api/v1/logs/levels", map[string]any{"global": "debug"}, true, http.StatusOK, "")
	// 先保持常驻，完成调度状态与比较基线，再经 API 开始自动休眠。
	h.expect("PUT", "/api/v1/plugins/"+id+"/sleep-policy", map[string]string{"policy": "never"}, true, http.StatusOK, "")
	since := time.Now()
	h.action("enable", id, http.StatusOK)
	original, ok := core.assembly.router.Caller(id)
	if !ok {
		t.Fatal("enabled plugin has no Caller")
	}
	firstSession := core.assembly.waker.Activity(id).Session
	injectIdleCommand(t, core, "/arm")
	select {
	case <-armed:
	case <-time.After(pipelineTimeout):
		t.Fatal("dynamic task was not armed")
	}
	if err := core.assembly.scheduler.SetJobPaused(t.Context(), id, "paused", true); err != nil {
		t.Fatal(err)
	}
	snapshot, err := core.store.ActiveRouteSnapshot(t.Context(), id)
	if err != nil || len(snapshot.Routes) == 0 {
		t.Fatalf("initial routes=%+v error=%v", snapshot, err)
	}
	jobs, err := core.store.Jobs(t.Context(), id)
	if err != nil || len(jobs) != 2 {
		t.Fatalf("initial jobs=%+v error=%v", jobs, err)
	}
	for _, job := range jobs {
		if job.Anchor.IsZero() || job.Paused != (job.JobID == "paused") {
			t.Fatalf("initial static state=%+v", job)
		}
	}
	tasks, err := core.store.Tasks(t.Context(), id, "remind")
	if err != nil || len(tasks) != 1 || tasks[0].Key != "once" || tasks[0].PayloadJSON != `{"Marker":"keep-args"}` || tasks[0].At.IsZero() {
		t.Fatalf("initial tasks=%+v error=%v", tasks, err)
	}
	db, err := sql.Open("sqlite", filepath.Join(h.dir, "plugins", id, "generations", original.GenerationID, "db.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	assertPreserved := func() {
		t.Helper()
		generation, err := core.store.Active(t.Context(), id)
		if err != nil || generation != original.GenerationID {
			t.Fatalf("active generation=%s error=%v", generation, err)
		}
		config, err := core.store.GenerationConfig(t.Context(), generation)
		if err != nil || config != `{"marker":"keep-original"}` {
			t.Fatalf("generation config=%s error=%v", config, err)
		}
		current, err := core.store.ActiveRouteSnapshot(t.Context(), id)
		if err != nil || current.PluginID != snapshot.PluginID || current.GenerationID != snapshot.GenerationID || !reflect.DeepEqual(current.RouteHash, snapshot.RouteHash) || !reflect.DeepEqual(current.BuildDigest, snapshot.BuildDigest) || len(current.Routes) != len(snapshot.Routes) {
			t.Fatalf("route snapshot changed: %+v error=%v", current, err)
		}
		for i, route := range current.Routes {
			if !proto.Equal(route, snapshot.Routes[i]) {
				t.Fatalf("route %d changed: %v", i, route)
			}
		}
		currentJobs, err := core.store.Jobs(t.Context(), id)
		if err != nil || !reflect.DeepEqual(currentJobs, jobs) {
			t.Fatalf("static state changed: %+v error=%v", currentJobs, err)
		}
		currentTasks, err := core.store.Tasks(t.Context(), id, "remind")
		if err != nil || !reflect.DeepEqual(currentTasks, tasks) {
			t.Fatalf("dynamic state changed: %+v error=%v", currentTasks, err)
		}
		var marker string
		if err := db.QueryRowContext(t.Context(), "SELECT value FROM wake_marker").Scan(&marker); err != nil || marker != "original" || migrations.Load() != 1 {
			t.Fatalf("business DB=%q migrations=%d error=%v", marker, migrations.Load(), err)
		}
	}
	assertCounts := func(cold, wakes uint64) {
		t.Helper()
		want := map[metrics.Reason]uint64{"cold_start": cold}
		if wakes != 0 {
			want["started"], want["ready"] = wakes, wakes
		}
		waitTaskState(t, func() bool {
			return reflect.DeepEqual(core.assembly.metrics.Snapshot()[id]["wake"], want)
		})
		logged := make(map[metrics.Reason]uint64)
		for _, record := range core.output.Recent("") {
			if !record.Time.Before(since) && strings.HasPrefix(record.Message, "Wake activity for component ") && record.Attrs["plugin_id"] == id {
				if record.Level != slog.LevelDebug {
					t.Fatalf("wake log level=%v", record)
				}
				logged[metrics.Reason(record.Attrs["reason"].(string))]++
			}
		}
		if !reflect.DeepEqual(logged, want) {
			t.Fatalf("wake logs=%v counters=%v", logged, want)
		}
	}

	// 1. HTTP 路由以显式时长允许自动休眠，管理读取保持原计数。
	h.expect("PUT", "/api/v1/plugins/"+id+"/sleep-policy", map[string]string{"policy": "150ms"}, true, http.StatusOK, "")
	waitIdleRuntime(t, core, id, state.RuntimeSuspended)
	h.expect("GET", "/api/v1/plugins/"+id+"/runtime", nil, true, http.StatusOK, `"state":"SUSPENDED"`)
	assertCounts(1, 0)
	assertPreserved()
	select {
	case <-firstSession.Done():
	default:
		t.Fatal("suspended Session is still open")
	}

	// 2. 百个请求先全部写出；OnReady 放行前保持零交付，放行后每个序号仅交付一次。
	sent, results := make(chan error, requests), make(chan error, requests)
	for i := range requests {
		go func() {
			index := strconv.Itoa(i)
			req, err := http.NewRequestWithContext(t.Context(), "GET", "http://"+core.http.Addr+"/plugins/"+id+"/probe?index="+index, nil)
			if err != nil {
				sent <- err
				results <- err
				return
			}
			req.Header.Set("Authorization", "Bearer "+h.token)
			trace := &httptrace.ClientTrace{WroteRequest: func(info httptrace.WroteRequestInfo) { sent <- info.Err }}
			req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
			response, err := h.client.Do(req)
			if err == nil {
				body, readErr := io.ReadAll(response.Body)
				closeErr := response.Body.Close()
				if readErr != nil || closeErr != nil || response.StatusCode != http.StatusOK || string(body) != index {
					err = fmt.Errorf("request %s: status=%d body=%q read=%v close=%v", index, response.StatusCode, body, readErr, closeErr)
				}
			}
			results <- err
		}()
	}
	for range requests {
		select {
		case err := <-sent:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(pipelineTimeout):
			t.Fatal("concurrent requests were not written")
		}
	}
	select {
	case <-readyBlocked:
	case <-time.After(pipelineTimeout):
		t.Fatal("shared wake did not reach OnReady")
	}
	mu.Lock()
	delivered := len(deliveries)
	mu.Unlock()
	if got := core.assembly.metrics.Snapshot()[id]["wake"]; got["started"] != 1 || got["cold_start"] != 2 || got["ready"] != 0 || delivered != 0 {
		t.Fatalf("before Ready: counters=%v deliveries=%d", got, delivered)
	}
	readyOpen.Do(func() { close(readyGate) })
	waitTaskState(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(deliveries) == requests
	})
	current, ok := core.assembly.router.Caller(id)
	if !ok || current == original || current.GenerationID != original.GenerationID || core.assembly.waker.Activity(id).Session == firstSession {
		t.Fatalf("shared wake Caller=%+v original=%+v", current, original)
	}
	secondSession := core.assembly.waker.Activity(id).Session
	assertCounts(2, 1)
	deliveryOpen.Do(func() { close(deliveryGate) })
	for range requests {
		select {
		case err := <-results:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(pipelineTimeout):
			t.Fatal("concurrent request did not complete")
		}
	}
	waitIdleRuntime(t, core, id, state.RuntimeSuspended)
	assertPreserved()
	assertCounts(2, 1)

	// 3. 认证管理 API 切回常驻，恢复原代次且持续保持 ACTIVE。
	h.expect("PUT", "/api/v1/core/sleep", map[string]string{"mode": "always_on", "idle_timeout": "150ms"}, true, http.StatusOK, "")
	waitIdleRuntime(t, core, id, state.RuntimeActive)
	resident, ok := core.assembly.router.Caller(id)
	if !ok || resident == current || resident.GenerationID != original.GenerationID || core.assembly.waker.Activity(id).Session == secondSession {
		t.Fatalf("always_on Caller=%+v previous=%+v", resident, current)
	}
	time.Sleep(2*idleCheckInterval + 150*time.Millisecond)
	h.expect("GET", "/api/v1/plugins/"+id+"/runtime", nil, true, http.StatusOK, `"state":"ACTIVE"`)
	h.expect("GET", "/api/v1/core/sleep", nil, true, http.StatusOK, `"mode":"always_on"`)
	assertPreserved()
	assertCounts(3, 2)
	mu.Lock()
	for i := range requests {
		if deliveries[strconv.Itoa(i)] != 1 {
			t.Errorf("request %d deliveries=%d", i, deliveries[strconv.Itoa(i)])
		}
	}
	mu.Unlock()
	if starts.Load() != 3 || readies.Load() != 3 || stops.Load() != 2 {
		t.Fatalf("lifecycle starts=%d readies=%d stops=%d", starts.Load(), readies.Load(), stops.Load())
	}

	// 4. 三个真实冷启动分别关联同一 Session 的两个实际回调及 Core 可见阶段。
	callbacks := make(map[string]map[string]int)
	for _, record := range core.output.Recent(id) {
		if !strings.HasPrefix(record.Message, "Lifecycle callback ") {
			continue
		}
		session, _ := record.Attrs["lifecycle_session_id"].(string)
		phase, _ := record.Attrs["phase"].(string)
		duration, err := strconv.ParseInt(fmt.Sprint(record.Attrs["duration_ns"]), 10, 64)
		if record.Level != slog.LevelDebug || record.Attrs["generation_id"] != original.GenerationID || record.Attrs["executed"] != true || record.Attrs["outcome"] != "success" || duration <= 0 || err != nil {
			t.Fatalf("callback metric=%+v parse=%v", record, err)
		}
		if phase == "OnReady" && record.Attrs["session_id"] != session {
			t.Fatalf("OnReady Session association=%+v", record)
		}
		if callbacks[session] == nil {
			callbacks[session] = make(map[string]int)
		}
		callbacks[session][phase]++
	}
	sessions := make(map[string]bool)
	for _, record := range core.output.Recent("") {
		if !strings.HasPrefix(record.Message, "Component "+id+" completed its cold start in ") || record.Attrs["plugin_id"] != id {
			continue
		}
		session, _ := record.Attrs["session_id"].(string)
		decoded, decodeErr := hex.DecodeString(session)
		wantReason := "resume"
		if len(sessions) == 0 {
			wantReason = "first_install"
		}
		if record.Level != slog.LevelDebug || record.Attrs["generation_id"] != original.GenerationID || record.Attrs["activation_reason"] != wantReason || decodeErr != nil || len(decoded) != 16 || sessions[session] || !reflect.DeepEqual(callbacks[session], map[string]int{"OnStart": 1, "OnReady": 1}) {
			t.Fatalf("cold-start association=%+v callbacks=%v", record, callbacks)
		}
		for _, key := range []string{"pid", "duration_ns", "prepare_ns", "activate_ns"} {
			value, err := strconv.ParseInt(fmt.Sprint(record.Attrs[key]), 10, 64)
			if err != nil || value <= 0 {
				t.Fatalf("cold-start %s=%v error=%v", key, record.Attrs[key], err)
			}
		}
		sessions[session] = true
	}
	if len(sessions) != 3 || len(callbacks) != 3 {
		t.Fatalf("cold-start Sessions=%v callbacks=%v", sessions, callbacks)
	}
}

const sleepMetricsPluginSource = `package main
import (
 "context"
 "database/sql"
 "fmt"
 "io"
 "net/http"
 "time"
 "github.com/AlonaBot/AlonaBot/sdk"
 "github.com/AlonaBot/AlonaBot/sdk/sqlitedb"
 "github.com/AlonaBot/AlonaBot/sdk/task"
)
func observe(event string) error {
 response,err:=http.Get("AUDIT_URL/"+event);if err!=nil{return err}
 _,readErr:=io.Copy(io.Discard,response.Body);closeErr:=response.Body.Close()
 if readErr!=nil{return readErr};if closeErr!=nil{return closeErr}
 if response.StatusCode!=http.StatusOK{return fmt.Errorf("audit status %d",response.StatusCode)}
 return nil
}
type ReminderArgs struct { Marker string }
func New() *sdk.Plugin {
 p:=sdk.NewPlugin()
 sqlitedb.Migrate(p,func(ctx context.Context,db *sql.DB)error{
  if _,err:=db.ExecContext(ctx,"CREATE TABLE wake_marker (value TEXT); INSERT INTO wake_marker VALUES ('original')");err!=nil{return err}
  return observe("migration")
 })
 p.OnStart(func(context.Context)error{return observe("start")})
 p.OnReady(func(ctx context.Context)error{
  db,err:=sqlitedb.Open(ctx);if err!=nil{return err};var marker string
  if err:=db.QueryRowContext(ctx,"SELECT value FROM wake_marker").Scan(&marker);err!=nil{return err}
  if marker!="original"{return fmt.Errorf("business marker changed: %q",marker)}
  return observe("ready")
 })
 p.OnStop(func(context.Context)error{return observe("stop")})
 p.Every("pulse",time.Hour).Handle(func(*sdk.Context)error{return fmt.Errorf("early static delivery")})
 p.Every("paused",2*time.Hour).Handle(func(*sdk.Context)error{return fmt.Errorf("paused job delivered")})
 reminder:=task.Define[ReminderArgs]("remind")
 task.On(p,reminder).Handle(func(*sdk.Context,ReminderArgs)error{return fmt.Errorf("early dynamic delivery")})
 p.Command("arm").Handle(func(c *sdk.Context)error{
  if _,err:=reminder.At(c,time.Now().Add(time.Hour),ReminderArgs{Marker:"keep-args"},task.Key("once"));err!=nil{return err}
  return observe("armed")
 })
 p.Route(http.MethodGet,"/probe").HandleFunc(func(w http.ResponseWriter,r *http.Request){
  index:=r.URL.Query().Get("index")
  if err:=observe("http?index="+index);err!=nil{http.Error(w,err.Error(),http.StatusInternalServerError);return}
  io.WriteString(w,index)
 })
 return p
}
func main(){if err:=sdk.Run(New());err!=nil{panic(err)}}
`
