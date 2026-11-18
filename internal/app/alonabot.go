// Package app 装配 Core，管理候选与组件进程，并通过 loopback Session 运行同进程开发组件。
//
// Start 等全部组件激活后返回，调用方取消 context 后通过 Core.Wait 等待停止。
// Options 描述一次运行的输入；开发 CLI 可通过 Core 的重启与设置方法热部署组件。
// 已安装组件仍由独立进程运行。规格见 CONTRACTS §15.4、§15.5。
package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adminauth"
	"github.com/AlonaBot/AlonaBot/internal/core/gc"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/httpproxy"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/secret"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/webhost"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/mattn/go-isatty"
	"google.golang.org/protobuf/proto"
)

// Options 汇总一次运行的装配参数；指针字段为 nil 时不覆盖 core_config.json。
type Options struct {
	DataDir    string      // 原样交给 filepath.Abs；不设缺省值
	Args       []string    // 调优参数与参数文件，后于 ALONABOT_ARGS 应用
	Components []Component // 开发组件，按登记顺序
	Listen     *string
	LogLevel   *slog.Level
	LogFormat  *string
	Stdio      bool    // Executable 组件使用 stdio 传输
	WebUIs     []fs.FS // 开发 WebUI；多于一个或含 nil 时返回原有的用法错误
}

// Core 持有已装配的运行资源；取消 Start 的 context 后调用 Wait 等待停止结果。
type Core struct {
	assembly      *assembly
	store         *state.Store
	output        *logging.Output
	logLevel      *slog.Level // 本次显式覆盖；nil 时随 core_config.json 的 log.level 刷新。
	values        tune.Values
	components    []*runningComponent
	development   []Component
	developmentMu sync.RWMutex       // HTTP 管理端读取 Schema 时，与开发资源替换同步。
	operations    chan coreOperation // Core.run 独占组件列表，HTTP 回调把变更交给同一所有者。
	stopping      chan struct{}      // 关闭后解除仍在等待管理操作的 HTTP 请求。
	dir           string
	environment   map[string]string
	http          *http.Server
	httpErrors    chan error
	monitors      sync.WaitGroup // 停止后等待退出观察者交还清理权，再关闭共享资源。
	done          chan struct{}  // err 写入后关闭，允许多个 Wait 读取同一结果。
	err           error

	stoppedInstances map[[2]string]bool // 包与实例的运行期停用选择，跨 Session 保留；同样由操作队列串行访问。
}

