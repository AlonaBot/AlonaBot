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

	"github.com/AlonaBot/AlonaBot/contracts/i18n"
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

// Plugin 提供派发所需的插件身份、激活状态、静态路由、文案与 Session 调用。
type Plugin interface {
	PluginID() string
	Activated() bool
	Routes() []*pb.RouteDescriptor
	// Messages 返回当前包的只读文案目录；无资源时为 nil。
	Messages() *i18n.Catalog
	// DefaultLocale 返回当前 Session 的 Bootstrap 默认语言快照。
	DefaultLocale() string
	Call(ctx context.Context, req wire.Request) (wire.Response, error)
	// SendEvent 等待 EVENT 帧写出，不等待订阅回调。
	SendEvent(ctx context.Context, target string, ev *pb.BusEvent) error
}

type instancePlugin struct{ instance *host.Instance }

type pluginSession struct {
	plugin Plugin
	caller Caller // 与该 Session 的 Mux.For 绑定相同身份
}

type inFlightKey struct {
	caller  Caller
	eventID string
}

// FromInstance 把真实宿主实例接入路由器，激活状态从宿主实时读取。
func FromInstance(in *host.Instance) Plugin { return &instancePlugin{instance: in} }

func (p *instancePlugin) PluginID() string              { return p.instance.PluginID }
func (p *instancePlugin) Activated() bool               { return p.instance.State() == host.StateActivated }
func (p *instancePlugin) Routes() []*pb.RouteDescriptor { return p.instance.Routes }
func (p *instancePlugin) Messages() *i18n.Catalog       { return p.instance.Messages }
func (p *instancePlugin) DefaultLocale() string         { return p.instance.DefaultLocale }
func (p *instancePlugin) Call(ctx context.Context, req wire.Request) (wire.Response, error) {
	return p.instance.Session.Call(ctx, req)
}

func (p *instancePlugin) SendEvent(ctx context.Context, target string, ev *pb.BusEvent) error {
	return p.instance.Session.SendEvent(ctx, target, ev)
}

// CommandPolicy 提供命令准入与 Core 管理员身份判定。
type CommandPolicy interface {
	CheckCommand(ctx context.Context, pluginID string, route *pb.RouteDescriptor, actor *pb.Actor, target *pb.Target) *pb.ErrorInfo
	IsAdmin(ctx context.Context, actor *pb.Actor) (bool, error)
}

// Router 并发执行有界的事件链；每条链内按稳定次序调用候选。
// mu 串行化接纳、出队和注册；链调用插件期间不持锁。
type Router struct {
	adapters    *adapter.Registry
	send        ServiceFunc
	prepare     func(*pb.Event, Caller) (*pb.Event, error)
	prefixes    func(adapterInstanceID string) []string
	policy      CommandPolicy
	matches     map[string][]commandRoute // 在 mu 下替换，路由快照发布后只读
	queue       []queuedEvent
	concurrency int
	queueEvents int
	queueBytes  int64 // 等待队列上限，单位为编码后的字节数
	global      *wire.Budget
	queuedBytes int64         // 等待中的事件占用，由 mu 保护
	stats       Stats         // 计数及运行状态，由 mu 保护
	drained     chan struct{} // 关闭接纳且全部运行链退出时关闭
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex // 保护注册、在途调用、命令索引与分发状态
	plugins     map[string]*pluginSession
	inFlight    map[inFlightKey]struct{} // 仅覆盖 Plugin.Call；排队与 prepare 期间不登记
	commands    *CommandIndex            // 在 mu 下发布；已发布快照只读
	started     bool
	closed      bool
}

