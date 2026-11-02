// Package bus 提供非持久 Event Bus 的发布准入与逐订阅插件的有界发送队列。
//
// [Bus.Register] 将发布入口接入 router.Mux，身份取自认证 Session 的 Caller。
// 装配处在插件激活和停止时调用 [Bus.Attach]、[Bus.Detach]；离线事件不缓存、不补发。
// 每个订阅插件有独立的发送 goroutine，写出结果仅供计数，无业务回执（CONTRACTS §3、WIRE §5）。
package bus

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

const (
	eventIDBytes = 16
	// 名字部分与 contracts/validate 的 Service 名称规则相同。
	maxEventNameLength = 64
)

// Options 配置逐订阅插件的等待队列；零值读取 tune.Defaults()。
type Options struct {
	QueueEvents int   // BUS_QUEUE_EVENTS：等待发送的条数上限
	QueueBytes  int64 // BUS_QUEUE_BYTES：等待发送的 BusEvent protobuf 载荷字节上限
}

// Stats 汇总当前 Core 运行期间的发布、写出与丢弃数。
type Stats struct {
	Published uint64 // 对发布插件计数，表示发布准入成功
	Delivered uint64 // 对订阅插件计数，表示 EVENT 已写出
	Dropped   uint64 // 对订阅插件计数，包含队列溢出、撤销与发送失败
}

// Bus 保存订阅队列与按插件累计的计数；New 创建，Register/Attach/Detach 接入装配。
// mu 串行化发布入队与订阅变更；发送和日志均在锁外进行，慢 Session 只占用自己的发送 goroutine。
type Bus struct {
	log         *slog.Logger
	opt         Options
	mu          sync.Mutex // 保护 subscribers、counts 和 subscriber 的队列；不跨 SendEvent 持锁
	subscribers map[string]*subscriber
	counts      map[string]*Stats // Detach 保留计数，新 Session 从空队列开始
}

type subscriber struct {
	plugin router.Plugin
	id     string
	routes []*pb.BusSubscriptionSpec // 已由 REGISTER 校验的不可变订阅描述
	ctx    context.Context
	cancel context.CancelFunc
	signal chan struct{} // 容量 1，唤醒独立 sender；队列本身由 Bus.mu 保护
	queue  []queuedEvent
	bytes  int64
	stats  *Stats // 由 Bus.mu 保护，与 Bus.counts 中的累计计数共享
}

type queuedEvent struct {
	target string
	event  *pb.BusEvent
	bytes  int64
}

type droppedEvent struct {
	pluginID  string
	eventName string
	count     uint64
}

// New 创建独立的非持久 Bus；log 由 Core 注入。
func New(log *slog.Logger, opt Options) *Bus {
	defaults := tune.Defaults()
	if opt.QueueEvents == 0 {
		opt.QueueEvents = defaults.BusQueueEvents
	}
	if opt.QueueBytes == 0 {
		opt.QueueBytes = defaults.BusQueueBytes
	}
	return &Bus{log: log, opt: opt, subscribers: make(map[string]*subscriber), counts: make(map[string]*Stats)}
}

// Register 将 core.bus.publish 注册到 m。
func (b *Bus) Register(m *router.Mux) { m.Handle(wire.ServiceBusPublish, b.publish) }

// Attach 为插件的 BUS_SUBSCRIPTION 路由建立共享队列和一个 sender。
// 同插件重新接入时撤销原队列及发送 context，事件不转交给新 Session。
func (b *Bus) Attach(p router.Plugin) {
	var routes []*pb.BusSubscriptionSpec
	for _, route := range p.Routes() {
		if route.GetKind() == pb.RouteKind_ROUTE_KIND_BUS_SUBSCRIPTION {
			routes = append(routes, route.GetBusSubscription())
		}
	}
	id := p.PluginID()
	b.mu.Lock()
	dropped := b.detachLocked(id)
	var sub *subscriber
	if len(routes) > 0 {
		ctx, cancel := context.WithCancel(context.Background())
		sub = &subscriber{plugin: p, id: id, routes: routes, ctx: ctx, cancel: cancel, signal: make(chan struct{}, 1), stats: b.counterLocked(id)}
		b.subscribers[id] = sub
	}
	b.mu.Unlock()
	for _, d := range dropped {
		b.warnDrop(d, "subscriber detached")
	}
	// sender 的生命周期归本次 Attach；Detach 取消后退出。
	if sub != nil {
		go b.send(sub)
	}
}

// Detach 撤销订阅，丢弃未发送事件并取消正在等待写出的 SendEvent。
func (b *Bus) Detach(pluginID string) {
	b.mu.Lock()
	dropped := b.detachLocked(pluginID)
	b.mu.Unlock()
	for _, d := range dropped {
		b.warnDrop(d, "subscriber detached")
	}
}

// detachLocked 丢弃旧队列；调用方持有 b.mu，发送中的事件由 sender 记录最终结果。
func (b *Bus) detachLocked(id string) []droppedEvent {
	sub, ok := b.subscribers[id]
	if !ok {
		return nil
	}
	delete(b.subscribers, id)
	sub.cancel()
	dropped := make([]droppedEvent, 0, len(sub.queue))
	for _, q := range sub.queue {
		dropped = append(dropped, b.dropLocked(sub, q.event))
	}
	sub.queue, sub.bytes = nil, 0
	return dropped
}

// Stats 返回插件的累计计数快照；Published 对发布者计数，其余对订阅者计数。
func (b *Bus) Stats(pluginID string) Stats {
	b.mu.Lock()
	defer b.mu.Unlock()
	if stats := b.counts[pluginID]; stats != nil {
		return *stats
	}
	return Stats{}
}

