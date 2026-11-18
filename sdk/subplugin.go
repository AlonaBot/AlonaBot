package sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/AlonaBot/AlonaBot/sdk/config"
	"github.com/AlonaBot/AlonaBot/sdk/internal/dbhook"
)

// SubPlugin 提供一层逻辑路由分组、命令与 Hook 的缺省权限和同名配置视图。
type SubPlugin struct {
	registrationScope
	scope      pluginScope
	commands   CommandGroup
	usesConfig bool
}

// Sub 创建逻辑子插件；ID 和重复声明在 Run 收集注册时检查。
func (p *Plugin) Sub(id string) *SubPlugin {
	p.checkRegistration()
	s := &SubPlugin{registrationScope: p.sharedScope(id)}
	s.commands = CommandGroup{plugin: p, subID: id}
	s.scope = pluginScope{plugin: p, subID: id, commands: &s.commands}
	p.subs = append(p.subs, s)
	return s
}

// Permission 设置子插件内命令、文本匹配和 Hook 的缺省权限。
func (s *SubPlugin) Permission(p Permission) *SubPlugin { s.commands.Permission(p); return s }

// Config 返回可在登记时保存的同名顶层 object 视图；准备完成前读取会 panic。
func (s *SubPlugin) Config() config.Values {
	if !s.scope.plugin.running.Load() {
		s.usesConfig = true
	}
	var values config.Values
	dbhook.BindConfig(&values, func() config.Values { return s.scope.plugin.Config().Object(s.scope.subID) })
	return values
}

func (p *Plugin) checkSubPlugins() error {
	var errs []error
	seen := make(map[string]bool)
	var schema map[string]json.RawMessage
	// 生成资源是声明边界；无配置视图时不要求存在 schema。
	for _, s := range p.subs {
		id := s.scope.subID
		valid := id != ""
		for _, c := range []byte(id) {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
				valid = false
			}
		}
		if !valid {
			errs = append(errs, fmt.Errorf("invalid sub-plugin ID %q", id))
		}
		if seen[id] {
			errs = append(errs, fmt.Errorf("sub-plugin %q registered twice", id))
		}
		seen[id] = true

		if s.usesConfig {
			if schema == nil && len(generatedFor(p).ConfigSchema) != 0 {
				if err := json.Unmarshal(generatedFor(p).ConfigSchema, &schema); err != nil {
					errs = append(errs, fmt.Errorf("sub-plugin %q config schema: %w", id, err))
					continue
				}
			}
			var field struct {
				Type string `json:"type"`
			}
			if raw := schema[id]; len(raw) != 0 {
				if err := json.Unmarshal(raw, &field); err != nil {
					errs = append(errs, fmt.Errorf("sub-plugin %q config declaration: %w", id, err))
					continue
				}
			}
			if field.Type != "object" {
				errs = append(errs, fmt.Errorf("sub-plugin %q requires a same-name top-level config object", id))
			}
		}
	}
	return errors.Join(errs...)
}

// Group 返回本作用域下共享路径与权限的命令组。
func (r *SubPlugin) Group(name string) *CommandGroup { return r.scope.commands.Group(name) }

// Command 开始登记相对于本作用域路径的命令。
func (r *SubPlugin) Command(path ...string) *CommandBuilder { return r.scope.commands.Command(path...) }

// Regex 开始登记 Go regexp 文本匹配。
func (r *SubPlugin) Regex(expr string) *CommandBuilder {
	return r.scope.commands.Regex(expr)
}

// Keyword 开始登记包含文本匹配。
func (r *SubPlugin) Keyword(patterns ...string) *CommandBuilder {
	return r.scope.commands.Keyword(patterns...)
}

// StartsWith 开始登记开头文本匹配。
func (r *SubPlugin) StartsWith(patterns ...string) *CommandBuilder {
	return r.scope.commands.StartsWith(patterns...)
}

// EndsWith 开始登记结尾文本匹配。
func (r *SubPlugin) EndsWith(patterns ...string) *CommandBuilder {
	return r.scope.commands.EndsWith(patterns...)
}

// FullMatch 开始登记完整文本匹配。
func (r *SubPlugin) FullMatch(patterns ...string) *CommandBuilder {
	return r.scope.commands.FullMatch(patterns...)
}

// Raw 开始登记原始消息路由。
func (r *SubPlugin) Raw() *CommandBuilder { return r.scope.commands.Raw() }

// On 开始登记事件路由。
func (r *SubPlugin) On(eventType EventType) *EventBuilder {
	return r.scope.on("", eventType)
}

// Route 开始构建 method+path 的 HTTP 路由。
func (r *SubPlugin) Route(method, path string) *RouteBuilder {
	return r.scope.plugin.http.Route(method, path)
}

// Every 开始构建固定间隔任务。
func (r *SubPlugin) Every(jobID string, d time.Duration) *ScheduleBuilder {
	return r.scope.every(jobID, d)
}

// Cron 开始构建五字段 cron 任务。
func (r *SubPlugin) Cron(jobID, expr string) *ScheduleBuilder {
	return r.scope.cron(jobID, expr)
}
