// Package button 声明与登记类型化按钮动作，编码按钮值并在点击时解码。
package button

import (
	"encoding/json"
	"reflect"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/internal/messagecodec"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// Action 将动作名与按钮值的 Go 类型绑定。
type Action[T any] struct {
	name    string
	version uint32
}

// Define 声明按钮动作；值类型实现 SchemaVersion 时使用该版本，否则为 1。
func Define[T any](name string) Action[T] {
	version := uint32(1)
	var value T
	if v, ok := any(value).(interface{ SchemaVersion() uint32 }); ok && reflect.TypeFor[T]().Kind() != reflect.Pointer {
		version = v.SchemaVersion()
	}
	return Action[T]{name: name, version: version}
}

// Button 编码按钮值；编码错误保留到 Send 发出请求前报告。
func (a Action[T]) Button(label string, value T) message.Button {
	encoded, err := json.Marshal(value)
	return messagecodec.NewButton(a.name, label, string(encoded), err)
}

// On 开始构建 button.clicked 路由，route_id 固定为 button. 加动作名。
func On[T any](r sdk.Registrar, a Action[T]) *Builder[T] {
	scope := registry.For[*sdk.Context, sdk.Permission](r)
	event, set := scope.Button(a.name, a.version)
	return &Builder[T]{event: event.(*sdk.EventBuilder), set: set}
}

// Builder 收集带类型的按钮回调，Handle 结束构建。
type Builder[T any] struct {
	event *sdk.EventBuilder
	set   func(func(string) (any, error), func(*sdk.Context, any) error)
}

// Priority 设置路由优先级。
func (b *Builder[T]) Priority(p int32) *Builder[T] {
	b.event.Priority(p)
	return b
}

// Permission 设置每次点击所需的权限。
func (b *Builder[T]) Permission(p sdk.Permission) *Builder[T] {
	b.event.Permission(p)
	return b
}

// Rule 按调用与参数顺序追加按钮回调条件。
func (b *Builder[T]) Rule(rules ...sdk.Rule) *Builder[T] {
	b.event.Rule(rules...)
	return b
}

// Continue 让按钮 Handler 在独立的延续调用中执行。
func (b *Builder[T]) Continue() *Builder[T] {
	b.event.Continue()
	return b
}

// Handle 登记接收解码后按钮值的处理函数。
func (b *Builder[T]) Handle(h func(*sdk.Context, T) error) {
	decode := func(raw string) (any, error) {
		var value T
		err := json.Unmarshal([]byte(raw), &value)
		return value, err
	}
	if h == nil {
		b.set(decode, nil)
		b.event.Handle(nil)
		return
	}
	b.set(decode, func(c *sdk.Context, value any) error { return h(c, value.(T)) })
	b.event.Handle(func(*sdk.Context) error { return nil })
}
