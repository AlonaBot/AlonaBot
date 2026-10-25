package contracts_test

import (
	"strings"
	"testing"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// 以下期望值逐字取自 spec/WIRE-FORMAT.md §1/§2；改动这里等于改动线上协议。

// Envelope 字段 1～16 的编号与名称和 WIRE §1 一致，且没有多余字段。
func TestEnvelopeFieldNumbers(t *testing.T) {
	want := []string{
		1: "protocol_major", 2: "protocol_minor", 3: "session_id", 4: "initiator_side",
		5: "call_id", 6: "parent_call_id", 7: "deadline_unix_ms", 8: "kind",
		9: "source_plugin_id", 10: "target", 11: "trace_id", 12: "payload_encoding",
		13: "schema_id", 14: "schema_version", 15: "payload", 16: "error",
	}
	fields := (&pb.Envelope{}).ProtoReflect().Descriptor().Fields()
	if got := fields.Len(); got != len(want)-1 {
		t.Fatalf("Envelope 字段数 = %d，WIRE 规定 %d", got, len(want)-1)
	}
	for num := 1; num < len(want); num++ {
		f := fields.ByNumber(protoreflect.FieldNumber(num))
		if f == nil || string(f.Name()) != want[num] {
			t.Errorf("Envelope 字段 %d 应为 %s，实际 %v", num, want[num], f)
			continue
		}
		// 6/7 用 optional 区分缺失与零值；16 是消息字段，天然有 presence。
		wantPresence := num == 6 || num == 7 || num == 16
		if f.HasPresence() != wantPresence {
			t.Errorf("Envelope.%s presence = %v，应为 %v", f.Name(), f.HasPresence(), wantPresence)
		}
		if f.ContainingOneof() != nil && !f.ContainingOneof().IsSynthetic() {
			t.Errorf("Envelope.%s 不得放进 oneof", f.Name())
		}
	}
}

// ErrorInfo 字段 1～3 的编号与名称和 WIRE §2 一致。
func TestErrorInfoFieldNumbers(t *testing.T) {
	fields := (&pb.ErrorInfo{}).ProtoReflect().Descriptor().Fields()
	for num, name := range map[int]string{1: "code", 2: "safe_message", 3: "trace_id"} {
		if f := fields.ByNumber(protoreflect.FieldNumber(num)); f == nil || string(f.Name()) != name {
			t.Errorf("ErrorInfo 字段 %d 应为 %s，实际 %v", num, name, f)
		}
	}
	if fields.Len() != 3 {
		t.Errorf("ErrorInfo 字段数 = %d，应为 3", fields.Len())
	}
}

// 规格列出取值的枚举逐项一致；全部已注册枚举的零值都是 *_UNSPECIFIED。
func TestEnumValues(t *testing.T) {
	cases := []struct {
		enum   protoreflect.EnumDescriptor
		prefix string // 枚举值名称的类型前缀
		// values 按数值顺序列出不含前缀的名称，下标即枚举数值。
		values []string
	}{
		{pb.InitiatorSide(0).Descriptor(), "INITIATOR_SIDE_", []string{"UNSPECIFIED", "CORE", "PLUGIN"}},
		{pb.FrameKind(0).Descriptor(), "FRAME_KIND_", []string{
			"UNSPECIFIED", "HELLO", "HELLO_ACK", "REGISTER", "PREPARED", "ACTIVATE", "ACTIVATED",
			"REQUEST", "RESPONSE", "CANCEL", "EVENT", "STOPPING", "HEARTBEAT", "SESSION_ERROR"}},
		{pb.PayloadEncoding(0).Descriptor(), "PAYLOAD_ENCODING_", []string{"UNSPECIFIED", "PROTOBUF", "JSON"}},
		{pb.ErrorCode(0).Descriptor(), "ERROR_CODE_", []string{
			"UNSPECIFIED", "INVALID_ARGUMENT", "UNAUTHENTICATED", "PERMISSION_DENIED", "NOT_FOUND",
			"ALREADY_EXISTS", "FAILED_PRECONDITION", "UNSUPPORTED", "RESOURCE_EXHAUSTED",
			"DEADLINE_EXCEEDED", "CANCELLED", "UNAVAILABLE", "INTERNAL", "PROTOCOL_ERROR",
			"CALL_CYCLE", "PLUGIN_NOT_READY", "NO_DEFAULT_TARGET", "NEEDS_MANUAL_MIGRATION"}},
		// MESSAGE-MODEL §4。
		{pb.SendStatus(0).Descriptor(), "SEND_STATUS_", []string{"UNSPECIFIED", "ACCEPTED", "PARTIAL", "REJECTED", "UNKNOWN"}},
		{pb.PartStatus(0).Descriptor(), "PART_STATUS_", []string{"UNSPECIFIED", "ACCEPTED", "REJECTED", "UNKNOWN", "NOT_ATTEMPTED"}},
		// MESSAGE-MODEL §3。
		{pb.Propagation(0).Descriptor(), "PROPAGATION_", []string{"UNSPECIFIED", "CONTINUE", "STOP"}},
		// WIRE §3.3 的 kind 与 CONTRACTS §2.3 的 Hook phase。
		{pb.RouteKind(0).Descriptor(), "ROUTE_KIND_", []string{
			"UNSPECIFIED", "COMMAND", "EVENT", "HOOK", "SERVICE", "SCHEDULE", "HTTP", "BUS_SUBSCRIPTION"}},
		{pb.HookPhase(0).Descriptor(), "HOOK_PHASE_", []string{"UNSPECIFIED", "PRE_EVENT", "BEFORE_SEND", "AFTER_SEND"}},
	}
	for _, c := range cases {
		vals := c.enum.Values()
		if vals.Len() != len(c.values) {
			t.Errorf("%s 取值数 = %d，WIRE 规定 %d", c.enum.Name(), vals.Len(), len(c.values))
		}
		for num, name := range c.values {
			v := vals.ByNumber(protoreflect.EnumNumber(num))
			if v == nil || string(v.Name()) != c.prefix+name {
				t.Errorf("%s 值 %d 应为 %s%s，实际 %v", c.enum.Name(), num, c.prefix, name, v)
			}
		}
	}
}

// 本包所有枚举的零值必须是 *_UNSPECIFIED，且每个值都带枚举类型前缀（含嵌套在消息内的枚举）。
func TestAllEnumsHaveUnspecifiedZero(t *testing.T) {
	_ = pb.File_alonabot_v1_envelope_proto // 确保本包已注册
	n := 0
	protoregistry.GlobalFiles.RangeFilesByPackage("alonabot.v1", func(fd protoreflect.FileDescriptor) bool {
		for i := 0; i < fd.Enums().Len(); i++ {
			checkEnum(t, fd.Enums().Get(i))
			n++
		}
		return true
	})
	if n < 20 {
		t.Fatalf("只检查了 %d 个枚举，注册表扫描可能失效", n)
	}
}

// checkEnum 检查枚举零值名为 <类型前缀>_UNSPECIFIED。
func checkEnum(t *testing.T, e protoreflect.EnumDescriptor) {
	t.Helper()
	prefix := screamingSnake(string(e.Name())) + "_"
	zero := e.Values().ByNumber(0)
	if zero == nil || string(zero.Name()) != prefix+"UNSPECIFIED" {
		t.Errorf("%s 零值应为 %sUNSPECIFIED，实际 %v", e.FullName(), prefix, zero)
	}
	for j := 0; j < e.Values().Len(); j++ {
		if n := string(e.Values().Get(j).Name()); !strings.HasPrefix(n, prefix) {
			t.Errorf("%s 的值 %s 缺少前缀 %s", e.FullName(), n, prefix)
		}
	}
}

// Service 调用必须同时携带可信上下文与业务载荷（WIRE §4）。
func TestServiceInvocationCarriesContext(t *testing.T) {
	fields := (&pb.ServiceInvocation{}).ProtoReflect().Descriptor().Fields()
	ctx := fields.ByName("context")
	if ctx == nil || ctx.Message().FullName() != "alonabot.v1.InvocationContext" {
		t.Errorf("ServiceInvocation.context 应为 InvocationContext，实际 %v", ctx)
	}
	for _, name := range []protoreflect.Name{"payload", "encoding", "schema_id", "schema_version"} {
		if fields.ByName(name) == nil {
			t.Errorf("ServiceInvocation 缺少 %s", name)
		}
	}
}

// HandlerOutcome 的字段与 MESSAGE-MODEL §3 对应；error 放在 Envelope，不在此消息内。
func TestHandlerOutcomeFields(t *testing.T) {
	fields := (&pb.HandlerOutcome{}).ProtoReflect().Descriptor().Fields()
	for _, name := range []protoreflect.Name{"matched", "propagation", "chain"} {
		if fields.ByName(name) == nil {
			t.Errorf("HandlerOutcome 缺少 %s", name)
		}
	}
	if fields.Len() != 3 {
		t.Errorf("HandlerOutcome 字段数 = %d，应为 3", fields.Len())
	}
}

// screamingSnake 把 FrameKind 转成 FRAME_KIND。
func screamingSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToUpper(b.String())
}
