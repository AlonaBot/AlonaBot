package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/generated"
	"github.com/AlonaBot/AlonaBot/sdk/config"
	"github.com/AlonaBot/AlonaBot/sdk/internal/dbhook"
)

type migrationState = dbhook.State

// Plugin 持有一次构造的路由、配置、生成资源和生命周期回调。
// Run 开始后登记状态固定；每次 NewPlugin 返回独立的对象。
type Plugin struct {
	migrationState
	commands      CommandGroup
	hooks         []*hookRoute
	events        eventRegistry
	services      serviceRegistry
	subscriptions subscriptionRegistry
	http          httpRegistry
	schedules     scheduleRegistry
	tasks         taskRegistry
	scope         pluginScope
	subs          []*SubPlugin
	jobSubPlugins map[string]string
	onStart       func(context.Context) error
	onReady       func(context.Context) error
	onStop        func(context.Context) error
	onConfigLoad  config.Loader
	callbacks     map[string]bool
	config        config.Values
	configReady   bool
	running       atomic.Bool
	errs          []error
}

// Registrar 约束类型化服务、订阅、按钮和任务只能登记到插件或子插件。
type Registrar interface{ registrar() *pluginScope }

type pluginScope struct {
	plugin   *Plugin
	subID    string
	commands *CommandGroup
}

type pluginContextKey struct{}

// NewPlugin 创建一个独立的插件对象。
func NewPlugin() *Plugin {
	pc, _, _, _ := runtime.Caller(1)
	pkgPath, _ := runtimeFunctionParts(runtime.FuncForPC(pc).Name())
	p := &Plugin{callbacks: make(map[string]bool), jobSubPlugins: make(map[string]string)}
	generated.SetEntry(p, pkgPath)
	p.commands.plugin = p
	p.events.plugin = p
	p.http.plugin = p
	p.schedules.plugin = p
	p.scope = pluginScope{plugin: p, commands: &p.commands}
	return p
}

func (p *Plugin) registrar() *pluginScope { return &p.scope }

func (p *Plugin) freeze() {
	if p != nil {
		p.running.Store(true)
		dbhook.Freeze(p)
	}
}

func (p *Plugin) checkRegistration() {
	if p.running.Load() {
		panic("plugin registration is closed after Run starts")
	}
}

func (p *Plugin) callback(name string) {
	p.checkRegistration()
	if p.callbacks[name] {
		p.errs = append(p.errs, fmt.Errorf("%s registered twice", name))
	}
	p.callbacks[name] = true
}

// OnStart 登记配置及数据库准备完成后、REGISTER 之前执行的回调。
func (p *Plugin) OnStart(fn func(context.Context) error) { p.callback("OnStart"); p.onStart = fn }

// OnReady 登记 ACTIVATE 后、ACTIVATED 之前执行的回调。
func (p *Plugin) OnReady(fn func(context.Context) error) { p.callback("OnReady"); p.onReady = fn }

// OnStop 登记停止预算内执行的资源关闭回调。
func (p *Plugin) OnStop(fn func(context.Context) error) { p.callback("OnStop"); p.onStop = fn }

// OnConfigLoad 登记声明式迁移之后执行的纯配置变换。
func (p *Plugin) OnConfigLoad(fn func(context.Context, uint32, json.RawMessage) (json.RawMessage, error)) {
	p.callback("OnConfigLoad")
	p.onConfigLoad = fn
}

// Config 返回整个插件的只读配置；准备完成前读取会 panic。
func (p *Plugin) Config() config.Values {
	if !p.configReady {
		panic("config is not prepared")
	}
	return p.config
}

func (p *Plugin) setConfig(values config.Values) { p.config = values; p.configReady = true }

func (r *pluginScope) on(routeID string, eventType EventType) *EventBuilder {
	b := r.plugin.events.On(routeID, eventType)
	b.route.subID = r.subID
	return b
}

func (r *pluginScope) every(jobID string, d time.Duration) *ScheduleBuilder {
	b := r.plugin.schedules.Every(jobID, d)
	r.plugin.jobSubPlugins[jobID] = r.subID
	return b
}

func (r *pluginScope) cron(jobID, expr string) *ScheduleBuilder {
	b := r.plugin.schedules.Cron(jobID, expr)
	r.plugin.jobSubPlugins[jobID] = r.subID
	return b
}

// Group 返回本作用域下共享路径与权限的命令组。
func (r *Plugin) Group(name string) *CommandGroup { return r.scope.commands.Group(name) }

// Command 开始登记相对于本作用域路径的命令。
func (r *Plugin) Command(path ...string) *CommandBuilder { return r.scope.commands.Command(path...) }

// Regex 开始登记 Go regexp 文本匹配。
func (r *Plugin) Regex(routeID, expr string) *CommandBuilder {
	return r.scope.commands.Regex(routeID, expr)
}

// Keyword 开始登记包含文本匹配。
func (r *Plugin) Keyword(routeID string, patterns ...string) *CommandBuilder {
	return r.scope.commands.Keyword(routeID, patterns...)
}

// StartsWith 开始登记开头文本匹配。
func (r *Plugin) StartsWith(routeID string, patterns ...string) *CommandBuilder {
	return r.scope.commands.StartsWith(routeID, patterns...)
}

// EndsWith 开始登记结尾文本匹配。
func (r *Plugin) EndsWith(routeID string, patterns ...string) *CommandBuilder {
	return r.scope.commands.EndsWith(routeID, patterns...)
}

// FullMatch 开始登记完整文本匹配。
func (r *Plugin) FullMatch(routeID string, patterns ...string) *CommandBuilder {
	return r.scope.commands.FullMatch(routeID, patterns...)
}

// Raw 开始登记原始消息路由。
func (r *Plugin) Raw(routeID string) *CommandBuilder { return r.scope.commands.Raw(routeID) }

// On 开始登记事件路由。
func (r *Plugin) On(routeID string, eventType EventType) *EventBuilder {
	return r.scope.on(routeID, eventType)
}

// Route 开始构建 method+path 的 HTTP 路由。
func (r *Plugin) Route(method, path string) *RouteBuilder {
	return r.scope.plugin.http.Route(method, path)
}

// Every 开始构建固定间隔任务。
func (r *Plugin) Every(jobID string, d time.Duration) *ScheduleBuilder {
	return r.scope.every(jobID, d)
}

// Cron 开始构建五字段 cron 任务。
func (r *Plugin) Cron(jobID, expr string) *ScheduleBuilder {
	return r.scope.cron(jobID, expr)
}

func entryPackage(p *Plugin) string {
	pkgPath, _ := generated.Entry(p)
	return pkgPath
}

func generatedFor(p *Plugin) generated.Resources {
	r, _ := generated.Lookup(p, entryPackage(p))
	return r
}

func componentResources(component any) (generated.Resources, string, bool) {
	pkgPath, ok := generated.Entry(component)
	if !ok {
		typ := reflect.TypeOf(component)
		if typ.Kind() == reflect.Pointer {
			typ = typ.Elem()
		}
		pkgPath = typ.PkgPath()
	}
	r, ok := generated.Lookup(component, pkgPath)
	return r, pkgPath, ok
}

func requireResources(component any) error {
	_, pkgPath, ok := componentResources(component)
	if !ok {
		return fmt.Errorf("entry package %q has no registered resources; use alonabot build or alonabot run", pkgPath)
	}
	return nil
}
