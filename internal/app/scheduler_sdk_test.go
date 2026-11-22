//go:build unix

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	genresources "github.com/AlonaBot/AlonaBot/internal/generated"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	taskapi "github.com/AlonaBot/AlonaBot/sdk/task"
)

const schedulerSDKMode = "tasks-sdk"

type schedulerSDKArgs struct {
	Key     string `json:"key"`
	Text    string `json:"text"`
	Version int64  `json:"version"`
	Session string `json:"session"`
}

// schedulerSDKPlugin 经公开作者 API 操作真实 Core，审计只保存调用结果。
func schedulerSDKPlugin(mode string) *sdk.Plugin {
	p := sdk.NewPlugin()
	genresources.Bind(p, genresources.Resources{})
	task := taskapi.Define[schedulerSDKArgs]("remind")
	checkGate := func(ctx context.Context) error {
		_, err := task.List(ctx)
		if errcode.Of(err) != errcode.FailedPrecondition {
			return fmt.Errorf("task lifecycle gate: %v", err)
		}
		return nil
	}
	p.OnStart(checkGate)
	p.OnReady(checkGate)
	p.OnStop(checkGate)
	p.Every("pulse", time.Second).Handle(func(c *sdk.Context) error { return pipelineObserve("sdk-static", "pulse", c) })
	taskapi.On(p.Sub("reminders"), task).Handle(func(c *sdk.Context, args schedulerSDKArgs) error {
		if c.User() != "" || c.Session() != "" || c.Event().Actor != nil {
			return fmt.Errorf("task received an actor or default target")
		}
		if _, err := c.Send(message.Text(args.Text)); errcode.Of(err) != errcode.NoDefaultTarget {
			return fmt.Errorf("task default target: %v", err)
		}
		raw, err := json.Marshal(args)
		if err != nil {
			return err
		}
		if err := pipelineAudit(pipelineRecord{Stage: "sdk-task", Value: c.Event().Payload.(*message.ScheduledPayload).TaskKey, Text: string(raw)}); err != nil {
			return err
		}
		if args.Session != "" {
			if _, err := c.SendTo(args.Session, message.Text(args.Text)); err != nil {
				return err
			}
		}
		return nil
	})
	p.Command("task").Args("operation").Handle(func(c *sdk.Context, operation string) error {
		ctx := c.Context()
		args := func(key, text string, version int64) schedulerSDKArgs {
			return schedulerSDKArgs{Key: key, Text: text, Version: version, Session: c.Session()}
		}
		var err error
		switch operation {
		case "start":
			if _, err = task.At(ctx, time.Now().Add(time.Second), args("at", "once 参数", 1), taskapi.Key("at")); err != nil {
				return err
			}
			if _, err = task.Every(ctx, time.Second, args("every", "periodic 参数", 2), taskapi.Key("every")); err != nil {
				return err
			}
			if _, err = task.Cron(ctx, "* * * * *", args("cron", "cron 参数", 3), taskapi.Key("cron"), taskapi.Timezone("UTC")); err != nil {
				return err
			}
		case "replace":
			if _, err = task.At(ctx, time.Now().Add(time.Hour), args("replace", "old", 1), taskapi.Key("replace")); err != nil {
				return err
			}
			if _, err = task.Every(ctx, time.Hour, args("replace", "new 参数", 2), taskapi.Key("replace")); err != nil {
				return err
			}
			instances, listErr := task.List(ctx)
			if listErr != nil {
				return listErr
			}
			found := false
			for _, instance := range instances {
				if instance.Key != "replace" {
					continue
				}
				found = true
				if instance.Args.Text != "new 参数" || instance.Args.Version != 2 || instance.Interval != time.Hour || !instance.At.IsZero() || instance.Cron != "" || instance.Timezone != "" || instance.NextAt == nil {
					return fmt.Errorf("replacement lost fields: %+v", instance)
				}
			}
			if !found {
				return fmt.Errorf("replacement missing from List")
			}
			err = task.Cancel(ctx, "replace")
		case "cancel":
			for _, key := range []string{"every", "cron"} {
				if err = task.Cancel(ctx, key); err != nil {
					return err
				}
			}
			instances, listErr := task.List(ctx)
			if listErr != nil {
				return listErr
			}
			if len(instances) != 0 {
				return fmt.Errorf("cancel left instances: %+v", instances)
			}
		case "pause", "resume", "status":
			if operation == "pause" {
				err = taskapi.Pause(ctx, "pulse")
			}
			if operation == "resume" {
				err = taskapi.Resume(ctx, "pulse")
			}
			if err != nil {
				return err
			}
			state, statusErr := taskapi.Status(ctx, "pulse")
			if statusErr != nil {
				return statusErr
			}
			if operation == "pause" && (!state.Paused || state.NextAt != nil || state.LastAt == nil) {
				return fmt.Errorf("paused status: %+v", state)
			}
			if operation == "resume" && (state.Paused || state.NextAt == nil || state.LastAt == nil) {
				return fmt.Errorf("resumed status: %+v", state)
			}
			raw, encodeErr := json.Marshal(state)
			if encodeErr != nil {
				return encodeErr
			}
			if err := pipelineAudit(pipelineRecord{Stage: "sdk-status", Value: operation, Text: string(raw)}); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown task operation %q in %s", operation, mode)
		}
		if err != nil {
			return err
		}
		return pipelineAudit(pipelineRecord{Stage: "sdk-command", Value: operation})
	})
	return p
}

