// Package conversation 提供按插件身份隔离的 Agent 会话存储服务。
//
// 插件请求经 router.Mux 绑定的 Caller 取得身份；管理层直接调用 List、Clear、Purge。
// list 分批读取状态库并按帧预算分页，条目原样保存。
package conversation

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

const (
	schemaVersion     = 1
	maxKeyBytes       = 256      // 会话键的字节上限。
	envelopeAllowance = 64 << 10 // 为信封字段预留的余量。
	readBatchEntries  = 128
)

var invalidKeyMessage = fmt.Sprintf("conversation key must be 1 to %d bytes of valid UTF-8", maxKeyBytes)

// Service 经 Core 状态库保存会话，插件身份由注册的 Mux 提供。
type Service struct{ store *state.Store }

// New 创建会话服务。
func New(store *state.Store) *Service { return &Service{store: store} }

// Register 将三个会话固定服务注册到 m。
func (s *Service) Register(m *router.Mux) {
	m.Handle(wire.ServiceConversationAppend, s.append)
	m.Handle(wire.ServiceConversationList, s.list)
	m.Handle(wire.ServiceConversationClear, s.clear)
}

// List 按会话键字典序列出插件的会话摘要。
func (s *Service) List(ctx context.Context, pluginID string) ([]state.ConversationSummary, error) {
	return s.store.Conversations(ctx, pluginID)
}

// Clear 删除指定插件的指定会话条目。
func (s *Service) Clear(ctx context.Context, pluginID, key string) error {
	if !validKey(key) {
		return errors.New(invalidKeyMessage)
	}
	return s.store.ClearConversation(ctx, pluginID, key)
}

// Purge 删除指定插件的全部会话条目。
func (s *Service) Purge(ctx context.Context, pluginID string) error {
	return s.store.PurgeConversations(ctx, pluginID)
}

func decode(in *wire.Inbound, request proto.Message) error {
	if in.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.Payload.SchemaID != validate.SchemaID(request) || in.Payload.SchemaVersion != schemaVersion {
		return fmt.Errorf("expected %s protobuf schema version %d", validate.SchemaID(request), schemaVersion)
	}
	if err := proto.Unmarshal(in.Payload.Bytes, request); err != nil {
		return fmt.Errorf("invalid conversation request protobuf: %w", err)
	}
	return nil
}

func validKey(key string) bool {
	return len(key) > 0 && len(key) <= maxKeyBytes && utf8.ValidString(key)
}

func failure(code pb.ErrorCode, message string) wire.Response {
	return wire.Response{Err: &pb.ErrorInfo{Code: code, SafeMessage: message}}
}

func response(result proto.Message) wire.Response {
	encoded, err := proto.Marshal(result)
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "encode conversation response: "+err.Error())
	}
	return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(result), SchemaVersion: schemaVersion, Bytes: encoded}}
}

func (s *Service) append(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	// 1. 整次校验完成后才进入写事务，避免部分追加。
	request := &pb.ConversationAppendRequest{}
	if err := decode(in, request); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	if !validKey(request.ConversationKey) {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, invalidKeyMessage)
	}
	if len(request.EntriesJson) == 0 {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "conversation append requires at least one entry")
	}
	for _, entry := range request.EntriesJson {
		if !validate.IsJSONObject([]byte(entry)) {
			return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "conversation entry must be a JSON object")
		}
	}
	// 2. 身份只取自 Mux，在单个事务中按请求顺序追加。
	lastSeq, err := s.store.AppendConversation(ctx, caller.PluginID, request.ConversationKey, request.EntriesJson)
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "append conversation: "+err.Error())
	}
	return response(&pb.ConversationAppendResponse{LastSeq: lastSeq})
}

func (s *Service) list(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	// 1. 校验请求并将零起点归一为首条序号。
	request := &pb.ConversationListRequest{}
	if err := decode(in, request); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	if !validKey(request.ConversationKey) {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, invalidKeyMessage)
	}
	fromSeq := max(request.FromSeq, 1)
	result := &pb.ConversationListResponse{}
	size := 0
	// 2. 每批固定条数，逐条计入 repeated 消息的 tag、长度前缀和内容。
	for {
		entries, err := s.store.ListConversation(ctx, caller.PluginID, request.ConversationKey, fromSeq, readBatchEntries)
		if err != nil {
			return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "list conversation: "+err.Error())
		}
		for _, entry := range entries {
			item := &pb.ConversationEntry{Seq: entry.Seq, EntryJson: entry.EntryJSON}
			itemSize := protowire.SizeTag(1) + protowire.SizeBytes(proto.Size(item))
			if len(result.Entries) > 0 && size+itemSize > wire.MaxFrame-envelopeAllowance {
				result.NextSeq = proto.Uint64(entry.Seq)
				return response(result)
			}
			result.Entries = append(result.Entries, item)
			size += itemSize
			fromSeq = entry.Seq + 1
		}
		if len(entries) < readBatchEntries {
			return response(result)
		}
	}
}

func (s *Service) clear(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	request := &pb.ConversationClearRequest{}
	if err := decode(in, request); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	if !validKey(request.ConversationKey) {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, invalidKeyMessage)
	}
	if err := s.store.ClearConversation(ctx, caller.PluginID, request.ConversationKey); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "clear conversation: "+err.Error())
	}
	return response(&pb.ConversationClearResponse{})
}
