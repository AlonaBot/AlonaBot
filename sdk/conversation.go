package sdk

import (
	"context"
	"encoding/json"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// ConversationEntry 保存会话内的序号与原样 JSON 条目。
type ConversationEntry struct {
	Seq   uint64
	Entry json.RawMessage
}

// AppendConversation 原子追加 JSON 对象，返回本次追加的最后一个序号。
func AppendConversation(ctx context.Context, key string, entries ...json.RawMessage) (lastSeq uint64, err error) {
	request := &pb.ConversationAppendRequest{ConversationKey: key, EntriesJson: make([]string, len(entries))}
	for i, entry := range entries {
		request.EntriesJson[i] = string(entry)
	}
	result := &pb.ConversationAppendResponse{}
	if err := callConversation(ctx, wire.ServiceConversationAppend, request, result); err != nil {
		return 0, err
	}
	return result.LastSeq, nil
}

// ListConversation 读取一页会话条目；more 为 true 时以 nextSeq 继续读取。
// fromSeq 为零时从 1 开始；全部读完时 more 为 false。
func ListConversation(ctx context.Context, key string, fromSeq uint64) (entries []ConversationEntry, nextSeq uint64, more bool, err error) {
	result := &pb.ConversationListResponse{}
	if err := callConversation(ctx, wire.ServiceConversationList, &pb.ConversationListRequest{ConversationKey: key, FromSeq: fromSeq}, result); err != nil {
		return nil, 0, false, err
	}
	for _, entry := range result.Entries {
		entries = append(entries, ConversationEntry{Seq: entry.Seq, Entry: json.RawMessage(entry.EntryJson)})
	}
	return entries, result.GetNextSeq(), result.NextSeq != nil, nil
}

// ClearConversation 删除会话的全部条目；会话不存在时也成功。
func ClearConversation(ctx context.Context, key string) error {
	return callConversation(ctx, wire.ServiceConversationClear, &pb.ConversationClearRequest{ConversationKey: key}, &pb.ConversationClearResponse{})
}

func callConversation(ctx context.Context, target string, request, result proto.Message) error {
	// 1. 复用就绪门禁；会话请求只需要插件身份，不设置父调用链。
	session, err := currentSession()
	if err != nil {
		return err
	}
	encoded, err := proto.Marshal(request)
	if err != nil {
		return &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "encode conversation request: " + err.Error()}
	}
	out, callError := session.Call(ctx, wire.Request{Target: target,
		Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: invocationSchemaVersion, Bytes: encoded}})
	// 2. Core 错误保留安全消息，协议码统一转为公开错误码。
	if callError != nil {
		info := callError.(*wire.CallError).ErrorInfo
		return &Error{Code: validate.PublicCode(info.GetCode()), Message: info.GetSafeMessage()}
	}
	if err := decodeInvocation(out.Payload, result); err != nil {
		return &Error{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, Message: err.Error()}
	}
	return nil
}