// Start 完成身份检查、握手、注册与激活后返回 Core；失败时清理已启动的部分。
// 启动期间 ctx 取消且清理成功时返回 context.Canceled，清理错误随启动错误一起返回。
func Start(ctx context.Context, o Options) (*Core, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(o.WebUIs) > 1 {
		return nil, startupUsage(errors.New("WebUI option may only be registered once"))
	}
	if len(o.WebUIs) == 1 && o.WebUIs[0] == nil {
		return nil, startupUsage(errors.New("WebUI file system is required"))
	}
	var webUI fs.FS
	if len(o.WebUIs) == 1 {
		webUI = o.WebUIs[0]
	}
	args := o.Args
	if env := os.Getenv("ALONABOT_ARGS"); env != "" {
		args = append([]string{env}, args...)
	}
	values, overrides, err := tune.Parse(args)
	if err != nil {
		return nil, startupUsage(err)
	}
	dir, err := filepath.Abs(o.DataDir)
	if err != nil {
		return nil, err
	}
	// 启动诊断先写 stderr，状态库核对与配置解析失败也保留 CRITICAL 文本。
	bootstrap := settings.Defaults()
	if o.LogLevel != nil {
		bootstrap.LogLevel = *o.LogLevel
	}
	if o.LogFormat != nil {
		bootstrap.LogFormat = *o.LogFormat
	}
	output, err := logging.Open(logging.Options{Format: logging.Format(bootstrap.LogFormat), Level: bootstrap.LogLevel, Stderr: os.Stderr, BufferRecords: values.LogBufferRecords})
	if err != nil {
		return nil, startupUsage(err)
	}
	slog.SetDefault(output.Logger())
	// 1. 取得实例锁后才打开日志和媒体目录，避免改动另一个 Core 的资源。
	store, err := state.Open(ctx, dir)
	if err != nil {
		return nil, err
	}
	c := &Core{store: store, output: output, values: values, dir: dir, logLevel: o.LogLevel, done: make(chan struct{}), stopping: make(chan struct{}), operations: make(chan coreOperation)}
	fail := func(err error) (*Core, error) {
		c.err = err
		if ctx.Err() == context.Canceled && startupCleanupError(err) == nil {
			c.err = nil
		}
		shutdownErr := c.shutdown()
		if c.err == nil && shutdownErr == nil {
			return nil, context.Canceled
		}
		return nil, errors.Join(err, shutdownErr)
	}
	file, err := settings.Load(dir)
	if err != nil {
		return fail(startupUsage(err))
	}
	configuration := file.Settings()
	baseURL, err := url.Parse(configuration.BaseURL)
	if err != nil {
		return fail(fmt.Errorf("parse base_url: %w", err))
	}
	if o.Listen != nil {
		if err := settings.ValidateListen(*o.Listen); err != nil {
			return fail(startupUsage(fmt.Errorf("--listen: %w", err)))
		}
		hostname, _, _ := net.SplitHostPort(*o.Listen)
		ip, _ := netip.ParseAddr(hostname)
		if !ip.IsLoopback() && !strings.EqualFold(strings.TrimSuffix(hostname, "."), "localhost") && baseURL.Scheme != "https" && len(configuration.TrustedProxies) == 0 {
			return fail(startupUsage(errors.New("--listen: non-loopback listen requires HTTPS base_url or trusted_proxies")))
		}
		configuration.Listen = *o.Listen
	}
	if o.LogLevel != nil {
		configuration.LogLevel = *o.LogLevel
	}
	if o.LogFormat != nil {
		configuration.LogFormat = *o.LogFormat
	}
	installs, err := store.Installs(ctx)
	if err != nil {
		return fail(err)
	}
	components := slices.Clone(o.Components)
	if err := validateComponents(components, installs); err != nil {
		return fail(err)
	}
	c.development = components
	for i := range c.development {
		c.development[i].stdio = o.Stdio
	}
	output, err = logging.Open(logging.Options{Format: logging.Format(configuration.LogFormat), Level: configuration.LogLevel, Stderr: os.Stderr, BufferRecords: values.LogBufferRecords})
	if err != nil {
		return fail(startupUsage(err))
	}
	c.output = output
	slog.SetDefault(output.Logger())
	for id, value := range configuration.LogSources {
		if value == "inherit" {
			continue
		}
		level := logging.LevelOff
		if value == "critical" {
			level = logging.LevelCritical
		} else if value != "off" {
			if err := level.UnmarshalText([]byte(value)); err != nil {
				return fail(err)
			}
		}
		output.Levels().SetSource(id, level)
	}
	if webUI != nil {
		slog.Warn("development WebUI has no manifest; Admin API version compatibility is not checked")
	}
	if configuration.LogFile.Enabled {
		if err := output.EnableFile(logging.FileOptions{Dir: filepath.Join(dir, "logs"), MaxSize: configuration.LogFile.MaxSize, MaxTotal: configuration.LogFile.MaxTotal, MaxAgeDays: configuration.LogFile.MaxAgeDays}); err != nil {
			return fail(err)
		}
	}
	interactive := isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stderr.Fd())
	token, err := adminauth.FirstToken(ctx, store, dir, interactive)
	if err != nil {
		return fail(err)
	}
	if token != "" {
		if _, err := fmt.Fprintln(os.Stderr, token); err != nil {
			return fail(fmt.Errorf("write admin token: %w", err))
		}
	} else if !interactive {
		name := filepath.Join(dir, adminauth.FirstTokenFile)
		if _, err := os.Stat(name); err == nil {
			slog.Info("first admin token is available", "path", name)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fail(fmt.Errorf("stat first admin token file: %w", err))
		}
	}
	if configuration.Proxy != "" {
		for _, name := range []string{"HTTP_PROXY", "HTTPS_PROXY"} {
			if err := os.Setenv(name, configuration.Proxy); err != nil {
				return fail(fmt.Errorf("set %s: %w", name, err))
			}
		}
	}
	c.environment = make(map[string]string)
	for _, name := range []string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY"} {
		if value, exists := os.LookupEnv(name); exists {
			c.environment[name] = value
		}
	}
	if len(overrides) > 0 {
		attrs := make([]any, 0, len(overrides))
		for _, override := range overrides {
			attrs = append(attrs, slog.Group(override.Name, "value", override.Value, "origin", override.Origin))
		}
		slog.Warn("tuning parameters overridden", attrs...)
	}
	if values.CoreMemorySoftLimit > 0 {
		debug.SetMemoryLimit(values.CoreMemorySoftLimit)
	}
	// 2. 装配公共服务，再依次启动已启用安装包与显式登记的开发组件。
	c.assembly, err = assemble(store, dir, values, file)
	if err != nil {
		return fail(err)
	}
	c.assembly.core = c
	// 恢复只信 Journal 与 Active；先清理候选，再启动任一正式或开发组件。
	recovery, err := store.Journals(ctx, "")
	if err != nil {
		return fail(err)
	}
	blocked, err := c.assembly.journal.Recover(ctx)
	if err != nil {
		return fail(err)
	}
	for _, id := range blocked {
		slog.Error("old process may still be running", "plugin_id", id)
	}
	authentication := adminauth.New(store, time.Now, baseURL, dir)
	adminMux := http.NewServeMux()
	authentication.Routes(adminMux)
	development := make(map[string]*manifest.Manifest)
	c.stoppedInstances = make(map[[2]string]bool)
	for _, entry := range components {
		development[entry.manifest.ID] = entry.manifest
	}
	version := coreVersion()
	management := manage.New(manage.Options{Logs: output, Stopping: c.stopping, NotifyLogLevel: func(ctx context.Context, id string, level pb.LogLevel) error {
		return c.operation(ctx, func(ctx context.Context) error {
			plugin, ok := c.assembly.router.Plugin(id)
			if !ok || !plugin.Activated() {
				return nil
			}
			request := &pb.LogSetLevelRequest{Level: level}
			encoded, err := proto.Marshal(request)
			if err != nil {
				return err
			}
			response, err := plugin.Call(ctx, wire.Request{Target: wire.ServiceLogSetLevel, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: 1, Bytes: encoded}})
			if err != nil {
				return err
			}
			if response.Err != nil {
				return &wire.CallError{ErrorInfo: response.Err}
			}
			reply := &pb.LogSetLevelResponse{}
			if response.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || response.Payload.SchemaID != validate.SchemaID(reply) || response.Payload.SchemaVersion != 1 {
				return errors.New("invalid log level response metadata")
			}
			return proto.Unmarshal(response.Payload.Bytes, reply)
		})
	}, Store: store, Settings: c.assembly.settings, Policy: c.assembly.policy, Router: c.assembly.router, Options: c.assembly.options, Tools: c.assembly.tools,
		CoreInfo: func() (string, string) { return version, c.http.Addr },
		Adapters: c.assembly.adapters, State: c.assembly.componentState, InstanceSchema: c.instanceSchema(components),
		Packages: filepath.Join(dir, "packages"), Limits: packages.NewLimits(values), Development: development, DevelopmentWebUI: webUI,
		Descriptor:        c.assembly.componentDescriptor,
		DevelopmentConfig: c.developmentConfig, RestartDevelopment: c.Restart,
		Candidates: managedCandidates{core: c},
		Start: func(ctx context.Context, id string) error {
			return c.operation(ctx, func(ctx context.Context) error {
				blocked, err := c.assembly.journal.Recover(ctx)
				if err != nil {
					return err
				}
				if slices.Contains(blocked, id) {
					return fmt.Errorf("plugin %s: old process may still be running", id)
				}
				return c.startManaged(ctx, id)
			})
		},
		Stop: func(ctx context.Context, id string, reason pb.StopReason) error {
			return c.operation(ctx, func(context.Context) error { return c.stopManaged(id, reason) })
		},
		StartInstance: func(ctx context.Context, id, instance string, config json.RawMessage) error {
			return c.operation(ctx, func(ctx context.Context) error {
				for _, running := range c.components {
					if running != nil && running.id == id && running.remote != nil {
						if err := running.remote.StartInstance(ctx, instance, config); err != nil {
							return err
						}
						delete(c.stoppedInstances, [2]string{id, instance})
						return nil
					}
				}
				return fmt.Errorf("adapter is not running: %s", id)
			})
		},
		StopInstance: func(ctx context.Context, id, instance string) error {
			return c.operation(ctx, func(ctx context.Context) error {
				c.stoppedInstances[[2]string{id, instance}] = true
				for _, running := range c.components {
					if running != nil && running.id == id && running.remote != nil {
						return running.remote.StopInstance(ctx, instance)
					}
				}
				return fmt.Errorf("adapter is not running: %s", id)
			})
		},
		ForgetInstance: func(ctx context.Context, id, instance string) error {
			return c.operation(ctx, func(context.Context) error {
				delete(c.stoppedInstances, [2]string{id, instance})
				return nil
			})
		},
	})
	managementHandler := management.Handler(authentication)
	var managementWrites sync.Mutex        // 覆盖元数据写入与随后的引用回收。
	managementReady := make(chan struct{}) // 启动回收完成后才接受管理写入。
	administration := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			managementHandler.ServeHTTP(w, r)
			return
		}
		select {
		case <-managementReady:
		case <-c.stopping:
			http.Error(w, "Core is stopping", http.StatusServiceUnavailable)
			return
		}
		managementWrites.Lock()
		defer managementWrites.Unlock()
		managementHandler.ServeHTTP(w, r)
		// 元数据操作也需要回收；统一排入运行循环，避免和候选准备交错。
		if err := c.operation(context.WithoutCancel(r.Context()), func(ctx context.Context) error {
			_, err := gc.Collect(ctx, c.store, secret.New(c.store), filepath.Join(c.dir, "packages"))
			return err
		}); err != nil {
			slog.Error("resource collection after management operation failed", "error", err)
		}
	})
	adminMux.Handle("/api/v1/", administration)
	adminMux.Handle("/api/", administration)
	adminMux.Handle("/api", administration)
	static := management.WebUI()
	adminMux.Handle("/", static)
	admin := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/"), "/")
		// 让所属入口拒绝原始路径，避免 ServeMux 先清理再重定向到另一份资源。
		if r.URL.Path != "/" && (!fs.ValidPath(name) || strings.ContainsAny(name, "\\:\x00")) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				administration.ServeHTTP(w, r)
			} else {
				static.ServeHTTP(w, r)
			}
			return
		}
		adminMux.ServeHTTP(w, r)
	})
	c.assembly.http = httpproxy.New(authentication)
	publicHost := webhost.New(webhost.Options{BaseURL: baseURL, Proxy: c.assembly.http, Admin: admin})
	listener, err := net.Listen("tcp", configuration.Listen)
	if err != nil {
		return fail(fmt.Errorf("listen HTTP: %w", err))
	}
	c.http = &http.Server{Addr: listener.Addr().String(), Handler: publicHost, ErrorLog: slog.NewLogLogger(slog.Default().Handler(), slog.LevelError)}
	c.httpErrors = make(chan error, 1)
	go func() { c.httpErrors <- c.http.Serve(listener) }()
	slog.Info("HTTP listener started", "listen", listener.Addr().String())
	slog.Info("Core started", "data_dir", dir)
	for _, install := range installs {
		if install.Source == devSource || install.Uninstalled || install.Role == state.RoleWebUI {
			continue
		}
		enabled, err := store.Enabled(ctx, install.PluginID)
		if err != nil {
			return fail(err)
		}
		if enabled {
			if slices.Contains(blocked, install.PluginID) {
				continue
			}
			active, err := store.Active(ctx, install.PluginID)
			if err != nil {
				return fail(err)
			}
			recovering := false
			for _, record := range recovery {
				if record.PluginID != install.PluginID || record.Phase == "cleaned" {
					continue
				}
				if record.Phase != "committed" || record.NewGeneration == active {
					recovering = true
				}
			}
			if recovering && active == "" {
				continue // 首装未提交时保留未激活状态。
			}
			running, err := c.startInstalled(ctx, dir, install)
			if running != nil {
				c.components = append(c.components, running)
			}
			if err != nil && recovering {
				slog.Log(ctx, logging.LevelCritical, "committed generation recovery failed", "plugin_id", install.PluginID, "generation_id", active, "error", err)
				if running != nil && running.instance.State() != host.StateDegraded {
					running.instance.Degrade(err)
					c.assembly.detach(running)
					if closeErr := running.instance.Session.Close(); closeErr != nil {
						slog.Error("close degraded session failed", "plugin_id", install.PluginID, "error", closeErr)
					}
				}
				continue
			}
			if err != nil && (running == nil || running.instance.State() != host.StateDegraded) {
				return fail(fmt.Errorf("start %s: %w", install.PluginID, err))
			}
		}
	}
	// 开发 Adapter 先完成 REGISTER，依赖检查才能读取真实的 extension 与动作声明。
	for i := range components {
		if components[i].role != state.RoleAdapter {
			continue
		}
		if err := ctx.Err(); err != nil {
			return fail(err)
		}
		if err := c.startDevelopment(ctx, dir, &components[i]); err != nil {
			return fail(fmt.Errorf("start %s: %w", components[i].manifest.ID, err))
		}
	}
	for i := range components {
		if components[i].role == state.RoleAdapter {
			continue
		}
		if err := ctx.Err(); err != nil {
			return fail(err)
		}
		snapshot, err := manage.DependencySnapshot(ctx, store, c.assembly.componentDescriptor)
		if err != nil {
			return fail(err)
		}
		// 同次运行的开发插件与 Agent 都参与版本检查，保留目录参数的先后自由。
		for _, entry := range components {
			if entry.role != state.RoleAdapter {
				snapshot.Plugins[entry.manifest.ID] = entry.manifest.Version
			}
		}
		if err := manage.CheckDependencies(components[i].manifest, snapshot); err != nil {
			return fail(fmt.Errorf("start %s: %w", components[i].manifest.ID, err))
		}
		if err := c.startDevelopment(ctx, dir, &components[i]); err != nil {
			return fail(fmt.Errorf("start %s: %w", components[i].manifest.ID, err))
		}
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if _, err := gc.Collect(ctx, store, secret.New(store), filepath.Join(dir, "packages")); err != nil {
		return fail(err)
	}
	c.assembly.router.Start()
	close(managementReady) // 发布启动完成，使管理写入与启动回收互斥。
	// 3. 一个所有者处理取消与异常，全部组件共用一次停止起点。
	go c.run(ctx)
	return c, nil
}

