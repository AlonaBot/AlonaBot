// Package task 声明、登记与管理带 JSON 对象参数的动态任务，也控制静态任务状态。
package task

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
)

const (
	taskRoutePrefix = "task."
	taskKeyBytes    = 16 // 十六进制编码后为 32 字节，符合 Core 的 key 约束。
)

// Type 绑定任务名与参数类型；零值无效。
type Type[T any] struct{ name string }

// Define 声明带 JSON 对象参数的任务，类型规则由生成器与 Service 共用。
func Define[T any](name string) Type[T] { return Type[T]{name: name} }

// Builder 收集类型化任务；Handle 结束登记。
type Builder[T any] struct {
	scope   *registry.Scope[*sdk.Context, sdk.Permission]
	handles int
	handler func(*sdk.Context, T) error
}

// On 开始在插件或子插件上登记任务，任务名在整个插件内共享。
func On[T any](r sdk.Registrar, t Type[T]) *Builder[T] {
	scope := registry.For[*sdk.Context, sdk.Permission](r)
	scope.Check()
	b := &Builder[T]{scope: scope}
	scope.Tasks.Pending = append(scope.Tasks.Pending, func() (registry.Task[*sdk.Context], error) {
		route := registry.Task[*sdk.Context]{Name: t.name, SubID: scope.SubID}
		if b.handles != 1 || b.handler == nil {
			return route, fmt.Errorf("task %q requires exactly one non-nil Handle", t.name)
		}
		if _, _, err := registry.JSONType[T](); err != nil {
			return route, fmt.Errorf("task %q: %w", t.name, err)
		}
		typ := reflect.TypeFor[T]()
		if typ.Kind() != reflect.Struct && typ.Kind() != reflect.Map {
			return route, fmt.Errorf("task %q parameters must encode a JSON object", t.name)
		}
		handler := b.handler // 注册收集阶段固定本次 Session 的处理函数。
		route.Invoke = func(c *sdk.Context) error {
			// 协议边界核对任务身份，解码成功后才进入用户函数。
			event, inbound := scope.TaskContext(c)
			payload := event.GetEventPayload().GetScheduled()
			if event.GetEventType() != "scheduled" || payload == nil || payload.GetTaskName() != t.name || payload.GetJobId() != taskRoutePrefix+t.name || inbound.Target != taskRoutePrefix+t.name {
				c.Logger().Warn("task invocation does not match registered task", "task_name", t.name)
				return errcode.InvalidArgument.Errorf("task %q invocation does not match registered task", t.name)
			}
			var args T
			if err := decodeTaskArgs(payload.GetTaskPayloadJson(), &args); err != nil {
				c.Logger().Warn("decode task payload failed", "task_name", t.name, "error", err)
				return errcode.InvalidArgument.Errorf("task %q payload: %v", t.name, err)
			}
			return handler(c, args)
		}
		return route, nil
	})
	return b
}

// Handle 结束任务登记；缺失、重复或 nil 函数在 Run 收集注册时报错。
func (b *Builder[T]) Handle(fn func(*sdk.Context, T) error) {
	b.scope.Check()
	b.handles++
	b.handler = fn
}

// Option 调整一次任务实例的 key 或 cron 时区。
type Option func(*taskOptions)
type taskOptions struct {
	key, timezone       string
	keySet, timezoneSet bool
}

// Key 指定实例 key；同任务名和 key 的实例由 Core 整体替换。
func Key(key string) Option { return func(o *taskOptions) { o.key, o.keySet = key, true } }

// Timezone 指定 cron 的 IANA 时区；At 与 Every 上显式设置时返回 INVALID_ARGUMENT。
func Timezone(name string) Option {
	return func(o *taskOptions) { o.timezone, o.timezoneSet = name, true }
}

// At 创建或替换一次性实例，成功返回实际使用的 key。
func (t Type[T]) At(ctx context.Context, when time.Time, args T, opts ...Option) (string, error) {
	return t.set(ctx, &pb.TaskTrigger{Trigger: &pb.TaskTrigger_AtUnixMs{AtUnixMs: when.UnixMilli()}}, args, opts)
}

// Every 创建或替换固定间隔实例，触发规则由 Core 校验。
func (t Type[T]) Every(ctx context.Context, d time.Duration, args T, opts ...Option) (string, error) {
	return t.set(ctx, &pb.TaskTrigger{Trigger: &pb.TaskTrigger_IntervalMs{IntervalMs: uint64(d / time.Millisecond)}}, args, opts)
}

