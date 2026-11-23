// Package adapter 定义 Core 接纳事件、发送消息与查找平台实例的接口。
//
// [Instance] 负责平台接入，[Sink] 只接纳事件，[Registry] 按实例 ID 保存接入。
// 发送结果与错误可以同时存在，调用方须保留二者。
package adapter

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"sync"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/proto"
)

// ErrQueueFull 表示 Core 的事件接纳队列已满。
var ErrQueueFull = errors.New("Core event queue is full")

// Sink 接纳事件而不等待 Handler；队列已满时返回 ErrQueueFull。
type Sink interface {
	Deliver(ctx context.Context, ev *pb.Event) error
}

// Instance 提供一个平台账号实例的接收与发送入口。
type Instance interface {
	AdapterID() string
	InstanceID() string
	AccountID() string
	Capabilities() []*pb.Capability
	// Start 完成凭证和冲突检查，启动接收后返回。
	Start(ctx context.Context, sink Sink) error
	Stop(ctx context.Context) error
	// Send 保留已确认结果，即使同时返回错误。
	Send(ctx context.Context, target *pb.Target, chain *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo)
}

// Registry 并发安全地按实例 ID 保存 Adapter 实例；使用 NewRegistry 创建。
type Registry struct {
	mu        sync.RWMutex // 保护 instances 与 defaults；不在锁内调用平台操作
	epochs    map[string]*instanceEpoch
	instances map[string]Instance
	defaults  map[string]string
}

// NewRegistry 创建空的实例注册表。
func NewRegistry() *Registry {
	return &Registry{instances: make(map[string]Instance), epochs: make(map[string]*instanceEpoch)}
}

// Add 注册实例；重复的 instance_id 返回错误，原实例保持不变。
func (r *Registry) Add(in Instance) error {
	id := in.InstanceID()
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.instances[id]; ok {
		return fmt.Errorf("adapter instance %q already registered", id)
	}
	r.instances[id] = in
	r.epochs[id] = &instanceEpoch{}
	return nil
}

// Get 查找实例 ID 对应的接入。
func (r *Registry) Get(instanceID string) (Instance, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	in, ok := r.instances[instanceID]
	return in, ok
}

// SetDefaults 复制并整体替换各 Adapter 的缺省实例，可与其他方法并发调用。
func (r *Registry) SetDefaults(defaults map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaults = maps.Clone(defaults)
}

// Complete 按在途会话或缺省实例补齐目标的空字段，返回副本且保留已有字段。
// inflight 可以为空；未登记的实例留给调用方按原有准入顺序检查。
func (r *Registry) Complete(target, inflight *pb.Target) (*pb.Target, *pb.ErrorInfo) {
	// 1. 插件输入须包含会话及至少一种 Adapter 标识。
	if target.GetChatId() == "" {
		return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: "chat_id"}
	}
	if target.GetAdapterId() == "" && target.GetAdapterInstanceId() == "" {
		return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: "adapter_id, adapter_instance_id"}
	}
	completed := proto.CloneOf(target)
	// 2. 同一锁内选择实例并读取登记快照，避免默认表与实例表交错更新。
	r.mu.RLock()
	if completed.AdapterInstanceId == "" {
		if inflight != nil && inflight.GetAdapterId() == completed.AdapterId && inflight.GetChatId() == completed.ChatId {
			completed.AdapterInstanceId = inflight.GetAdapterInstanceId()
		} else {
			completed.AdapterInstanceId = r.defaults[completed.AdapterId]
		}
		if completed.AdapterInstanceId == "" {
			r.mu.RUnlock()
			return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, SafeMessage: "no adapter instance selected; specify one or set default_instance"}
		}
	}
	instance, ok := r.instances[completed.AdapterInstanceId]
	r.mu.RUnlock()
	// 3. 身份由实例提供，锁外读取，已给出的字段由调用方执行绑定检查。
	if ok {
		if completed.AdapterId == "" {
			completed.AdapterId = instance.AdapterID()
		}
		if completed.AccountId == "" {
			completed.AccountId = instance.AccountID()
		}
	}
	return completed, nil
}

// Remove 撤销实例登记；不存在的 ID 无副作用，可与 Add、Get 并发调用。
func (r *Registry) Remove(instanceID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.instances, instanceID)
	delete(r.epochs, instanceID)
}

// NodesReader 查询当前目标内的完整递归节点，保留结果与错误。
type NodesReader interface {
	ReadNodes(context.Context, *pb.Target, *pb.ForwardSegment) ([]*pb.ForwardItem, *pb.ErrorInfo)
}