// New 按显式的并发和队列限额创建路由器；限额由装配处从 tune.Values 传入。
func New(opt Options) *Router {
	if opt.Prefixes == nil {
		opt.Prefixes = func(string) []string { return []string{"/"} }
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Router{
		adapters: opt.Adapters, send: opt.Send, prepare: opt.Prepare, prefixes: opt.Prefixes, policy: opt.Policy,
		concurrency: opt.Concurrency, queueEvents: opt.QueueEvents, queueBytes: opt.QueueBytes, global: opt.Global,
		ctx: ctx, cancel: cancel, drained: make(chan struct{}),
		matches: make(map[string][]commandRoute), plugins: make(map[string]*pluginSession),
		inFlight: make(map[inFlightKey]struct{}), commands: &CommandIndex{},
	}
}

// Attach 注册一个接收派发的插件并重建命令索引；相同插件 ID 替换为新实例。
// c 必须与该 Session 的 Mux.For 使用同一个 Caller。
func (r *Router) Attach(p Plugin, c Caller) {
	id := p.PluginID()
	r.mu.Lock()
	defer r.mu.Unlock()
	session := &pluginSession{plugin: p, caller: c}
	r.plugins[id] = session
	r.cacheCommands(session)
	r.rebuildCommands()
}

// Plugin 返回最近一次 Attach 的插件实例及查找结果；激活状态由调用方即时检查。
func (r *Router) Plugin(pluginID string) (Plugin, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	session, ok := r.plugins[pluginID]
	if !ok {
		return nil, false
	}
	return session.plugin, true
}

// Caller 返回当前已激活绑定的可信身份；替换、移除或退出后的身份不再返回。
func (r *Router) Caller(pluginID string) (Caller, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	session, ok := r.plugins[pluginID]
	if !ok || !session.plugin.Activated() {
		return Caller{}, false
	}
	return session.caller, true
}

// HasInFlight 按 event_id 检查当前 Session 是否正在调用该事件的 Handler。
// 调用返回、Router 取消或插件绑定失效后，关联立即失效。
func (r *Router) HasInFlight(c Caller, ev *pb.Event) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	session, ok := r.plugins[c.PluginID]
	if !ok || session.caller != c || !session.plugin.Activated() || r.ctx.Err() != nil {
		return false
	}
	_, ok = r.inFlight[inFlightKey{caller: c, eventID: ev.GetEventId()}]
	return ok
}

// Detach 移除插件并重建命令索引，使其退出后续事件的候选集。
func (r *Router) Detach(pluginID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.plugins, pluginID)
	delete(r.matches, pluginID)
	r.rebuildCommands()
}

type candidate struct {
	*pluginSession
	pluginID string
	route    *pb.RouteDescriptor
}

