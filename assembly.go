package alonabot

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/contracts/i18n"
	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	coreadapter "github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter/remote"
	"github.com/AlonaBot/AlonaBot/internal/core/broker"
	"github.com/AlonaBot/AlonaBot/internal/core/bus"
	"github.com/AlonaBot/AlonaBot/internal/core/buttons"
	"github.com/AlonaBot/AlonaBot/internal/core/conversation"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/httpproxy"
	"github.com/AlonaBot/AlonaBot/internal/core/media"
	"github.com/AlonaBot/AlonaBot/internal/core/message"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/process"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/scheduler"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
	"github.com/AlonaBot/AlonaBot/internal/core/tool"
	"github.com/AlonaBot/AlonaBot/internal/inproc"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
)

const (
	devSource            = "dev"
	devGenerationPrefix  = "dev:"
	launchRandomBytes    = 32 // Hello 的认证 Secret 与本次开发摘要均为 256 bit 随机值。
	defaultLocale        = "en-US"
	emptyConfig          = "{}"
	initialConfigVersion = 1
	packageDigestPrefix  = 12 // 与 packages.Candidate.MoveTo 的不可变目录命名一致。
)

type component struct {
	root      sdk.Root
	adapter   adapter.Adapter
	role      state.Role
	manifest  *manifest.Manifest
	resources sdk.Generated
}

func validateComponents(components []component, installs []state.Install) error {
	ids := make(map[string]bool)
	adapters := make(map[string]bool)
	for i := range components {
		entry := &components[i]
		var object any = entry.root
		if entry.role == state.RoleAdapter {
			object = entry.adapter
		}
		if object == nil || (reflect.ValueOf(object).Kind() == reflect.Pointer && reflect.ValueOf(object).IsNil()) {
			return fmt.Errorf("component %T is nil", object)
		}
		generated, ok := sdk.GeneratedFor(object)
		if !ok || len(generated.Manifest) == 0 {
			return fmt.Errorf("component %T has no generated project manifest", object)
		}
		parsed, err := manifest.ParseProject(generated.Manifest)
		if err != nil {
			return fmt.Errorf("component %T: %w", object, err)
		}
		m := parsed.Manifest
		for _, warning := range parsed.Warnings {
			slog.Warn("project manifest warning", "plugin_id", m.ID, "warning", warning)
		}
		if state.Role(m.Role) != entry.role {
			return fmt.Errorf("component %s has role %s, expected %s", m.ID, m.Role, entry.role)
		}
		if ids[m.ID] {
			return fmt.Errorf("duplicate component ID %s", m.ID)
		}
		ids[m.ID] = true
		if m.AdapterID != "" {
			if adapters[m.AdapterID] {
				return fmt.Errorf("duplicate adapter_id %s", m.AdapterID)
			}
			adapters[m.AdapterID] = true
			_, problems := configschema.Parse(generated.AdapterConfigSchema)
			if len(problems) != 0 {
				return fmt.Errorf("adapter %s instance schema: %v", m.ID, problems)
			}
		}
		for _, installed := range installs {
			if installed.PluginID == m.ID {
				if installed.Source != devSource || installed.Role != entry.role || installed.AdapterID != m.AdapterID {
					return fmt.Errorf("component %s conflicts with installed component", m.ID)
				}
			} else if m.AdapterID != "" && installed.AdapterID == m.AdapterID {
				return fmt.Errorf("adapter_id %s conflicts with installed component %s", m.AdapterID, installed.PluginID)
			}
		}
		entry.manifest, entry.resources = m, generated
	}
	return nil
}

type assembly struct {
	adapters  *coreadapter.Registry
	router    *router.Router
	mux       *router.Mux
	media     *media.Store
	buttons   *buttons.Service
	bus       *bus.Bus
	scheduler *scheduler.Scheduler
	http      *httpproxy.Server
	global    *wire.Budget
	store     *state.Store
}

