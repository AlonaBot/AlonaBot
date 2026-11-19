//go:build unix

package app

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// TestSchedulerWake 验证正常空闲退出后静态与动态到期唤醒，并保留原代次、数据和调度状态。
func TestSchedulerWake(t *testing.T) {
	const id = "wakeprobe"
	type observation struct {
		event       string
		scheduledAt int64
	}
	events := make(chan observation, 16)
	finish := make(chan struct{}, 2)
	audit := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		event := strings.TrimPrefix(r.URL.Path, "/")
		var at int64
		if event == "static" || event == "dynamic" {
			var err error
			at, err = strconv.ParseInt(r.URL.Query().Get("at"), 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		events <- observation{event, at}
		if event == "static" || event == "dynamic" {
			select {
			case <-finish:
			case <-r.Context().Done():
			}
		}
	}))
	source := strings.Replace(wakePluginSource, ` "io"`, ` "io"
 "time"
 "github.com/AlonaBot/AlonaBot/sdk/task"
 "github.com/AlonaBot/AlonaBot/sdk/message"`, 1)
	source = strings.Replace(source, "func New()", `type ReminderArgs struct { Marker string }
func scheduled(c *sdk.Context, kind string)error {
 event:=c.Event()
 if event.Actor!=nil || c.User()!="" || c.Session()!="" || event.Source!=message.SourceScheduler {return fmt.Errorf("scheduled identity changed")}
 payload,ok:=event.Payload.(*message.ScheduledPayload)
 if !ok || payload.RunID=="" {return fmt.Errorf("missing scheduled payload")}
 return observe(fmt.Sprintf("%s?at=%d",kind,payload.ScheduledAt.UnixMilli()))
}
func New()`, 1)
	source = strings.Replace(source, " return p", ` p.Every("pulse",4*time.Second).Handle(func(c *sdk.Context)error{return scheduled(c,"static")})
 p.Every("paused",time.Hour).Handle(func(*sdk.Context)error{return fmt.Errorf("paused job executed")})
 reminder:=task.Define[ReminderArgs]("remind")
 task.On(p,reminder).Handle(func(c *sdk.Context,args ReminderArgs)error{
  if args.Marker!="keep-args"{return fmt.Errorf("task arguments changed: %+v",args)}
  return scheduled(c,"dynamic")
 })
 p.Command("arm").Handle(func(c *sdk.Context)error{
  if _,err:=reminder.At(c,time.Now().Add(3*time.Second),ReminderArgs{Marker:"keep-args"},task.Key("once"));err!=nil{return err}
  return observe("armed")
 })
 p.OnStop(func(context.Context)error{return observe("stop")})
 return p`, 1)
	h, core := startIdleProbe(t, source, audit.URL)
	t.Cleanup(func() { close(finish); h.stop(); audit.Close() })
	await := func(want string) observation {
		t.Helper()
		select {
		case got := <-events:
			if got.event != want {
				t.Fatalf("event=%+v want=%s", got, want)
			}
			return got
		case <-time.After(pipelineTimeout):
			t.Fatalf("waiting for %s timed out", want)
			return observation{}
		}
	}
	h.action("enable", id, http.StatusOK)
	original, ok := core.assembly.router.Caller(id)
	if !ok {
		t.Fatal("enabled plugin has no Caller")
	}
	if err := core.assembly.scheduler.SetJobPaused(t.Context(), id, "paused", true); err != nil {
		t.Fatal(err)
	}
	jobs, err := core.store.Jobs(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	var anchor time.Time
	for _, job := range jobs {
		if job.JobID == "pulse" {
			anchor = job.Anchor
		}
	}
	if anchor.IsZero() {
		t.Fatal("missing static anchor")
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

	// 1. 真实定时器从首次正常休眠唤醒；Handler 同步点让活动引用可观察。
	await("stop")
	waitIdleRuntime(t, core, id, state.RuntimeSuspended)
	first := await("static")
	if first.scheduledAt != anchor.Add(4*time.Second).UnixMilli() {
		t.Fatalf("scheduled_at=%d anchor=%v", first.scheduledAt, anchor)
	}
	current, ok := core.assembly.router.Caller(id)
	if !ok || current == original || current.GenerationID != original.GenerationID {
		t.Fatalf("wake Caller: original=%+v current=%+v", original, current)
	}
	assertActive := func() {
		t.Helper()
		if a := core.assembly.waker.Activity(id); !a.Ready || a.References == 0 {
			t.Fatalf("activity=%+v", a)
		}
		select {
		case unexpected := <-events:
			t.Fatalf("running task received %+v", unexpected)
		case <-time.After(2*idleCheckInterval + 50*time.Millisecond):
		}
		status, err := core.store.RuntimeStatus(t.Context(), id)
		if err != nil || status.State != state.RuntimeActive {
			t.Fatalf("running status=%+v error=%v", status, err)
		}
	}
	assertActive()
	if err := core.assembly.scheduler.SetJobPaused(t.Context(), id, "pulse", true); err != nil {
		t.Fatal(err)
	}

	// 2. 通过插件任务 API 创建 at，执行结束后先休眠，再自然唤醒消费一次。
	injectIdleCommand(t, core, "/arm")
	await("armed")
	tasks, err := core.store.Tasks(t.Context(), id, "remind")
	if err != nil || len(tasks) != 1 {
		t.Fatalf("tasks=%v error=%v", tasks, err)
	}
	at := tasks[0].At
	finish <- struct{}{}
	waitTaskState(t, func() bool { return !core.assembly.scheduler.Stats(id, "pulse").Running })
	await("stop")
	waitIdleRuntime(t, core, id, state.RuntimeSuspended)
	dynamic := await("dynamic")
	if dynamic.scheduledAt != at.UnixMilli() {
		t.Fatalf("dynamic scheduled_at=%d want=%d", dynamic.scheduledAt, at.UnixMilli())
	}
	next, ok := core.assembly.router.Caller(id)
	if !ok || next == current || next.GenerationID != original.GenerationID {
		t.Fatalf("dynamic Caller=%+v previous=%+v", next, current)
	}
	tasks, err = core.store.Tasks(t.Context(), id, "")
	if err != nil || len(tasks) != 0 {
		t.Fatalf("consumed at=%v error=%v", tasks, err)
	}
	assertActive()
	finish <- struct{}{}
	waitTaskState(t, func() bool { return core.assembly.scheduler.TasksStats(id).Running == 0 })
	await("stop")
	waitIdleRuntime(t, core, id, state.RuntimeSuspended)

	// 3. 三次 Session 使用同一代次和业务 DB；静态起点与暂停状态跨两次唤醒保持。
	generation, err := core.store.Active(t.Context(), id)
	if err != nil || generation != original.GenerationID {
		t.Fatalf("generation=%s error=%v", generation, err)
	}
	config, err := core.store.GenerationConfig(t.Context(), generation)
	if err != nil || config != `{"marker":"keep-original"}` {
		t.Fatalf("config=%s error=%v", config, err)
	}
	var marker string
	if err := db.QueryRowContext(t.Context(), "SELECT value FROM wake_marker").Scan(&marker); err != nil || marker != "original" {
		t.Fatalf("business DB=%s error=%v", marker, err)
	}
	restored, err := core.store.Jobs(t.Context(), id)
	if err != nil || len(restored) != 2 {
		t.Fatalf("jobs=%v error=%v", restored, err)
	}
	for _, job := range restored {
		if !job.Paused || job.JobID == "pulse" && !job.Anchor.Equal(anchor) {
			t.Fatalf("restored job=%+v anchor=%v", job, anchor)
		}
	}
	if stats := core.assembly.scheduler.Stats(id, "pulse"); stats.Runs != 1 || stats.Running {
		t.Fatal(stats)
	}
	select {
	case unexpected := <-events:
		t.Fatalf("duplicate delivery=%+v", unexpected)
	default:
	}
	// 最后一次正常休眠仍保留计划供下一轮使用。
	if jobs := core.assembly.scheduler.Jobs(id); len(jobs) != 2 {
		t.Fatalf("suspended plans=%v", jobs)
	}
}
