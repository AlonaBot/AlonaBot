package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
	"google.golang.org/protobuf/proto"
)

// Effect 提示工具的副作用类别，零值表示未声明。
type Effect int

const (
	// Unspecified 保留未声明类别的状态。
	Unspecified Effect = iota
	// ReadOnly 声明只读取数据。
	ReadOnly
	// Write 声明会写入数据。
	Write
	// Destructive 声明可能破坏或删除数据。
	Destructive
)

// Info 保存 Core 返回的可调用工具描述。
type Info struct {
	Name, Description string
	InputSchema       json.RawMessage // Service 输入 Schema 的原文
	Policy            string          // "allow" 或 "confirm"
	Effect            Effect
}

// List 读取当前插件可访问且策略允许列出的工具，按完整名称排序。
// Session 尚未就绪或已停止时返回 FAILED_PRECONDITION。
func List(ctx context.Context) ([]Info, error) {
	result, err := registry.Proto[*pb.ToolListRequest, *pb.ToolListResponse](wire.ServiceToolList).Call(ctx, &pb.ToolListRequest{})
	if err != nil {
		return nil, err
	}
	tools := make([]Info, 0, len(result.GetTools()))
	for _, entry := range result.GetTools() {
		var policy string
		switch entry.GetPolicy() {
		case pb.ToolPolicy_TOOL_POLICY_ALLOW:
			policy = "allow"
		case pb.ToolPolicy_TOOL_POLICY_CONFIRM:
			policy = "confirm"
		default:
			return nil, &sdk.Error{Code: errcode.Internal, Message: fmt.Sprintf("unexpected tool policy %d for %q", entry.GetPolicy(), entry.GetName())}
		}
		tools = append(tools, Info{Name: entry.GetName(), Description: entry.GetDescription(), InputSchema: json.RawMessage(entry.GetInputSchemaJson()), Policy: policy, Effect: Effect(entry.GetEffect())})
	}
	return tools, nil
}

// Invoke 在 Handler 或其延续调用内执行工具，保留 JSON 参数与结果原文。
// 策略要求确认时等待当前用户决定；调用随父请求结束而取消。
func Invoke(ctx context.Context, name string, arguments json.RawMessage) (json.RawMessage, error) {
	result := &pb.ToolInvokeResponse{}
	err := callTool(ctx, wire.ServiceToolInvoke, &pb.ToolInvokeRequest{Name: name, ArgumentsJson: string(arguments)}, result)
	return json.RawMessage(result.GetResultJson()), err
}

// Confirm 在当前父调用链中请求用户确认操作，允许时返回 nil。
func Confirm(ctx context.Context, operation, summary string, effect Effect) error {
	if effect < Unspecified || effect > Destructive {
		return errcode.InvalidArgument.Errorf("unknown tool effect %d", effect)
	}
	return callTool(ctx, wire.ServiceToolConfirm, &pb.ToolConfirmRequest{Operation: operation, Summary: summary, Effect: pb.ToolEffect(effect)}, &pb.ToolConfirmResponse{})
}

// ConfirmRequired 要求可信交互用户确认，即使当前操作策略为 allow；deny 仍优先拒绝。
func ConfirmRequired(ctx context.Context, operation, summary string, effect Effect) error {
	if effect < Unspecified || effect > Destructive {
		return errcode.InvalidArgument.Errorf("unknown tool effect %d", effect)
	}
	return callTool(ctx, wire.ServiceToolConfirm, &pb.ToolConfirmRequest{Operation: operation, Summary: summary, Effect: pb.ToolEffect(effect), RequireConfirmation: true}, &pb.ToolConfirmResponse{})
}

func callTool(ctx context.Context, target string, request, result proto.Message) error {
	// 1. 工具操作须关联活跃父调用；后台调用在 SDK 边界直接拒绝。
	parent, active := registry.Parent(ctx)
	if !active {
		return errcode.FailedPrecondition.New("tool operation requires an active handler context")
	}
	session, err := registry.Session(ctx)
	if err != nil {
		return err
	}
	encoded, err := proto.Marshal(request)
	if err != nil {
		return errcode.InvalidArgument.New("encode tool request: " + err.Error())
	}
	// 2. 确认等待使用调用 context 的期限，避免普通请求缺省期限提前结束等待。
	deadline, _ := ctx.Deadline()
	response, callError := session.Call(ctx, wire.Request{Target: target, ParentCallID: parent.CallID, TraceID: parent.TraceID, Deadline: deadline,
		Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: registry.SchemaVersion, Bytes: encoded}})
	var publicError error
	if callError != nil {
		info := callError.(*wire.CallError).ErrorInfo
		publicError = &sdk.Error{Code: errcode.Code(validate.PublicCode(info.GetCode())), Message: info.GetSafeMessage()}
	}
	// 3. 错误可能同时携带工具结果，先解码有效载荷再保留原始业务错误。
	if callError == nil || response.Payload.SchemaID != "" {
		if err := registry.Decode(response.Payload, result); err != nil {
			return errors.Join(publicError, errcode.Internal.New(err.Error()))
		}
	}
	return publicError
}