func assemble(store *state.Store, dir string, v tune.Values) (*assembly, error) {
	a := &assembly{store: store, adapters: coreadapter.NewRegistry(), mux: router.NewMux(), global: wire.NewBudget(v.CoreMaxPayload)}
	var err error
	a.media, err = media.Open(filepath.Join(dir, "media"), media.NewLimits(v), time.Now)
	if err != nil {
		return nil, err
	}
	p := policy.New(store)
	if err := p.LoadCommandNames(context.Background()); err != nil {
		return nil, fmt.Errorf("load command names: %w", err)
	}
	a.buttons = buttons.New(store, time.Now, v.ButtonTokenTTL, func(id string) (string, []*pb.RouteDescriptor, bool) {
		plugin, ok := a.router.Plugin(id)
		if !ok || !plugin.Activated() {
			return "", nil, false
		}
		caller, ok := a.router.Caller(id)
		return caller.GenerationID, plugin.Routes(), ok
	})
	messages := message.New(a.adapters, p, a.media.LeaseForSend, a.buttons.Issue, v.RequestTimeout)
	a.router = router.New(router.Options{Concurrency: v.DispatchConcurrency, QueueEvents: v.DispatchQueueEvents, QueueBytes: v.DispatchQueueBytes,
		Global: a.global, Adapters: a.adapters, Send: messages.Send, Prepare: a.media.PrepareIncoming, Buttons: a.buttons.Resolve, Policy: p, CommandState: p.CommandState, Prefixes: nil, CommandNames: p.CommandNames})
	p.SetCommandsChanged(a.router.CommandsChanged)
	a.mux.Handle(wire.ServiceSend, messages.Send)
	a.mux.Handle(wire.ServiceEdit, messages.Edit)
	a.mux.Handle(wire.ServiceDelete, messages.Delete)
	a.mux.Handle(wire.ServiceAction, messages.Action)
	a.media.AddServices(a.mux)
	a.media.AddReadServices(a.mux, a.adapters, a.router, store.AdapterInstanceAllowed)
	conversation.New(store).Register(a.mux)
	tool.New(store, a.router).Register(a.mux)
	a.bus = bus.New(bus.Options{QueueEvents: v.BusQueueEvents, QueueBytes: v.BusQueueBytes})
	a.bus.Register(a.mux)
	a.mux.Forward(broker.New(a.router, v.RequestTimeout).Call)
	a.scheduler = scheduler.New(scheduler.Options{JobTimeout: v.JobTimeout})
	a.http = httpproxy.New(nil)
	return a, nil
}

// runningComponent 保存一次外部进程或同进程 Session，退出观察者结束后才由 Core 替换。
type runningComponent struct {
	id       string
	install  state.Install // 独立进程恢复时复用安装身份与已提交代次。
	caller   router.Caller
	instance *host.Instance
	process  *supervisor.Running // 仅独立进程设置。
	remote   *remote.Process
	listener net.Listener
	conn     *wire.Conn
	finished chan struct{} // SDK goroutine 返回后关闭，并发布 runErr。
	runErr   error
}

func (r *runningComponent) exited() <-chan struct{} {
	if r.process != nil {
		return r.process.Process.Exited()
	}
	return r.finished
}

// WaitTree 与 KillTree 让 host 的同一停止状态机等待 goroutine；总截止只报告超时。
func (r *runningComponent) Exited() <-chan struct{} { return r.finished }
func (r *runningComponent) WaitTree(deadline time.Time) (bool, error) {
	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()
	select {
	case <-r.finished:
		return true, nil
	case <-timer.C:
		return false, nil
	}
}
func (r *runningComponent) KillTree() error {
	select {
	case <-r.finished:
		return nil
	default:
		return fmt.Errorf("in-process component %s did not exit within STOP_TOTAL", r.id)
	}
}

func (a *assembly) handler(r *runningComponent, install state.Install) wire.Handler {
	if install.Role != state.RoleAdapter {
		return a.mux.For(r.caller)
	}
	r.remote = remote.New(remote.Options{AdapterID: install.AdapterID, PluginID: install.PluginID, Store: a.store, Registry: a.adapters, Sink: a.router, RegisterIncoming: a.media.RegisterIncoming,
		StartCommandSync: a.router.StartCommandSync, StopCommandSync: a.router.StopCommandSync})
	// Adapter 私有服务按 Session 建独立 Mux，多个 Adapter 不会重复登记固定服务名。
	local := router.NewMux()
	r.remote.Services(local)
	private := local.For(r.caller)
	shared := a.mux.For(r.caller)
	return func(ctx context.Context, in *wire.Inbound) wire.Response {
		switch in.Target {
		case wire.ServiceCoreAdapterDeliver, wire.ServiceCoreAdapterCursor, wire.ServiceCoreAdapterStatus:
			return private(ctx, in)
		default:
			return shared(ctx, in)
		}
	}
}