// TestSchedulerSDK 守住作者 API→协议→SQLite/Core 调度→类型化 Handler 的真实进程主路径。
func TestSchedulerSDK(t *testing.T) {
	h := newPipeline(t)
	h.assembly.adapters.SetDefaults(map[string]string{"loop": pipelineInstanceID})
	h.install("platform", "adapter")
	h.launch("platform", "adapter", 0)
	h.install("scheduler", schedulerSDKMode)
	h.launch("scheduler", schedulerSDKMode, 0)
	command := func(operation string) {
		t.Helper()
		h.inject(pipelineInput{Text: "/task " + operation, User: "alice"})
		h.await(func(records []pipelineRecord) bool { return hasTaskRecord(records, "sdk-command", operation) })
	}
	command("start")
	command("replace")
	// 五字段 cron 按真实分钟边界触发，等待上限包含下一分钟与现有进程测试预算。
	timer := time.NewTimer(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)) + pipelineTimeout)
	defer timer.Stop()
	for !(hasTaskRecord(h.snapshot(), "sdk-task", "at") && hasTaskRecord(h.snapshot(), "sdk-task", "every") && hasTaskRecord(h.snapshot(), "sdk-task", "cron")) {
		select {
		case <-h.changed:
		case <-timer.C:
			t.Fatalf("typed task triggers timed out: %+v", h.snapshot())
		}
	}
	for _, record := range h.snapshot() {
		if record.Stage != "sdk-task" {
			continue
		}
		var args schedulerSDKArgs
		if err := json.Unmarshal([]byte(record.Text), &args); err != nil {
			t.Fatal(err)
		}
		want := map[string]schedulerSDKArgs{
			"at":    {Key: "at", Text: "once 参数", Version: 1, Session: "loop:group#topic"},
			"every": {Key: "every", Text: "periodic 参数", Version: 2, Session: "loop:group#topic"},
			"cron":  {Key: "cron", Text: "cron 参数", Version: 3, Session: "loop:group#topic"},
		}
		if args != want[record.Value] {
			t.Fatalf("typed args=%+v want=%+v", args, want[record.Value])
		}
	}
	command("cancel")
	command("pause")
	// 给取消前的一轮留出实际返回时间，然后越过两个 interval，检查没有新触发。
	waitTaskState(t, func() bool {
		return h.assembly.scheduler.TasksStats("scheduler").Running == 0 && !h.assembly.scheduler.Stats("scheduler", "pulse").Running
	})
	before := h.snapshot()
	time.Sleep(2 * time.Second)
	after := h.snapshot()
	for _, record := range after[len(before):] {
		if record.Stage == "sdk-task" || record.Stage == "sdk-static" {
			t.Fatalf("trigger after cancel/pause: %+v", record)
		}
	}
	command("resume")
	count := len(pipelineStages(h.snapshot(), "sdk-static"))
	h.await(func(records []pipelineRecord) bool { return len(pipelineStages(records, "sdk-static")) > count })
	command("status")
	// 显式目标发送经过真实消息流水线，三种 typed 参数都到达回环 Adapter。
	for _, text := range []string{"once 参数", "periodic 参数", "cron 参数"} {
		found := false
		for _, record := range h.snapshot() {
			if record.Stage == "send" && record.Text == text {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing explicit task send %q", text)
		}
	}
}
