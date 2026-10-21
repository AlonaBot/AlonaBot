package validate

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"testing"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/proto"
)

var update = flag.Bool("update", false, "重写 contracts/testdata/frames 下的 golden 文件")

var (
	sid   = bytes.Repeat([]byte{0xA1}, 16)
	trace = bytes.Repeat([]byte{0xB2}, 16)
)

func mustMarshal(t testing.TB, m proto.Message) []byte {
	t.Helper()
	b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func control(t testing.TB, k pb.FrameKind, m proto.Message) *pb.Envelope {
	env := &pb.Envelope{
		ProtocolMajor: 1, Kind: k, SessionId: sid,
		PayloadEncoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF,
		SchemaId:        SchemaID(m), SchemaVersion: 1, Payload: mustMarshal(t, m),
	}
	if k == pb.FrameKind_FRAME_KIND_HELLO {
		env.SessionId = nil
		env.SourcePluginId = "alonabot_plugin_demo"
	}
	return env
}

func request(side pb.InitiatorSide) *pb.Envelope {
	return &pb.Envelope{
		ProtocolMajor: 1, Kind: pb.FrameKind_FRAME_KIND_REQUEST, SessionId: sid,
		InitiatorSide: side, CallId: 7, DeadlineUnixMs: proto.Int64(1_800_000_000_000),
		Target: "alonabot_plugin_demo/echo", TraceId: trace,
		PayloadEncoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON,
		SchemaId:        "alonabot_plugin_demo.echo.v1", SchemaVersion: 1, Payload: []byte(`{"text":"hi"}`),
	}
}

func partialResponse(t testing.TB) *pb.Envelope {
	result := &pb.SendResult{
		Status:             pb.SendStatus_SEND_STATUS_PARTIAL,
		PlatformMessageIds: []string{"101"},
		PartResults: []*pb.PartResult{
			{PartIndex: 0, Status: pb.PartStatus_PART_STATUS_ACCEPTED, PlatformMessageIds: []string{"101"}},
			{PartIndex: 1, Status: pb.PartStatus_PART_STATUS_UNKNOWN},
		},
	}
	return &pb.Envelope{
		ProtocolMajor: 1, Kind: pb.FrameKind_FRAME_KIND_RESPONSE, SessionId: sid,
		InitiatorSide: pb.InitiatorSide_INITIATOR_SIDE_PLUGIN, CallId: 7, TraceId: trace,
		PayloadEncoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF,
		SchemaId:        SchemaID(result), SchemaVersion: 1, Payload: mustMarshal(t, result),
		Error: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, SafeMessage: "第 2 部分结果未知", TraceId: trace},
	}
}

func code(err error) pb.ErrorCode {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return pb.ErrorCode_ERROR_CODE_UNSPECIFIED
}

