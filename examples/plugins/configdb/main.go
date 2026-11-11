// configdb 演示配置迁移、受管 SQLite 与本地化的笔记命令。
//
// 默认构建只链接 sqlitedb，使用 configdb_gorm 标签可切换到共用连接的 GORM 入口。
// fail_migration 与 fail_start 供临时 data_dir 中的候选失败验收使用。
package main

//go:generate go run ../../../cmd/alonabot plugin generate

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

const (
	legacyConfigVersion  = 1
	createNotesMigration = 1
	indexNotesMigration  = 2
	startupKey           = "startup"
	notesTable           = "configdb_note"
	startupTable         = "configdb_startup"
)

//go:embed notes.sql
var notesDDL string

//go:embed note_index.sql
var noteIndexDDL string

type plugin struct {
	sdk.Base
}

// OnConfigLoad 将版本 1 的问候词数组拼接成版本 2 的字符串，保留完整原始对象。
func (*plugin) OnConfigLoad(ctx context.Context, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if fromVersion != legacyConfigVersion {
		return raw, nil
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, fmt.Errorf("decode config object: %w", err)
	}
	if greeting, exists := object["greeting"]; exists {
		var words []string
		if err := json.Unmarshal(greeting, &words); err != nil {
			return nil, fmt.Errorf("decode legacy greeting array: %w", err)
		}
		converted, err := json.Marshal(strings.Join(words, " "))
		if err != nil {
			return nil, fmt.Errorf("encode greeting: %w", err)
		}
		object["greeting"] = converted
	}
	return json.Marshal(object)
}

// Commands 注册 /note set、get、delete 与语言选择示例 /note hello。
func (*plugin) Commands(r *sdk.CommandRegistry) {
	g := r.Group("note")
	g.Command("set").Handle(func(c *sdk.Context, key, value string) (message.Chain, error) {
		if err := saveNote(c.Context(), key, value); err != nil {
			return nil, err
		}
		return reply(c, "note.saved", "key", key)
	})
	g.Command("get").Handle(func(c *sdk.Context, key string) (message.Chain, error) {
		value, err := readNote(c.Context(), key)
		if err != nil {
			return nil, err
		}
		return reply(c, "note.value", "key", key, "value", value)
	})
	g.Command("delete").Handle(func(c *sdk.Context, key string) (message.Chain, error) {
		if err := deleteNote(c.Context(), key); err != nil {
			return nil, err
		}
		return reply(c, "note.deleted", "key", key)
	})
	g.Command("hello").Handle(func(c *sdk.Context, name string) (message.Chain, error) {
		return reply(c, "note.hello", "name", name)
	})
}

func reply(c *sdk.Context, key string, args ...any) (message.Chain, error) {
	text, err := c.T(key, args...)
	if err != nil {
		return nil, err
	}
	return message.Chain{message.Text(text)}, nil
}

func (p *plugin) started() error {
	if p.Config().Bool("fail_start") {
		return errors.New("requested failure after startup database write")
	}
	slog.Info("configdb started")
	return nil
}

func main() {
	if err := sdk.Run(&plugin{}); err != nil {
		slog.Error("configdb stopped with an error", "err", err)
		os.Exit(1)
	}
}
