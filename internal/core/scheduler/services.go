package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/schedule"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

const maxTaskKeyBytes = 128 // CONTRACTS §4.3。

// Services 将六个固定服务绑定到 Mux；current 核对调用者的当前 Session。
func (s *Scheduler) Services(m *router.Mux, current func(router.Caller) (router.Plugin, bool)) {
	for _, name := range []string{wire.ServiceTaskSet, wire.ServiceTaskCancel, wire.ServiceTaskList, wire.ServiceJobPause, wire.ServiceJobResume, wire.ServiceJobStatus} {
		m.Handle(name, func(ctx context.Context, c router.Caller, in *wire.Inbound) wire.Response {
			s.mu.Lock()
			defer s.mu.Unlock()
			p, ok := current(c)
			plugin := s.plugins[c.PluginID]
			if !ok || !p.Activated() || s.closed || plugin == nil || plugin.current == nil || !plugin.current.plugin.Activated() || !sameRoutes(p, plugin.current.plugin) {
				return failure(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "task services require the current activated Session")
			}
			switch in.Target {
			case wire.ServiceTaskSet:
				return s.set(ctx, c.PluginID, plugin, in)
			case wire.ServiceTaskCancel:
				return s.cancel(ctx, c.PluginID, plugin, in)
			case wire.ServiceTaskList:
				return s.list(ctx, c.PluginID, plugin, in)
			default:
				return s.job(ctx, c.PluginID, plugin, in)
			}
		})
	}
}

// Routes 中的描述对象随 Session 解码；核对它们可避免 Router 已切换、Scheduler 尚未发布时使用旧注册。
func sameRoutes(a, b router.Plugin) bool {
	return slices.Equal(a.Routes(), b.Routes())
}

func decode(in *wire.Inbound, request proto.Message) error {
	if in.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.Payload.SchemaID != validate.SchemaID(request) || in.Payload.SchemaVersion != handlerSchemaVersion {
		return fmt.Errorf("expected %s protobuf schema version %d", validate.SchemaID(request), handlerSchemaVersion)
	}
	if err := proto.Unmarshal(in.Payload.Bytes, request); err != nil {
		return fmt.Errorf("invalid task protobuf: %w", err)
	}
	return nil
}
func failure(code pb.ErrorCode, message string) wire.Response {
	return wire.Response{Err: &pb.ErrorInfo{Code: code, SafeMessage: message}}
}
func response(result proto.Message) wire.Response {
	encoded, err := proto.Marshal(result)
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "encode task response: "+err.Error())
	}
	return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(result), SchemaVersion: handlerSchemaVersion, Bytes: encoded}}
}
func storageFailure(pluginID string, err error) wire.Response {
	var validation *validate.Error
	if errors.As(err, &validation) {
		return failure(validation.Code, validation.Message)
	}
	// 完整错误留在 Core 诊断中；安全错误通道不泄露宿主路径或任务参数。
	slog.Error("task storage operation failed", "plugin_id", pluginID, "error", err)
	return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "task storage operation failed")
}
func validKey(key string) bool {
	return len(key) > 0 && len(key) <= maxTaskKeyBytes && utf8.ValidString(key)
}

func (s *Scheduler) set(ctx context.Context, pluginID string, plugin *pluginJobs, in *wire.Inbound) wire.Response {
	// 1. 所有输入校验完毕后才写入；失败保留旧定义、timer 和队列。
	request := &pb.TaskSetRequest{}
	if err := decode(in, request); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	if !validKey(request.Key) {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "task key must be 1 to 128 bytes of valid UTF-8")
	}
	if int64(len(request.PayloadJson)) > s.payloadBytes || !utf8.ValidString(request.PayloadJson) || !validate.IsJSONObject([]byte(request.PayloadJson)) {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "task payload must be a JSON object within TASK_PAYLOAD_BYTES")
	}
	now := s.now()
	task := state.Task{PluginID: pluginID, Name: request.TaskName, Key: request.Key, PayloadJSON: request.PayloadJson, CreatedAt: now}
	if request.Trigger == nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "task trigger is required")
	}
	var trigger schedule.Trigger
	var err error
	switch t := request.Trigger.GetTrigger().(type) {
	case *pb.TaskTrigger_AtUnixMs:
		task.At = time.UnixMilli(t.AtUnixMs)
		if !task.At.After(now) {
			return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "task at must be in the future")
		}
	case *pb.TaskTrigger_IntervalMs:
		if t.IntervalMs > uint64(math.MaxInt64/int64(time.Millisecond)) {
			return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "task interval exceeds duration range")
		}
		task.Interval = time.Duration(t.IntervalMs) * time.Millisecond
		task.Anchor = now
		trigger, err = schedule.Interval(task.Interval)
	case *pb.TaskTrigger_Cron:
		task.Cron, task.Timezone = t.Cron, request.Trigger.Timezone
		if task.Timezone == "" {
			task.Timezone = "UTC"
		}
		trigger, err = schedule.Cron(task.Cron, task.Timezone)
	default:
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "task requires exactly one trigger")
	}
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	if task.Cron == "" && request.Trigger.Timezone != "" {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "timezone only applies to cron")
	}
	if !plugin.current.names[request.TaskName] {
		return failure(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "task name is not registered")
	}
	if err := s.store.SetTask(ctx, task, s.maxTasks); err != nil {
		return storageFailure(pluginID, err)
	}
	// 2. 提交成功才替换本 Session 的未来安排；同键旧在途调用继续持有稳定槽。
	key := taskKey{task.Name, task.Key}
	removeTask(plugin.current, key)
	registered := &registeredTask{definition: task, trigger: trigger, state: s.taskState(plugin, key)}
	plugin.current.tasks[key] = registered
	s.armTask(pluginID, plugin.current, registered, now)
	return response(&pb.TaskSetResponse{NextAtUnixMs: registered.next.UnixMilli()})
}

