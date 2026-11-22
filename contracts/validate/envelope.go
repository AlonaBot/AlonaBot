// Package validate 提供协议 1.0 的纯函数校验。
//
// 它检查 Envelope 结构、控制帧载荷、REGISTER 载荷、MessageChain 结构，以及握手时的版本与能力协商。
//
// 依赖 Session 状态的检查不在这里做，例如 Session Secret、source_plugin_id 是否匹配、
// 父调用是否仍在途；这些由 internal/wire 与 internal/core/host 负责。
package validate

import (
	"encoding/json"
	"fmt"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// 本实现支持的传输协议版本。
const (
	ProtocolMajor uint32 = 1 // 传输协议主版本
	ProtocolMinor uint32 = 0 // 本实现支持的最高次版本
	// SchemaVersion 固定控制帧与 Core 固定载荷的首发版本。
	SchemaVersion uint32 = 1
)

// session_id 与 trace_id 的固定字节长度。
const (
	SessionIDLen = 16
	TraceIDLen   = 16
)

// Error 携带协议错误码，供调用方直接映射为对端可见的 ErrorInfo。
type Error struct {
	Code    pb.ErrorCode // 对端或调用者应看到的协议错误码
	Message string       // 不含敏感信息的说明
}

// Error 返回「错误码: 说明」。
func (e *Error) Error() string { return e.Code.String() + ": " + e.Message }

// errf 用格式化说明构造 *Error。
func errf(code pb.ErrorCode, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

// PublicCode 把未知错误码映射为 INTERNAL（原数字只应保留在内部诊断中）。
func PublicCode(c pb.ErrorCode) pb.ErrorCode {
	if c == pb.ErrorCode_ERROR_CODE_UNSPECIFIED || !knownEnum(c) {
		return pb.ErrorCode_ERROR_CODE_INTERNAL
	}
	return c
}

// knownEnum 报告枚举值是否在当前 Schema 中定义。
func knownEnum(e protoreflect.Enum) bool {
	return e.Descriptor().Values().ByNumber(e.Number()) != nil
}

// SchemaID 返回控制帧与 Core 固定载荷使用的 schema_id：alonabot.<snake_case 消息名>.v1。
func SchemaID(m proto.Message) string {
	return "alonabot." + snake(string(m.ProtoReflect().Descriptor().Name())) + ".v1"
}

// snake 把 CamelCase 消息名转为 snake_case。
func snake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			r += 'a' - 'A'
		}
		b.WriteRune(r)
	}
	return b.String()
}

// controlPayload 给出各控制帧种类的载荷类型；REQUEST/RESPONSE 的载荷由业务 schema 决定。
var controlPayload = map[pb.FrameKind]func() proto.Message{
	pb.FrameKind_FRAME_KIND_HELLO:         func() proto.Message { return &pb.Hello{} },
	pb.FrameKind_FRAME_KIND_HELLO_ACK:     func() proto.Message { return &pb.HelloAck{} },
	pb.FrameKind_FRAME_KIND_REGISTER:      func() proto.Message { return &pb.Register{} },
	pb.FrameKind_FRAME_KIND_PREPARED:      func() proto.Message { return &pb.Prepared{} },
	pb.FrameKind_FRAME_KIND_ACTIVATE:      func() proto.Message { return &pb.Activate{} },
	pb.FrameKind_FRAME_KIND_ACTIVATED:     func() proto.Message { return &pb.Activated{} },
	pb.FrameKind_FRAME_KIND_CANCEL:        func() proto.Message { return &pb.Cancel{} },
	pb.FrameKind_FRAME_KIND_EVENT:         func() proto.Message { return &pb.BusEvent{} },
	pb.FrameKind_FRAME_KIND_STOPPING:      func() proto.Message { return &pb.Stopping{} },
	pb.FrameKind_FRAME_KIND_HEARTBEAT:     func() proto.Message { return &pb.Heartbeat{} },
	pb.FrameKind_FRAME_KIND_SESSION_ERROR: func() proto.Message { return &pb.SessionError{} },
}

// isCall 报告帧种类是否属于调用帧（REQUEST/RESPONSE/CANCEL）。
func isCall(k pb.FrameKind) bool {
	return k == pb.FrameKind_FRAME_KIND_REQUEST || k == pb.FrameKind_FRAME_KIND_RESPONSE || k == pb.FrameKind_FRAME_KIND_CANCEL
}

