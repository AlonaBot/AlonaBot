package router

import (
	"cmp"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

const (
	traceIDBytes         = 16
	handlerSchemaVersion = 1
)

// Plugin 提供派发所需的插件身份、激活状态、静态路由与 Session 调用。
type Plugin interface {
	PluginID() string
	Activated() bool
	Routes() []*pb.RouteDescriptor
	Call(ctx context.Context, req wire.Request) (wire.Response, error)
}

type instancePlugin struct{ instance *host.Instance }

// FromInstance 把真实宿主实例接入路由器，激活状态从宿主实时读取。
func FromInstance(in *host.Instance) Plugin { return &instancePlugin{instance: in} }

func (p *instancePlugin) PluginID() string              { return p.instance.PluginID }
func (p *instancePlugin) Activated() bool               { return p.instance.State() == host.StateActivated }
func (p *instancePlugin) Routes() []*pb.RouteDescriptor { return p.instance.Routes }
func (p *instancePlugin) Call(ctx context.Context, req wire.Request) (wire.Response, error) {
	return p.instance.Session.Call(ctx, req)
}

// Router 按顺序派发已接纳事件；New 创建，Start 启动，Close 取消在途派发并等待 worker 退出。
// 队列接纳与插件注册由 mu 串行化，worker 调用插件时不持锁；队列不关闭，避免并发投递遇到已关闭 channel。
type Router struct {
	log     *slog.Logger
	queue   chan *pb.Event
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.Mutex // 保护 plugins、started、closed 与 worker 的启动
	plugins map[string]Plugin
	started bool
	closed  bool
	worker  sync.WaitGroup
}

// New 创建容量为 queueEvents 的事件队列，log 使用 Core 的 Logger。
// queueEvents 由装配处传入 tune.Values.DispatchQueueEvents。
func New(log *slog.Logger, queueEvents int) *Router {
	ctx, cancel := context.WithCancel(context.Background())
	return &Router{log: log, queue: make(chan *pb.Event, queueEvents), ctx: ctx, cancel: cancel, plugins: make(map[string]Plugin)}
}

// Attach 注册一个接收派发的插件；相同插件 ID 替换为新实例。
func (r *Router) Attach(p Plugin) {
	id := p.PluginID()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.plugins[id] = p
}

// Detach 使插件不再进入后续事件的候选集。
func (r *Router) Detach(pluginID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.plugins, pluginID)
}

// Start 启动唯一的派发 worker；重复启动或关闭后的启动不产生 worker。
func (r *Router) Start() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.started || r.closed {
		return
	}
	r.started = true
	r.worker.Add(1)
	// shortcut: 单 worker 且仅按条数限额，有界并发派发（DISPATCH_CONCURRENCY、DISPATCH_QUEUE_BYTES）接入后替换。
	go func() {
		defer r.worker.Done()
		for {
			select {
			case <-r.ctx.Done():
				return
			case ev := <-r.queue:
				r.dispatch(ev)
			}
		}
	}()
}

// Close 停止接纳、取消在途调用，并等待派发 worker 退出；未派发事件不再处理。
func (r *Router) Close() {
	r.mu.Lock()
	r.closed = true
	r.cancel()
	r.mu.Unlock()
	// 不持注册锁等待，派发中的 worker 仍可完成状态读取并退出。
	r.worker.Wait()
}

// Deliver 接纳事件快照，不等待 Handler；满队列立即返回 adapter.ErrQueueFull。
func (r *Router) Deliver(ctx context.Context, ev *pb.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return errors.New("event router is closed")
	}
	select {
	case r.queue <- proto.CloneOf(ev):
		return nil
	default:
		return adapter.ErrQueueFull
	}
}

type candidate struct {
	plugin   Plugin
	pluginID string
	route    *pb.RouteDescriptor
}

func (r *Router) candidates(ev *pb.Event) []candidate {
	r.mu.Lock()
	plugins := make([]Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		plugins = append(plugins, p)
	}
	r.mu.Unlock()
	var candidates []candidate
	for _, p := range plugins {
		if !p.Activated() {
			continue
		}
		for _, route := range p.Routes() {
			if route.GetKind() == pb.RouteKind_ROUTE_KIND_EVENT && slices.Contains(route.GetEvent().GetEventTypeFilter(), ev.GetEventType()) {
				candidates = append(candidates, candidate{plugin: p, pluginID: p.PluginID(), route: route})
			}
		}
	}
	slices.SortFunc(candidates, func(a, b candidate) int {
		if order := cmp.Compare(a.route.GetPriority(), b.route.GetPriority()); order != 0 {
			return order
		}
		if order := cmp.Compare(a.pluginID, b.pluginID); order != 0 {
			return order
		}
		return cmp.Compare(a.route.GetRouteId(), b.route.GetRouteId())
	})
	return candidates
}

func (r *Router) dispatch(ev *pb.Event) {
	for _, candidate := range r.candidates(ev) {
		if r.ctx.Err() != nil {
			return
		}
		if !candidate.plugin.Activated() {
			continue
		}
		trace := make([]byte, traceIDBytes)
		rand.Read(trace)
		routeID := candidate.route.GetRouteId()
		log := r.log.With("plugin_id", candidate.pluginID, "route_id", routeID, "trace_id", hex.EncodeToString(trace))
		request := &pb.HandlerRequest{RouteId: routeID, Event: ev, Context: &pb.InvocationContext{Actor: ev.GetActor(), DefaultTarget: ev.GetTarget()}}
		encoded, err := proto.Marshal(request)
		if err != nil {
			log.Warn("encode event handler request failed", "error", err)
			continue
		}
		response, err := candidate.plugin.Call(r.ctx, wire.Request{
			Target: routeID, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: handlerSchemaVersion, Bytes: encoded},
			TraceID: trace, Attach: &Invocation{PluginID: candidate.pluginID, Event: ev, DefaultTarget: ev.GetTarget()},
		})
		if err != nil {
			log.Warn("event handler call failed", "error", err)
			continue
		}
		if response.Err != nil {
			log.Warn("event handler returned an error", "code", response.Err.GetCode(), "error", response.Err.GetSafeMessage())
			continue
		}
		outcome, err := decodeOutcome(response.Payload)
		if err != nil {
			log.Warn("decode event handler outcome failed", "error", err)
			continue
		}
		if outcome.GetPropagation() == pb.Propagation_PROPAGATION_STOP {
			return
		}
	}
}

func decodeOutcome(payload wire.Payload) (*pb.HandlerOutcome, error) {
	outcome := &pb.HandlerOutcome{}
	if payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || payload.SchemaID != validate.SchemaID(outcome) || payload.SchemaVersion != handlerSchemaVersion {
		return nil, errors.New("expected HandlerOutcome protobuf schema version 1")
	}
	if err := proto.Unmarshal(payload.Bytes, outcome); err != nil {
		return nil, fmt.Errorf("decode HandlerOutcome: %w", err)
	}
	return outcome, nil
}
