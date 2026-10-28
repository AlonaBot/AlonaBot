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
// to 为 nil 时由 Core 返回 NO_DEFAULT_TARGET；出错时仍保留有效的部分发送结果。
func Send(ctx context.Context, to *pb.Target, chain message.Chain) (*pb.SendResult, error) {
	// 1. Session 就绪后编码发送请求，目标许可交由 Core 判断。
	session, err := currentSession()
	if err != nil {
		return nil, err
	}
	request := &pb.SendRequest{To: to, Chain: chain.Proto()}
	encoded, err := proto.Marshal(request)
	if err != nil {
		return nil, &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "encode SendRequest: " + err.Error()}
	}
	// 2. 独立后台请求不关联父调用，由 Core 检查显式目标许可。
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
	if response.Err != nil {
		return result, &Error{Code: validate.PublicCode(response.Err.GetCode()), Message: response.Err.GetSafeMessage()}
	}
	if callError != nil {
		return result, &Error{Code: validate.PublicCode(wire.CodeOf(callError)), Message: callError.Error()}
	}
	return result, nil
}
