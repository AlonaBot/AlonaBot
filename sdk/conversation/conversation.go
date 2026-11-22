// Package conversation 追加、分页读取与清除原样 JSON 会话历史。
package conversation

import (
	"context"
	"encoding/json"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
	"google.golang.org/protobuf/proto"
)

// Entry 保存会话内的序号与原样 JSON 条目。
type Entry struct {
	Seq   uint64
	Entry json.RawMessage
}

// Append 原子追加 JSON 对象，返回本次追加的最后一个序号。
func Append(ctx context.Context, key string, entries ...json.RawMessage) (lastSeq uint64, err error) {
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

// List 读取一页会话条目；more 为 true 时以 nextSeq 继续读取。
// fromSeq 为零时从 1 开始；全部读完时 more 为 false。
func List(ctx context.Context, key string, fromSeq uint64) (entries []Entry, nextSeq uint64, more bool, err error) {
	result := &pb.ConversationListResponse{}
	if err := callConversation(ctx, wire.ServiceConversationList, &pb.ConversationListRequest{ConversationKey: key, FromSeq: fromSeq}, result); err != nil {
		return nil, 0, false, err
	}
	for _, entry := range result.Entries {
		entries = append(entries, Entry{Seq: entry.Seq, Entry: json.RawMessage(entry.EntryJson)})
	}
	return entries, result.GetNextSeq(), result.NextSeq != nil, nil
}

// Clear 删除会话的全部条目；会话不存在时也成功。
func Clear(ctx context.Context, key string) error {
	return callConversation(ctx, wire.ServiceConversationClear, &pb.ConversationClearRequest{ConversationKey: key}, &pb.ConversationClearResponse{})
}

func callConversation(ctx context.Context, target string, request, result proto.Message) error {
	// 1. 复用就绪门禁；会话请求只需要插件身份，不设置父调用链。
	session, err := registry.Session(ctx)
	if err != nil {
		return err
	}
	encoded, err := proto.Marshal(request)
	if err != nil {
		return &sdk.Error{Code: errcode.InvalidArgument, Message: "encode conversation request: " + err.Error()}
	}
	out, callError := session.Call(ctx, wire.Request{Target: target,
		Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: registry.SchemaVersion, Bytes: encoded}})
	// 2. Core 错误保留安全消息，协议码统一转为公开错误码。
	if callError != nil {
		info := callError.(*wire.CallError).ErrorInfo
		return &sdk.Error{Code: errcode.Code(validate.PublicCode(info.GetCode())), Message: info.GetSafeMessage()}
	}
	if err := registry.Decode(out.Payload, result); err != nil {
		return &sdk.Error{Code: errcode.Internal, Message: err.Error()}
	}
	return nil
}
