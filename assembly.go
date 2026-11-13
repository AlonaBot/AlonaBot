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
	"sync"
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
	"github.com/AlonaBot/AlonaBot/internal/core/config"
	"github.com/AlonaBot/AlonaBot/internal/core/continuation"
	"github.com/AlonaBot/AlonaBot/internal/core/conversation"
	"github.com/AlonaBot/AlonaBot/internal/core/hook"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/httpproxy"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/media"
	"github.com/AlonaBot/AlonaBot/internal/core/message"
	"github.com/AlonaBot/AlonaBot/internal/core/metrics"
	coreoptions "github.com/AlonaBot/AlonaBot/internal/core/options"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/process"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/scheduler"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
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
	root      *sdk.Plugin
	adapter   adapter.Adapter
	role      state.Role
	manifest  *manifest.Manifest
	resources sdk.Generated
}

func validateComponents(components []component, installs []state.Install) error {
	ids := make(map[string]bool)
	adapters := make(map[string]bool)
	agent := false
	for i := range components {
		entry := &components[i]
		var object any = entry.root
		if entry.role == state.RoleAdapter {
			object = entry.adapter
		}
		if object == nil || (reflect.ValueOf(object).Kind() == reflect.Pointer && reflect.ValueOf(object).IsNil()) {
			return fmt.Errorf("component %T is nil", object)
		}
		var generated sdk.Generated
		if entry.role == state.RoleAdapter {
			generated, _ = sdk.GeneratedFor(entry.adapter)
		} else {
			generated = entry.root.Generated()
		}
		if len(generated.Manifest) == 0 {
			return fmt.Errorf("component %T has no generated project manifest", object)
		}
		parsed, err := manifest.ParseProject(generated.Manifest)
		if err != nil {
			return fmt.Errorf("component %T: %w", object, err)
		}
		m := parsed.Manifest
		if entry.role == state.RolePlugin && m.Role == string(state.RoleAgent) {
			entry.role = state.RoleAgent
		}
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
		if entry.role == state.RoleAgent {
			if agent {
				return fmt.Errorf("duplicate agent component %s", m.ID)
			}
			agent = true
		}
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
			if installed.Uninstalled {
				continue
			}
			if installed.PluginID == m.ID {
				if installed.Source != devSource || installed.Role != entry.role || installed.AdapterID != m.AdapterID {
					return fmt.Errorf("component %s conflicts with installed component", m.ID)
				}
			} else if m.AdapterID != "" && installed.AdapterID == m.AdapterID {
				return fmt.Errorf("adapter_id %s conflicts with installed component %s", m.AdapterID, installed.PluginID)
			} else if entry.role == state.RoleAgent && installed.Role == state.RoleAgent {
				return fmt.Errorf("agent conflicts with installed component %s", installed.PluginID)
			}
		}
		entry.manifest, entry.resources = m, generated
	}
	return nil
}

type assembly struct {
	adapters   *coreadapter.Registry
	router     *router.Router
	mux        *router.Mux
	media      *media.Store
	buttons    *buttons.Service
	bus        *bus.Bus
	scheduler  *scheduler.Scheduler
	http       *httpproxy.Server
	global     *wire.Budget
	metrics    *metrics.Counters
	store      *state.Store
	policy     *policy.Policy
	components sync.Map // 运行状态查询借用已激活的 host.Instance，退出时按同一指针撤销。
	options    *coreoptions.Service
	settings   *settings.File
	schemaMu   sync.RWMutex // 保护按 Session 发布的只读选项声明。
	schemas    map[string]componentOptionSchemas
}

type componentOptionSchemas struct {
	session string
	config  *configschema.Schema
	adapter *configschema.Schema
}

