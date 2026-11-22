package registry

import (
	"context"
	"encoding/json"
	"log/slog"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"google.golang.org/protobuf/proto"
)

// SchemaVersion 固定 SDK 调用封装的版本。
const SchemaVersion = 1

// Session 复用根包的就绪门禁，在 sdk 初始化时设置。
var Session func(context.Context) (*wire.Session, error)

// Call 复用根包的父调用关联与公开错误转换。
var Call func(context.Context, *wire.Session, string, wire.Payload) (wire.Payload, error)

// Decode 校验 Protobuf 响应的编码、Schema 和版本。
var Decode func(wire.Payload, proto.Message) error

// Logger 返回所属组件的日志出口。
var Logger func(context.Context) *slog.Logger

// Runtime 读取已通过就绪门禁的组件身份与 Schema。
var Runtime func(context.Context) (pluginID, entry string, schemas map[string][]byte)

// Parent 返回 Handler 的入站调用和当前有效状态；后台返回 nil。
var Parent func(context.Context) (*wire.Inbound, bool)

// Action 连接后台平台动作与根包的发送选项；options 保持 sdk.SendOption 的原始切片。
var Action func(context.Context, string, string, uint32, any, any) (json.RawMessage, error)

// ProtoSpec 保存 SDK 内部固定服务的目标与消息类型。
type ProtoSpec[In, Out any] struct{ target string }

// Proto 声明 SDK 内部的 Protobuf 固定服务。
func Proto[In, Out any](target string) ProtoSpec[In, Out] { return ProtoSpec[In, Out]{target: target} }

// NewMessage 按 Protobuf 类型创建新的输出消息。
func NewMessage[T any]() proto.Message {
	var message T
	return any(message).(proto.Message).ProtoReflect().Type().New().Interface()
}

// Call 沿用 SDK 的就绪门禁与父调用关联。
func (s ProtoSpec[In, Out]) Call(ctx context.Context, in In) (Out, error) {
	var zero Out
	session, err := Session(ctx)
	if err != nil {
		return zero, err
	}
	input := any(in).(proto.Message)
	encoded, err := proto.Marshal(input)
	if err != nil {
		return zero, &errcode.Error{Code: errcode.InvalidArgument, Message: "encode service input: " + err.Error()}
	}
	response, err := Call(ctx, session, s.target, wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(input), SchemaVersion: SchemaVersion, Bytes: encoded})
	if err != nil {
		return zero, err
	}
	out := NewMessage[Out]()
	if err := Decode(response, out); err != nil {
		return zero, &errcode.Error{Code: errcode.Internal, Message: "decode service output: " + err.Error()}
	}
	return out.(Out), nil
}