func (a *assembly) attach(ctx context.Context, r *runningComponent) error {
	plugin := router.FromInstance(r.instance)
	a.router.Attach(plugin, r.caller)
	a.bus.Attach(plugin)
	if r.process == nil {
		if r.instance.HTTPListener != nil {
			a.http.Bind(r.id, r.instance.GenerationID, r.instance.HTTPListener, r.instance.HTTPProxyToken, r.instance.Routes)
		}
		a.scheduler.Attach(plugin, time.Now())
	}
	if r.remote != nil {
		return r.remote.Bind(ctx, r.instance, r.caller)
	}
	return nil
}

func (c *Core) startDevelopment(ctx context.Context, dir string, entry *component) error {
	id := entry.manifest.ID
	digest := make([]byte, launchRandomBytes)
	rand.Read(digest)
	install, err := c.store.GetInstall(ctx, id)
	if errors.Is(err, state.ErrNotFound) {
		install = state.Install{Role: entry.role, PluginID: id, AdapterID: entry.manifest.AdapterID, Source: devSource, BuildDigest: digest}
		err = c.store.AddInstall(ctx, install)
	} else if err == nil {
		err = c.store.SetDevBuildDigest(ctx, id, digest)
	}
	if err != nil {
		return err
	}
	generation := devGenerationPrefix + id
	active, err := c.store.Active(ctx, id)
	if err != nil {
		return err
	}
	if active != "" && active != generation {
		return fmt.Errorf("dev component %s has unexpected active generation %s", id, active)
	}
	snapshots, err := c.store.PluginGenerationConfigs(ctx, id)
	if err != nil {
		return err
	}
	if _, exists := snapshots[generation]; !exists {
		if err := c.store.AddGeneration(ctx, id, generation, emptyConfig, initialConfigVersion); err != nil {
			return err
		}
	}
	pluginDir := filepath.Join(dir, "plugins", id)
	if err := os.MkdirAll(pluginDir, 0o700); err != nil {
		return err
	}
	listener, err := wire.ListenLoopback()
	if err != nil {
		return err
	}
	running := &runningComponent{id: id, caller: router.NewCaller(id, generation), listener: listener, finished: make(chan struct{})}
	c.components = append(c.components, running)
	l := &inproc.Launch{PluginID: id, Secret: make([]byte, launchRandomBytes), Addr: listener.Addr().String(), BuildDigest: digest, RuntimeLockPath: filepath.Join(pluginDir, wire.RuntimeLockFile)}
	rand.Read(l.Secret)
	go func() {
		if entry.role == state.RoleAdapter {
			running.runErr = sdk.RunAdapterInProcess(entry.adapter, l)
		} else {
			running.runErr = sdk.RunInProcess(entry.root, l)
		}
		close(running.finished)
		_ = listener.Close() // SDK 若在拨号前失败，解除 Accept 的等待。
	}()
	stopAccept := context.AfterFunc(ctx, func() { _ = listener.Close() })
	defer stopAccept()
	if err := listener.(*net.TCPListener).SetDeadline(time.Now().Add(c.values.PhaseHello)); err != nil {
		return err
	}
	conn, err := listener.Accept()
	if err != nil {
		return fmt.Errorf("accept %s: %w", id, err)
	}
	running.conn = wire.NewStreamConn(conn)
	limits := host.NewLimits(c.values)
	boot := bootstrap(generation, initialConfigVersion)
	messages, problems := i18n.Parse(entry.resources.I18n)
	if len(problems) != 0 {
		return fmt.Errorf("component %s messages: %v", id, problems)
	}
	prep, err := host.Prepare(ctx, running.conn, host.Expect{Role: entry.role, AdapterID: entry.manifest.AdapterID, PluginID: id, Secret: l.Secret, BuildDigest: digest,
		Offer: validate.Offer{Major: validate.ProtocolMajor, Minor: validate.ProtocolMinor}}, boot, limits)
	if err != nil {
		return err
	}
	activation := host.ActivateOptions{PluginID: id, GenerationID: generation, Handler: c.assembly.handler(running, install), Global: c.assembly.global, Limits: limits, Messages: messages}
	if active == "" {
		activation.Commit = func(ctx context.Context) error { return c.store.CommitActive(ctx, id, "", generation) }
	}
	running.instance, err = host.Activate(ctx, running.conn, prep, activation)
	if err != nil {
		return err
	}
	return c.assembly.attach(ctx, running)
}