// Envelope 按帧种类检查结构合法性。失败返回 *Error，code 为 PROTOCOL_ERROR，
// 只有 JSON 载荷语法错误返回 INVALID_ARGUMENT。
func Envelope(env *pb.Envelope) error {
	// 1. 帧种类与主版本。
	k := env.GetKind()
	if k == pb.FrameKind_FRAME_KIND_UNSPECIFIED || !knownEnum(k) {
		return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "unknown or unset frame kind %d", int32(k))
	}
	if env.GetProtocolMajor() != ProtocolMajor {
		return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "protocol_major=%d, only %d is supported", env.GetProtocolMajor(), ProtocolMajor)
	}

	// 2. session_id 与 trace_id 的长度。
	sid := len(env.GetSessionId())
	if k == pb.FrameKind_FRAME_KIND_HELLO && sid != 0 {
		return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "HELLO must not carry session_id")
	}
	if k != pb.FrameKind_FRAME_KIND_HELLO && sid != SessionIDLen {
		return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "session_id length is %d, want %d", sid, SessionIDLen)
	}
	if n := len(env.GetTraceId()); n != 0 && n != TraceIDLen {
		return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "trace_id length is %d, want %d", n, TraceIDLen)
	}

	// 3. 调用帧的 call_id、initiator_side、parent_call_id 与 deadline。
	side := env.GetInitiatorSide()
	if isCall(k) {
		if env.GetCallId() == 0 {
			return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "call_id of %s must not be 0", k)
		}
		if side == pb.InitiatorSide_INITIATOR_SIDE_UNSPECIFIED || !knownEnum(side) {
			return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "%s is missing a valid initiator_side", k)
		}
	} else if env.GetCallId() != 0 {
		return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "call_id of %s must be 0", k)
	}

	if env.ParentCallId != nil {
		if k != pb.FrameKind_FRAME_KIND_REQUEST || side != pb.InitiatorSide_INITIATOR_SIDE_PLUGIN {
			return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "only plugin-initiated REQUEST may set parent_call_id")
		}
		if env.GetParentCallId() == 0 {
			return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "parent_call_id must not be 0")
		}
	}
	if k == pb.FrameKind_FRAME_KIND_REQUEST && env.GetDeadlineUnixMs() <= 0 {
		return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "REQUEST must set deadline_unix_ms")
	}

	// 4. error 与载荷。
	if err := checkError(k, env.GetError()); err != nil {
		return err
	}
	return checkPayload(env)
}

// checkError 校验 error 只出现在 RESPONSE/SESSION_ERROR 上（SESSION_ERROR 必须有），且错误码非零、trace_id 长度合法。
func checkError(k pb.FrameKind, e *pb.ErrorInfo) error {
	switch {
	case e == nil:
		if k == pb.FrameKind_FRAME_KIND_SESSION_ERROR {
			return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "SESSION_ERROR must carry error")
		}
		return nil
	case k != pb.FrameKind_FRAME_KIND_RESPONSE && k != pb.FrameKind_FRAME_KIND_SESSION_ERROR:
		return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "%s must not carry error", k)
	case e.GetCode() == pb.ErrorCode_ERROR_CODE_UNSPECIFIED:
		return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "ErrorInfo.code must not be UNSPECIFIED")
	case len(e.GetTraceId()) != 0 && len(e.GetTraceId()) != TraceIDLen:
		return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "ErrorInfo.trace_id has invalid length")
	}
	return nil
}

// checkPayload 校验载荷相关字段的一致性，以及控制帧载荷的编码与 schema_id。
// 空 proto 消息序列化为 0 字节，所以「是否有载荷」以 schema_id 判断，不以 payload 长度判断。
func checkPayload(env *pb.Envelope) error {
	k := env.GetKind()
	if env.GetSchemaId() == "" {
		if len(env.GetPayload()) != 0 {
			return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "payload is non-empty but schema_id is missing")
		}
		if env.GetPayloadEncoding() != pb.PayloadEncoding_PAYLOAD_ENCODING_UNSPECIFIED || env.GetSchemaVersion() != 0 {
			return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "encoding and schema_version must not be set without schema_id")
		}
		if _, control := controlPayload[k]; control {
			return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "%s is missing its payload", k)
		}
		if k == pb.FrameKind_FRAME_KIND_RESPONSE && env.GetError() == nil {
			return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "RESPONSE has neither a result nor an error")
		}
		return nil
	}

	enc := env.GetPayloadEncoding()
	if enc == pb.PayloadEncoding_PAYLOAD_ENCODING_UNSPECIFIED || !knownEnum(enc) {
		return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "a payload requires a valid payload_encoding")
	}
	if env.GetSchemaVersion() == 0 {
		return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "a payload requires a non-zero schema_version")
	}
	if k != pb.FrameKind_FRAME_KIND_REQUEST && k != pb.FrameKind_FRAME_KIND_RESPONSE && env.GetSchemaVersion() != SchemaVersion {
		return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "schema_version=%d, only %d is supported", env.GetSchemaVersion(), SchemaVersion)
	}
	if newMessage, control := controlPayload[k]; control {
		if enc != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF {
			return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "payload of %s must use PROTOBUF encoding", k)
		}
		if want := SchemaID(newMessage()); env.GetSchemaId() != want {
			return errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "schema_id of %s should be %s, got %s", k, want, env.GetSchemaId())
		}
	}
	if enc == pb.PayloadEncoding_PAYLOAD_ENCODING_JSON && !json.Valid(env.GetPayload()) {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "JSON payload is not valid UTF-8 JSON")
	}
	return nil
}

// DecodeControl 先做 Envelope 校验，再把控制帧载荷解码为对应消息。REQUEST/RESPONSE 不适用。
func DecodeControl(env *pb.Envelope) (proto.Message, error) {
	if err := Envelope(env); err != nil {
		return nil, err
	}
	newMsg, ok := controlPayload[env.GetKind()]
	if !ok {
		return nil, errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "%s is not a control frame", env.GetKind())
	}
	m := newMsg()
	if err := proto.Unmarshal(env.GetPayload(), m); err != nil {
		return nil, errf(pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "cannot decode %s payload: %v", env.GetKind(), err)
	}
	return m, nil
}
