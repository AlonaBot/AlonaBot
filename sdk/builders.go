package sdk

// CommandBuilder 按匹配类型提供可用的注册选项。
type CommandBuilder struct{ state *commandBuilder }

// Permission 设置当前路由的 Permission 选项。
func (b *CommandBuilder) Permission(p Permission) *CommandBuilder { b.state.Permission(p); return b }

// Priority 设置当前路由的 Priority 选项。
func (b *CommandBuilder) Priority(p int32) *CommandBuilder { b.state.Priority(p); return b }

// Rule 设置当前路由的 Rule 选项。
func (b *CommandBuilder) Rule(rules ...Rule) *CommandBuilder { b.state.Rule(rules...); return b }

// Continue 设置当前路由的 Continue 选项。
func (b *CommandBuilder) Continue() *CommandBuilder { b.state.Continue(); return b }

// Menu 设置当前路由的 Menu 选项。
func (b *CommandBuilder) Menu() *CommandBuilder { b.state.Menu(); return b }

// Alias 设置当前路由的 Alias 选项。
func (b *CommandBuilder) Alias(path ...string) *CommandBuilder { b.state.Alias(path...); return b }

// Args 设置当前路由的 Args 选项。
func (b *CommandBuilder) Args(names ...string) *CommandBuilder { b.state.Args(names...); return b }

// Optional 设置当前路由的 Optional 选项。
func (b *CommandBuilder) Optional(name string, def ...string) *CommandBuilder {
	b.state.Optional(name, def...)
	return b
}

// RawTail 设置当前路由的 RawTail 选项。
func (b *CommandBuilder) RawTail() *CommandBuilder { b.state.RawTail(); return b }

// Description 设置当前路由的 Description 选项。
func (b *CommandBuilder) Description(text string) *CommandBuilder {
	b.state.Description(text)
	return b
}

// UsageKey 设置当前路由的 UsageKey 选项。
func (b *CommandBuilder) UsageKey(key string) *CommandBuilder { b.state.UsageKey(key); return b }

// Silent 设置当前路由的 Silent 选项。
func (b *CommandBuilder) Silent() *CommandBuilder { b.state.Silent(); return b }

// Handle 完成路由登记并保留既有签名校验。
func (b *CommandBuilder) Handle(handler any) { b.state.Handle(handler) }

// TextBuilder 按匹配类型提供可用的注册选项。
type TextBuilder struct{ state *commandBuilder }

// Permission 设置当前路由的 Permission 选项。
func (b *TextBuilder) Permission(p Permission) *TextBuilder { b.state.Permission(p); return b }

// Priority 设置当前路由的 Priority 选项。
func (b *TextBuilder) Priority(p int32) *TextBuilder { b.state.Priority(p); return b }

// Rule 设置当前路由的 Rule 选项。
func (b *TextBuilder) Rule(rules ...Rule) *TextBuilder { b.state.Rule(rules...); return b }

// Continue 设置当前路由的 Continue 选项。
func (b *TextBuilder) Continue() *TextBuilder { b.state.Continue(); return b }

// ID 设置当前路由的 ID 选项。
func (b *TextBuilder) ID(id string) *TextBuilder { b.state.ID(id); return b }

// IgnoreCase 设置当前路由的 IgnoreCase 选项。
func (b *TextBuilder) IgnoreCase() *TextBuilder { b.state.IgnoreCase(); return b }

// Handle 完成路由登记并保留既有签名校验。
func (b *TextBuilder) Handle(handler any) { b.state.Handle(handler) }

// RegexBuilder 按匹配类型提供可用的注册选项。
type RegexBuilder struct{ state *commandBuilder }

// Permission 设置当前路由的 Permission 选项。
func (b *RegexBuilder) Permission(p Permission) *RegexBuilder { b.state.Permission(p); return b }

// Priority 设置当前路由的 Priority 选项。
func (b *RegexBuilder) Priority(p int32) *RegexBuilder { b.state.Priority(p); return b }

// Rule 设置当前路由的 Rule 选项。
func (b *RegexBuilder) Rule(rules ...Rule) *RegexBuilder { b.state.Rule(rules...); return b }

// Continue 设置当前路由的 Continue 选项。
func (b *RegexBuilder) Continue() *RegexBuilder { b.state.Continue(); return b }

// ID 设置当前路由的 ID 选项。
func (b *RegexBuilder) ID(id string) *RegexBuilder { b.state.ID(id); return b }

// Handle 完成路由登记并保留既有签名校验。
func (b *RegexBuilder) Handle(handler any) { b.state.Handle(handler) }

// RawBuilder 按匹配类型提供可用的注册选项。
type RawBuilder struct{ state *commandBuilder }

// Permission 设置当前路由的 Permission 选项。
func (b *RawBuilder) Permission(p Permission) *RawBuilder { b.state.Permission(p); return b }

// Priority 设置当前路由的 Priority 选项。
func (b *RawBuilder) Priority(p int32) *RawBuilder { b.state.Priority(p); return b }

// Rule 设置当前路由的 Rule 选项。
func (b *RawBuilder) Rule(rules ...Rule) *RawBuilder { b.state.Rule(rules...); return b }

// Continue 设置当前路由的 Continue 选项。
func (b *RawBuilder) Continue() *RawBuilder { b.state.Continue(); return b }

// ID 设置当前路由的 ID 选项。
func (b *RawBuilder) ID(id string) *RawBuilder { b.state.ID(id); return b }

// Handle 完成路由登记并保留既有签名校验。
func (b *RawBuilder) Handle(handler any) { b.state.Handle(handler) }