func startupUsage(err error) error {
	return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: err.Error()}
}

// Wait 等待 Core 停止；组件停止超时或开发组件异常退出时返回含组件 ID 的错误。
func (c *Core) Wait() error { <-c.done; return c.err }

// componentExit 在旧进程树的清理完成后交回退出结果，Core.run 独占替换组件列表。
type componentExit struct {
	running  *runningComponent
	abnormal bool
	state    string
	err      error
}

func (c *Core) watch(running *runningComponent, exits chan<- componentExit) {
	c.monitors.Go(func() {
		select {
		case <-running.exited():
		case <-running.instance.Session.Done():
		}
		result := componentExit{running: running}
		if running.process != nil {
			result.abnormal, result.state, result.err = running.process.WaitExit()
		} else {
			result.err = fmt.Errorf("component %s exited before Core shutdown", running.id)
		}
		// 停止时归还观察者，不让 HTTP 动态增减组件改变退出队列容量的要求。
		select {
		case exits <- result:
		case <-c.stopping:
		}
	})
}

func (c *Core) run(ctx context.Context) {
	exits := make(chan componentExit)
	policy := host.NewRestartPolicy(c.values)
	for _, running := range c.components {
		c.watch(running, exits)
	}
	ticker := time.NewTicker(c.values.MediaTTL)
	defer ticker.Stop()
running:
	for {
		select {
		case <-ctx.Done():
			break running
		case err := <-c.httpErrors:
			c.err = fmt.Errorf("serve HTTP: %w", err)
			break running
		case result := <-exits:
			if ctx.Err() != nil {
				break running
			}
			index := -1
			for i, running := range c.components {
				if running == result.running {
					index = i
					break
				}
			}
			if index < 0 {
				continue
			} // 管理停止的旧 Session 已移出，不再影响新 Session。
			old := result.running
			if old.instance.State() == host.StateDegraded {
				// 已提交失败保留供管理员恢复；观察者已清理进程树，不触发自动重启。
				c.assembly.media.ReleaseSession(old.caller.SessionKey)
				if result.err != nil {
					slog.Error("degraded process cleanup failed", "plugin_id", old.id, "error", result.err)
				}
				continue
			}
			c.assembly.detach(old)
			c.assembly.media.ReleaseSession(old.caller.SessionKey)
			if old.process == nil {
				c.err = result.err
				break running
			}
			if result.abnormal {
				slog.Error("plugin exited abnormally", "plugin_id", old.id, "state", result.state)
			} else {
				slog.Info("plugin exited", "plugin_id", old.id, "state", result.state)
			}
			if result.err != nil {
				slog.Error("process tree not confirmed empty, not starting a replacement", "plugin_id", old.id, "error", result.err)
				continue
			}
			enabled, err := c.store.Enabled(ctx, old.id)
			if err != nil {
				slog.Error("read component enabled state failed", "plugin_id", old.id, "error", err)
				continue
			}
			if !enabled {
				continue
			}
			switch policy.OnExit(old.id, result.abnormal, time.Now()) {
			case host.DecisionRestart:
				slog.Info("restarting plugin", "plugin_id", old.id)
				if old.install.Source == devSource {
					c.components[index] = nil
					before := len(c.components)
					if err := c.startManaged(ctx, old.id); err != nil {
						slog.Error("restart plugin failed", "plugin_id", old.id, "error", err)
					}
					for _, running := range c.components[before:] {
						if running != nil {
							c.watch(running, exits)
						}
					}
					continue
				}
				replacement, err := c.startInstalled(ctx, c.dir, old.install)
				if replacement != nil {
					c.components[index] = replacement
				}
				if err != nil {
					slog.Error("restart plugin failed", "plugin_id", old.id, "error", err)
					if replacement != nil {
						if stopErr := c.stopComponent(replacement, time.Now(), host.NewStopTiming(c.values), pb.StopReason_STOP_REASON_DISABLE); stopErr != nil {
							slog.Error("cleanup failed restart", "plugin_id", old.id, "error", stopErr)
						}
					}
					continue
				}
				c.watch(replacement, exits)
			case host.DecisionDisable:
				if err := c.store.SetEnabled(ctx, old.id, false); err != nil {
					slog.Error("disable plugin failed", "plugin_id", old.id, "error", err)
				} else {
					slog.Error("plugin disabled after repeated abnormal exits", "plugin_id", old.id, "crashes", c.values.RestartMaxCrashes, "window", c.values.RestartWindow)
				}
			}
		case operation := <-c.operations:
			before := len(c.components)
			operationContext, cancel := context.WithCancel(operation.ctx)
			stopCancel := context.AfterFunc(ctx, cancel)
			err := operation.apply(operationContext)
			stopCancel()
			cancel()
			for _, running := range c.components[before:] {
				if running != nil {
					c.watch(running, exits)
				}
			}
			operation.result <- err
		case <-ticker.C:
			c.assembly.media.Sweep()
		}
	}
	c.err = errors.Join(c.err, c.shutdown())
	close(c.done)
}

