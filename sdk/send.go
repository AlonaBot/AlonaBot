package sdk

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/proto"
)

// Send 通过已激活的 Session 独立向 to 发送，不关联入站调用。
// Session 尚未就绪或已停止时返回 FAILED_PRECONDITION。
// to 为 nil 时由 Core 返回 NO_DEFAULT_TARGET；出错时仍保留有效的部分发送结果。
func Send(ctx context.Context, to *pb.Target, chain message.Chain) (*pb.SendResult, error) {
	if err := checkMessageEncoding(chain); err != nil {
		return nil, err
	}
	// 1. 取得已就绪的 Session 并编码 SendRequest。
	session, err := currentSession()
	if err != nil {
		return nil, err
	}
	request := &pb.SendRequest{To: to, Chain: chain.Proto()}
	encoded, err := proto.Marshal(request)
	if err != nil {
		return nil, &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "encode SendRequest: " + err.Error()}
	}
	// 2. 不设父调用，发起独立调用；Core 按显式目标检查实例许可。
	response, callError := session.Call(ctx, wire.Request{Target: wire.ServiceSend,
		Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: invocationSchemaVersion, Bytes: encoded}})
	// 3. 即使出现 ErrorInfo，仍先解码并保留有效的部分发送结果。
	var result *pb.SendResult
	if response.Payload.SchemaID != "" {
		result = &pb.SendResult{}
		if err := decodeInvocation(response.Payload, result); err != nil {
			return nil, &Error{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, Message: err.Error()}
		}
	}
	if callError != nil {
		// Session.Call 的远端错误与本地准入错误均为 *wire.CallError，保留其原始消息。
		info := callError.(*wire.CallError).ErrorInfo
		return result, &Error{Code: validate.PublicCode(info.GetCode()), Message: info.GetSafeMessage()}
	}
	return result, nil
}