func (s *Scheduler) cancel(ctx context.Context, pluginID string, plugin *pluginJobs, in *wire.Inbound) wire.Response {
	request := &pb.TaskCancelRequest{}
	if err := decode(in, request); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	if request.TaskName == "" || !validKey(request.Key) {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "task name and valid key are required")
	}
	key := taskKey{request.TaskName, request.Key}
	js := plugin.tasks[key]
	err := s.store.DeleteTask(ctx, pluginID, key.name, key.key)
	if err != nil {
		if !errors.Is(err, state.ErrNotFound) {
			return storageFailure(pluginID, err)
		}
		if js == nil || js.running == nil {
			return failure(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "task instance not found")
		}
	} else if !plugin.current.names[key.name] {
		plugin.taskStats.Unregistered--
	}
	removeTask(plugin.current, key)
	if js != nil && js.running != nil {
		js.running.cancel()
	} else {
		delete(plugin.tasks, key)
	}
	return response(&pb.TaskCancelResponse{})
}

func taskTrigger(task state.Task) *pb.TaskTrigger {
	switch {
	case !task.At.IsZero():
		return &pb.TaskTrigger{Trigger: &pb.TaskTrigger_AtUnixMs{AtUnixMs: task.At.UnixMilli()}}
	case task.Interval != 0:
		return &pb.TaskTrigger{Trigger: &pb.TaskTrigger_IntervalMs{IntervalMs: uint64(task.Interval / time.Millisecond)}}
	default:
		return &pb.TaskTrigger{Trigger: &pb.TaskTrigger_Cron{Cron: task.Cron}, Timezone: task.Timezone}
	}
}

// list 按完整 protobuf 和 RESPONSE Envelope 的真实编码大小分页，游标只在还有数据时出现。
func (s *Scheduler) list(ctx context.Context, pluginID string, plugin *pluginJobs, in *wire.Inbound) wire.Response {
	request := &pb.TaskListRequest{}
	if err := decode(in, request); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	if request.TaskName == "" || (request.AfterKey != nil && !validKey(*request.AfterKey)) {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "task name and valid cursor are required")
	}
	tasks, err := s.store.Tasks(ctx, pluginID, request.TaskName)
	if err != nil {
		return storageFailure(pluginID, err)
	}
	slices.SortFunc(tasks, func(a, b state.Task) int { return strings.Compare(a.Key, b.Key) })
	tasks = slices.DeleteFunc(tasks, func(t state.Task) bool { return request.AfterKey != nil && t.Key <= *request.AfterKey })
	result := &pb.TaskListResponse{}
	base := proto.Size(&pb.Envelope{ProtocolMajor: validate.ProtocolMajor, ProtocolMinor: validate.ProtocolMinor, SessionId: make([]byte, validate.SessionIDLen), Kind: pb.FrameKind_FRAME_KIND_RESPONSE, InitiatorSide: pb.InitiatorSide_INITIATOR_SIDE_PLUGIN, CallId: in.CallID, PayloadEncoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaId: validate.SchemaID(result), SchemaVersion: handlerSchemaVersion})
	size := 0
	for i, task := range tasks {
		instance := &pb.TaskInstance{Key: task.Key, Trigger: taskTrigger(task), PayloadJson: task.PayloadJSON}
		if registered := plugin.current.tasks[taskKey{task.Name, task.Key}]; registered != nil && !registered.next.IsZero() {
			instance.NextAtUnixMs = proto.Int64(registered.next.UnixMilli())
		}
		itemSize := protowire.SizeTag(1) + protowire.SizeBytes(proto.Size(instance))
		candidate := size + itemSize
		if i+1 < len(tasks) {
			candidate += protowire.SizeTag(2) + protowire.SizeBytes(len(task.Key))
		}
		if base+protowire.SizeTag(15)+protowire.SizeBytes(candidate) > wire.MaxFrame {
			if len(result.Instances) == 0 {
				return failure(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "one task instance exceeds frame limit")
			}
			result.NextAfterKey = proto.String(result.Instances[len(result.Instances)-1].Key)
			return response(result)
		}
		result.Instances = append(result.Instances, instance)
		size += itemSize
	}
	return response(result)
}
