package state

import (
	"context"
	"database/sql"
	"math"
	"time"
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
	err = s.tx(ctx, func(tx *sql.Tx) error {
		// 单连接在事务期间独占，读末尾和写入之间不会穿插其他追加。
		if err := tx.QueryRowContext(ctx,
			"SELECT coalesce(max(seq), 0) FROM conversation_entries WHERE plugin_id = ? AND conversation_key = ?",
			pluginID, key).Scan(&lastSeq); err != nil {
			return err
		}
		createdAt := time.Now().UnixMilli()
		for _, entry := range entries {
			lastSeq++
			if _, err := tx.ExecContext(ctx,
				"INSERT INTO conversation_entries (plugin_id, conversation_key, seq, entry_json, created_at) VALUES (?, ?, ?, ?, ?)",
				pluginID, key, lastSeq, entry, createdAt); err != nil {
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
	rows, err := s.db.QueryContext(ctx,
		"SELECT seq, entry_json FROM conversation_entries WHERE plugin_id = ? AND conversation_key = ? AND seq >= ? ORDER BY seq LIMIT ?",
		pluginID, key, fromSeq, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []ConversationEntry
	for rows.Next() {
		var entry ConversationEntry
		if err := rows.Scan(&entry.Seq, &entry.EntryJSON); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

// ClearConversation 删除该插件指定会话的全部条目；不存在时也成功。
func (s *Store) ClearConversation(ctx context.Context, pluginID, key string) error {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM conversation_entries WHERE plugin_id = ? AND conversation_key = ?", pluginID, key)
	return err
}

// Conversations 按会话键字典序返回该插件的非空会话摘要。
func (s *Store) Conversations(ctx context.Context, pluginID string) ([]ConversationSummary, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT conversation_key, count(*), max(seq), max(created_at) FROM conversation_entries WHERE plugin_id = ? GROUP BY conversation_key ORDER BY conversation_key",
		pluginID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var summaries []ConversationSummary
	for rows.Next() {
		var summary ConversationSummary
		var updatedAt int64
		if err := rows.Scan(&summary.Key, &summary.Entries, &summary.LastSeq, &updatedAt); err != nil {
			return nil, err
		}
		summary.UpdatedAt = time.UnixMilli(updatedAt)
		summaries = append(summaries, summary)
	}
	return summaries, rows.Err()
}

// PurgeConversations 删除该插件的全部会话条目。
func (s *Store) PurgeConversations(ctx context.Context, pluginID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM conversation_entries WHERE plugin_id = ?", pluginID)
	return err
}