func assemble(store *state.Store, dir string, v tune.Values, file *settings.File) (*assembly, error) {
	a := &assembly{store: store, settings: file, adapters: coreadapter.NewRegistry(), mux: router.NewMux(), global: wire.NewBudget(v.CoreMaxPayload), metrics: &metrics.Counters{}}
	var err error
	a.media, err = media.Open(filepath.Join(dir, "media"), media.NewLimits(v), time.Now)
	if err != nil {
		return nil, err
	}
	p := policy.New(store)
	a.policy = p
	var admins []policy.Admin
	for _, admin := range a.settings.Settings().Admins {
		admins = append(admins, policy.Admin{AdapterID: admin.AdapterID, UserID: admin.UserID})
	}
	p.SetAdmins(admins)
	a.adapters.SetDefaults(a.settings.Settings().DefaultInstances())
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
	// Router、Hook Runner 与消息服务互相引用：Router 先建，Send 与 PreEvent 在调用时才读取之后赋值的变量。
	var messages *message.Service
	var hooks *hook.Runner
	// 延续调用的 Manager 先于 Router 创建，回调在调用时才读取 a.router。
	continuations := continuation.New(continuation.Options{Max: v.ContinuationMax, Timeout: v.ContinuationTimeout, Prepare: a.media.PrepareIncoming,
		Caller: func(id string) (router.Caller, bool) { return a.router.Caller(id) },
		Call: func(ctx context.Context, inv *router.Invocation, parent uint64, deadline time.Time) (wire.Response, error) {
			return a.router.CallContinuation(ctx, inv, parent, deadline)
		}})
	a.router = router.New(router.Options{Concurrency: v.DispatchConcurrency, QueueEvents: v.DispatchQueueEvents, QueueBytes: v.DispatchQueueBytes,
		Global: a.global, Adapters: a.adapters, Prepare: a.media.PrepareIncoming, Buttons: a.buttons.Resolve, Policy: p, CommandState: p.CommandState, Prefixes: a.settings.CommandPrefixes, CommandNames: p.CommandNames,
		Continuations: continuations, Metrics: a.metrics,
		Send: func(ctx context.Context, c router.Caller, in *wire.Inbound) wire.Response {
			return messages.Send(ctx, c, in)
		},
		PreEvent: func(ctx context.Context, ev *pb.Event) (*pb.EventPayload, bool, error) {
			return hooks.PreEvent(ctx, ev)
		}})
	hooks = hook.New(a.router, hook.Options{Prepare: a.media.PrepareIncoming, CheckEvent: a.checkAdapterEvent})
	messages = message.New(a.adapters, p, a.media.LeaseForSend, a.buttons.Issue, v.RequestTimeout, hooks)
	p.SetCommandsChanged(a.router.CommandsChanged)
	continuations.Register(a.mux)
	a.mux.Handle(wire.ServiceSend, messages.Send)
	a.mux.Handle(wire.ServiceEdit, messages.Edit)
	a.mux.Handle(wire.ServiceDelete, messages.Delete)
	a.mux.Handle(wire.ServiceAction, messages.Action)
	a.media.AddServices(a.mux)
	a.media.AddReadServices(a.mux, a.adapters, a.router, store.AdapterInstanceAllowed)
	conversation.New(store).Register(a.mux)
	tool.New(store, a.router).Register(a.mux)
	a.bus = bus.New(bus.Options{QueueEvents: v.BusQueueEvents, QueueBytes: v.BusQueueBytes, Metrics: a.metrics})
	a.bus.Register(a.mux)
	serviceBroker := broker.New(broker.Options{Router: a.router, RequestTimeout: v.RequestTimeout, MaxDepth: v.CallChainMaxDepth, Metrics: a.metrics})
	a.mux.Forward(serviceBroker.Call)
	a.schemas = make(map[string]componentOptionSchemas)
	a.options = coreoptions.New(coreoptions.Options{Schema: a.optionSchema, Router: a.router, Broker: serviceBroker, Page: v.OptionsPage, Instances: func(ctx context.Context) ([]coreoptions.AdapterInstance, error) {
		installed, err := store.Installs(ctx)
		if err != nil {
			return nil, err
		}
		owners := make(map[string]string)
		for _, in := range installed {
			if in.Role == state.RoleAdapter && !in.Uninstalled {
				owners[in.AdapterID] = in.PluginID
			}
		}
		var instances []coreoptions.AdapterInstance
		for adapterID, configured := range a.settings.Settings().Adapters {
			owner := owners[adapterID]
			if owner == "" || configured.Package != "" && configured.Package != owner {
				continue
			}
			for instanceID := range configured.Instances {
				instances = append(instances, coreoptions.AdapterInstance{InstanceID: instanceID, AdapterID: adapterID})
			}
		}
		return instances, nil
	}})
	a.scheduler = scheduler.New(scheduler.Options{JobTimeout: v.JobTimeout})
	return a, nil
}