func (r *Router) candidates(ev *pb.Event) []candidate {
	r.mu.Lock()
	plugins := make([]*pluginSession, 0, len(r.plugins))
	for _, p := range r.plugins {
		plugins = append(plugins, p)
	}
	r.mu.Unlock()
	var candidates []candidate
	for _, p := range plugins {
		if !p.plugin.Activated() {
			continue
		}
		for _, route := range p.plugin.Routes() {
			if route.GetKind() == pb.RouteKind_ROUTE_KIND_EVENT && slices.Contains(route.GetEvent().GetEventTypeFilter(), ev.GetEventType()) {
				candidates = append(candidates, candidate{pluginSession: p, pluginID: p.plugin.PluginID(), route: route})
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

func (r *Router) dispatch(event *trustedEvent) {
	ev := event.retained
	// 1. Core 权限在进入任一路由前统一计算，来源只提供平台权限及其展开结果。
	if actor := ev.GetActor(); actor != nil {
		permissions := []string{"user"}
		for _, permission := range actor.GetPermissions() {
			if permission != "user" && permission != "admin" {
				permissions = append(permissions, permission)
			}
		}
		if r.policy != nil {
			admin, err := r.policy.IsAdmin(r.ctx, actor)
			if err != nil {
				slog.Warn("read Core Admin status failed", "adapter_instance_id", actor.GetAdapterInstanceId(), "error", err)
			} else if admin {
				permissions = append(permissions, "admin")
			}
		}
		slices.Sort(permissions)
		actor.Permissions = slices.Compact(permissions)
	}
	ev = event.rebuild()
	// 2. 命令层的 STOP 终止整条分发链；消费但继续的命令仍进入事件层。
	if ev.GetEventType() == "message.created" && r.dispatchCommands(ev, event) {
		return
	}
	// 3. 事件路由独立判断权限，拒绝当前候选后继续后继路由。
	for _, candidate := range r.candidates(ev) {
		if r.ctx.Err() != nil {
			return
		}
		if !candidate.plugin.Activated() || !r.permitted(candidate, ev) {
			continue
		}
		trace := make([]byte, traceIDBytes)
		rand.Read(trace)
		routeID := candidate.route.GetRouteId()
		response, err := r.call(candidate, event, nil, trace)
		if err != nil {
			slog.With("plugin_id", candidate.pluginID, "route_id", routeID, "trace_id", hex.EncodeToString(trace)).Warn("event handler call failed", "error", err)
		}
		if response.Err != nil && err == nil {
			slog.With("plugin_id", candidate.pluginID, "route_id", routeID, "trace_id", hex.EncodeToString(trace)).Warn("event handler returned an error", "code", response.Err.GetCode(), "error", response.Err.GetSafeMessage())
		}
		// 业务错误仍可能携带有效 STOP；先读取传播结果，再决定是否继续。
		outcome, err := decodeOutcome(response.Payload)
		if err != nil {
			slog.With("plugin_id", candidate.pluginID, "route_id", routeID, "trace_id", hex.EncodeToString(trace)).Warn("decode event handler outcome failed", "error", err)
			continue
		}
		if outcome.GetPropagation() == pb.Propagation_PROPAGATION_STOP {
			return
		}
	}
}

func (r *Router) permitted(c candidate, ev *pb.Event) bool {
	if len(c.route.GetRequiredPermissions()) == 0 {
		return true
	}
	for _, required := range c.route.GetRequiredPermissions() {
		if slices.Contains(ev.GetActor().GetPermissions(), required) {
			return true
		}
	}
	slog.Debug("route permission denied", "plugin_id", c.pluginID, "route_id", c.route.GetRouteId())
	return false
}

// current 在持有 mu 时检查候选仍属于当前的激活绑定，隔离旧候选快照。
func (r *Router) current(c candidate) bool {
	return r.ctx.Err() == nil && r.plugins[c.pluginID] == c.pluginSession && c.plugin.Activated()
}

func (r *Router) call(c candidate, event *trustedEvent, bind func(*pb.Event) (*pb.CommandInvocation, error), trace []byte) (wire.Response, error) {
	// 1. 每次调用先重建可信字段，再在锁外为接收 Session 签发媒体。
	ev := event.rebuild()
	r.mu.Lock()
	current := r.current(c)
	r.mu.Unlock()
	if !current {
		return wire.Response{}, errors.New("plugin Session is no longer active")
	}
	prepared := ev
	if r.prepare != nil {
		var err error
		prepared, err = r.prepare(ev, c.caller)
		if err != nil {
			return wire.Response{}, fmt.Errorf("prepare handler event: %w", err)
		}
	}
	// 参数从接收方事件绑定，使媒体令牌与该 Session 的事件授权一致。
	var command *pb.CommandInvocation
	if bind != nil {
		var err error
		command, err = bind(prepared)
		if err != nil {
			return wire.Response{}, fmt.Errorf("%w: %w", errCommandArguments, err)
		}
	}
	request := &pb.HandlerRequest{RouteId: c.route.GetRouteId(), Event: prepared, Command: command, Context: &pb.InvocationContext{Actor: proto.CloneOf(event.retained.GetActor()), DefaultTarget: proto.CloneOf(event.retained.GetTarget())}}
	encoded, err := proto.Marshal(request)
	if err != nil {
		return wire.Response{}, fmt.Errorf("encode handler request: %w", err)
	}
	// 2. prepare 期间可能发生替换或移除；实际调用前再核对，并只为本次 Call 登记关联。
	key := inFlightKey{caller: c.caller, eventID: event.retained.GetEventId()}
	r.mu.Lock()
	if !r.current(c) {
		r.mu.Unlock()
		return wire.Response{}, errors.New("plugin Session is no longer active")
	}
	r.inFlight[key] = struct{}{}
	r.mu.Unlock()
	defer func() {
		r.mu.Lock()
		delete(r.inFlight, key)
		r.mu.Unlock()
	}()
	return c.plugin.Call(r.ctx, wire.Request{
		Target: request.RouteId, TraceID: trace,
		Attach:  &Invocation{PluginID: c.pluginID, Event: prepared, DefaultTarget: proto.CloneOf(event.retained.GetTarget())},
		Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: handlerSchemaVersion, Bytes: encoded},
	})
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
