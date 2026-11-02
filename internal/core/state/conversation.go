package state

import (
	"context"
	"math"
	"time"

	"gorm.io/gorm"
)

// ConversationEntry 保存会话内的连续序号和原样 JSON 条目。
type ConversationEntry struct {
	Seq       uint64
	EntryJSON string
}

// ConversationSummary 汇总一个会话的条目数量、末尾序号和最后追加时间。
type ConversationSummary struct {
	Key       string
	Entries   int
	LastSeq   uint64
	UpdatedAt time.Time
}

// AppendConversation 在一个事务内按输入顺序追加，返回提交后的末尾序号。
func (s *Store) AppendConversation(ctx context.Context, pluginID, key string, entries []string) (lastSeq uint64, err error) {
	err = s.tx(ctx, func(tx *gorm.DB) error {
		// 单连接在事务期间独占，读末尾和写入之间不会穿插其他追加。
		if err := tx.Model(&conversationEntry{}).Select("coalesce(max(seq), 0)").
			Where("plugin_id = ? AND conversation_key = ?", pluginID, key).Scan(&lastSeq).Error; err != nil {
			return err
		}
		createdAt := time.Now().UnixMilli()
		for _, entry := range entries {
			lastSeq++
			if err := tx.Create(&conversationEntry{PluginID: pluginID, ConversationKey: key, Seq: lastSeq, EntryJSON: entry, CreatedAt: createdAt}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return lastSeq, nil
}

// ListConversation 按序号升序读取 seq >= fromSeq 的条目，至多返回 limit 条。
func (s *Store) ListConversation(ctx context.Context, pluginID, key string, fromSeq uint64, limit int) ([]ConversationEntry, error) {
	// SQLite INTEGER 的序号最大为 MaxInt64，更大的起点没有匹配条目。
	if fromSeq > math.MaxInt64 {
		return nil, nil
	}
	var rows []conversationEntry
	err := s.db.WithContext(ctx).Select("seq", "entry_json").
		Where("plugin_id = ? AND conversation_key = ? AND seq >= ?", pluginID, key, fromSeq).
		Order("seq").Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	var entries []ConversationEntry
	for _, row := range rows {
		entries = append(entries, ConversationEntry{Seq: row.Seq, EntryJSON: row.EntryJSON})
	}
	return entries, nil
}

// ClearConversation 删除该插件指定会话的全部条目；不存在时也成功。
func (s *Store) ClearConversation(ctx context.Context, pluginID, key string) error {
	return s.db.WithContext(ctx).Where("plugin_id = ? AND conversation_key = ?", pluginID, key).Delete(&conversationEntry{}).Error
}

// Conversations 按会话键字典序返回该插件的非空会话摘要。
func (s *Store) Conversations(ctx context.Context, pluginID string) ([]ConversationSummary, error) {
	var rows []struct {
		Key       string
		Entries   int
		LastSeq   uint64
		UpdatedAt int64
	}
	err := s.db.WithContext(ctx).Model(&conversationEntry{}).
		Select("conversation_key AS key, count(*) AS entries, max(seq) AS last_seq, max(created_at) AS updated_at").
		Where("plugin_id = ?", pluginID).Group("conversation_key").Order("conversation_key").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	var summaries []ConversationSummary
	for _, row := range rows {
		summaries = append(summaries, ConversationSummary{Key: row.Key, Entries: row.Entries, LastSeq: row.LastSeq, UpdatedAt: time.UnixMilli(row.UpdatedAt)})
	}
	return summaries, nil
}

// PurgeConversations 删除该插件的全部会话条目。
func (s *Store) PurgeConversations(ctx context.Context, pluginID string) error {
	return s.db.WithContext(ctx).Where("plugin_id = ?", pluginID).Delete(&conversationEntry{}).Error
}
