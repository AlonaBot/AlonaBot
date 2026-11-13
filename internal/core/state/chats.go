package state

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Chat 保存 Core 见过的聊天；同一聊天的所有话题共用一条记录。
type Chat struct {
	AdapterID, ChatID   string
	ChatType            string // private、group、channel；空串表示未知。
	Title, InstanceID   string
	FirstSeen, LastSeen time.Time // Core 接纳事件的时刻，精确到毫秒。
}

// ChatKey 指定平台内唯一的聊天，省略实例与话题。
type ChatKey struct{ AdapterID, ChatID string }

// ChatQuery 指定搜索、筛选或精确聊天键，以及返回条数。
type ChatQuery struct {
	Search, ChatType, AdapterID string
	Keys                        []ChatKey // 非 nil 时只按键取回，忽略搜索与筛选。
	Limit                       int       // 正数限制条数；0 返回全部匹配记录。
}

type chatRecord struct {
	AdapterID                   string `gorm:"primaryKey"`
	ChatID                      string `gorm:"primaryKey"`
	ChatType, Title, InstanceID string
	FirstSeen, LastSeen         int64
}

func (chatRecord) TableName() string { return "chat" }

func putChat(tx *gorm.DB, chat *Chat, now time.Time) error {
	columns := []string{"instance_id", "last_seen"}
	if chat.Title != "" {
		columns = append(columns, "title")
	}
	if chat.ChatType != "" {
		columns = append(columns, "chat_type")
	}
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "adapter_id"}, {Name: "chat_id"}},
		DoUpdates: clause.AssignmentColumns(columns),
	}).Create(&chatRecord{AdapterID: chat.AdapterID, ChatID: chat.ChatID, ChatType: chat.ChatType, Title: chat.Title, InstanceID: chat.InstanceID, FirstSeen: now.UnixMilli(), LastSeen: now.UnixMilli()}).Error
}

// Chats 按最近活跃时刻从新到旧返回聊天；more 表示匹配数超出 Limit。
// 搜索按 Unicode 小写匹配名称与 chat_id，保留搜索词中的通配符原文。
func (s *Store) Chats(ctx context.Context, q ChatQuery) (chats []Chat, more bool, err error) {
	// 1. 在状态库筛选身份与类型，按值取回时忽略搜索条件。
	query := s.db.WithContext(ctx).Model(&chatRecord{})
	if q.Keys != nil {
		if len(q.Keys) == 0 {
			return nil, false, nil
		}
		keys := make([][]any, 0, len(q.Keys))
		for _, key := range q.Keys {
			keys = append(keys, []any{key.AdapterID, key.ChatID})
		}
		query = query.Where("(adapter_id, chat_id) IN ?", keys)
		q.Search = ""
	} else {
		if q.AdapterID != "" {
			query = query.Where("adapter_id = ?", q.AdapterID)
		}
		if q.ChatType != "" {
			query = query.Where("chat_type = ?", q.ChatType)
		}
	}
	query = query.Order("last_seen DESC, adapter_id, chat_id")
	if q.Search == "" && q.Limit > 0 {
		query = query.Limit(q.Limit + 1)
	}
	rows, err := query.Rows()
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	// 2. SQLite 的 lower 只处理 ASCII；逐行匹配支持非 ASCII 名称，找到多一条后结束。
	search := strings.ToLower(q.Search)
	for rows.Next() {
		var row chatRecord
		if err := s.db.ScanRows(rows, &row); err != nil {
			return chats, false, err
		}
		if !strings.Contains(strings.ToLower(row.Title), search) && !strings.Contains(strings.ToLower(row.ChatID), search) {
			continue
		}
		if q.Limit > 0 && len(chats) == q.Limit {
			return chats, true, nil
		}
		chats = append(chats, Chat{AdapterID: row.AdapterID, ChatID: row.ChatID, ChatType: row.ChatType, Title: row.Title, InstanceID: row.InstanceID, FirstSeen: time.UnixMilli(row.FirstSeen).UTC(), LastSeen: time.UnixMilli(row.LastSeen).UTC()})
	}
	return chats, false, rows.Err()
}