// Restart 检查新资源的依赖后重新启动一个开发组件，沿用 dev:<id> 代次；依赖失败时保持旧进程运行。
func (c *Core) Restart(ctx context.Context, id string) error {
	return c.operation(ctx, func(ctx context.Context) error {
		for i := range c.development {
			entry := &c.development[i]
			if entry.manifest.ID != id {
				continue
			}
			next := *entry
			if next.executable != "" {
				if err := next.loadExecutableResources(); err != nil {
					return err
				}
			}
			// 停止前用已启用组件的当前版本和运行中 Adapter 声明检查新资源。
			snapshot, err := manage.DependencySnapshot(ctx, c.store, c.assembly.componentDescriptor)
			if err != nil {
				return err
			}
			if err := manage.CheckDependencies(next.manifest, snapshot); err != nil {
				return err
			}
			if err := c.stopManaged(id, pb.StopReason_STOP_REASON_RESTART); err != nil {
				return err
			}
			c.developmentMu.Lock()
			*entry = next
			c.developmentMu.Unlock()
			return c.startManaged(ctx, id)
		}
		return fmt.Errorf("component %s is not a development component", id)
	})
}

// ReloadSettings 重新读取 core_config.json，更新 Core Admin、缺省实例和日志级别；命令前缀立即读取新值，
// default_locale 在之后建立的 Session 生效。与其他管理操作串行，组件保持运行。
func (c *Core) ReloadSettings(ctx context.Context) error {
	return c.operation(ctx, func(context.Context) error { return c.reloadSettings() })
}

