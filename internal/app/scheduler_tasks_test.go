//go:build unix

package app

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/httpproxy"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// persistentTaskProcess 在现有测试二进制内实现原始协议插件，避免依赖尚未交付的 SDK Task API。
func persistentTaskProcess(mode string) error {
	socket, err := net.Dial("tcp", os.Getenv(wire.EnvCoreAddr))
	if err != nil {
		return err
	}
	conn := wire.NewStreamConn(socket)
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
	send := func(sid []byte, kind pb.FrameKind, message proto.Message) error {
		env, err := wire.ControlEnvelope(sid, 0, id, kind, message)
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
	ack := &pb.HelloAck{}
	if err := proto.Unmarshal(env.Payload, ack); err != nil {
		return err
	}
	registration := &pb.Register{CandidateConfigJson: "{}", CandidateConfigVersion: ack.GetBootstrap().GetTargetConfigVersion(), ConfigSchemaHash: ack.GetBootstrap().GetConfigSchemaHash(), Migration: &pb.MigrationState{Mode: pb.MigrationMode_MIGRATION_MODE_NONE}, Routes: []*pb.RouteDescriptor{
		{RouteId: "job.pulse", Kind: pb.RouteKind_ROUTE_KIND_SCHEDULE, Spec: &pb.RouteDescriptor_Schedule{Schedule: &pb.ScheduleSpec{JobId: "pulse", Trigger: &pb.ScheduleSpec_IntervalMs{IntervalMs: 1000}}}},
	}}
	if mode != "tasks-missing" {
		registration.Routes = append(registration.Routes, &pb.RouteDescriptor{RouteId: "task.remind", Kind: pb.RouteKind_ROUTE_KIND_TASK, Spec: &pb.RouteDescriptor_Task{Task: &pb.TaskSpec{Name: "remind"}}})
	}
	for _, service := range []struct {
		name              string
		request, response proto.Message
	}{
		{"task.set", &pb.TaskSetRequest{}, &pb.TaskSetResponse{}}, {"task.cancel", &pb.TaskCancelRequest{}, &pb.TaskCancelResponse{}}, {"task.list", &pb.TaskListRequest{}, &pb.TaskListResponse{}},
		{"job.pause", &pb.JobPauseRequest{}, &pb.JobPauseResponse{}}, {"job.resume", &pb.JobResumeRequest{}, &pb.JobResumeResponse{}}, {"job.status", &pb.JobStatusRequest{}, &pb.JobStatusResponse{}},
	} {
		registration.Routes = append(registration.Routes, &pb.RouteDescriptor{RouteId: "svc.probe." + service.name, Kind: pb.RouteKind_ROUTE_KIND_SERVICE, InputSchemaId: validate.SchemaID(service.request), InputSchemaVersion: 1, OutputSchemaId: validate.SchemaID(service.response), OutputSchemaVersion: 1, Spec: &pb.RouteDescriptor_Service{Service: &pb.ServiceSpec{ServiceName: "probe." + service.name, Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, Access: pb.ServiceAccess_SERVICE_ACCESS_PRIVATE}}})
	}
	if err := send(ack.SessionId, pb.FrameKind_FRAME_KIND_REGISTER, registration); err != nil {
		return err
	}
	if env, err = conn.ReadFrame(); err != nil {
		return err
	} else if env.Kind != pb.FrameKind_FRAME_KIND_PREPARED {
		return fmt.Errorf("expected PREPARED: %v", env.Kind)
	}
	if env, err = conn.ReadFrame(); err != nil {
		return err
	} else if env.Kind != pb.FrameKind_FRAME_KIND_ACTIVATE {
		return fmt.Errorf("expected ACTIVATE: %v", env.Kind)
	}
	activate := &pb.Activate{}
	if err := proto.Unmarshal(env.Payload, activate); err != nil {
		return err
	}
	if err := send(ack.SessionId, pb.FrameKind_FRAME_KIND_ACTIVATED, &pb.Activated{GenerationId: activate.GenerationId, Ready: true}); err != nil {
		return err
	}
	stopping := make(chan *pb.Stopping, 1)
	var session *wire.Session
	session = wire.New(conn, wire.Options{Local: pb.InitiatorSide_INITIATOR_SIDE_PLUGIN, SessionID: ack.SessionId, SourcePluginID: id,
		Control: func(env *pb.Envelope) error {
			if env.Kind != pb.FrameKind_FRAME_KIND_STOPPING {
				return fmt.Errorf("unexpected control %v", env.Kind)
			}
			stop := &pb.Stopping{}
			if err := proto.Unmarshal(env.Payload, stop); err != nil {
				return err
			}
			stopping <- stop
			return nil
		},
		Handler: func(ctx context.Context, in *wire.Inbound) wire.Response {
			if strings.HasPrefix(in.Target, "svc.probe.") {
				// 后台身份调用固定服务，刻意不携 parent_call_id。
				result, err := session.Call(ctx, wire.Request{Target: "core." + strings.TrimPrefix(in.Target, "svc.probe."), Payload: in.Payload})
				if err != nil && result.Err == nil {
					return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}}
				}
				return result
			}
			request := &pb.HandlerRequest{}
			if err := proto.Unmarshal(in.Payload.Bytes, request); err != nil {
				return taskProcessError(err)
			}
			payload := request.Event.GetEventPayload().GetScheduled()
			if request.Event.GetActor() != nil || request.Event.GetTarget() != nil || request.Context.GetActor() != nil || request.Context.GetDefaultTarget() != nil || !slices.Equal(request.Context.Scopes, []string{"job"}) {
				return taskProcessError(errors.New("scheduled invocation has invalid identity"))
			}
			stage := "task"
			if payload.TaskName == "" {
				stage = "static"
			}
			record := pipelineRecord{Stage: stage, Route: request.RouteId, Value: payload.TaskKey, Text: payload.TaskPayloadJson, Job: payload.JobId, RunID: payload.RunId}
			if err := pipelineAudit(record); err != nil {
				return taskProcessError(err)
			}
			if strings.HasPrefix(payload.TaskKey, "hold") {
				<-ctx.Done() // Core Cancel 与 STOPPING 的协作取消均要留下实际观察证据。
				record.Stage = "cancelled"
				if err := pipelineAudit(record); err != nil {
					return taskProcessError(err)
				}
			}
			record.Stage = "completed"
			if err := pipelineAudit(record); err != nil {
				return taskProcessError(err)
			}
			return taskProcessResponse(&pb.HandlerOutcome{Matched: true, Propagation: pb.Propagation_PROPAGATION_CONTINUE})
		},
	})
	session.Start()
	select {
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
func taskProcessError(err error) wire.Response {
	return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}}
}
func taskProcessResponse(message proto.Message) wire.Response {
	b, err := proto.Marshal(message)
	if err != nil {
		return taskProcessError(err)
	}
	return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(message), SchemaVersion: 1, Bytes: b}}
}

