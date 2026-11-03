// Package memory 提供测试和示例使用的内存 Adapter，不进入生产装配。
//
// [Adapter] 经 Start 接收事件接纳入口，Inject 投递事件，Sent 返回发送快照。
// 默认发送为每个片段生成已确认回执；SetSendFunc 可模拟部分成功与错误。
package memory

import (
	"context"
	"crypto/rand"
	"errors"
	"strconv"
	"sync"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"google.golang.org/protobuf/proto"
)

// Sent 保存一次逻辑发送的目标与消息链快照。
type Sent struct {
	Target *pb.Target
	Chain  *pb.MessageChain
}

// Adapter 并发安全地保存接纳入口、发送记录与模拟发送函数。
// 构造时的身份保持不变；自定义发送函数在锁外执行，由调用方保证自身并发安全。
type Adapter struct {
	adapterID  string
	instanceID string
	accountID  string
	sessionKey string     // 构造时生成，实例整个生命周期固定
	mu         sync.Mutex // 保护 sink、sent、nextID 和 sendFunc
	sink       adapter.Sink
	sent       []Sent
	nextID     uint64
	sendFunc   func(*pb.Target, *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo)
}

var (
	_ adapter.Instance     = (*Adapter)(nil)
	_ adapter.MediaSession = (*Adapter)(nil)
)

// New 创建指定平台、实例与账号身份的内存接入。
func New(adapterID, instanceID, accountID string) *Adapter {
	return &Adapter{adapterID: adapterID, instanceID: instanceID, accountID: accountID, sessionKey: rand.Text()}
}

// AdapterID 返回平台类型。
func (a *Adapter) AdapterID() string { return a.adapterID }

// InstanceID 返回实例 ID。
func (a *Adapter) InstanceID() string { return a.instanceID }

// AccountID 返回平台账号 ID。
func (a *Adapter) AccountID() string { return a.accountID }

// MediaSessionKey 返回构造时固定的媒体消费 Session 身份。
func (a *Adapter) MediaSessionKey() string { return a.sessionKey }

// Capabilities 返回该内存接入的能力声明；当前没有额外平台能力。
func (a *Adapter) Capabilities() []*pb.Capability { return nil }

// Start 保存事件接纳入口，之后 Inject 可投递事件。
func (a *Adapter) Start(_ context.Context, sink adapter.Sink) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sink = sink
	return nil
}

// Stop 移除接纳入口；已经取得入口的 Inject 由原调用继续完成。
func (a *Adapter) Stop(context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sink = nil
	return nil
}

// Inject 经 Start 提供的 Sink 投递事件；未启动或已停止时返回错误。
func (a *Adapter) Inject(ctx context.Context, ev *pb.Event) error {
	a.mu.Lock()
	sink := a.sink
	a.mu.Unlock()
	// 锁外调用 Sink，避免下游派发阻塞实例的其他操作。
	if sink == nil {
		return errors.New("memory adapter is not started")
	}
	return sink.Deliver(ctx, ev)
}

// Sent 返回发送记录的独立快照，调用方修改它不会影响内部记录。
func (a *Adapter) Sent() []Sent {
	a.mu.Lock()
	defer a.mu.Unlock()
	sent := make([]Sent, len(a.sent))
	for i, s := range a.sent {
		sent[i] = Sent{Target: proto.CloneOf(s.Target), Chain: proto.CloneOf(s.Chain)}
	}
	return sent
}

// SetSendFunc 设置模拟发送函数；nil 恢复默认的已确认回执。
func (a *Adapter) SetSendFunc(fn func(*pb.Target, *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sendFunc = fn
}

// Send 先保存发送快照，再执行模拟函数或为每个片段返回 ACCEPTED 回执。
func (a *Adapter) Send(_ context.Context, target *pb.Target, chain *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo) {
	a.mu.Lock()
	a.sent = append(a.sent, Sent{Target: proto.CloneOf(target), Chain: proto.CloneOf(chain)})
	fn := a.sendFunc
	if fn != nil {
		a.mu.Unlock()
		// 回调可以读取 Sent，不在持锁时执行用户代码。
		return fn(target, chain)
	}
	result := &pb.SendResult{Status: pb.SendStatus_SEND_STATUS_ACCEPTED, Target: proto.CloneOf(target)}
	for i := range chain.GetSegments() {
		a.nextID++
		id := "m" + strconv.FormatUint(a.nextID, 10)
		result.PlatformMessageIds = append(result.PlatformMessageIds, id)
		result.PartResults = append(result.PartResults, &pb.PartResult{
			PartIndex: uint32(i), SegmentIndexes: []uint32{uint32(i)},
			Status: pb.PartStatus_PART_STATUS_ACCEPTED, PlatformMessageIds: []string{id},
		})
	}
	a.mu.Unlock()
	if len(result.PlatformMessageIds) > 0 {
		result.PrimaryPlatformMessageId = &result.PlatformMessageIds[0]
	}
	return result, nil
}