func (c *Core) reloadSettings() error {
	if _, err := c.assembly.settings.Reload(); err != nil {
		return err
	}
	// 管理操作可能已通过 Settings.Update 发布快照，仍须同步实际权限。
	configuration := c.assembly.settings.Settings()
	var admins []policy.Admin
	for _, admin := range configuration.Admins {
		admins = append(admins, policy.Admin{AdapterID: admin.AdapterID, UserID: admin.UserID})
	}
	c.assembly.policy.SetAdmins(admins)
	c.assembly.adapters.SetDefaults(configuration.DefaultInstances())
	level := configuration.LogLevel
	if c.logLevel != nil {
		level = *c.logLevel
	}
	c.output.Levels().SetGlobal(level)
	return nil
}

// RestartInstance 只重新准备并启动开发 Adapter 的指定实例。
func (c *Core) RestartInstance(ctx context.Context, adapterID, instanceID string) error {
	return c.operation(ctx, func(ctx context.Context) error {
		for _, entry := range c.development {
			if entry.manifest.AdapterID != adapterID || entry.role != state.RoleAdapter {
				continue
			}
			if err := c.reloadSettings(); err != nil {
				return err
			}
			for _, running := range c.components {
				if running == nil || running.id != entry.manifest.ID || running.remote == nil {
					continue
				}
				schema, problems := configschema.Parse(entry.resources.AdapterConfigSchema)
				if len(problems) != 0 {
					return fmt.Errorf("adapter instance schema: %v", problems)
				}
				owned, err := c.assembly.settings.BindAdapter(adapterID, running.id)
				if err != nil {
					return err
				}
				if !owned {
					return fmt.Errorf("adapter %s configuration belongs to another component", adapterID)
				}
				instances, diagnostics, err := c.assembly.settings.PrepareInstances(adapterID, schema)
				if err != nil {
					return err
				}
				for _, problem := range diagnostics {
					if problem.InstanceID == "" || problem.InstanceID == instanceID {
						return fmt.Errorf("instance %s: %s: %s", instanceID, problem.Path, problem.Reason)
					}
				}
				if err := running.remote.StopInstance(ctx, instanceID); err != nil {
					var detail *validate.Error
					if !errors.As(err, &detail) || detail.Code != pb.ErrorCode_ERROR_CODE_NOT_FOUND {
						return err
					}
				}
				c.assembly.adapters.SetDefaults(c.assembly.settings.Settings().DefaultInstances())
				if configuration, exists := instances[instanceID]; exists {
					return running.remote.StartInstance(ctx, instanceID, configuration)
				}
				return nil
			}
			return fmt.Errorf("adapter %s is not running", adapterID)
		}
		return fmt.Errorf("adapter %s is not a development component", adapterID)
	})
}

