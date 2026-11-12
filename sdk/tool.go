package sdk

import (
	"context"
	"encoding/json"
	"fmt"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
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
	result, err := CallProto[*pb.ToolListResponse](ctx, wire.ServiceToolList, &pb.ToolListRequest{})
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