// taskPipelineAt 复用生产装配和 pipelineHarness 的进程/观察入口，让两个测试阶段重开同一个 data_dir。
func taskPipelineAt(t *testing.T, dir string) *pipelineHarness {
	t.Helper()
	store, err := state.Open(t.Context(), filepath.Join(dir, "data"))
	if err != nil {
		t.Fatal(err)
	}
	h := &pipelineHarness{t: t, dir: dir, store: store, values: tune.Defaults(), changed: make(chan struct{}, 1), release: make(chan struct{}, 1), components: make(map[string]*runningComponent)}
	h.values.StopDrain = pipelineStopDrain
	h.values.TaskConcurrency = 1
	file, err := settings.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	h.assembly, err = assemble(store, dir, h.values, file)
	if err != nil {
		t.Fatal(err)
	}
	h.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var record pipelineRecord
		if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.mu.Lock()
		h.records = append(h.records, record)
		h.mu.Unlock()
		select {
		case h.changed <- struct{}{}:
		default:
		} // 审计只负责唤醒，断言读取完整快照。
		w.WriteHeader(http.StatusNoContent)
	}))
	h.assembly.http = httpproxy.New(nil)
	h.assembly.router.Start()
	t.Cleanup(func() {
		for _, component := range h.components {
			h.stop(component)
		}
		h.assembly.scheduler.Close()
		ctx, cancel := context.WithTimeout(context.Background(), pipelineTimeout)
		defer cancel()
		if err := h.assembly.router.Close(ctx); err != nil {
			t.Error(err)
		}
		h.server.Close()
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	return h
}
func taskCall(t *testing.T, c *runningComponent, name string, in, out proto.Message) {
	t.Helper()
	payload := taskProcessResponse(in).Payload
	ctx, cancel := context.WithTimeout(t.Context(), pipelineTimeout)
	defer cancel()
	r, err := c.instance.Session.Call(ctx, wire.Request{Target: "svc.probe." + name, Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	if r.Payload.SchemaID != validate.SchemaID(out) {
		t.Fatalf("response schema=%s", r.Payload.SchemaID)
	}
	if err := proto.Unmarshal(r.Payload.Bytes, out); err != nil {
		t.Fatal(err)
	}
}
func taskAt(key string, at time.Time) *pb.TaskSetRequest {
	return &pb.TaskSetRequest{TaskName: "remind", Key: key, Trigger: &pb.TaskTrigger{Trigger: &pb.TaskTrigger_AtUnixMs{AtUnixMs: at.UnixMilli()}}, PayloadJson: `{ "phase": 2 }`}
}
func taskEvery(key string) *pb.TaskSetRequest {
	return &pb.TaskSetRequest{TaskName: "remind", Key: key, Trigger: &pb.TaskTrigger{Trigger: &pb.TaskTrigger_IntervalMs{IntervalMs: 1000}}, PayloadJson: `{ "phase": 2 }`}
}
func hasTaskRecord(records []pipelineRecord, stage, key string) bool {
	for _, r := range records {
		if r.Stage == stage && r.Value == key {
			return true
		}
	}
	return false
}
func waitTaskState(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(pipelineTimeout)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatal("task condition timed out")
		}
		time.Sleep(time.Millisecond)
	}
}

// taskArchive 提供同一个真实测试进程的发布包，供管理安装入口核对来源及归档完整性。
func taskArchive(t *testing.T, id string) manage.InstallRequest {
	t.Helper()
	executable, err := os.ReadFile(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	info, err := buildinfo.ReadFile(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	const source = "github.com/AlonaBot/testtasks"
	manifest := fmt.Sprintf("manifest_version: 1\nname: Tasks\nid: %s\nversion: 1.0.0\nprotocol_min_major: 1\nprotocol_max_major: 1\nrequired_capabilities: []\nos: %s\narch: %s\nentrypoint: plugin\nconfig_version: 1\nsource: %s\nmigration_mode: none\nprepare_effects: none\n", id, runtime.GOOS, runtime.GOARCH, source)
	var archive bytes.Buffer
	compressed := gzip.NewWriter(&archive)
	writer := tar.NewWriter(compressed)
	for _, entry := range []struct {
		name string
		body []byte
		mode int64
	}{{"plugin.yaml", []byte(manifest), 0600}, {"plugin", executable, 0700}, {"build-info", []byte(info.String()), 0600}} {
		if err := writer.WriteHeader(&tar.Header{Name: entry.name, Mode: entry.mode, Size: int64(len(entry.body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write(entry.body); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := compressed.Close(); err != nil {
		t.Fatal(err)
	}
	archivePath := filepath.Join(t.TempDir(), "tasks.tar.gz")
	if err := os.WriteFile(archivePath, archive.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	// 安装协议要求发布清单携带整个归档的摘要，供生产预检核对。
	digest := sha256.Sum256(archive.Bytes())
	raw, err := json.Marshal(release.Release{Source: source, PluginID: id, Version: "1.0.0", Assets: []release.Asset{{OS: runtime.GOOS, Arch: runtime.GOARCH, URL: "https://example.com/tasks.tar.gz", Size: int64(archive.Len()), SHA256: hex.EncodeToString(digest[:])}}})
	if err != nil {
		t.Fatal(err)
	}
	return manage.InstallRequest{Archive: archivePath, ReleaseJSON: string(raw), Confirm: true}
}

func taskManager(h *pipelineHarness) *manage.Service {
	return manage.New(manage.Options{Store: h.store, Settings: h.assembly.settings, Policy: h.assembly.policy, Adapters: h.assembly.adapters, Packages: filepath.Join(h.dir, "packages"), Limits: packages.NewLimits(h.values),
		State: func(id string) string {
			if h.components[id] != nil {
				return "Activated"
			}
			return ""
		},
		Stop: func(ctx context.Context, id string, _ pb.StopReason) error {
			if component := h.components[id]; component != nil {
				h.assembly.detach(component)
				result := component.process.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
				delete(h.components, id)
				return result.Err
			}
			return nil
		},
	})
}

// TestPersistentTasks 经真实原始协议进程、Core 装配与 SQLite 验证任务的持久化、取消和卸载清理。
func TestPersistentTasks(t *testing.T) {
	dir := t.TempDir()
	const id = "persistent"
	installation := taskArchive(t, id)
	var anchor, jobAnchor time.Time
	// 1. 创建/替换、完整参数、无 Actor 和已派发 at 的协作取消。
	t.Run("create_replace_cancel", func(t *testing.T) {
		h := taskPipelineAt(t, dir)
		if _, err := taskManager(h).Install(t.Context(), installation); err != nil {
			t.Fatal(err)
		}
		if err := h.store.AddGeneration(t.Context(), id, pipelineGeneration+":"+id, "{}", 1); err != nil {
			t.Fatal(err)
		}
		c := h.launch(id, "tasks", 0)
		first := taskEvery("cycle")
		first.PayloadJson = `{"phase":1}`
		taskCall(t, c, "task.set", first, &pb.TaskSetResponse{})
		taskCall(t, c, "task.set", taskEvery("cycle"), &pb.TaskSetResponse{})
		taskCall(t, c, "task.set", taskAt("once", time.Now().Add(300*time.Millisecond)), &pb.TaskSetResponse{})
		list := &pb.TaskListResponse{}
		taskCall(t, c, "task.list", &pb.TaskListRequest{TaskName: "remind"}, list)
		if len(list.Instances) != 2 || list.Instances[0].Key != "cycle" || list.Instances[0].PayloadJson != `{ "phase": 2 }` || list.Instances[0].Trigger.GetIntervalMs() != 1000 || list.NextAfterKey != nil {
			t.Fatal(list)
		}
		h.await(func(r []pipelineRecord) bool {
			return hasTaskRecord(r, "completed", "once") && hasTaskRecord(r, "completed", "cycle")
		})

		for _, r := range h.snapshot() {
			if r.Stage == "task" && (r.Route != "task.remind" || r.Job != "task.remind" || r.RunID == "" || r.Text != `{ "phase": 2 }`) {
				t.Fatalf("task payload=%+v", r)
			}
		}
		taskCall(t, c, "job.pause", &pb.JobPauseRequest{JobId: "pulse"}, &pb.JobPauseResponse{})
		status := &pb.JobStatusResponse{}
		taskCall(t, c, "job.status", &pb.JobStatusRequest{JobId: "pulse"}, status)
		if !status.Paused || status.NextAtUnixMs != nil {
			t.Fatal(status)
		}
		tasks, err := h.store.Tasks(t.Context(), id, "remind")
		if err != nil || len(tasks) != 1 {
			t.Fatalf("tasks=%v %v", tasks, err)
		}
		anchor = tasks[0].Anchor
		jobs, err := h.store.Jobs(t.Context(), id)
		if err != nil || len(jobs) != 1 {
			t.Fatalf("jobs=%v %v", jobs, err)
		}
		jobAnchor = jobs[0].Anchor
		taskCall(t, c, "task.set", taskAt("expired", time.Now().Add(time.Second)), &pb.TaskSetResponse{})
		taskCall(t, c, "task.set", taskAt("hold.cancel", time.Now().Add(100*time.Millisecond)), &pb.TaskSetResponse{})
		h.await(func(r []pipelineRecord) bool { return hasTaskRecord(r, "task", "hold.cancel") })
		taskCall(t, c, "task.cancel", &pb.TaskCancelRequest{TaskName: "remind", Key: "hold.cancel"}, &pb.TaskCancelResponse{})
		h.await(func(r []pipelineRecord) bool {
			return hasTaskRecord(r, "cancelled", "hold.cancel") && hasTaskRecord(r, "completed", "hold.cancel")
		})
		deadline := time.Now().Add(time.Second)
		for h.assembly.scheduler.TasksStats(id).Running != 0 && time.Now().Before(deadline) {
			time.Sleep(time.Millisecond)
		}
		if got := h.assembly.scheduler.TasksStats(id).Running; got != 0 {
			t.Errorf("actual cancelled Response did not release dynamic slot: running=%d", got)
		}

	})
	// Core 完全停止期间越过一次性时刻与周期 tick；下一阶段重新 Open 同一状态库。
	time.Sleep(2200 * time.Millisecond)
	t.Run("reopen_resume_and_uninstall", func(t *testing.T) {
		h := taskPipelineAt(t, dir)
		c := h.launch(id, "tasks", 0)
		tasks, err := h.store.Tasks(t.Context(), id, "remind")
		if err != nil || len(tasks) != 1 || !tasks[0].Anchor.Equal(anchor) {
			t.Fatalf("reopen tasks=%v %v", tasks, err)
		}
		jobs, err := h.store.Jobs(t.Context(), id)
		if err != nil || len(jobs) != 1 || !jobs[0].Anchor.Equal(jobAnchor) || !jobs[0].Paused {
			t.Fatalf("reopen jobs=%v %v", jobs, err)
		}
		if h.assembly.scheduler.TasksStats(id).Expired != 1 {
			t.Fatal("expired at was not diagnosed")
		}
		status := &pb.JobStatusResponse{}
		taskCall(t, c, "job.status", &pb.JobStatusRequest{JobId: "pulse"}, status)
		if !status.Paused || status.NextAtUnixMs != nil || status.LastAtUnixMs != nil {
			t.Fatal(status)
		}
		taskCall(t, c, "job.resume", &pb.JobResumeRequest{JobId: "pulse"}, &pb.JobResumeResponse{})
		taskCall(t, c, "job.status", &pb.JobStatusRequest{JobId: "pulse"}, status)
		if status.NextAtUnixMs == nil || *status.NextAtUnixMs <= time.Now().UnixMilli() {
			t.Fatal(status)
		}
		h.await(func(r []pipelineRecord) bool {
			return hasTaskRecord(r, "completed", "cycle") && hasTaskRecord(r, "static", "")
		})
		manager := taskManager(h)
		for _, scope := range []manage.PurgeScope{manage.PurgeNone, manage.PurgeData, manage.PurgeAll} {
			c = h.components[id]
			taskCall(t, c, "task.set", taskEvery("cycle"), &pb.TaskSetResponse{})
			before, err := h.store.Tasks(t.Context(), id, "remind")
			if err != nil || len(before) != 1 {
				t.Fatalf("before=%v %v", before, err)
			}
			config := filepath.Join(dir, "plugins", id, "config.json")
			if err := os.MkdirAll(filepath.Dir(config), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(config, []byte(`{}`), 0600); err != nil {
				t.Fatal(err)
			}
			if scope == manage.PurgeNone {
				if err := h.store.AddSecret(t.Context(), state.Secret{PluginID: id, RefID: "task-secret", Value: []byte("retained")}); err != nil {
					t.Fatal(err)
				}
			}
			hold := fmt.Sprintf("hold.uninstall.%d", scope)
			queued := fmt.Sprintf("queued.%d", scope)
			taskCall(t, c, "task.set", taskAt(hold, time.Now().Add(100*time.Millisecond)), &pb.TaskSetResponse{})
			h.await(func(r []pipelineRecord) bool { return hasTaskRecord(r, "task", hold) })
			taskCall(t, c, "task.set", taskAt(queued, time.Now().Add(100*time.Millisecond)), &pb.TaskSetResponse{})
			waitTaskState(t, func() bool { return h.assembly.scheduler.TasksStats(id).Queued > 0 })
			if scope != manage.PurgeNone {
				err := manager.Uninstall(t.Context(), id, scope, false)
				var confirmation *manage.ErrNeedsConfirm
				if !errors.As(err, &confirmation) || !slices.Contains(confirmation.Affected, id+":dynamic_task") || !slices.Contains(confirmation.Affected, id+":scheduled_job") {
					t.Fatalf("confirmation=%v", err)
				}
			}
			if err := manager.Uninstall(t.Context(), id, scope, true); err != nil {
				t.Fatal(err)
			}
			if !hasTaskRecord(h.snapshot(), "cancelled", hold) || hasTaskRecord(h.snapshot(), "task", queued) {
				t.Fatalf("uninstall cancellation=%v", h.snapshot())
			}
			if stats := h.assembly.scheduler.TasksStats(id); stats.Running != 0 || stats.Queued != 0 {
				t.Fatal(stats)
			}
			saved, err := h.store.Tasks(t.Context(), id, "")
			if err != nil {
				t.Fatal(err)
			}
			savedJobs, err := h.store.Jobs(t.Context(), id)
			if err != nil {
				t.Fatal(err)
			}
			if scope == manage.PurgeNone {
				if len(saved) == 0 || len(savedJobs) != 1 {
					t.Fatalf("retained=%v %v", saved, savedJobs)
				}
			} else if len(saved) != 0 || len(savedJobs) != 0 {
				t.Fatalf("purged=%v %v", saved, savedJobs)
			}
			if scope == manage.PurgeData {
				if _, err := os.Stat(config); err != nil {
					t.Fatal(err)
				}
				secrets, err := h.store.ListSecrets(t.Context(), id)
				if err != nil || len(secrets) != 1 {
					t.Fatalf("secrets=%v %v", secrets, err)
				}
			}
			if scope == manage.PurgeAll {
				if _, err := os.Stat(config); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("purged config=%v", err)
				}
				secrets, err := h.store.ListSecrets(t.Context(), id)
				if err != nil || len(secrets) != 0 {
					t.Fatalf("purged secrets=%v %v", secrets, err)
				}
			}
			// 管理入口重新校验同来源发布包，随后沿生产 Supervisor/Attach 恢复。
			reinstalled, err := manager.Install(t.Context(), installation)
			if err != nil {
				t.Fatal(err)
			}
			if reinstalled.Source != "github.com/AlonaBot/testtasks" {
				t.Fatal(reinstalled)
			}
			active, err := h.store.Active(t.Context(), id)
			if err != nil {
				t.Fatal(err)
			}
			if active == "" {
				if err := h.store.AddGeneration(t.Context(), id, pipelineGeneration+":"+id, "{}", 1); err != nil {
					t.Fatal(err)
				}
			}
			c = h.launch(id, "tasks", 0)
			restored, err := h.store.Tasks(t.Context(), id, "remind")
			if err != nil {
				t.Fatal(err)
			}
			if scope == manage.PurgeNone {
				if len(restored) != 1 || !restored[0].Anchor.Equal(before[0].Anchor) {
					t.Fatalf("reinstalled=%v", restored)
				}
				count := len(h.snapshot())
				h.await(func(r []pipelineRecord) bool { return hasTaskRecord(r[count:], "completed", "cycle") })
			} else if len(restored) != 0 {
				t.Fatalf("purged task returned: %v", restored)
			}
		}
	})
}