func TestEnvelope(t *testing.T) {
	hello := &pb.Hello{PluginId: "alonabot_plugin_demo", Secret: bytes.Repeat([]byte{1}, 32), SupportedMajor: 1}
	const proto_, invalid = pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
	cases := []struct {
		name string
		env  func() *pb.Envelope
		want pb.ErrorCode // UNSPECIFIED 表示应当通过
	}{
		{"合法 HELLO", func() *pb.Envelope { return control(t, pb.FrameKind_FRAME_KIND_HELLO, hello) }, 0},
		{"合法插件 REQUEST（带父调用）", func() *pb.Envelope {
			e := request(pb.InitiatorSide_INITIATOR_SIDE_PLUGIN)
			e.ParentCallId = proto.Uint64(3)
			return e
		}, 0},
		{"RESPONSE 同时带 payload 与 error", func() *pb.Envelope { return partialResponse(t) }, 0},
		{"只有 error 的 RESPONSE", func() *pb.Envelope {
			e := partialResponse(t)
			e.SchemaId, e.SchemaVersion, e.Payload, e.PayloadEncoding = "", 0, nil, 0
			return e
		}, 0},
		{"空消息载荷（0 字节）仍是有效载荷", func() *pb.Envelope {
			e := partialResponse(t)
			e.Error, e.SchemaId, e.Payload = nil, SchemaID(&pb.AfterSendResult{}), nil
			return e
		}, 0},
		{"kind 为 UNSPECIFIED", func() *pb.Envelope { e := request(1); e.Kind = 0; return e }, proto_},
		{"未知 kind", func() *pb.Envelope { e := request(1); e.Kind = 99; return e }, proto_},
		{"Major 不是 1", func() *pb.Envelope { e := request(1); e.ProtocolMajor = 2; return e }, proto_},
		{"HELLO 带 session_id", func() *pb.Envelope {
			e := control(t, pb.FrameKind_FRAME_KIND_HELLO, hello)
			e.SessionId = sid
			return e
		}, proto_},
		{"session_id 长度错误", func() *pb.Envelope { e := request(1); e.SessionId = sid[:8]; return e }, proto_},
		{"REQUEST 的 call_id 为 0", func() *pb.Envelope { e := request(1); e.CallId = 0; return e }, proto_},
		{"REQUEST 的 initiator_side 为 UNSPECIFIED", func() *pb.Envelope { e := request(0); return e }, proto_},
		{"未知 initiator_side", func() *pb.Envelope { e := request(9); return e }, proto_},
		{"控制帧 call_id 非零", func() *pb.Envelope {
			e := control(t, pb.FrameKind_FRAME_KIND_HELLO, hello)
			e.CallId = 1
			return e
		}, proto_},
		{"Core 发起的 REQUEST 设置 parent_call_id", func() *pb.Envelope {
			e := request(pb.InitiatorSide_INITIATOR_SIDE_CORE)
			e.ParentCallId = proto.Uint64(3)
			return e
		}, proto_},
		{"parent_call_id 显式设为 0（字段存在性）", func() *pb.Envelope {
			e := request(pb.InitiatorSide_INITIATOR_SIDE_PLUGIN)
			e.ParentCallId = proto.Uint64(0)
			return e
		}, proto_},
		{"REQUEST 缺 deadline", func() *pb.Envelope { e := request(1); e.DeadlineUnixMs = nil; return e }, proto_},
		{"REQUEST 带 error", func() *pb.Envelope {
			e := request(1)
			e.Error = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL}
			return e
		}, proto_},
		{"ErrorInfo.code 为 UNSPECIFIED", func() *pb.Envelope {
			e := partialResponse(t)
			e.Error.Code = 0
			return e
		}, proto_},
		{"SESSION_ERROR 缺 error", func() *pb.Envelope {
			return control(t, pb.FrameKind_FRAME_KIND_SESSION_ERROR, &pb.SessionError{FailedPhase: pb.SessionPhase_SESSION_PHASE_ON_START})
		}, proto_},
		{"RESPONSE 既无结果也无 error", func() *pb.Envelope {
			e := partialResponse(t)
			e.Error, e.SchemaId, e.SchemaVersion, e.Payload, e.PayloadEncoding = nil, "", 0, nil, 0
			return e
		}, proto_},
		{"payload 缺 schema_id", func() *pb.Envelope { e := request(1); e.SchemaId = ""; return e }, proto_},
		{"有载荷但编码为 UNSPECIFIED", func() *pb.Envelope { e := request(1); e.PayloadEncoding = 0; return e }, proto_},
		{"schema_version 不是 1", func() *pb.Envelope { e := request(1); e.SchemaVersion = 2; return e }, proto_},
		{"控制帧使用 JSON 编码", func() *pb.Envelope {
			e := control(t, pb.FrameKind_FRAME_KIND_HELLO, hello)
			e.PayloadEncoding = pb.PayloadEncoding_PAYLOAD_ENCODING_JSON
			return e
		}, proto_},
		{"控制帧 schema_id 与 kind 不符", func() *pb.Envelope {
			e := control(t, pb.FrameKind_FRAME_KIND_HELLO, hello)
			e.SchemaId = SchemaID(&pb.HelloAck{})
			return e
		}, proto_},
		{"控制帧缺载荷", func() *pb.Envelope {
			return &pb.Envelope{ProtocolMajor: 1, Kind: pb.FrameKind_FRAME_KIND_STOPPING, SessionId: sid}
		}, proto_},
		{"JSON 载荷语法错误", func() *pb.Envelope { e := request(1); e.Payload = []byte(`{"text":`); return e }, invalid},
		{"JSON 编码但载荷是 Protobuf 字节", func() *pb.Envelope {
			e := request(1)
			e.Payload = mustMarshal(t, hello)
			return e
		}, invalid},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := Envelope(c.env())
			if got := code(err); got != c.want {
				t.Fatalf("错误码 = %v，期望 %v（err=%v）", got, c.want, err)
			}
		})
	}
}