// counterLocked 取得插件的累计计数；调用方持有 b.mu。
func (b *Bus) counterLocked(id string) *Stats {
	stats := b.counts[id]
	if stats == nil {
		stats = &Stats{}
		b.counts[id] = stats
	}
	return stats
}

func (b *Bus) publish(_ context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	// 1. 认证身份只取自 Caller；先完成所有准入检查，再分发。
	req := &pb.PublishRequest{}
	payload := in.Payload
	if payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || payload.SchemaID != validate.SchemaID(req) || payload.SchemaVersion != validate.SchemaVersion {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "expected PublishRequest protobuf schema version 1")
	}
	if err := proto.Unmarshal(payload.Bytes, req); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid publish request protobuf: "+err.Error())
	}
	name, owned := strings.CutPrefix(req.GetEventName(), caller.PluginID+"/")
	if !owned || !validEventName(name) {
		return failure(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "event name must belong to the caller namespace and use a valid service name")
	}
	if req.GetSchemaId() == "" || req.GetSchemaVersion() == 0 {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "event requires a schema ID and non-zero version")
	}
	switch req.GetEncoding() {
	case pb.PayloadEncoding_PAYLOAD_ENCODING_JSON:
		if !utf8.Valid(req.GetPayload()) || !json.Valid(req.GetPayload()) {
			return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "event payload is not valid UTF-8 JSON")
		}
	case pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF:
	default:
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "event requires JSON or PROTOBUF encoding")
	}
	// 2. Core 分配标识和接纳时刻；BusEvent 没有 Actor 或用户权限字段。
	var id [eventIDBytes]byte
	rand.Read(id[:])
	eventID := hex.EncodeToString(id[:])
	ev := &pb.BusEvent{EventName: req.GetEventName(), EventId: eventID, SourcePluginId: caller.PluginID, OccurredAtUnixMs: time.Now().UnixMilli(),
		SchemaId: req.GetSchemaId(), SchemaVersion: req.GetSchemaVersion(), Encoding: req.GetEncoding(), Payload: req.GetPayload()}
	result := &pb.PublishResponse{EventId: eventID}
	encoded, err := proto.Marshal(result)
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "encode publish response: "+err.Error())
	}
	n := int64(proto.Size(ev))
	var dropped []droppedEvent
	// 3. 同一次发布的各订阅者入队顺序由 mu 串行化，绝不等待发送完成。
	b.mu.Lock()
	b.counterLocked(caller.PluginID).Published++
	for _, sub := range b.subscribers {
		if !sub.plugin.Activated() {
			continue
		}
		for _, route := range sub.routes {
			if route.GetEventName() != ev.GetEventName() || route.GetSchemaId() != ev.GetSchemaId() || route.GetSchemaVersion() != ev.GetSchemaVersion() {
				continue
			}
			if len(sub.queue) >= b.opt.QueueEvents || sub.bytes+n > b.opt.QueueBytes {
				dropped = append(dropped, b.dropLocked(sub, ev))
				continue
			}
			sub.queue = append(sub.queue, queuedEvent{target: route.GetCallbackRouteId(), event: ev, bytes: n})
			sub.bytes += n
			// 已有唤醒信号时不等待 sender。
			select {
			case sub.signal <- struct{}{}:
			default:
			}
		}
	}
	b.mu.Unlock()
	for _, d := range dropped {
		b.warnDrop(d, "subscriber queue is full")
	}
	return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(result), SchemaVersion: validate.SchemaVersion, Bytes: encoded}}
}

func (b *Bus) send(sub *subscriber) {
	for {
		// 1. 锁内取出一条；空队列只等待信号或本次订阅被撤销。
		b.mu.Lock()
		if sub.ctx.Err() != nil {
			b.mu.Unlock()
			return
		}
		if len(sub.queue) == 0 {
			b.mu.Unlock()
			select {
			case <-sub.signal:
			case <-sub.ctx.Done():
				return
			}
			continue
		}
		q := sub.queue[0]
		sub.queue[0] = queuedEvent{}
		sub.queue = sub.queue[1:]
		sub.bytes -= q.bytes
		active := sub.plugin.Activated()
		b.mu.Unlock()
		// 2. 单个 Session 的写出可以阻塞，只占用该订阅者的 sender。
		var err error
		if active {
			err = sub.plugin.SendEvent(sub.ctx, q.target, q.event)
		} else {
			err = errors.New("subscriber is not active")
		}
		b.mu.Lock()
		var dropped droppedEvent
		if err == nil {
			sub.stats.Delivered++
		} else {
			dropped = b.dropLocked(sub, q.event)
		}
		b.mu.Unlock()
		if err != nil {
			b.warnDrop(dropped, err.Error())
		}
	}
}

// dropLocked 记录丢弃并捕获日志字段；调用方持有 b.mu，日志在锁外写入。
func (b *Bus) dropLocked(sub *subscriber, ev *pb.BusEvent) droppedEvent {
	sub.stats.Dropped++
	return droppedEvent{pluginID: sub.id, eventName: ev.GetEventName(), count: sub.stats.Dropped}
}

func (b *Bus) warnDrop(d droppedEvent, reason string) {
	b.log.Warn("bus event dropped", "plugin_id", d.pluginID, "event_name", d.eventName, "dropped", d.count, "reason", reason)
}

func failure(code pb.ErrorCode, message string) wire.Response {
	return wire.Response{Err: &pb.ErrorInfo{Code: code, SafeMessage: message}}
}

func validEventName(name string) bool {
	if len(name) == 0 || len(name) > maxEventNameLength || name[0] < 'a' || name[0] > 'z' {
		return false
	}
	for _, c := range name {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '.') {
			return false
		}
	}
	return true
}
