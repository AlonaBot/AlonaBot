package contracts_test

import (
	"strings"
	"testing"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"google.golang.org/protobuf/proto"
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
		t.Fatalf("Envelope has %d fields, WIRE specifies %d", got, len(want)-1)
	}
	for num := 1; num < len(want); num++ {
		f := fields.ByNumber(protoreflect.FieldNumber(num))
		if f == nil || string(f.Name()) != want[num] {
			t.Errorf("Envelope field %d should be %s, got %v", num, want[num], f)
			continue
		}
		// 6/7 用 optional 区分缺失与零值；16 是消息字段，天然有 presence。
		wantPresence := num == 6 || num == 7 || num == 16
		if f.HasPresence() != wantPresence {
			t.Errorf("Envelope.%s presence = %v, want %v", f.Name(), f.HasPresence(), wantPresence)
		}
		if f.ContainingOneof() != nil && !f.ContainingOneof().IsSynthetic() {
			t.Errorf("Envelope.%s must not be in a oneof", f.Name())
		}
	}
}

// ErrorInfo 字段 1～3 的编号与名称和 WIRE §2 一致。
func TestErrorInfoFieldNumbers(t *testing.T) {
	fields := (&pb.ErrorInfo{}).ProtoReflect().Descriptor().Fields()
	for num, name := range map[int]string{1: "code", 2: "safe_message", 3: "trace_id"} {
		if f := fields.ByNumber(protoreflect.FieldNumber(num)); f == nil || string(f.Name()) != name {
			t.Errorf("ErrorInfo field %d should be %s, got %v", num, name, f)
		}
	}
	if fields.Len() != 3 {
		t.Errorf("ErrorInfo has %d fields, want 3", fields.Len())
	}
}

// TestButtonCoreTokenFieldNumber 固定 Core 回调令牌的字段编号、类型与 optional presence。
func TestButtonCoreTokenFieldNumber(t *testing.T) {
	field := (&pb.Button{}).ProtoReflect().Descriptor().Fields().ByNumber(4)
	if field == nil || field.Name() != "core_token" || field.Kind() != protoreflect.StringKind || !field.HasOptionalKeyword() || !field.HasPresence() || field.IsList() {
		t.Errorf("Button field 4 should be optional string core_token, got %v", field)
	}
}

