package sdk

import (
	"encoding/json"
	"reflect"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk/internal/messagecodec"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

const buttonRoutePrefix = "button."

// ButtonAction 将动作名与按钮值的 Go 类型绑定。
type ButtonAction[T any] struct {
	name    string
	version uint32
}

// Action 声明按钮动作；值类型实现 SchemaVersion 时使用该版本，否则为 1。
func Action[T any](name string) ButtonAction[T] {
	version := uint32(1)
	var value T
	if v, ok := any(value).(interface{ SchemaVersion() uint32 }); ok && reflect.TypeFor[T]().Kind() != reflect.Pointer {
		version = v.SchemaVersion()
	}
	return ButtonAction[T]{name: name, version: version}
}

// Button 编码按钮值；编码错误保留到 Send 发出请求前报告。
func (a ButtonAction[T]) Button(label string, value T) message.Button {
	encoded, err := json.Marshal(value)
	return messagecodec.NewButton(a.name, label, string(encoded), err)
}

// OnButton 开始构建 button.clicked 路由，route_id 固定为 button. 加动作名。
func OnButton[T any](r *EventRegistry, a ButtonAction[T]) *ButtonBuilder[T] {
	b := r.On(buttonRoutePrefix+a.name, "button.clicked")
	b.route.actionID, b.route.actionVersion = a.name, a.version
	b.button = true
	return &ButtonBuilder[T]{event: b}
}

// ButtonBuilder 收集带类型的按钮回调，Handle 结束构建。
type ButtonBuilder[T any] struct{ event *EventBuilder }

// Priority 设置路由优先级。
func (b *ButtonBuilder[T]) Priority(p int32) *ButtonBuilder[T] {
	b.event.Priority(p)
	return b
}

// Permission 设置每次点击所需的权限。
func (b *ButtonBuilder[T]) Permission(p Permission) *ButtonBuilder[T] {
	b.event.Permission(p)
	return b
}

// Handle 登记接收解码后按钮值的处理函数。
func (b *ButtonBuilder[T]) Handle(h func(*Context, T) error) {
	b.event.route.decodeButton = func(raw string) (any, error) {
		var value T
		err := json.Unmarshal([]byte(raw), &value)
		return value, err
	}
	if h == nil {
		b.event.Handle(nil)
		return
	}
	b.event.route.handleButton = func(c *Context, value any) error { return h(c, value.(T)) }
	b.event.Handle(func(*Context) error { return nil })
}

func checkMessageEncoding(chain message.Chain) error {
	for _, segment := range chain {
		if err := messagecodec.Error(segment); err != nil {
			return &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "encode message: " + err.Error()}
		}
	}
	return nil
}
