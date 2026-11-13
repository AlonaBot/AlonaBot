package alonabot

import (
	"context"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
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
	"github.com/AlonaBot/AlonaBot/internal/core/journal"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/media"
	"github.com/AlonaBot/AlonaBot/internal/core/message"
	"github.com/AlonaBot/AlonaBot/internal/core/metrics"
	coreoptions "github.com/AlonaBot/AlonaBot/internal/core/options"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/scheduler"
	"github.com/AlonaBot/AlonaBot/internal/core/secret"
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
	root            *sdk.Plugin
	adapter         adapter.Adapter
	role            state.Role
	manifest        *manifest.Manifest
	resources       sdk.Generated
	executable      string
	packageManifest []byte
	digest          []byte
	stdio           bool
}

func validateComponents(components []component, installs []state.Install) error {
	ids := make(map[string]bool)
	adapters := make(map[string]bool)
	agent := false
	for i := range components {
		entry := &components[i]
		var generated sdk.Generated
		var m *manifest.Manifest
		var warnings []string
		if entry.executable != "" {
			parsed, err := manifest.Parse(entry.packageManifest)
			if err != nil {
				return err
			}
			m = parsed.Manifest
			warnings = parsed.Warnings
			if m.Role == string(state.RoleWebUI) {
				return fmt.Errorf("component %s has no executable role", m.ID)
			}
			entry.role, entry.manifest = state.Role(m.Role), m
			if err := entry.loadExecutableResources(); err != nil {
				return err
			}
			generated = entry.resources
		} else {
			var object any = entry.root
			if entry.role == state.RoleAdapter {
				object = entry.adapter
			}
			if object == nil || (reflect.ValueOf(object).Kind() == reflect.Pointer && reflect.ValueOf(object).IsNil()) {
				return fmt.Errorf("component %T is nil", object)
			}
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
			m = parsed.Manifest
			warnings = parsed.Warnings
		}
		if entry.role == state.RolePlugin && m.Role == string(state.RoleAgent) {
			entry.role = state.RoleAgent
		}
		for _, warning := range warnings {
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
	core       *Core // assemble 返回后由 Start 绑定，候选回调在运行循环内读取。
	journal    *journal.Orchestrator
	candidate  *runningComponent // 仅运行循环访问；Spec 与提交结果合并前的本次运行上下文。
	adapters   *coreadapter.Registry
	router     *router.Router
	mux        *router.Mux
	media      *media.Store
	buttons    *buttons.Service
	bus        *bus.Bus
	broker     *broker.Broker
	scheduler  *scheduler.Scheduler
	http       *httpproxy.Server
	global     *wire.Budget
	metrics    *metrics.Counters
	store      *state.Store
	policy     *policy.Policy
	components sync.Map // 运行状态查询借用 host.Instance，激活失败时保留 Degraded 诊断。
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
	var dispatched func(router.Dispatch)
	if harness := inproc.HarnessFor(dir); harness != nil && harness.Dispatched != nil {
		dispatched = func(d router.Dispatch) {
			result := inproc.Dispatch{InstanceID: d.InstanceID, SourceEventID: d.SourceEventID, Waiting: d.Waiting}
			for _, outcome := range d.Outcomes {
				result.Outcomes = append(result.Outcomes, inproc.Outcome{PluginID: outcome.PluginID, RouteID: outcome.RouteID, ErrorCode: int32(outcome.Err.GetCode()), SafeMessage: outcome.Err.GetSafeMessage()})
			}
			harness.Dispatched(result)
		}
	}
	a.router = router.New(router.Options{Dispatched: dispatched, Concurrency: v.DispatchConcurrency, QueueEvents: v.DispatchQueueEvents, QueueBytes: v.DispatchQueueBytes,
		Global: a.global, Adapters: a.adapters, Prepare: a.media.PrepareIncoming, Buttons: a.buttons.Resolve, Intercept: a.buttons.InterceptConfirm, Policy: p, CommandState: p.CommandState, Prefixes: a.settings.CommandPrefixes, CommandNames: p.CommandNames,
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
	tools := tool.New(store, a.router)
	tools.Register(a.mux)
	a.bus = bus.New(bus.Options{QueueEvents: v.BusQueueEvents, QueueBytes: v.BusQueueBytes, Metrics: a.metrics})
	a.bus.Register(a.mux)
	a.broker = broker.New(broker.Options{Router: a.router, RequestTimeout: v.RequestTimeout, MaxDepth: v.CallChainMaxDepth, Metrics: a.metrics, Agent: a.agentID})
	tools.RegisterInvoke(a.mux, tool.InvokeOptions{Call: a.broker.Call, Adapters: a.adapters, Buttons: a.buttons,
		ConfirmTimeout: v.ToolConfirmTimeout, DefaultLocale: func() string { return a.settings.Settings().DefaultLocale }})
	a.mux.Forward(a.broker.Call)
	a.schemas = make(map[string]componentOptionSchemas)
	a.options = coreoptions.New(coreoptions.Options{Schema: a.optionSchema, Router: a.router, Broker: a.broker, Chats: store.Chats, Page: v.OptionsPage, Agent: a.agentID, Instances: func(ctx context.Context) ([]coreoptions.AdapterInstance, error) {
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
	a.scheduler = scheduler.New(scheduler.Options{JobTimeout: v.JobTimeout, Store: store, TasksPerPlugin: v.TasksPerPlugin, TaskPayloadBytes: v.TaskPayloadBytes, TaskConcurrency: v.TaskConcurrency})
	a.scheduler.Services(a.mux, a.router.PluginFor)
	a.journal = journal.New(journal.Options{Store: store, Secrets: secret.New(store), Config: config.New(store, dir), Settings: file, DataDir: dir, State: a.componentState,
		Spec: func(ctx context.Context, in state.Install, packageDir string, boot *pb.Bootstrap) (supervisor.Spec, error) {
			spec, running, err := a.core.installedSpec(ctx, in, packageDir, boot)
			a.candidate = running
			return spec, err
		},
		Stop: func(ctx context.Context, id string, reason pb.StopReason) error {
			return a.core.stopManaged(id, reason)
		},
		Resume: func(ctx context.Context, id string) error { return a.core.startManaged(ctx, id) },
	})
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
	id        string
	install   state.Install // 独立进程恢复时复用安装身份与已提交代次。
	caller    router.Caller
	instance  *host.Instance
	process   *supervisor.Running // 仅独立进程设置。
	remote    *remote.Process
	listener  net.Listener
	conn      *wire.Conn
	finished  chan struct{} // SDK goroutine 返回后关闭，并发布 runErr。
	runErr    error
	schemas   componentOptionSchemas
	attached  chan struct{} // 本 Session 的 Router/Scheduler 装配完成后关闭，发布 attachErr。
	attachErr error
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
	r.attached = make(chan struct{})
	bound := a.mux.For(r.caller)
	shared := func(ctx context.Context, in *wire.Inbound) wire.Response {
		switch in.Target {
		case wire.ServiceTaskSet, wire.ServiceTaskCancel, wire.ServiceTaskList, wire.ServiceJobPause, wire.ServiceJobResume, wire.ServiceJobStatus:
			// ACTIVATED 后读循环立即接纳请求；等待本 Session 装配，避免首请求越过发布。
			select {
			case <-r.attached:
				if r.attachErr != nil {
					return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "task services could not attach this Session"}}
				}
			case <-ctx.Done():
				code := pb.ErrorCode_ERROR_CODE_CANCELLED
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					code = pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED
				}
				return wire.Response{Err: &pb.ErrorInfo{Code: code, SafeMessage: "task service ended while waiting for Session attachment"}}
			}
		}
		return bound(ctx, in)
	}
	if install.Role != state.RoleAdapter {
		return shared
	}
	r.remote = remote.New(remote.Options{AdapterID: install.AdapterID, PluginID: install.PluginID, Store: a.store, Registry: a.adapters, Sink: a.router, RegisterIncoming: a.media.RegisterIncoming,
		StartCommandSync: a.router.StartCommandSync, StopCommandSync: a.router.StopCommandSync})
	// Adapter 私有服务按 Session 建独立 Mux，多个 Adapter 不会重复登记固定服务名。
	local := router.NewMux()
	r.remote.Services(local)
	private := local.For(r.caller)
	return func(ctx context.Context, in *wire.Inbound) wire.Response {
		switch in.Target {
		case wire.ServiceCoreAdapterDeliver, wire.ServiceCoreAdapterCursor, wire.ServiceCoreAdapterStatus:
			return private(ctx, in)
		default:
			return shared(ctx, in)
		}
	}
}

func (a *assembly) attach(ctx context.Context, r *runningComponent) (err error) {
	defer func() {
		r.attachErr = err
		close(r.attached)
	}()
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
		if err := a.scheduler.Attach(ctx, plugin, time.Now()); err != nil {
			a.detach(r)
			return fmt.Errorf("attach scheduler: %w", err)
		}
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
	if entry.executable != "" {
		version = *entry.manifest.ConfigVersion
	}
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
	if _, explicit := current.LogSources[id]; !explicit {
		c.output.Levels().SetSource(id, slog.LevelDebug)
	} else if current.LogSources[id] == "inherit" {
		c.output.Levels().InheritSource(id)
	}
	boot.LogLevel = logging.ProtoLevel(c.output.Levels().Source(id))

	schemas, err := optionSchemas(entry.resources)
	if err != nil {
		return fmt.Errorf("component %s options schema: %w", id, err)
	}

	// 2. 保留固定代次与已有安装身份，本次启动只刷新构建摘要。
	digest := entry.digest
	if entry.executable == "" {
		digest = make([]byte, launchRandomBytes)
		rand.Read(digest)
	}
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
	if harness := inproc.HarnessFor(dir); harness != nil {
		if instances, restricted := harness.AllowedInstances[id]; restricted {
			if err := c.store.SetAdapterInstanceAccess(ctx, id, false, instances); err != nil {
				return err
			}
		}
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
	// 3. 独立进程与同进程共用开发配置、数据库和固定代次。
	if entry.executable != "" {
		boot.DatabaseDescriptor.MigrationMode = pb.MigrationMode_MIGRATION_MODE_NONE
		switch entry.manifest.MigrationMode {
		case manifest.MigrationSelf:
			boot.DatabaseDescriptor.MigrationMode = pb.MigrationMode_MIGRATION_MODE_SELF
		case manifest.MigrationVersioned:
			boot.DatabaseDescriptor.MigrationMode = pb.MigrationMode_MIGRATION_MODE_VERSIONED
		}
		running := &runningComponent{id: id, install: install, caller: router.NewCaller(id, generation), schemas: schemas}
		running.process, err = supervisor.Launch(ctx, c.store, supervisor.Spec{
			Role: entry.role, AdapterID: entry.manifest.AdapterID, PluginID: id, PluginDir: pluginDir,
			Executable: entry.executable, BuildDigest: digest, GenerationID: generation, Bootstrap: boot,
			ConfigSchema: entry.resources.ConfigSchema, I18n: entry.resources.I18n,
			Development: true, NewCandidate: active == "", Stdio: entry.stdio,
			DefaultLocale: current.DefaultLocale, Handler: c.assembly.handler(running, install), Global: c.assembly.global,
			Limits: host.NewLimits(c.values), Stop: host.NewStopTiming(c.values), Log: c.output.Plugin(id, generation, componentLogAttrs(install)...),
			Scheduler: c.assembly.scheduler, HTTP: c.assembly.http, MemorySoftLimit: c.values.PluginMemorySoftLimit, ExtraEnv: c.environment,
		})
		if err != nil {
			return err
		}
		running.instance = running.process.Instance
		c.components = append(c.components, running)
		if err := c.assembly.attach(ctx, running); err != nil {
			return err
		}
		if err := c.startInstances(ctx, running, entry.resources.AdapterConfigSchema); err != nil {
			return err
		}
		mode := "restart"
		if active == "" {
			mode = "new"
		}
		if err := c.store.SetEnabled(ctx, id, true); err != nil {
			return err
		}
		slog.Info("plugin activated", "plugin_id", id, "generation_id", generation, "mode", mode, "pid", running.process.Process.Pid())
		return nil
	}
	// 4. 同进程调用方保留原有 SDK 入口。

	listener, err := wire.ListenLoopback()
	if err != nil {
		return err
	}
	running := &runningComponent{id: id, install: install, caller: router.NewCaller(id, generation), listener: listener, finished: make(chan struct{}), schemas: schemas}
	c.components = append(c.components, running)
	l := &inproc.Launch{PluginID: id, Secret: make([]byte, launchRandomBytes), Addr: listener.Addr().String(), BuildDigest: digest, RuntimeLockPath: filepath.Join(pluginDir, wire.RuntimeLockFile), Log: c.output.Component(id, generation, componentLogAttrs(install)...)}
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
	// STOPPING 和退出收尾都可能到达；旧 Session 的迟到清理保留新绑定。
	if a.components.CompareAndDelete(r.id, r.instance) {
		if r.instance.State() == host.StateDegraded {
			// 撤销业务绑定时保留管理诊断，迟到清理不得覆盖新 Session。
			a.components.LoadOrStore(r.id, r.instance)
		}
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
	a.broker.Release(r.caller)
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
	a.components.CompareAndDelete(r.id, r.instance)
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
		return c.prepareCandidate(ctx, install.PluginID, func(ctx context.Context) (*journal.Result, error) {
			return c.assembly.journal.Prepare(ctx, install.PluginID, journal.KindFirstInstall, nil)
		})
	}
	packageDir, err := installedDirectory(dir, install)
	if err != nil {
		return nil, err
	}
	pluginDir := filepath.Join(dir, "plugins", install.PluginID)
	if err := os.MkdirAll(pluginDir, 0o700); err != nil {
		return nil, err
	}
	boot := bootstrap(generation, initialConfigVersion)
	boot.DatabaseDescriptor.ActivationReason = pb.ActivationReason_ACTIVATION_REASON_RESUME
	version, err := c.store.GenerationConfigVersion(ctx, generation)
	if err != nil {
		return nil, err
	}
	boot.PreviousConfigVersion = &version
	spec, running, err := c.installedSpec(ctx, install, packageDir, boot)
	if err != nil {
		return nil, err
	}
	running.process, err = supervisor.Launch(ctx, c.store, spec)
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
	slog.Info("plugin activated", "plugin_id", running.id, "generation_id", generation, "mode", "restart", "pid", running.process.Process.Pid())
	return running, nil
}

// prepareCandidate 由运行循环独占调用，合并准备上下文与显式交接的进程资源。
func (c *Core) prepareCandidate(ctx context.Context, id string, prepare func(context.Context) (*journal.Result, error)) (*runningComponent, error) {
	a := c.assembly
	defer func() { a.candidate = nil }()
	// 1. 准备失败由 Journal 中止，恢复直接在当前运行循环执行。
	r, err := prepare(ctx)
	if err != nil {
		if r != nil {
			err = a.journal.Abort(ctx, r, err)
		}
		return nil, err
	}
	running := a.candidate
	record, err := c.store.GetJournal(ctx, r.OpID)
	if err != nil {
		return nil, a.journal.Abort(ctx, r, err)
	}
	kind := journal.Kind(record.Kind)
	// 2. 提交返回进程所有权，激活失败也登记诊断并交给退出观察者。
	process, err := a.journal.Commit(ctx, r)
	if process == nil {
		return nil, err
	}
	running.process, running.instance = process, process.Instance
	if kind == journal.KindFirstInstall {
		err = errors.Join(err, c.store.SetEnabled(context.WithoutCancel(ctx), id, true))
	}
	if (kind == journal.KindUpdate || kind == journal.KindRollback) && (!record.WasEnabled || !record.WasRunning) {
		// 提交只切换权威代次；先停止候选，之后显式 enable 才建立业务 Session。
		if stopErr := c.stopComponent(running, time.Now(), host.NewStopTiming(c.values), pb.StopReason_STOP_REASON_DISABLE); stopErr != nil {
			return running, errors.Join(err, stopErr)
		}
		return nil, err
	}
	if running.instance.State() == host.StateDegraded {
		a.components.Store(id, running.instance)
		return running, err
	}
	degraded := func(cause error) (*runningComponent, error) {
		// 元数据已提交，装配失败沿同一 Degraded 状态交还资源，保留管理员恢复入口。
		failure := errors.Join(err, cause)
		running.instance.Degrade(failure)
		a.detach(running)
		slog.Error("committed candidate assembly failed", "plugin_id", id, "op_id", r.OpID, "generation_id", r.Generation, "kind", kind, "error", failure)
		return running, errors.Join(failure, running.instance.Session.Close())
	}
	// 3. ACTIVATED 成功后发布业务入口，再启动 Adapter 实例。
	if attachErr := a.attach(ctx, running); attachErr != nil {
		return degraded(attachErr)
	}
	var instanceSchema []byte
	if running.install.Role == state.RoleAdapter {
		directory, directoryErr := installedDirectory(c.dir, running.install)
		if directoryErr != nil {
			return degraded(directoryErr)
		}
		var readErr error
		instanceSchema, readErr = os.ReadFile(filepath.Join(directory, "adapter_config_schema.json"))
		if readErr != nil {
			return degraded(readErr)
		}
	}
	if startErr := c.startInstances(ctx, running, instanceSchema); startErr != nil {
		return degraded(startErr)
	}
	mode := string(kind)
	if kind == journal.KindFirstInstall {
		mode = "new"
	}
	slog.Info("plugin activated", "plugin_id", id, "generation_id", r.Generation, "mode", mode, "pid", process.Process.Pid())
	return running, err
}

// managedCandidates 只在外部管理入口排队，Journal 的停止和恢复回调直接执行。
type managedCandidates struct{ core *Core }

// Reload 串行导入正式组件候选，开发组件沿直接停止与启动路径执行。
func (m managedCandidates) Reload(ctx context.Context, id string) error {
	return m.core.operation(ctx, func(ctx context.Context) error {
		in, err := m.core.store.GetInstall(ctx, id)
		if err != nil {
			return err
		}
		if in.Source == devSource {
			if err := m.core.stopManaged(id, pb.StopReason_STOP_REASON_RELOAD); err != nil {
				return err
			}
			return m.core.startManaged(ctx, id)
		}
		return m.apply(ctx, id, func(ctx context.Context) (*journal.Result, error) {
			return m.core.assembly.journal.Prepare(ctx, id, journal.KindReload, nil)
		})
	})
}

// Update 把已预检的候选交给一次 Core 操作。
func (m managedCandidates) Update(ctx context.Context, id string, staged *packages.Candidate) error {
	return m.core.operation(ctx, func(ctx context.Context) error {
		return m.apply(ctx, id, func(ctx context.Context) (*journal.Result, error) {
			return m.core.assembly.journal.Prepare(ctx, id, journal.KindUpdate, staged)
		})
	})
}

func (m managedCandidates) apply(ctx context.Context, id string, prepare func(context.Context) (*journal.Result, error)) error {
	running, err := m.core.prepareCandidate(ctx, id, prepare)
	if running != nil {
		m.core.components = append(m.core.components, running)
		if err != nil && running.instance.State() != host.StateDegraded {
			err = errors.Join(err, m.core.stopManaged(id, pb.StopReason_STOP_REASON_DISABLE))
		}
	}
	return err
}

// Rollback 串行准备保留代次的副本，沿用候选接入、提交与失败停止路径。
func (m managedCandidates) Rollback(ctx context.Context, id, generation string) error {
	return m.core.operation(ctx, func(ctx context.Context) error {
		return m.apply(ctx, id, func(ctx context.Context) (*journal.Result, error) {
			return m.core.assembly.journal.PrepareRollback(ctx, id, generation)
		})
	})
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
	if instance.State() == host.StateDegraded {
		return host.StateDegraded.String()
	}
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
		c.developmentMu.RLock()
		for _, entry := range components {
			if entry.manifest.ID == id {
				raw = entry.resources.AdapterConfigSchema
				break
			}
		}
		c.developmentMu.RUnlock()
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

// loadExecutableResources 将资源与当前二进制摘要配对；重建改变身份时保留旧 Session。
func (entry *component) loadExecutableResources() error {
	path, err := filepath.Abs(entry.executable)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	digest := sha512.Sum512_256(content)
	raw, err := os.ReadFile(fmt.Sprintf("%s.%x.json", path, digest[:]))
	resources := sdk.Generated{}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err == nil {
		if err := json.Unmarshal(raw, &resources); err != nil {
			return err
		}
		parsed, err := manifest.Parse(resources.Manifest)
		if err != nil {
			return err
		}
		m := parsed.Manifest
		if m.ID != entry.manifest.ID || m.Role != entry.manifest.Role || m.AdapterID != entry.manifest.AdapterID {
			return fmt.Errorf("development component identity changed: %s", entry.manifest.ID)
		}
		entry.manifest = m
	}
	entry.executable, entry.digest, entry.resources = path, digest[:], resources
	return nil
}

// componentLogAttrs 从 Core 的安装记录绑定可信来源类别，Adapter 的实例字段由 SDK 记录携带。
func componentLogAttrs(install state.Install) []slog.Attr {
	if install.Role == state.RoleAdapter {
		return []slog.Attr{slog.String(logging.KeyComponent, logging.ComponentAdapter), slog.String(logging.KeyAdapterID, install.AdapterID)}
	}
	return nil
}
