// Package router 提供 Core 的可信调用身份、固定服务分流与标准事件派发。
//
// [Caller] 绑定一次插件 Session；[Invocation] 保存在 Core 发起的父请求附件中。
// [Mux] 将插件请求交给固定服务或 Service Broker，身份由装配时的 For 参数决定（CONTRACTS §1.2）。
// [Router] 按优先级、插件与路由 ID 顺序调用已激活的 [Plugin]，遵循传播停止标记。
package router

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

const sessionKeyBytes = 16

// Caller 保存 Core 认证的插件身份与本次 Session 标识。
type Caller struct {
	PluginID     string
	GenerationID string
	SessionKey   string // 每次启动唯一，资源所有权按本次 Session 绑定
}

// NewCaller 为插件的本次启动生成 16 字节随机值的十六进制 SessionKey。
func NewCaller(pluginID, generationID string) Caller {
	key := make([]byte, sessionKeyBytes)
	rand.Read(key)
	return Caller{PluginID: pluginID, GenerationID: generationID, SessionKey: hex.EncodeToString(key)}
}

// Invocation 保存派发时的可信事件与默认目标，通过 Request.Attach 绑定父调用。
type Invocation struct {
	OnContinuation func() func()   // 登记时保留分发关联，返回的函数在取消或延续完成后释放。
	OnWait         func()          // 实际登记等待后通知。
	OnWaitMatch    func(*pb.Event) // 交付等待结果前切换到命中事件的关联。
	dispatch       *dispatchCall
	PluginID       string
	Event          *pb.Event
	DefaultTarget  *pb.Target
	Chain          []string // 祖先插件链，最早者在前，不含当前插件。
	Scopes         []string // Core 授予本次调用的权限范围。
	TraceID        []byte   // Core 为本次调用记录的 trace，只由 Core 设置。
	SendForbidden  bool     // 出站 Hook 调用为 true；只由 Core 设置
	RouteID        string   // 被调用的 route_id。
	Continuation   bool     // 本调用为延续调用。
	candidate      candidate
	request        *pb.HandlerRequest // 保留原参数与已为所属 Session 准备的事件。
}

// ServiceFunc 接收固定 Caller 与经过 Session 校验的插件请求。
type ServiceFunc func(ctx context.Context, c Caller, in *wire.Inbound) wire.Response

// Mux 并发安全地保存固定服务注册；服务函数在锁外执行。
type Mux struct {
	mu       sync.RWMutex // 保护 services 与 forward，不跨服务调用持锁
	services map[string]ServiceFunc
	forward  ServiceFunc
}

// NewMux 创建空的固定服务分流器。
func NewMux() *Mux { return &Mux{services: make(map[string]ServiceFunc)} }

// Handle 注册固定服务；同名重复注册时 panic。
func (m *Mux) Handle(name string, fn ServiceFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.services[name]; exists {
		panic("duplicate Core service: " + name)
	}
	m.services[name] = fn
}

// Forward 设置跨插件 Service 的转交入口；固定服务始终优先。
func (m *Mux) Forward(fn ServiceFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.forward = fn
}

// For 将调用身份固定到一个 Session；含 / 的未知目标交给 Forward，其余返回 NOT_FOUND。
func (m *Mux) For(c Caller) wire.Handler {
	return func(ctx context.Context, in *wire.Inbound) wire.Response {
		m.mu.RLock()
		fn, ok := m.services[in.Target]
		if !ok && strings.Contains(in.Target, "/") {
			fn = m.forward
		}
		m.mu.RUnlock()
		if fn == nil {
			return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, SafeMessage: "Core service not found"}}
		}
		return fn(ctx, c, in)
	}
}