func bootstrap(generation string, version uint32) *pb.Bootstrap {
	return &pb.Bootstrap{GenerationId: generation, CandidateConfigJson: emptyConfig, TargetConfigVersion: version, DefaultLocale: defaultLocale,
		DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_NONE}, LogLevel: pb.LogLevel_LOG_LEVEL_INFO}
}

func (a *assembly) detach(r *runningComponent) {
	a.router.Detach(r.id)
	a.bus.Detach(r.id)
	a.scheduler.Detach(r.id)
	if r.instance != nil {
		a.http.Unbind(r.id, r.instance.GenerationID)
	}
	if r.remote != nil {
		r.remote.Unbind()
	}
}

func (c *Core) stopComponent(r *runningComponent, t0 time.Time, timing host.StopTiming) error {
	a := c.assembly
	frozen := func(deadline time.Time) { a.scheduler.Drain(r.id); a.http.Drain(r.id, deadline); a.bus.Detach(r.id) }
	detach := func() { a.detach(r) }
	var err error
	if r.process != nil {
		// supervisor 处理进程树、HTTP 与 Scheduler；其余映射由装配撤销。
		a.bus.Detach(r.id)
		result := r.process.Stop(t0, pb.StopReason_STOP_REASON_CORE_SHUTDOWN)
		err = result.Err
		detach()
	} else if r.instance != nil {
		result := r.instance.Stop(t0, pb.StopReason_STOP_REASON_CORE_SHUTDOWN, r, timing, host.StopHooks{Frozen: frozen, Stopping: detach})
		err = result.Err
	} else {
		if r.conn != nil {
			_ = r.conn.Close()
		}
		_ = r.listener.Close()
		empty, waitErr := r.WaitTree(t0.Add(timing.Total))
		err = waitErr
		if !empty {
			err = errors.Join(err, r.KillTree())
		}
		detach()
	}
	if r.listener != nil {
		_ = r.listener.Close()
	}
	a.media.ReleaseSession(r.caller.SessionKey)
	select {
	case <-r.exited():
		if r.process == nil {
			err = errors.Join(err, r.runErr)
		}
		a.scheduler.Released(r.id)
	default:
	}
	if err != nil {
		slog.Error("component shutdown failed", "plugin_id", r.id, "error", err)
		return fmt.Errorf("component %s: %w", r.id, err)
	}
	return nil
}

