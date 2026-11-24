package message

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/libtnb/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

const replyDatabase = "reply.sqlite"

//go:embed reply.sql
var replySchema string

type storedReply struct {
	Session   string `gorm:"column:session;primaryKey"`
	TargetKey string `gorm:"column:target_key;primaryKey"`
	Prefix    string `gorm:"column:prefix"`
	Mention   []byte `gorm:"column:mention"`
}

func (storedReply) TableName() string { return "reply_decoration" }

// OpenReplyStore 在持有 Core 实例锁时创建本次运行的关联库，清理上次 Session 遗留的数据。
// 连接只保留 SQLite 有界页缓存，历史消息按键查询；现有 Core 状态库保持独立。
func (s *Service) OpenReplyStore(dir string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create reply runtime directory: %w", err)
	}
	path := filepath.Join(dir, replyDatabase)
	for _, suffix := range []string{"", "-wal", "-shm", "-journal"} {
		if err := os.Remove(path + suffix); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove previous reply database: %w", err)
		}
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create reply database: %w", err)
	}
	if err := file.Close(); err != nil {
		return err
	}
	uri := (&url.URL{Scheme: "file", Path: path}).String()
	db, err := gorm.Open(sqlite.Open(uri), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return fmt.Errorf("open reply database: %w", err)
	}
	conn, err := db.DB()
	if err != nil {
		return err
	}
	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)
	if err := db.Exec(replySchema).Error; err != nil {
		return errors.Join(fmt.Errorf("initialize reply database: %w", err), conn.Close())
	}
	s.replyDB = db
	return nil
}

// Close 释放运行期回复数据库连接；调用者先排空消息与观察任务。
func (s *Service) Close() error {
	if s.replyDB == nil {
		return nil
	}
	conn, err := s.replyDB.DB()
	if err != nil {
		return err
	}
	return conn.Close()
}

func replySessionKey(c router.Caller) string {
	data, _ := json.Marshal(c) // Caller 只含 string 字段。
	return string(data)
}

func storedReplyKey(c router.Caller, key replyKey) storedReply {
	data, _ := json.Marshal([]any{key.adapter, key.instance, key.account, key.chat, key.thread, key.threadPresent, key.message})
	return storedReply{Session: replySessionKey(c), TargetKey: string(data)}
}

func (s *Service) deleteReplies(c router.Caller) error {
	if s.replyDB == nil {
		return nil
	}
	return s.replyDB.WithContext(context.Background()).Where(&storedReply{Session: replySessionKey(c)}).Delete(&storedReply{}).Error
}

func (s *Service) saveReply(row *storedReply) error {
	if s.replyDB == nil {
		return errors.New("reply database is not open")
	}
	return s.replyDB.Clauses(clause.OnConflict{UpdateAll: true}).Create(row).Error
}
