package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"google.golang.org/protobuf/proto"
)

// ToolDefault 指定工具未被管理策略覆盖时的调用方式。
type ToolDefault int

const (
	// ToolAllow 允许工具调用直接执行。
	ToolAllow ToolDefault = iota
	// ToolConfirm 要求工具调用先取得用户确认。
	ToolConfirm
)

// ToolInfo 保存 Core 返回的可调用工具描述。
type ToolInfo struct {
	Name, Description string
	InputSchema       json.RawMessage // Service 输入 Schema 的原文
	Policy            string          // "allow" 或 "confirm"
}

// ListTools 读取当前插件可访问且策略允许列出的工具，按完整名称排序。
// Session 尚未就绪或已停止时返回 FAILED_PRECONDITION。
func ListTools(ctx context.Context) ([]ToolInfo, error) {
	result, err := Proto[*pb.ToolListRequest, *pb.ToolListResponse](wire.ServiceToolList).Call(ctx, &pb.ToolListRequest{})
	if err != nil {
		return nil, err
	}
	tools := make([]ToolInfo, 0, len(result.GetTools()))
	for _, entry := range result.GetTools() {
		var policy string
		switch entry.GetPolicy() {
		case pb.ToolPolicy_TOOL_POLICY_ALLOW:
			policy = "allow"
		case pb.ToolPolicy_TOOL_POLICY_CONFIRM:
			policy = "confirm"
		default:
			return nil, &Error{Code: errcode.Internal, Message: fmt.Sprintf("unexpected tool policy %d for %q", entry.GetPolicy(), entry.GetName())}
		}
		tools = append(tools, ToolInfo{Name: entry.GetName(), Description: entry.GetDescription(), InputSchema: json.RawMessage(entry.GetInputSchemaJson()), Policy: policy})
	}
	return tools, nil
}

// InvokeTool 在 Handler 或其延续调用内执行工具，保留 JSON 参数与结果原文。
// 策略要求确认时等待当前用户决定；调用随父请求结束而取消。
func InvokeTool(ctx context.Context, name string, arguments json.RawMessage) (json.RawMessage, error) {
	result := &pb.ToolInvokeResponse{}
	err := callTool(ctx, wire.ServiceToolInvoke, &pb.ToolInvokeRequest{Name: name, ArgumentsJson: string(arguments)}, result)
	return json.RawMessage(result.GetResultJson()), err
}

// ConfirmOperation 在当前父调用链中请求用户确认操作，允许时返回 nil。
func ConfirmOperation(ctx context.Context, operation, summary string) error {
	return callTool(ctx, wire.ServiceToolConfirm, &pb.ToolConfirmRequest{Operation: operation, Summary: summary}, &pb.ToolConfirmResponse{})
}

func callTool(ctx context.Context, target string, request, result proto.Message) error {
	// 1. 工具操作须关联活跃父调用；后台调用在 SDK 边界直接拒绝。
	c, ok := ctx.Value(serviceContextKey{}).(*Context)
	if !ok || !c.active.Load() {
		return errcode.FailedPrecondition.New("tool operation requires an active handler context")
	}
	session, err := currentSession(ctx)
	if err != nil {
		return err
	}
	encoded, err := proto.Marshal(request)
	if err != nil {
		return errcode.InvalidArgument.New("encode tool request: " + err.Error())
	}
	// 2. 确认等待使用调用 context 的期限，避免普通请求缺省期限提前结束等待。
	deadline, _ := ctx.Deadline()
	response, callError := session.Call(ctx, wire.Request{Target: target, ParentCallID: c.inbound.CallID, TraceID: c.inbound.TraceID, Deadline: deadline,
		Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: invocationSchemaVersion, Bytes: encoded}})
	var publicError error
	if callError != nil {
		info := callError.(*wire.CallError).ErrorInfo
		publicError = &Error{Code: errcode.Code(validate.PublicCode(info.GetCode())), Message: info.GetSafeMessage()}
	}
	// 3. 错误可能同时携带工具结果，先解码有效载荷再保留原始业务错误。
	if callError == nil || response.Payload.SchemaID != "" {
		if err := decodeInvocation(response.Payload, result); err != nil {
			return errors.Join(publicError, errcode.Internal.New(err.Error()))
		}
	}
	return publicError
}