// Cron 创建或替换五字段 cron 实例，时区缺省为 UTC。
func (t Type[T]) Cron(ctx context.Context, expr string, args T, opts ...Option) (string, error) {
	return t.set(ctx, &pb.TaskTrigger{Trigger: &pb.TaskTrigger_Cron{Cron: expr}}, args, opts)
}

func (t Type[T]) set(ctx context.Context, trigger *pb.TaskTrigger, args T, opts []Option) (string, error) {
	// 1. 先走既有业务门禁，再编码作者输入；非法组合不发出请求。
	if _, err := registry.Session(ctx); err != nil {
		return "", err
	}
	options := taskOptions{}
	for _, option := range opts {
		option(&options)
	}
	if _, cron := trigger.Trigger.(*pb.TaskTrigger_Cron); options.timezoneSet && !cron {
		return "", errcode.InvalidArgument.Errorf("task %q can only set Timezone on Cron", t.name)
	}
	trigger.Timezone = options.timezone
	encoded, err := json.Marshal(args)
	if err != nil {
		return "", errcode.InvalidArgument.Errorf("encode task %q parameters: %v", t.name, err)
	}
	if !validate.IsJSONObject(encoded) {
		return "", errcode.InvalidArgument.Errorf("task %q parameters must encode a JSON object", t.name)
	}
	if !options.keySet {
		var random [taskKeyBytes]byte
		if _, err := io.ReadFull(rand.Reader, random[:]); err != nil {
			return "", fmt.Errorf("generate task %q key: %w", t.name, err)
		}
		options.key = hex.EncodeToString(random[:])
	}
	// 2. 固定服务保留父调用关联、协议错误码与唯一的响应解码路径。
	_, err = registry.Proto[*pb.TaskSetRequest, *pb.TaskSetResponse](wire.ServiceTaskSet).Call(ctx, &pb.TaskSetRequest{
		TaskName: t.name, Key: options.key, Trigger: trigger, PayloadJson: string(encoded),
	})
	if err != nil {
		return "", err
	}
	return options.key, nil
}

// Cancel 删除实例并协作取消其在途调用，保留 Core 的错误码。
func (t Type[T]) Cancel(ctx context.Context, key string) error {
	_, err := registry.Proto[*pb.TaskCancelRequest, *pb.TaskCancelResponse](wire.ServiceTaskCancel).Call(ctx, &pb.TaskCancelRequest{TaskName: t.name, Key: key})
	return err
}

// Instance 保留实例的完整参数与互斥触发规则；未选的触发字段为零值。
type Instance[T any] struct {
	Key            string
	Args           T
	At             time.Time
	Interval       time.Duration
	Cron, Timezone string
	NextAt         *time.Time // nil 表示协议中没有安排下一次触发。
}

// List 按 Core 的 key 顺序读取全部分页；任一页失败返回错误。
func (t Type[T]) List(ctx context.Context) ([]Instance[T], error) {
	request := &pb.TaskListRequest{TaskName: t.name}
	var instances []Instance[T]
	for {
		response, err := registry.Proto[*pb.TaskListRequest, *pb.TaskListResponse](wire.ServiceTaskList).Call(ctx, request)
		if err != nil {
			return nil, err
		}
		for _, value := range response.Instances {
			instance := Instance[T]{Key: value.Key, NextAt: optionalTaskTime(value.NextAtUnixMs)}
			if err := decodeTaskArgs(value.PayloadJson, &instance.Args); err != nil {
				return nil, errcode.InvalidArgument.Errorf("decode task %q instance %q: %v", t.name, value.Key, err)
			}
			switch trigger := value.GetTrigger().GetTrigger().(type) {
			case *pb.TaskTrigger_AtUnixMs:
				instance.At = time.UnixMilli(trigger.AtUnixMs)
			case *pb.TaskTrigger_IntervalMs:
				instance.Interval = time.Duration(trigger.IntervalMs) * time.Millisecond
			case *pb.TaskTrigger_Cron:
				instance.Cron, instance.Timezone = trigger.Cron, value.Trigger.Timezone
			default:
				return nil, errcode.Internal.Errorf("task %q instance %q has no trigger", t.name, value.Key)
			}
			instances = append(instances, instance)
		}
		if response.NextAfterKey == nil {
			return instances, nil
		}
		request.AfterKey = response.NextAfterKey
	}
}

func decodeTaskArgs[T any](payload string, args *T) error {
	if !validate.IsJSONObject([]byte(payload)) {
		return errors.New("parameters must be a JSON object")
	}
	return json.Unmarshal([]byte(payload), args)
}

func optionalTaskTime(ms *int64) *time.Time {
	if ms == nil {
		return nil
	}
	value := time.UnixMilli(*ms)
	return &value
}
