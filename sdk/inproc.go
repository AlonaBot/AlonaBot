package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/inproc"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/internal/dbhook"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
)

// runInProcess 使用 Core 显式提供的参数运行插件；环境变量、工作目录与默认 Logger 保持进程级含义。
func runInProcess(root *Plugin, l *inproc.Launch) error {
	if err := requireResources(root); err != nil {
		return err
	}
	root.freeze()
	launch, err := inProcessLaunch(l)
	if err != nil {
		return err
	}
	return run(root, launch)
}

// runAdapterInProcess 使用显式参数运行 Adapter，复用独立进程入口的握手、激活与停止流程。
func runAdapterInProcess(a adapter.Adapter, l *inproc.Launch) error {
	if err := requireResources(a); err != nil {
		return err
	}
	launch, err := inProcessLaunch(l)
	if err != nil {
		return err
	}
	return runAdapter(a, launch)
}

func inProcessLaunch(l *inproc.Launch) (launch, error) {
	if l == nil {
		return launch{}, errors.New("in-process launch is nil")
	}
	return launch{pluginID: l.PluginID, transport: wire.TransportTCP, addr: l.Addr, secret: l.Secret, digest: l.BuildDigest, runtimeLockPath: l.RuntimeLockPath, log: l.Log, inProcess: true}, nil
}

type componentContextKey struct{}

// componentSession 持有一次启动的全部 SDK 状态；只在独立进程入口发布默认归属。
// callbacks 保留仍在执行的生命周期回调，同进程入口等它们退出后才释放 runtime.lock。
type componentSession struct {
	plugin       *Plugin // 准备期绑定，HTTP 与生命周期共享同一覆盖值所有者。
	ctx          context.Context
	cancel       context.CancelFunc // 准备期断链或 Session 关闭时取消阶段回调。
	database     *dbhook.Session
	active       atomic.Pointer[activation]
	translations atomic.Pointer[translationSession]
	level        slog.LevelVar
	log          slog.Handler // Core 提供的日志目标；nil 时使用默认 Handler。
	source       []slog.Attr
	inProcess    bool
	callbacks    sync.WaitGroup
	callbackMu   sync.Mutex // 与关闭同步，防止看门狗超时后又开始一个用户回调。
	closing      bool
}

var defaultSession atomic.Pointer[componentSession]

func init() {
	inproc.Plugin = func(plugin any, l *inproc.Launch) error { return runInProcess(plugin.(*Plugin), l) }
	inproc.Adapter = func(a any, l *inproc.Launch) error { return runAdapterInProcess(a.(adapter.Adapter), l) }
	dbhook.Logger = componentLogger
	registry.Parent = func(ctx context.Context) (*wire.Inbound, bool) {
		c, ok := ctx.Value(serviceContextKey{}).(*Context)
		if !ok {
			return nil, false
		}
		return c.inbound, c.active.Load()
	}
	registry.Action = func(ctx context.Context, to, name string, version uint32, params, options any) (json.RawMessage, error) {
		return callAction(ctx, nil, to, name, version, params, options.([]SendOption)...)
	}
	registry.Session = currentSession
	registry.Call = callService
	registry.Decode = decodeInvocation
	registry.Logger = componentLogger
	registry.Runtime = func(ctx context.Context) (string, string, map[string][]byte) {
		l := componentFor(ctx).active.Load().l
		return l.pluginID, l.entryPackage, l.serviceSchemas
	}
}

func newComponentSession(inProcess bool) *componentSession {
	s := &componentSession{inProcess: inProcess, database: dbhook.New(!inProcess)}
	s.ctx, s.cancel = context.WithCancel(s.context(context.Background()))
	if !inProcess {
		defaultSession.Store(s)
	}
	return s
}

func (s *componentSession) context(ctx context.Context) context.Context {
	return s.database.Context(context.WithValue(ctx, componentContextKey{}, s))
}

func componentFor(ctx context.Context) *componentSession {
	if s, ok := ctx.Value(componentContextKey{}).(*componentSession); ok {
		return s
	}
	return defaultSession.Load()
}

func (s *componentSession) close() error {
	s.cancel()
	s.callbackMu.Lock()
	s.closing = true
	s.callbackMu.Unlock()
	if s.inProcess {
		// 看门狗结束协议等待后，用户回调可能仍持有资源；锁保留到回调实际退出。
		s.callbacks.Wait()
	}
	defaultSession.CompareAndSwap(s, nil)
	return s.database.Close()
}

func (s *componentSession) beginCallback() bool {
	s.callbackMu.Lock()
	defer s.callbackMu.Unlock()
	if s.closing {
		return false
	}
	s.callbacks.Add(1)
	return true
}

func (s *componentSession) handler(next wire.Handler) wire.Handler {
	return func(ctx context.Context, in *wire.Inbound) wire.Response {
		if !s.beginCallback() {
			return dispatchError(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "plugin session has stopped")
		}
		defer s.callbacks.Done()
		if in.Target == wire.ServiceLogSetLevel {
			return s.changeLogLevel(in)
		}
		return next(s.context(ctx), in)
	}
}

func componentLogger(ctx context.Context) *slog.Logger {
	s := componentFor(ctx)
	if s == nil || !s.inProcess {
		return slog.Default()
	}
	handler := s.log
	if handler == nil {
		handler = slog.Default().Handler()
	}
	return slog.New(&componentLogHandler{Handler: handler.WithAttrs(s.source), session: s})
}

type componentLogHandler struct {
	slog.Handler
	session *componentSession
}

func (h *componentLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.session.level.Level()
}

func (h *componentLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &componentLogHandler{Handler: h.Handler.WithAttrs(attrs), session: h.session}
}

func (h *componentLogHandler) WithGroup(name string) slog.Handler {
	return &componentLogHandler{Handler: h.Handler.WithGroup(name), session: h.session}
}
