package debug

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/AlonaBot/AlonaBot/contracts/debugapi"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/encoding/protojson"
)

// Reply 保留真实 Adapter 的发送或编辑结果。
type Reply struct {
	TraceID, Operation, InstanceID, AccountID, Session string
	Chain                                              message.Chain
	Result                                             *message.SendResult
	Err                                                *errcode.Error
}

// Outcome 保存实际调用的 Handler 结果，包括 Handler 返回的业务错误。
type Outcome struct {
	TraceID           string
	PluginID, RouteID string
	Waiting           bool
	Err               *errcode.Error
}

const (
	// PhaseAdmission 标识 Handler 调用前的权限、生命周期和路由准入检查。
	PhaseAdmission = debugapi.PhaseAdmission
	// PhaseArguments 标识 Handler 调用前的命令参数解析。
	PhaseArguments = debugapi.PhaseArguments
)

// Diagnostic 保留 Handler 调用前的失败；Phase 为 PhaseAdmission 或 PhaseArguments。
type Diagnostic struct {
	Phase, TraceID, PluginID, RouteID string
	Err                               *errcode.Error
}

// EventResult 保存标准事件的 trace、回复、Handler 结果与调用前诊断。
type EventResult struct {
	TraceIDs    []string
	Waiting     bool
	Replies     []Reply
	Outcomes    []Outcome
	Diagnostics []Diagnostic
}

// Inject 使用管理认证注入标准事件；消息匹配和权限判断由运行中的 Core 执行。
func (c *Client) Inject(ctx context.Context, event *message.Event) (EventResult, error) {
	encoded, err := convert.EventToProto(event)
	if err != nil {
		return EventResult{}, err
	}
	raw, err := protojson.Marshal(encoded)
	if err != nil {
		return EventResult{}, err
	}
	raw, err = c.request(ctx, http.MethodPost, "/api/v1/debug/events", debugapi.EventRequest{Event: raw})
	if err != nil {
		return EventResult{}, err
	}
	var response debugapi.EventResult
	if err := json.Unmarshal(raw, &response); err != nil {
		return EventResult{}, err
	}
	result := EventResult{TraceIDs: response.TraceIDs, Waiting: response.Waiting}
	for _, o := range response.Outcomes {
		failure, err := eventError(o.Error)
		if err != nil {
			return result, err
		}
		result.Outcomes = append(result.Outcomes, Outcome{TraceID: o.TraceID, PluginID: o.PluginID, RouteID: o.RouteID, Waiting: o.Waiting, Err: failure})
	}
	for _, diagnostic := range response.Diagnostics {
		failure, err := eventError(diagnostic.Error)
		if err != nil {
			return result, err
		}
		result.Diagnostics = append(result.Diagnostics, Diagnostic{Phase: diagnostic.Phase, TraceID: diagnostic.TraceID, PluginID: diagnostic.PluginID, RouteID: diagnostic.RouteID, Err: failure})
	}
	for _, reply := range response.Replies {
		target := &pb.Target{}
		chain := &pb.MessageChain{}
		if err := protojson.Unmarshal(reply.To, target); err != nil {
			return result, err
		}
		if err := protojson.Unmarshal(reply.Chain, chain); err != nil {
			return result, err
		}
		var receipt *message.SendResult
		if len(reply.Result) > 0 {
			value := &pb.SendResult{}
			if err := protojson.Unmarshal(reply.Result, value); err != nil {
				return result, err
			}
			receipt = convert.SendResultFromProto(value)
		}
		failure, err := eventError(reply.Error)
		if err != nil {
			return result, err
		}
		result.Replies = append(result.Replies, Reply{TraceID: reply.TraceID, Operation: reply.Operation, InstanceID: target.AdapterInstanceId, AccountID: target.AccountId, Session: convert.SessionFromProto(target), Chain: convert.ChainFromProto(chain, target), Result: receipt, Err: failure})
	}
	return result, nil
}

func eventError(failure *debugapi.Error) (*errcode.Error, error) {
	if failure == nil {
		return nil, nil
	}
	value := &pb.ErrorInfo{}
	if err := protojson.Unmarshal(failure.Details, value); err != nil {
		return nil, err
	}
	return convert.ErrorFromProto(value), nil
}