func (c *Core) shutdown() error {
	close(c.stopping)
	var result error
	if c.http != nil {
		ctx, cancel := context.WithTimeout(context.Background(), c.values.StopTotal)
		result = c.http.Shutdown(ctx)
		cancel()
		result = errors.Join(result, c.http.Close())
	}
	t0 := time.Now()
	timing := host.NewStopTiming(c.values)
	var wg sync.WaitGroup
	errs := make(chan error, len(c.components)+1)
	// Router 的排空与组件冻结并行，避免把停止预算逐个累加。
	if c.assembly != nil {
		wg.Go(func() {
			ctx, cancel := context.WithDeadline(context.Background(), t0.Add(timing.Drain))
			defer cancel()
			if err := c.assembly.router.Close(ctx); err != nil {
				slog.Warn("event router drain ended", "error", err)
			}
		})
	}
	for _, running := range c.components {
		if running == nil {
			continue
		}
		wg.Go(func() {
			if err := c.stopComponent(running, t0, timing, pb.StopReason_STOP_REASON_CORE_SHUTDOWN); err != nil {
				errs <- err
			}
		})
	}
	wg.Wait()
	c.monitors.Wait()
	close(errs)
	for err := range errs {
		result = errors.Join(result, err)
	}
	if c.assembly != nil {
		c.assembly.scheduler.Close()
	}
	if c.output != nil {
		if err := errors.Join(c.err, result); err != nil {
			slog.Log(context.Background(), logging.LevelCritical, "Core stopped", "err", err)
		} else {
			slog.Info("Core stopped")
		}
	}
	// 同进程用户回调超时时保留实例锁，实际退出后再关闭状态库与日志。
	pending := false
	for _, running := range c.components {
		if running == nil {
			continue
		}
		select {
		case <-running.exited():
		default:
			pending = true
		}
	}
	if pending {
		go func() {
			for _, running := range c.components {
				if running == nil {
					continue
				}
				<-running.exited()
			}
			_ = c.close() // close 自己记录延迟清理的故障。
		}()
		return result
	}
	return errors.Join(result, c.close())
}