func TestDecodeControl(t *testing.T) {
	want := &pb.Stopping{DeadlineUnixMs: 42, Reason: pb.StopReason_STOP_REASON_DISABLE}
	got, err := DecodeControl(control(t, pb.FrameKind_FRAME_KIND_STOPPING, want))
	if err != nil || !proto.Equal(got, want) {
		t.Fatalf("DecodeControl = %v, %v", got, err)
	}
	bad := control(t, pb.FrameKind_FRAME_KIND_STOPPING, want)
	bad.Payload = []byte{0xFF, 0xFF, 0xFF}
	if _, err := DecodeControl(bad); code(err) != pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR {
		t.Fatalf("无法解码的载荷应返回 PROTOCOL_ERROR，实际 %v", err)
	}
	if _, err := DecodeControl(request(1)); code(err) != pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR {
		t.Fatalf("REQUEST 不是控制帧，实际 %v", err)
	}
}

func TestSchemaID(t *testing.T) {
	for m, want := range map[proto.Message]string{
		&pb.Hello{}: "alonabot.hello.v1", &pb.HelloAck{}: "alonabot.hello_ack.v1",
		&pb.SessionError{}: "alonabot.session_error.v1", &pb.BusEvent{}: "alonabot.bus_event.v1",
	} {
		if got := SchemaID(m); got != want {
			t.Errorf("SchemaID(%T) = %s，期望 %s", m, got, want)
		}
	}
}

func TestPublicCode(t *testing.T) {
	if got := PublicCode(pb.ErrorCode(250)); got != pb.ErrorCode_ERROR_CODE_INTERNAL {
		t.Errorf("未知错误码应映射为 INTERNAL，实际 %v", got)
	}
	if got := PublicCode(pb.ErrorCode_ERROR_CODE_CALL_CYCLE); got != pb.ErrorCode_ERROR_CODE_CALL_CYCLE {
		t.Errorf("已知错误码应保持不变，实际 %v", got)
	}
}

// golden fixture：固定字节，防止生成代码或字段号变化悄悄改变线上编码。
func TestGoldenFrames(t *testing.T) {
	frames := map[string]*pb.Envelope{
		"hello.bin": control(t, pb.FrameKind_FRAME_KIND_HELLO, &pb.Hello{
			PluginId: "alonabot_plugin_demo", Secret: bytes.Repeat([]byte{1}, 32),
			SupportedMajor: 1, SupportedMinor: 0, BuildDigest: bytes.Repeat([]byte{2}, 32),
			SupportedCapabilities: []*pb.ProtocolCapability{{Name: "core.send", Version: 1}},
			RequiredCapabilities:  []*pb.ProtocolCapability{{Name: "core.send", Version: 1}},
		}),
		"request_with_parent.bin": func() *pb.Envelope {
			e := request(pb.InitiatorSide_INITIATOR_SIDE_PLUGIN)
			e.ParentCallId = proto.Uint64(3)
			return e
		}(),
		"response_partial_with_error.bin": partialResponse(t),
		"session_error.bin": func() *pb.Envelope {
			e := control(t, pb.FrameKind_FRAME_KIND_SESSION_ERROR, &pb.SessionError{
				FailedPhase: pb.SessionPhase_SESSION_PHASE_ON_START, FailedKind: pb.FrameKind_FRAME_KIND_REGISTER})
			e.Error = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, SafeMessage: "OnStart 超时"}
			return e
		}(),
	}
	dir := filepath.Join("..", "testdata", "frames")
	for name, env := range frames {
		t.Run(name, func(t *testing.T) {
			if err := Envelope(env); err != nil {
				t.Fatalf("fixture 本身不合法: %v", err)
			}
			b := mustMarshal(t, env)
			path := filepath.Join(dir, name)
			if *update {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, b, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			golden, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("读取 golden 失败（首次请用 -update 生成）: %v", err)
			}
			if !bytes.Equal(b, golden) {
				t.Fatalf("编码与 golden 不一致")
			}
			var back pb.Envelope
			// proto.Equal 同时比较 optional 字段（parent_call_id/deadline）的存在性。
			if err := proto.Unmarshal(golden, &back); err != nil || !proto.Equal(&back, env) {
				t.Fatalf("golden 反解后与原值不同: %v", err)
			}
		})
	}
}