// TestExtensionFieldNumbers 固定 extension 分支与载荷的编号、类型和 oneof 归属。
func TestExtensionFieldNumbers(t *testing.T) {
	field := (&pb.Segment{}).ProtoReflect().Descriptor().Fields().ByNumber(9)
	if field == nil || field.Name() != "extension" || field.Kind() != protoreflect.MessageKind || field.IsList() || field.Message().FullName() != "alonabot.v1.ExtensionSegment" {
		t.Fatalf("Segment field 9 should be ExtensionSegment extension, got %v", field)
	}
	if oneof := field.ContainingOneof(); oneof == nil || oneof.Name() != "data" {
		t.Errorf("extension oneof = %v, want data", oneof)
	}
	names := []protoreflect.Name{"adapter_id", "name", "version", "payload_json"}
	kinds := []protoreflect.Kind{protoreflect.StringKind, protoreflect.StringKind, protoreflect.Uint32Kind, protoreflect.StringKind}
	fields := (&pb.ExtensionSegment{}).ProtoReflect().Descriptor().Fields()
	if fields.Len() != len(names) {
		t.Fatalf("ExtensionSegment has %d fields, want %d", fields.Len(), len(names))
	}
	for i, name := range names {
		f := fields.ByNumber(protoreflect.FieldNumber(i + 1))
		if f == nil || f.Name() != name || f.Kind() != kinds[i] || f.IsList() || f.HasPresence() || f.ContainingOneof() != nil {
			t.Errorf("ExtensionSegment field %d should be singular %s %s, got %v", i+1, kinds[i], name, f)
		}
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
		{pb.AdapterInstanceState(0).Descriptor(), "ADAPTER_INSTANCE_STATE_", []string{"UNSPECIFIED", "FAILED"}},
		{pb.BotOfflineReason(0).Descriptor(), "BOT_OFFLINE_REASON_", []string{"UNSPECIFIED", "STOPPED", "ADAPTER_EXITED", "INSTANCE_FAILED"}},
		{pb.TextUnit(0).Descriptor(), "TEXT_UNIT_", []string{"UNSPECIFIED", "UTF16_CODE_UNITS", "UNICODE_CODE_POINTS", "BYTES"}},
		{pb.SendStatus(0).Descriptor(), "SEND_STATUS_", []string{"UNSPECIFIED", "ACCEPTED", "PARTIAL", "REJECTED", "UNKNOWN"}},
		{pb.PartStatus(0).Descriptor(), "PART_STATUS_", []string{"UNSPECIFIED", "ACCEPTED", "REJECTED", "UNKNOWN", "NOT_ATTEMPTED"}},
		// MESSAGE-MODEL §3。
		{pb.Propagation(0).Descriptor(), "PROPAGATION_", []string{"UNSPECIFIED", "CONTINUE", "STOP"}},
		// WIRE §3.3 的 kind 与 CONTRACTS §2.3 的 Hook phase。
		{pb.RouteKind(0).Descriptor(), "ROUTE_KIND_", []string{
			"UNSPECIFIED", "COMMAND", "EVENT", "HOOK", "SERVICE", "SCHEDULE", "HTTP", "BUS_SUBSCRIPTION"}},
		{pb.HookPhase(0).Descriptor(), "HOOK_PHASE_", []string{"UNSPECIFIED", "PRE_EVENT", "BEFORE_SEND", "AFTER_SEND"}},
		{pb.ParamType(0).Descriptor(), "PARAM_TYPE_", []string{"UNSPECIFIED", "STRING", "INT", "FLOAT", "BOOL", "USER", "IMAGE", "MESSAGE"}},
		{pb.CommandMatch(0).Descriptor(), "COMMAND_MATCH_", []string{"UNSPECIFIED", "PATH", "REGEX", "RAW", "KEYWORD", "STARTSWITH", "ENDSWITH", "FULLMATCH"}},
		// WIRE §3.2 的 log_level。
		{pb.LogLevel(0).Descriptor(), "LOG_LEVEL_", []string{"UNSPECIFIED", "DEBUG", "INFO", "WARN", "ERROR", "OFF"}},
		// CONTRACTS §13.1、WIRE §3.5 的工具策略。
		{pb.ToolPolicy(0).Descriptor(), "TOOL_POLICY_", []string{"UNSPECIFIED", "ALLOW", "CONFIRM", "DENY"}},
	}
	for _, c := range cases {
		vals := c.enum.Values()
		if vals.Len() != len(c.values) {
			t.Errorf("%s has %d values, WIRE specifies %d", c.enum.Name(), vals.Len(), len(c.values))
		}
		for num, name := range c.values {
			v := vals.ByNumber(protoreflect.EnumNumber(num))
			if v == nil || string(v.Name()) != c.prefix+name {
				t.Errorf("%s value %d should be %s%s, got %v", c.enum.Name(), num, c.prefix, name, v)
			}
		}
	}
}

