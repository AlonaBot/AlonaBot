package sdk

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/proto"
)

// Edit 修改一条已发送的平台文本消息；目标缺省时使用当前调用链的默认 Target。
// chain 仅接受一个 text 片段和至多一个 buttons 片段，准入由 Core 检查。
// 出错时保留有效的编辑回执，失败后不发送新消息。
func (c *Context) Edit(platformMessageID string, chain message.Chain, opts ...SendOption) (*pb.SendResult, error) {
	if !c.active.Load() {
		return nil, &Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "handler context has expired"}
	}
	options := &sendOptions{}
	for _, option := range opts {
		option(options)
	}
	return callEdit(c.ctx, options.to, platformMessageID, chain, c.inbound)
}

// Edit 通过已激活的 Session 独立编辑指定消息，不关联入站调用。
// to 必须显式给出；缺少时由 Core 返回 NO_DEFAULT_TARGET。
// Session 尚未就绪或已停止时返回 FAILED_PRECONDITION，出错时保留有效回执。
func Edit(ctx context.Context, to *pb.Target, platformMessageID string, chain message.Chain) (*pb.SendResult, error) {
	return callEdit(ctx, to, platformMessageID, chain, nil)
}

func callEdit(ctx context.Context, to *pb.Target, platformMessageID string, chain message.Chain, parent *wire.Inbound) (*pb.SendResult, error) {
	// 1. 复用 Session 就绪门禁，编码编辑请求。
	session, err := currentSession()
	if err != nil {
		return nil, err
	}
	request := &pb.EditRequest{To: to, PlatformMessageId: platformMessageID, Chain: chain.Proto()}
	encoded, err := proto.Marshal(request)
	if err != nil {
		return nil, &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "encode EditRequest: " + err.Error()}
	}
	// 2. 仅 Handler 入口关联父调用；后台入口始终使用显式目标准入。
	call := wire.Request{Target: wire.ServiceEdit,
		Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: invocationSchemaVersion, Bytes: encoded}}
	if parent != nil {
		call.ParentCallID = parent.CallID
		call.TraceID = parent.TraceID
	}
	response, callError := session.Call(ctx, call)
	// 3. 先解码有效回执，再保留远端或本地调用的完整安全错误。
	var result *pb.SendResult
	if response.Payload.SchemaID != "" {
		result = &pb.SendResult{}
		if err := decodeInvocation(response.Payload, result); err != nil {
			return nil, &Error{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, Message: err.Error()}
		}
	}
	if callError != nil {
		info := callError.(*wire.CallError).ErrorInfo
		return result, &Error{Code: validate.PublicCode(info.GetCode()), Message: info.GetSafeMessage()}
	}
	return result, nil
}