func (a *assembly) optionSchema(componentID string, kind coreoptions.SchemaKind) (*configschema.Schema, bool) {
	caller, active := a.router.Caller(componentID)
	if !active {
		return nil, false
	}
	a.schemaMu.RLock()
	defer a.schemaMu.RUnlock()
	current := a.schemas[componentID]
	if current.session != caller.SessionKey {
		return nil, false
	}
	var schema *configschema.Schema
	switch kind {
	case coreoptions.ConfigSchema:
		schema = current.config
	case coreoptions.AdapterInstanceSchema:
		schema = current.adapter
	}
	return schema, schema != nil
}

func optionSchemas(resources sdk.Generated) (componentOptionSchemas, error) {
	var result componentOptionSchemas
	for _, entry := range []struct {
		name   string
		raw    []byte
		schema **configschema.Schema
	}{{"config_schema.json", resources.ConfigSchema, &result.config}, {"adapter_config_schema.json", resources.AdapterConfigSchema, &result.adapter}} {
		if len(entry.raw) == 0 {
			continue
		}
		schema, problems := configschema.Parse(entry.raw)
		if len(problems) != 0 {
			return result, fmt.Errorf("%s: %v", entry.name, problems)
		}
		*entry.schema = schema
	}
	return result, nil
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
	schemas  componentOptionSchemas
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
	a.components.Store(r.id, r.instance)
	plugin := router.FromInstance(r.instance)
	r.schemas.session = r.caller.SessionKey
	a.schemaMu.Lock()
	a.schemas[r.id] = r.schemas
	a.schemaMu.Unlock()
	a.router.Attach(plugin, r.caller)
	a.bus.Attach(plugin)
	if r.process == nil {
		if r.instance.HTTPListener != nil {
			a.http.Bind(r.id, r.instance.GenerationID, r.instance.HTTPListener, r.instance.HTTPProxyToken, r.instance.Routes, r.install.Role == state.RoleAdapter)
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
	const configFile = "config.json"
	const databaseFile = "dev.sqlite"
	// 1. 每次启动读取原配置，SDK 随后按首装顺序准备；文件内容保持原样。
	pluginDir := filepath.Join(dir, "plugins", id)
	configPath := filepath.Join(pluginDir, configFile)
	configJSON, err := os.ReadFile(configPath)
	if errors.Is(err, os.ErrNotExist) {
		configJSON = []byte(emptyConfig)
	} else if err != nil {
		return fmt.Errorf("read %s: %w", configPath, err)
	}
	if !validate.IsJSONObject(configJSON) {
		return fmt.Errorf("config %s must be a JSON object", configPath)
	}
	version := uint32(initialConfigVersion)
	var secretPaths []string
	if len(entry.resources.ConfigSchema) != 0 {
		schema, problems := configschema.Parse(entry.resources.ConfigSchema)
		if len(problems) != 0 {
			return fmt.Errorf("component %s config schema: %v", id, problems)
		}
		version = uint32(schema.Version())
		for _, path := range schema.SecretPaths() {
			secretPaths = append(secretPaths, strings.Join(path, "."))
		}
	}
	generation := devGenerationPrefix + id
	boot := bootstrap(generation, version)
	boot.CandidateConfigJson = string(configJSON)
	if entry.role != state.RoleAdapter {
		path := filepath.Join(pluginDir, databaseFile)
		reason := pb.ActivationReason_ACTIVATION_REASON_RESUME
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			reason = pb.ActivationReason_ACTIVATION_REASON_FIRST_INSTALL
		} else if err != nil {
			return fmt.Errorf("stat %s: %w", path, err)
		}
		boot.DatabaseDescriptor = &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_SQLITE,
			SqlitePath: path, ActivationReason: reason, MigrationMode: pb.MigrationMode_MIGRATION_MODE_UNSPECIFIED}
	}
	current := c.assembly.settings.Settings()
	level := slog.LevelDebug
	if explicit, ok := current.LogSources[id]; ok {
		switch explicit {
		case "off":
			level = logging.LevelOff
		case "inherit":
			level = current.LogLevel
		default:
			if err := level.UnmarshalText([]byte(explicit)); err != nil {
				return err
			}
		}
	}
	boot.LogLevel = logging.ProtoLevel(level)
	c.output.Levels().SetSource(id, level)

	schemas, err := optionSchemas(entry.resources)
	if err != nil {
		return fmt.Errorf("component %s options schema: %w", id, err)
	}

	// 2. 保留固定代次与已有安装身份，本次启动只刷新构建摘要。
	digest := make([]byte, launchRandomBytes)
	rand.Read(digest)
	install, err := c.store.GetInstall(ctx, id)
	if errors.Is(err, state.ErrNotFound) {
		install = state.Install{Role: entry.role, PluginID: id, AdapterID: entry.manifest.AdapterID, Source: devSource, BuildDigest: digest, Version: entry.manifest.Version}
		err = c.store.AddInstall(ctx, install)
	} else if err == nil {
		err = c.store.SetDevBuild(ctx, id, entry.manifest.Version, digest)
	}
	if err != nil {
		return err
	}
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
		if err := c.store.AddGeneration(ctx, id, generation, emptyConfig, version); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(pluginDir, 0o700); err != nil {
		return err
	}
	// 3. 走共用的握手、准备与激活流程。
	listener, err := wire.ListenLoopback()
	if err != nil {
		return err
	}
	running := &runningComponent{id: id, install: install, caller: router.NewCaller(id, generation), listener: listener, finished: make(chan struct{}), schemas: schemas}
	c.components = append(c.components, running)
	l := &inproc.Launch{PluginID: id, Secret: make([]byte, launchRandomBytes), Addr: listener.Addr().String(), BuildDigest: digest, RuntimeLockPath: filepath.Join(pluginDir, wire.RuntimeLockFile), Log: c.output.Component(id, generation)}
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
	boot.DefaultLocale = current.DefaultLocale
	messages, problems := i18n.Parse(entry.resources.I18n)
	if len(problems) != 0 {
		return fmt.Errorf("component %s messages: %v", id, problems)
	}
	prep, err := host.Prepare(ctx, running.conn, host.Expect{Role: entry.role, AdapterID: entry.manifest.AdapterID, PluginID: id, Secret: l.Secret, BuildDigest: digest,
		Offer: validate.Offer{Major: validate.ProtocolMajor, Minor: validate.ProtocolMinor}}, boot, limits)
	if err != nil {
		return err
	}
	if err := c.store.SetDevGenerationConfig(ctx, generation, string(prep.ConfigSnapshot), prep.Register.GetCandidateConfigVersion(), secretPaths); err != nil {
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
	if err := c.assembly.attach(ctx, running); err != nil {
		return err
	}
	if err := c.startInstances(ctx, running, entry.resources.AdapterConfigSchema); err != nil {
		return err
	}
	return c.store.SetEnabled(ctx, id, true)
}

func bootstrap(generation string, version uint32) *pb.Bootstrap {
	return &pb.Bootstrap{GenerationId: generation, CandidateConfigJson: emptyConfig, TargetConfigVersion: version, DefaultLocale: defaultLocale,
		DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_NONE}, LogLevel: pb.LogLevel_LOG_LEVEL_INFO}
}