func (c *Core) close() error {
	err := c.store.Close()
	if c.output != nil {
		err = errors.Join(err, c.output.Close())
	}
	if err != nil {
		slog.Log(context.Background(), logging.LevelCritical, "close Core resources failed", "error", err)
	}
	return err
}

type coreOperation struct {
	ctx    context.Context
	apply  func(context.Context) error
	result chan error
}

func (c *Core) operation(ctx context.Context, apply func(context.Context) error) error {
	op := coreOperation{ctx: ctx, apply: apply, result: make(chan error, 1)}
	select {
	case c.operations <- op:
	case <-ctx.Done():
		return ctx.Err()
	case <-c.stopping:
		return errors.New("Core is stopping")
	}
	// 已接纳的操作在取消后仍须交还资源所有权，调用者才可清理暂存包。
	select {
	case err := <-op.result:
		return err
	case <-c.stopping:
		return errors.New("Core is stopping")
	}
}

func (c *Core) startManaged(ctx context.Context, id string) error {
	for _, running := range c.components {
		if running != nil && running.id == id && c.assembly.componentState(id) == "Activated" {
			return nil
		}
	}
	if err := c.stopManaged(id, pb.StopReason_STOP_REASON_RESTART); err != nil {
		return err
	}
	for i := range c.development {
		if c.development[i].manifest.ID == id {
			if err := c.startDevelopment(ctx, c.dir, &c.development[i]); err != nil {
				return errors.Join(err, c.stopManaged(id, pb.StopReason_STOP_REASON_DISABLE))
			}
			return nil
		}
	}
	in, err := c.store.GetInstall(ctx, id)
	if err != nil {
		return err
	}
	running, err := c.startInstalled(ctx, c.dir, in)
	if running != nil {
		c.components = append(c.components, running)
	}
	if err != nil && running != nil && running.instance.State() != host.StateDegraded {
		return errors.Join(err, c.stopManaged(id, pb.StopReason_STOP_REASON_DISABLE))
	}
	return err
}

func (c *Core) stopManaged(id string, reason pb.StopReason) error {
	for i, running := range c.components {
		if running == nil || running.id != id {
			continue
		}
		if err := c.stopComponent(running, time.Now(), host.NewStopTiming(c.values), reason); err != nil {
			return err
		}
		c.components[i] = nil
	}
	return nil
}
