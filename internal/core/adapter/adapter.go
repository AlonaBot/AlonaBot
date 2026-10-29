// Package adapter 定义 Core 接纳事件、发送消息与查找平台实例的接口。
//
// [Instance] 负责平台接入，[Sink] 只接纳事件，[Registry] 按实例 ID 保存接入。
// 发送结果与错误可以同时存在，调用方须保留二者（CONTRACTS §8）。
package adapter

import (
	"context"
	"errors"
	"fmt"
	"sync"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
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
	mu        sync.RWMutex // 保护 instances；不在锁内调用平台操作
	instances map[string]Instance
}

// NewRegistry 创建空的实例注册表。
func NewRegistry() *Registry { return &Registry{instances: make(map[string]Instance)} }

// Add 注册实例；重复的 instance_id 返回错误，原实例保持不变。
func (r *Registry) Add(in Instance) error {
	id := in.InstanceID()
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.instances[id]; ok {
		return fmt.Errorf("adapter instance %q already registered", id)
	}
	r.instances[id] = in
	return nil
}

// Get 查找实例 ID 对应的接入。
func (r *Registry) Get(instanceID string) (Instance, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	in, ok := r.instances[instanceID]
	return in, ok
}

// Remove 撤销实例登记；不存在的 ID 无副作用，可与 Add、Get 并发调用。
func (r *Registry) Remove(instanceID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.instances, instanceID)
}