// TestCommandSpecFieldNumbers 固定命令描述及追加文本匹配字段的编号、类型与重复性。
func TestCommandSpecFieldNumbers(t *testing.T) {
	names := []protoreflect.Name{"match", "path", "aliases", "params", "handler_mode", "rule", "regex", "usage_mode", "", "patterns", "ignore_case", "usage_key"}
	kinds := []protoreflect.Kind{protoreflect.EnumKind, protoreflect.StringKind, protoreflect.MessageKind, protoreflect.MessageKind, protoreflect.EnumKind, protoreflect.MessageKind, protoreflect.StringKind, protoreflect.EnumKind, protoreflect.StringKind, protoreflect.StringKind, protoreflect.BoolKind, protoreflect.StringKind}
	fields := (&pb.CommandSpec{}).ProtoReflect().Descriptor().Fields()
	if fields.Len() != len(names)-1 {
		t.Fatalf("CommandSpec has %d fields, want %d", fields.Len(), len(names)-1)
	}
	for i, name := range names {
		f := fields.ByNumber(protoreflect.FieldNumber(i + 1))
		if name == "" {
			if f != nil {
				t.Errorf("CommandSpec field %d must be absent, got %v", i+1, f)
			}
			continue
		}
		if f == nil || f.Name() != name || f.Kind() != kinds[i] {
			t.Errorf("CommandSpec field %d should be %s %s, got %v", i+1, kinds[i], name, f)
			continue
		}
		repeated := name == "path" || name == "aliases" || name == "params" || name == "patterns"
		if f.IsList() != repeated {
			t.Errorf("CommandSpec.%s list = %v, want %v", name, f.IsList(), repeated)
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
		t.Fatalf("only %d enums were checked; the registry scan may be broken", n)
	}
}

// checkEnum 检查枚举零值名为 <类型前缀>_UNSPECIFIED。
func checkEnum(t *testing.T, e protoreflect.EnumDescriptor) {
	t.Helper()
	prefix := screamingSnake(string(e.Name())) + "_"
	zero := e.Values().ByNumber(0)
	if zero == nil || string(zero.Name()) != prefix+"UNSPECIFIED" {
		t.Errorf("zero value of %s should be %sUNSPECIFIED, got %v", e.FullName(), prefix, zero)
	}
	for j := 0; j < e.Values().Len(); j++ {
		if n := string(e.Values().Get(j).Name()); !strings.HasPrefix(n, prefix) {
			t.Errorf("%s value %s is missing prefix %s", e.FullName(), n, prefix)
		}
	}
}

// Service 调用必须同时携带可信上下文与业务载荷（WIRE §4）。
func TestServiceInvocationCarriesContext(t *testing.T) {
	fields := (&pb.ServiceInvocation{}).ProtoReflect().Descriptor().Fields()
	ctx := fields.ByName("context")
	if ctx == nil || ctx.Message().FullName() != "alonabot.v1.InvocationContext" {
		t.Errorf("ServiceInvocation.context should be InvocationContext, got %v", ctx)
	}
	for _, name := range []protoreflect.Name{"payload", "encoding", "schema_id", "schema_version"} {
		if fields.ByName(name) == nil {
			t.Errorf("ServiceInvocation is missing %s", name)
		}
	}
}

// HandlerOutcome 的字段与 MESSAGE-MODEL §3 对应；error 放在 Envelope，不在此消息内。
func TestHandlerOutcomeFields(t *testing.T) {
	fields := (&pb.HandlerOutcome{}).ProtoReflect().Descriptor().Fields()
	for _, name := range []protoreflect.Name{"matched", "propagation", "chain"} {
		if fields.ByName(name) == nil {
			t.Errorf("HandlerOutcome is missing %s", name)
		}
	}
	if fields.Len() != 3 {
		t.Errorf("HandlerOutcome has %d fields, want 3", fields.Len())
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

// SessionLimits 的字段名称、类型与编号遵循 WIRE §3.2。
func TestSessionLimitsFieldNumbers(t *testing.T) {
	fields := (&pb.SessionLimits{}).ProtoReflect().Descriptor().Fields()
	names := []string{"max_inflight", "max_payload_bytes", "request_timeout_ms"}
	if fields.Len() != len(names) {
		t.Fatalf("SessionLimits has %d fields, want %d", fields.Len(), len(names))
	}
	for i, name := range names {
		field := fields.ByNumber(protoreflect.FieldNumber(i + 1))
		if field == nil || string(field.Name()) != name || field.Kind() != protoreflect.Uint32Kind {
			t.Errorf("SessionLimits field %d should be uint32 %s, got %v", i+1, name, field)
		}
	}
}

// Bootstrap.session_limits 固定为字段 11，并引用 SessionLimits（WIRE §3.2）。
func TestBootstrapSessionLimitsFieldNumber(t *testing.T) {
	field := (&pb.Bootstrap{}).ProtoReflect().Descriptor().Fields().ByNumber(11)
	if field == nil || field.Name() != "session_limits" || field.Kind() != protoreflect.MessageKind || field.Message().FullName() != "alonabot.v1.SessionLimits" {
		t.Errorf("Bootstrap field 11 should be SessionLimits session_limits, got %v", field)
	}
}

// TestAdapterFieldNumbers 固定 Adapter 载荷与追加字段的线上编号。
func TestAdapterFieldNumbers(t *testing.T) {
	cases := map[string][]string{
		"Capability":                   {"name", "version", "supported", "text_limit"},
		"TextLimit":                    {"max", "unit"},
		"AdapterDescriptor":            {"adapter_id", "capabilities", "extensions", "permissions", "events"},
		"PermissionDeclaration":        {"name", "implies", "label_key"},
		"EventDeclaration":             {"name", "version", "schema_json"},
		"AdapterEventPayload":          {"version", "payload_json"},
		"ExtensionDeclaration":         {"name", "version", "schema_json"},
		"AdapterInstanceStartRequest":  {"instance_id", "config_json", "cursor"},
		"AdapterInstanceStartResponse": {"account_id", "account_display_name", "capabilities", "account_username"},
		"AdapterInstanceStopRequest":   {"instance_id"},
		"AdapterInstanceStopResponse":  {},
		"AdapterSendRequest":           {"instance_id", "target", "chain"},
		"EditRequest":                  {"to", "platform_message_id", "chain"},
		"AdapterEditRequest":           {"instance_id", "target", "platform_message_id", "chain"},
		"AdapterDeliverRequest":        {"instance_id", "event", "cursor"},
		"AdapterDeliverResponse":       {"accepted"},
		"AdapterCursorRequest":         {"instance_id", "cursor"},
		"AdapterCursorResponse":        {},
		"AdapterStatusRequest":         {"instance_id", "state", "safe_error"},
		"AdapterStatusResponse":        {},
		"AdapterMediaReadRequest":      {"instance_id", "file_id", "offset", "length"},
		"AdapterMediaReadResponse":     {"data", "next_offset", "eof"},
		"BotOnlinePayload":             {"adapter_id", "adapter_instance_id", "account_id", "account_display_name"},
		"BotOfflinePayload":            {"adapter_id", "adapter_instance_id", "account_id", "reason", "safe_error"},
	}
	for name, names := range cases {
		d, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName("alonabot.v1." + name))
		if err != nil {
			t.Fatal(err)
		}
		fields := d.(protoreflect.MessageDescriptor).Fields()
		if fields.Len() != len(names) {
			t.Errorf("%s field count = %d, want %d", name, fields.Len(), len(names))
		}
		for i, name := range names {
			f := fields.ByNumber(protoreflect.FieldNumber(i + 1))
			if f == nil || string(f.Name()) != name {
				t.Errorf("%s field %d should be %s, got %v", d.FullName(), i+1, name, f)
			}
		}
	}
	for _, c := range []struct {
		fields protoreflect.FieldDescriptors
		number protoreflect.FieldNumber
		name   protoreflect.Name
	}{
		{(&pb.Register{}).ProtoReflect().Descriptor().Fields(), 7, "adapter"},
		{(&pb.EventPayload{}).ProtoReflect().Descriptor().Fields(), 9, "bot_online"},
		{(&pb.EventPayload{}).ProtoReflect().Descriptor().Fields(), 10, "bot_offline"},
	} {
		f := c.fields.ByNumber(c.number)
		if f == nil || f.Name() != c.name {
			t.Errorf("field %d should be %s, got %v", c.number, c.name, f)
		}
	}
}

// TestTelegramMetadataFieldNumbers 固定入站语言与消息指向字段的编号和 presence。
func TestTelegramMetadataFieldNumbers(t *testing.T) {
	for _, c := range []struct {
		message  protoreflect.MessageDescriptor
		name     protoreflect.Name
		number   protoreflect.FieldNumber
		kind     protoreflect.Kind
		presence bool
	}{
		{(&pb.Actor{}).ProtoReflect().Descriptor(), "locale", 5, protoreflect.StringKind, true},
		{(&pb.MessageCreatedPayload{}).ProtoReflect().Descriptor(), "to_me", 3, protoreflect.BoolKind, false},
	} {
		f := c.message.Fields().ByNumber(c.number)
		if f == nil || f.Name() != c.name || f.Kind() != c.kind || f.HasPresence() != c.presence {
			t.Errorf("%s field %d = %v, want %s (%s, presence=%v)", c.message.Name(), c.number, f, c.name, c.kind, c.presence)
		}
	}
}

// 会话消息固定字段编号、类型、重复性与分页游标的 presence。
func TestConversationFieldNumbers(t *testing.T) {
	cases := []struct {
		message protoreflect.Message
		names   []protoreflect.Name
		kinds   []protoreflect.Kind
	}{
		{(&pb.ConversationAppendRequest{}).ProtoReflect(), []protoreflect.Name{"conversation_key", "entries_json"}, []protoreflect.Kind{protoreflect.StringKind, protoreflect.StringKind}},
		{(&pb.ConversationAppendResponse{}).ProtoReflect(), []protoreflect.Name{"last_seq"}, []protoreflect.Kind{protoreflect.Uint64Kind}},
		{(&pb.ConversationListRequest{}).ProtoReflect(), []protoreflect.Name{"conversation_key", "from_seq"}, []protoreflect.Kind{protoreflect.StringKind, protoreflect.Uint64Kind}},
		{(&pb.ConversationEntry{}).ProtoReflect(), []protoreflect.Name{"seq", "entry_json"}, []protoreflect.Kind{protoreflect.Uint64Kind, protoreflect.StringKind}},
		{(&pb.ConversationListResponse{}).ProtoReflect(), []protoreflect.Name{"entries", "next_seq"}, []protoreflect.Kind{protoreflect.MessageKind, protoreflect.Uint64Kind}},
		{(&pb.ConversationClearRequest{}).ProtoReflect(), []protoreflect.Name{"conversation_key"}, []protoreflect.Kind{protoreflect.StringKind}},
		{(&pb.ConversationClearResponse{}).ProtoReflect(), nil, nil},
	}
	for _, c := range cases {
		t.Run(string(c.message.Descriptor().Name()), func(t *testing.T) {
			fields := c.message.Descriptor().Fields()
			if fields.Len() != len(c.names) {
				t.Fatalf("field count = %d, want %d", fields.Len(), len(c.names))
			}
			for i, name := range c.names {
				f := fields.ByNumber(protoreflect.FieldNumber(i + 1))
				if f == nil || f.Name() != name || f.Kind() != c.kinds[i] {
					t.Fatalf("field %d should be %s %s, got %v", i+1, c.kinds[i], name, f)
				}
				repeated := name == "entries" || name == "entries_json"
				wantCardinality := protoreflect.Optional
				if repeated {
					wantCardinality = protoreflect.Repeated
				}
				if f.Cardinality() != wantCardinality {
					t.Errorf("%s cardinality = %v, want %v", name, f.Cardinality(), wantCardinality)
				}
				if f.HasPresence() != (name == "next_seq") {
					t.Errorf("%s presence = %v", name, f.HasPresence())
				}
				if name == "entries" && f.Message().FullName() != "alonabot.v1.ConversationEntry" {
					t.Errorf("entries type = %s, want ConversationEntry", f.Message().FullName())
				}
			}
		})
	}
}

// TestEnvelopeSchemaVersions 区分业务版本和固定控制帧版本（WIRE §1）。
func TestEnvelopeSchemaVersions(t *testing.T) {
	for _, kind := range []pb.FrameKind{pb.FrameKind_FRAME_KIND_REQUEST, pb.FrameKind_FRAME_KIND_RESPONSE, pb.FrameKind_FRAME_KIND_HEARTBEAT} {
		for _, version := range []uint32{0, 1, 2} {
			env := &pb.Envelope{ProtocolMajor: 1, SessionId: make([]byte, 16), Kind: kind, PayloadEncoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaId: "test.service.v1", SchemaVersion: version}
			if kind == pb.FrameKind_FRAME_KIND_HEARTBEAT {
				env.SchemaId = validate.SchemaID(&pb.Heartbeat{})
			} else {
				env.InitiatorSide = pb.InitiatorSide_INITIATOR_SIDE_PLUGIN
				env.CallId = 1
			}
			if kind == pb.FrameKind_FRAME_KIND_REQUEST {
				env.Target = "a/echo"
				env.DeadlineUnixMs = proto.Int64(1)
			}
			wantValid := version >= 1 && (kind != pb.FrameKind_FRAME_KIND_HEARTBEAT || version == 1)
			err := validate.Envelope(env)
			if (err == nil) != wantValid {
				t.Errorf("kind=%s version=%d: error=%v want_valid=%v", kind, version, err, wantValid)
			}
		}
	}
}

// TestPermissionProtocolFields 固定权限、扩展事件和片段参数字段的编号、类型与 oneof 归属。
func TestPermissionProtocolFields(t *testing.T) {
	for _, c := range []struct {
		message     proto.Message
		number      protoreflect.FieldNumber
		name        protoreflect.Name
		kind        protoreflect.Kind
		repeated    bool
		oneof       protoreflect.Name
		messageName protoreflect.FullName
	}{
		{&pb.Actor{}, 6, "permissions", protoreflect.StringKind, true, "", ""},
		{&pb.RouteDescriptor{}, 17, "required_permissions", protoreflect.StringKind, true, "", ""},
		{&pb.CommandSpec{}, 12, "usage_key", protoreflect.StringKind, false, "", ""},
		{&pb.AdapterDescriptor{}, 4, "permissions", protoreflect.MessageKind, true, "", "alonabot.v1.PermissionDeclaration"},
		{&pb.AdapterDescriptor{}, 5, "events", protoreflect.MessageKind, true, "", "alonabot.v1.EventDeclaration"},
		{&pb.PermissionDeclaration{}, 1, "name", protoreflect.StringKind, false, "", ""},
		{&pb.PermissionDeclaration{}, 2, "implies", protoreflect.StringKind, true, "", ""},
		{&pb.PermissionDeclaration{}, 3, "label_key", protoreflect.StringKind, false, "", ""},
		{&pb.EventDeclaration{}, 1, "name", protoreflect.StringKind, false, "", ""},
		{&pb.EventDeclaration{}, 2, "version", protoreflect.Uint32Kind, false, "", ""},
		{&pb.EventDeclaration{}, 3, "schema_json", protoreflect.StringKind, false, "", ""},
		{&pb.EventPayload{}, 11, "adapter_event", protoreflect.MessageKind, false, "kind", "alonabot.v1.AdapterEventPayload"},
		{&pb.AdapterEventPayload{}, 1, "version", protoreflect.Uint32Kind, false, "", ""},
		{&pb.AdapterEventPayload{}, 2, "payload_json", protoreflect.StringKind, false, "", ""},
		{&pb.ArgValue{}, 6, "user_id", protoreflect.StringKind, false, "value", ""},
		{&pb.ArgValue{}, 7, "image", protoreflect.MessageKind, false, "value", "alonabot.v1.ImageSegment"},
		{&pb.ArgValue{}, 8, "message", protoreflect.MessageKind, false, "value", "alonabot.v1.MessageChain"},
	} {
		d := c.message.ProtoReflect().Descriptor()
		t.Run(string(d.Name())+"."+string(c.name), func(t *testing.T) {
			f := d.Fields().ByNumber(c.number)
			if f == nil || f.Name() != c.name || f.Kind() != c.kind || f.IsList() != c.repeated {
				t.Fatalf("field %d = %v, want %s (%s, repeated=%v)", c.number, f, c.name, c.kind, c.repeated)
			}
			var oneof protoreflect.Name
			if o := f.ContainingOneof(); o != nil {
				oneof = o.Name()
			}
			if oneof != c.oneof {
				t.Errorf("oneof = %q, want %q", oneof, c.oneof)
			}
			if c.messageName != "" && f.Message().FullName() != c.messageName {
				t.Errorf("message type = %s, want %s", f.Message().FullName(), c.messageName)
			}
		})
	}
}

// TestRemovedPermissionFieldsReserved 固定已删除字段的编号与名称保留，检查旧枚举已删除。
func TestRemovedPermissionFieldsReserved(t *testing.T) {
	for _, c := range []struct {
		message proto.Message
		number  protoreflect.FieldNumber
		name    protoreflect.Name
	}{
		{&pb.Actor{}, 4, "trusted_roles"},
		{&pb.RouteDescriptor{}, 5, "required_permission"},
		{&pb.RuleDescriptor{}, 2, "require_chat_admin"},
		{&pb.CommandSpec{}, 9, "custom_usage"},
	} {
		d := c.message.ProtoReflect().Descriptor()
		if d.Fields().ByNumber(c.number) != nil || d.Fields().ByName(c.name) != nil || !d.ReservedRanges().Has(c.number) || !d.ReservedNames().Has(c.name) {
			t.Errorf("%s field %s (%d) must be removed and reserved", d.Name(), c.name, c.number)
		}
	}
	for _, name := range []protoreflect.FullName{"alonabot.v1.TrustedRole", "alonabot.v1.Permission"} {
		if _, err := protoregistry.GlobalFiles.FindDescriptorByName(name); err != protoregistry.NotFound {
			t.Errorf("removed enum %s: descriptor lookup error = %v, want NotFound", name, err)
		}
	}
}

// TestBootstrapDefaultLocaleFieldNumber 固定默认语言字段的线上编号与类型。
func TestBootstrapDefaultLocaleFieldNumber(t *testing.T) {
	field := (&pb.Bootstrap{}).ProtoReflect().Descriptor().Fields().ByNumber(12)
	if field == nil || field.Name() != "default_locale" || field.Kind() != protoreflect.StringKind {
		t.Errorf("Bootstrap field 12 should be string default_locale, got %v", field)
	}
}

// TestToolFieldNumbers 固定 Service 工具描述与目录载荷的编号、类型和重复性。
func TestToolFieldNumbers(t *testing.T) {
	for _, c := range []struct {
		message proto.Message
		names   []protoreflect.Name
		kinds   []protoreflect.Kind
	}{
		{&pb.ServiceSpec{}, []protoreflect.Name{"service_name", "access", "allowlist_plugin_ids", "encoding", "tool"}, []protoreflect.Kind{protoreflect.StringKind, protoreflect.EnumKind, protoreflect.StringKind, protoreflect.EnumKind, protoreflect.MessageKind}},
		{&pb.ToolSpec{}, []protoreflect.Name{"description", "default_policy", "input_schema_json"}, []protoreflect.Kind{protoreflect.StringKind, protoreflect.EnumKind, protoreflect.StringKind}},
		{&pb.ToolInfo{}, []protoreflect.Name{"name", "description", "input_schema_json", "policy"}, []protoreflect.Kind{protoreflect.StringKind, protoreflect.StringKind, protoreflect.StringKind, protoreflect.EnumKind}},
		{&pb.ToolListRequest{}, nil, nil},
		{&pb.ToolListResponse{}, []protoreflect.Name{"tools"}, []protoreflect.Kind{protoreflect.MessageKind}},
	} {
		t.Run(string(c.message.ProtoReflect().Descriptor().Name()), func(t *testing.T) {
			fields := c.message.ProtoReflect().Descriptor().Fields()
			if fields.Len() != len(c.names) {
				t.Fatalf("field count = %d, want %d", fields.Len(), len(c.names))
			}
			for i, name := range c.names {
				f := fields.ByNumber(protoreflect.FieldNumber(i + 1))
				repeated := name == "tools" || name == "allowlist_plugin_ids"
				if f == nil || f.Name() != name || f.Kind() != c.kinds[i] || f.IsList() != repeated {
					t.Fatalf("field %d = %v, want %s (%s, repeated=%v)", i+1, f, name, c.kinds[i], repeated)
				}
				switch name {
				case "tool", "tools":
					want := protoreflect.FullName("alonabot.v1.ToolSpec")
					if name == "tools" {
						want = "alonabot.v1.ToolInfo"
					}
					if f.Message().FullName() != want {
						t.Errorf("%s message = %s, want %s", name, f.Message().FullName(), want)
					}
				case "default_policy", "policy":
					if f.Enum().FullName() != "alonabot.v1.ToolPolicy" {
						t.Errorf("%s enum = %s, want ToolPolicy", name, f.Enum().FullName())
					}
				}
			}
		})
	}
}