func (a *assembly) detach(r *runningComponent) {
	a.components.CompareAndDelete(r.id, r.instance)
	a.router.Detach(r.id)
	a.schemaMu.Lock()
	delete(a.schemas, r.id)
	a.schemaMu.Unlock()
	a.bus.Detach(r.id)
	a.scheduler.Detach(r.id)
	if r.instance != nil {
		a.http.Unbind(r.id, r.instance.GenerationID, r.instance.HTTPProxyToken)
	}
	if r.remote != nil {
		r.remote.Unbind()
	}
}

func (c *Core) stopComponent(r *runningComponent, t0 time.Time, timing host.StopTiming, reason pb.StopReason) error {
	a := c.assembly
	frozen := func(deadline time.Time) { a.scheduler.Drain(r.id); a.http.Drain(r.id, deadline); a.bus.Detach(r.id) }
	detach := func() { a.detach(r) }
	var err error
	if r.process != nil {
		// supervisor 处理进程树、HTTP 与 Scheduler；其余映射由装配撤销。
		a.bus.Detach(r.id)
		result := r.process.Stop(t0, reason)
		slog.Info("plugin stopped", "plugin_id", r.id, "drained", result.Drained, "killed", result.Killed)
		err = result.Err
		detach()
	} else if r.instance != nil {
		result := r.instance.Stop(t0, reason, r, timing, host.StopHooks{Frozen: frozen, Stopping: detach})
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
	isNew := generation == ""
	packageDir, err := installedDirectory(dir, install)
	if err != nil {
		return nil, err
	}
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
	schemas, err := optionSchemas(resources)
	if err != nil {
		return nil, err
	}
	pluginDir := filepath.Join(dir, "plugins", install.PluginID)
	if err := os.MkdirAll(pluginDir, 0o700); err != nil {
		return nil, err
	}
	configuration := []byte(emptyConfig)
	if isNew {
		configuration, err = os.ReadFile(filepath.Join(pluginDir, "config.json"))
		if errors.Is(err, os.ErrNotExist) {
			configuration, err = []byte(emptyConfig), nil
		}
		if err != nil {
			return nil, err
		}
		configuration, err = config.Canonical(configuration)
		if err != nil {
			return nil, err
		}
		generation = "gen-" + rand.Text()
		// shortcut: 首装候选直接交给 supervisor 提交，候选 Journal 接入后由编排器接管。
		// 原文只交给 Bootstrap；supervisor 提取 secret 后才写最终快照。
		if err := c.store.AddGeneration(ctx, install.PluginID, generation, emptyConfig, *m.ConfigVersion); err != nil {
			return nil, err
		}
	}
	running := &runningComponent{id: install.PluginID, install: install, caller: router.NewCaller(install.PluginID, generation), schemas: schemas}
	boot := bootstrap(generation, *m.ConfigVersion)
	boot.LogLevel = logging.ProtoLevel(c.output.Levels().Source(install.PluginID))
	boot.DatabaseDescriptor.ActivationReason = pb.ActivationReason_ACTIVATION_REASON_RESUME
	if isNew {
		boot.CandidateConfigJson = string(configuration)
		boot.DatabaseDescriptor.ActivationReason = pb.ActivationReason_ACTIVATION_REASON_FIRST_INSTALL
	}
	switch m.MigrationMode {
	case manifest.MigrationSelf:
		boot.DatabaseDescriptor.MigrationMode = pb.MigrationMode_MIGRATION_MODE_SELF
	case manifest.MigrationVersioned:
		boot.DatabaseDescriptor.MigrationMode = pb.MigrationMode_MIGRATION_MODE_VERSIONED
	}
	if (install.Role == state.RolePlugin || install.Role == state.RoleAgent) && (m.PrepareEffects == manifest.EffectsManagedDB || m.MigrationMode == manifest.MigrationSelf || m.MigrationMode == manifest.MigrationVersioned) {
		boot.DatabaseDescriptor.Backend = pb.DatabaseBackend_DATABASE_BACKEND_SQLITE
	}
	running.process, err = supervisor.Launch(ctx, c.store, supervisor.Spec{Role: install.Role, AdapterID: install.AdapterID, PluginID: install.PluginID, PluginDir: pluginDir,
		Executable: executable, BuildDigest: install.BuildDigest, GenerationID: generation, Bootstrap: boot, ConfigSchema: resources.ConfigSchema, I18n: resources.I18n,
		NewCandidate:  isNew,
		DefaultLocale: c.assembly.settings.Settings().DefaultLocale, Handler: c.assembly.handler(running, install), Global: c.assembly.global, Limits: host.NewLimits(c.values), Stop: host.NewStopTiming(c.values),
		Log: c.output.Plugin(install.PluginID, generation), Scheduler: c.assembly.scheduler, HTTP: c.assembly.http, MemorySoftLimit: c.values.PluginMemorySoftLimit, ExtraEnv: c.environment})
	if err != nil {
		return nil, err
	}
	running.instance = running.process.Instance
	if err := c.assembly.attach(ctx, running); err != nil {
		return running, err
	}
	var instanceSchema []byte
	if install.Role == state.RoleAdapter {
		instanceSchema, err = os.ReadFile(filepath.Join(packageDir, "adapter_config_schema.json"))
		if err != nil {
			return running, err
		}
	}
	if err := c.startInstances(ctx, running, instanceSchema); err != nil {
		return running, err
	}
	mode := "restart"
	if isNew {
		mode = "new"
	}
	slog.Info("plugin activated", "plugin_id", running.id, "generation_id", generation, "mode", mode, "pid", running.process.Process.Pid())
	return running, nil
}

func installedDirectory(dir string, install state.Install) (string, error) {
	// MoveTo 已将版本与归档摘要前缀写入目录名；摘要绑定当前安装记录。
	digest := hex.EncodeToString(install.BuildDigest)
	if len(digest) < packageDigestPrefix {
		return "", fmt.Errorf("component %s has invalid installed build digest", install.PluginID)
	}
	parent := filepath.Join(dir, "packages", install.PluginID)
	entries, err := os.ReadDir(parent)
	if err != nil {
		return "", err
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
		return "", fmt.Errorf("component %s: expected one installed package for build digest, found %d", install.PluginID, len(paths))
	}
	return paths[0], nil
}

func (a *assembly) componentState(id string) string {
	value, ok := a.components.Load(id)
	if !ok {
		return ""
	}
	instance := value.(*host.Instance)
	select {
	case <-instance.Session.Done():
		return ""
	default:
		return instance.State().String()
	}
}

func (a *assembly) componentDescriptor(id string) *pb.AdapterDescriptor {
	value, ok := a.components.Load(id)
	if !ok {
		return nil
	}
	instance := value.(*host.Instance)
	if !router.FromInstance(instance).Activated() {
		return nil
	}
	return instance.Adapter
}

func (c *Core) instanceSchema(components []component) func(context.Context, string) (*configschema.Schema, error) {
	return func(ctx context.Context, id string) (*configschema.Schema, error) {
		var raw []byte
		for _, entry := range components {
			if entry.manifest.ID == id {
				raw = entry.resources.AdapterConfigSchema
				break
			}
		}
		if raw == nil {
			install, err := c.store.GetInstall(ctx, id)
			if err != nil {
				return nil, err
			}
			directory, err := installedDirectory(c.dir, install)
			if err != nil {
				return nil, err
			}
			raw, err = os.ReadFile(filepath.Join(directory, "adapter_config_schema.json"))
			if err != nil {
				return nil, err
			}
		}
		schema, problems := configschema.Parse(raw)
		if len(problems) != 0 {
			return nil, fmt.Errorf("adapter instance schema: %v", problems)
		}
		return schema, nil
	}
}

func (c *Core) startInstances(ctx context.Context, running *runningComponent, raw []byte) error {
	if running.remote == nil {
		return nil
	}
	unlock := c.store.LockPlugin(running.id)
	defer unlock()
	owned, err := c.assembly.settings.BindAdapter(running.install.AdapterID, running.id)
	if err != nil || !owned {
		return err
	}
	schema, problems := configschema.Parse(raw)
	if len(problems) != 0 {
		return fmt.Errorf("adapter instance schema: %v", problems)
	}
	instances, diagnostics, err := c.assembly.settings.PrepareInstances(running.install.AdapterID, schema)
	if err != nil {
		return err
	}
	for _, problem := range diagnostics {
		slog.Warn("adapter instance configuration failed", "plugin_id", running.id, "instance_id", problem.InstanceID, "path", problem.Path, "reason", problem.Reason)
	}
	c.assembly.adapters.SetDefaults(c.assembly.settings.Settings().DefaultInstances())
	for id, configuration := range instances {
		if c.stoppedInstances[[2]string{running.id, id}] {
			continue
		}
		if err := running.remote.StartInstance(ctx, id, configuration); err != nil {
			slog.Warn("adapter instance start failed", "plugin_id", running.id, "instance_id", id, "error", err)
		}
	}
	return nil
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
	raw, err = os.ReadFile(filepath.Join(dir, "adapter_config_schema.json"))
	if err == nil {
		schema, problems := configschema.Parse(raw)
		if len(problems) != 0 {
			return g, fmt.Errorf("installed adapter instance schema: %v", problems)
		}
		g.AdapterConfigSchema = schema.Canonical()
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

// checkAdapterEvent 按事件所属实例当前 Session 的 Adapter 声明复检 PreEvent 改写后的扩展事件。
func (a *assembly) checkAdapterEvent(ev *pb.Event) error {
	instance, ok := a.adapters.Get(ev.GetAdapterInstanceId())
	if !ok {
		return fmt.Errorf("adapter instance %q is unavailable", ev.GetAdapterInstanceId())
	}
	checker, ok := instance.(interface{ CheckEvent(*pb.Event) error })
	if !ok {
		return fmt.Errorf("adapter instance %q cannot validate extension events", ev.GetAdapterInstanceId())
	}
	return checker.CheckEvent(ev)
}
