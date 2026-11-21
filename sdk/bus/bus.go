// Package bus 声明、发布与订阅带类型的非持久 JSON 事件。
//
// [Event] 绑定名字与载荷，[Subscribe] 登记无 Actor 和默认目标的串行回调。
package bus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
	"google.golang.org/protobuf/proto"
)

// Event 绑定完整事件名与载荷类型，零值无效。
type Event[T any] struct {
	name string
}

// Define 声明 plugin_id/event_name 的 JSON 事件。
func Define[T any](name string) Event[T] { return Event[T]{name: name} }

// Subscription 构建一条类型化订阅；Handle 结束声明。
type Subscription[T any] struct {
	scope       *registry.Scope[*sdk.Context, sdk.Permission]
	routeID     string
	eventName   string
	handles     int
	afterHandle bool
	handler     func(*sdk.Context, T) error
}

// Subscribe 开始构建声明的 JSON 订阅，路由 ID 缺省为完整事件名。
// 回调按 Session 串行执行，sdk.Context 无 Actor 与默认 Target。
func Subscribe[T any](owner sdk.Registrar, e Event[T]) *Subscription[T] {
	scope := registry.For[*sdk.Context, sdk.Permission](owner)
	scope.Check()
	r := scope.Subscriptions
	s := &Subscription[T]{scope: scope, routeID: e.name, eventName: e.name}
	r.Pending = append(r.Pending, func() error {
		if s.handles != 1 || s.afterHandle {
			return fmt.Errorf("subscription route %q has invalid Handle sequence", s.routeID)
		}
		if s.handler == nil {
			return fmt.Errorf("handler for subscription route %q is nil", s.routeID)
		}
		handler := s.handler // 收集后冻结 Handler，运行时注册描述保持固定。
		id, version, err := registry.JSONType[T]()
		if err != nil {
			return fmt.Errorf("subscription route %q: %w", s.routeID, err)
		}
		id = registry.SchemaID(scope.Resources().ServiceSchemas, scope.Entry(), id)
		r.Subscriptions = append(r.Subscriptions, registry.Subscription[*sdk.Context]{Plugin: scope.Plugin, SubID: scope.SubID, Descriptor: &pb.RouteDescriptor{
			RouteId: s.routeID, Kind: pb.RouteKind_ROUTE_KIND_BUS_SUBSCRIPTION, DefaultEnabled: true,
			Spec: &pb.RouteDescriptor_BusSubscription{BusSubscription: &pb.BusSubscriptionSpec{EventName: s.eventName, SchemaId: id, SchemaVersion: version, CallbackRouteId: s.routeID}},
		}, Invoke: func(c *sdk.Context, payload []byte) error {
			var input T
			if err := json.Unmarshal(payload, &input); err != nil {
				c.Logger().Warn("decode bus event payload failed", "error", err)
				return nil
			}
			return handler(c, input)
		}})
		return nil
	})
	return s
}

// ID 设置订阅路由 ID；Handle 之后调用在 Run 收集注册时报错。
func (s *Subscription[T]) ID(routeID string) *Subscription[T] {
	s.scope.Check()
	if s.handles != 0 {
		s.afterHandle = true
	}
	s.routeID = routeID
	return s
}

// Handle 结束订阅声明，签名由类型参数约束。
func (s *Subscription[T]) Handle(h func(*sdk.Context, T) error) {
	s.scope.Check()
	s.handles++
	s.handler = h
}

// Publish 从具名载荷类型推导 Schema 标识与版本，发布本插件命名空间内的事件。
// 返回值只确认 Core 已完成发布准入。Handler 内外均可调用；ctx 只传递取消和期限，不继承用户权限。
func (e Event[T]) Publish(ctx context.Context, payload T) (eventID string, err error) {
	pluginID, eventName, ok := strings.Cut(e.name, "/")
	if !ok || pluginID == "" || eventName == "" {
		return "", errcode.InvalidArgument.Errorf("invalid bus event name %q", e.name)
	}
	schemaID, version, err := registry.JSONType[T]()
	if err != nil {
		return "", err
	}
	// 1. 复用 Session ready gate，以当前插件身份编码 JSON 和发布请求。
	session, err := registry.Session(ctx)
	if err != nil {
		return "", err
	}
	identity, entry, schemas := registry.Runtime(ctx)
	if pluginID != identity {
		return "", errcode.InvalidArgument.Errorf("bus event %q does not belong to this plugin", e.name)
	}
	schemaID = registry.SchemaID(schemas, entry, schemaID)
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", &sdk.Error{Code: errcode.InvalidArgument, Message: "encode bus event payload: " + err.Error()}
	}
	request := &pb.PublishRequest{EventName: e.name, SchemaId: schemaID, SchemaVersion: version, Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, Payload: encoded}
	encoded, err = proto.Marshal(request)
	if err != nil {
		return "", &sdk.Error{Code: errcode.InvalidArgument, Message: "encode publish request: " + err.Error()}
	}
	// 2. 独立请求不设置 parent_call_id、trace_id 或可信调用附件。
	response, err := session.Call(ctx, wire.Request{Target: wire.ServiceBusPublish,
		Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: registry.SchemaVersion, Bytes: encoded}})
	if err != nil {
		message := err.Error()
		var callError *wire.CallError
		if errors.As(err, &callError) {
			message = callError.ErrorInfo.GetSafeMessage()
		}
		return "", &sdk.Error{Code: errcode.Code(validate.PublicCode(wire.CodeOf(err))), Message: message}
	}
	result := &pb.PublishResponse{}
	if err := registry.Decode(response.Payload, result); err != nil {
		return "", &sdk.Error{Code: errcode.Internal, Message: err.Error()}
	}
	return result.GetEventId(), nil
}
