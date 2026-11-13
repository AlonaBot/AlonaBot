package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/AlonaBot/AlonaBot/sdk/config"
	"github.com/AlonaBot/AlonaBot/sdk/internal/dbhook"
)

type migrationState = dbhook.State

// Plugin 持有一次构造的路由、配置、生成资源和生命周期回调。
// Run 开始后登记状态固定；每次 NewPlugin 返回独立的对象。
type Plugin struct {
	migrationState
	commands       CommandGroup
	hooks          []*hookRoute
	events         eventRegistry
	services       serviceRegistry
	subscriptions  subscriptionRegistry
	http           httpRegistry
	schedules      scheduleRegistry
	scope          pluginScope
	subs           []*SubPlugin
	jobSubPlugins  map[string]string
	generated      Generated
	generatedCount int
	onStart        func(context.Context) error
	onReady        func(context.Context) error
	onStop         func(context.Context) error
	onConfigLoad   config.Loader
	callbacks      map[string]bool
	config         config.Values
	configReady    bool
	running        atomic.Bool
	errs           []error
}

// Registrar 约束类型化服务、订阅和按钮只能登记到插件或子插件。
type Registrar interface{ registrar() *pluginScope }

type pluginScope struct {
	plugin   *Plugin
	subID    string
	commands *CommandGroup
}

type pluginContextKey struct{}

// NewPlugin 创建一个独立的插件对象。
func NewPlugin() *Plugin {
	p := &Plugin{callbacks: make(map[string]bool), jobSubPlugins: make(map[string]string)}
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

func (r *pluginScope) every(jobID string, d time.Duration, fn func(*Context) error) {
	r.plugin.schedules.Every(jobID, d, fn)
	r.plugin.jobSubPlugins[jobID] = r.subID
}

func (r *pluginScope) cron(jobID, expr, timezone string, fn func(*Context) error) {
	r.plugin.schedules.Cron(jobID, expr, timezone, fn)
	r.plugin.jobSubPlugins[jobID] = r.subID
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

// Handle 登记 HTTP 路由。
func (r *Plugin) Handle(method, path string, h http.Handler, opts ...HTTPOption) {
	r.scope.plugin.http.Handle(method, path, h, opts...)
}

// Every 登记固定间隔任务。
func (r *Plugin) Every(jobID string, d time.Duration, fn func(*Context) error) {
	r.scope.every(jobID, d, fn)
}

// Cron 登记 cron 任务。
func (r *Plugin) Cron(jobID, expr, timezone string, fn func(*Context) error) {
	r.scope.cron(jobID, expr, timezone, fn)
}