func (c *Core) startInstalled(ctx context.Context, dir string, install state.Install) (*runningComponent, error) {
	generation, err := c.store.Active(ctx, install.PluginID)
	if err != nil {
		return nil, err
	}
	if generation == "" {
		return nil, fmt.Errorf("enabled component %s has no committed generation", install.PluginID)
	}
	// MoveTo 已将版本与归档摘要前缀写入目录名；摘要绑定当前安装记录。
	digest := hex.EncodeToString(install.BuildDigest)
	if len(digest) < packageDigestPrefix {
		return nil, fmt.Errorf("component %s has invalid installed build digest", install.PluginID)
	}
	parent := filepath.Join(dir, "packages", install.PluginID)
	entries, err := os.ReadDir(parent)
	if err != nil {
		return nil, err
	}
	var paths []string
	suffix := "-" + digest[:packageDigestPrefix]
	for _, entry := range entries {
		version, matches := strings.CutSuffix(entry.Name(), suffix)
		if entry.IsDir() && matches && manifest.ValidVersion(version) {
			paths = append(paths, filepath.Join(parent, entry.Name()))
		}
	}
	if len(paths) != 1 {
		return nil, fmt.Errorf("component %s: expected one installed package for build digest, found %d", install.PluginID, len(paths))
	}
	packageDir := paths[0]
	raw, err := os.ReadFile(filepath.Join(packageDir, "plugin.yaml"))
	if err != nil {
		return nil, err
	}
	parsed, err := manifest.Parse(raw)
	if err != nil {
		return nil, err
	}
	m := parsed.Manifest
	if m.ID != install.PluginID || state.Role(m.Role) != install.Role || m.AdapterID != install.AdapterID || m.Entrypoint != install.Entrypoint {
		return nil, fmt.Errorf("installed manifest identity differs from record for %s", install.PluginID)
	}
	for _, warning := range parsed.Warnings {
		slog.Warn("installed manifest warning", "plugin_id", m.ID, "warning", warning)
	}
	executable, err := process.Entrypoint(packageDir, install.Entrypoint)
	if err != nil {
		return nil, err
	}
	resources, err := packageResources(packageDir)
	if err != nil {
		return nil, err
	}
	pluginDir := filepath.Join(dir, "plugins", install.PluginID)
	if err := os.MkdirAll(pluginDir, 0o700); err != nil {
		return nil, err
	}
	running := &runningComponent{id: install.PluginID, install: install, caller: router.NewCaller(install.PluginID, generation)}
	boot := bootstrap(generation, *m.ConfigVersion)
	boot.DatabaseDescriptor.ActivationReason = pb.ActivationReason_ACTIVATION_REASON_RESUME
	switch m.MigrationMode {
	case manifest.MigrationSelf:
		boot.DatabaseDescriptor.MigrationMode = pb.MigrationMode_MIGRATION_MODE_SELF
	case manifest.MigrationVersioned:
		boot.DatabaseDescriptor.MigrationMode = pb.MigrationMode_MIGRATION_MODE_VERSIONED
	}
	if install.Role == state.RolePlugin && (m.PrepareEffects == manifest.EffectsManagedDB || m.MigrationMode == manifest.MigrationSelf || m.MigrationMode == manifest.MigrationVersioned) {
		boot.DatabaseDescriptor.Backend = pb.DatabaseBackend_DATABASE_BACKEND_SQLITE
	}
	running.process, err = supervisor.Launch(ctx, c.store, supervisor.Spec{Role: install.Role, AdapterID: install.AdapterID, PluginID: install.PluginID, PluginDir: pluginDir,
		Executable: executable, BuildDigest: install.BuildDigest, GenerationID: generation, Bootstrap: boot, ConfigSchema: resources.ConfigSchema, I18n: resources.I18n,
		DefaultLocale: defaultLocale, Handler: c.assembly.handler(running, install), Global: c.assembly.global, Limits: host.NewLimits(c.values), Stop: host.NewStopTiming(c.values),
		Log: c.output.Plugin(install.PluginID, generation), Scheduler: c.assembly.scheduler, HTTP: c.assembly.http, MemorySoftLimit: c.values.PluginMemorySoftLimit})
	if err != nil {
		return nil, err
	}
	running.instance = running.process.Instance
	return running, c.assembly.attach(ctx, running)
}

func packageResources(dir string) (sdk.Generated, error) {
	g := sdk.Generated{I18n: make(map[string][]byte)}
	raw, err := os.ReadFile(filepath.Join(dir, "config_schema.json"))
	if err == nil {
		schema, problems := configschema.Parse(raw)
		if len(problems) != 0 {
			return g, fmt.Errorf("installed config schema: %v", problems)
		}
		g.ConfigSchema = schema.Canonical()
	} else if !errors.Is(err, os.ErrNotExist) {
		return g, err
	}
	files, err := os.ReadDir(filepath.Join(dir, "i18n"))
	if errors.Is(err, os.ErrNotExist) {
		return g, nil
	}
	if err != nil {
		return g, err
	}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, "i18n", file.Name()))
		if err != nil {
			return g, err
		}
		g.I18n[file.Name()] = raw
	}
	return g, nil
}
