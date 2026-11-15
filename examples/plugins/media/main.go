// media 命令演示上传内置图片、相册与带类型按钮的文本编辑。
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	_ "embed"
	"errors"
	"log/slog"
	"os"
	"sync"

	"github.com/AlonaBot/AlonaBot/sdk"
	buttonapi "github.com/AlonaBot/AlonaBot/sdk/button"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	mediaapi "github.com/AlonaBot/AlonaBot/sdk/media"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"github.com/AlonaBot/AlonaBot/sdk/sqlitedb"
)

//go:embed assets/sample.png
var sample []byte

const (
	captionKey = "media.caption"
	promptKey  = "media.prompt"
	buttonKey  = "media.button"
	editedKey  = "media.edited"
)

// New 绑定资源并登记媒体命令与一次性编辑按钮。
func New() *sdk.Plugin {
	p := sdk.NewPlugin()

	p.Sub("captions").Permission(sdk.User).Command("caption").Handle(caption)
	edit := buttonapi.Define[string]("media.edit")
	var mu sync.Mutex
	sqlitedb.Migrate(p, func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS messages (button_key TEXT PRIMARY KEY, ref TEXT NOT NULL)")
		return err
	})
	p.Command("media").Handle(func(c *sdk.Context) error {
		c.StopPropagation()
		// 1. 通过 Upload 取得媒体引用，再用同一来源构造相邻图片。
		ref, err := mediaapi.Upload(c.Context(), bytes.NewReader(sample), "image/png")
		if err != nil {
			return err
		}
		// 锁覆盖发送与保存，确保先到达的点击能读到已确认句柄。
		mu.Lock()
		defer mu.Unlock()
		key := rand.Text()
		result, err := c.Send(message.Chain{
			message.ImageSegment{Source: ref.Source(), Caption: c.T(captionKey)},
			message.ImageSegment{Source: ref.Source(), Caption: c.T(captionKey)},
			message.Text(c.T(promptKey)),
			message.Buttons([]message.Button{edit.Button(c.T(buttonKey), key)}),
		})
		if err != nil {
			return err
		}
		// 2. 相册 ID 排在文本之前；末项没有 ID 时保留已发送事实，跳过编辑。
		if result.PartResults[len(result.PartResults)-1].MessageIDsUnavailable {
			return errcode.FailedPrecondition.New("telegram accepted the text without an editable message ID")
		}
		messages := result.Refs()
		db, err := sqlitedb.Open(c.Context())
		if err != nil {
			return err
		}
		_, err = db.ExecContext(c.Context(), "INSERT INTO messages (button_key, ref) VALUES (?, ?)", key, string(messages[len(messages)-1]))
		return err
	})
	buttonapi.On(p, edit).Handle(func(c *sdk.Context, key string) error {
		mu.Lock()
		defer mu.Unlock()
		db, err := sqlitedb.Open(c.Context())
		if err != nil {
			return err
		}
		var ref string
		if err := db.QueryRowContext(c.Context(), "SELECT ref FROM messages WHERE button_key = ?", key).Scan(&ref); errors.Is(err, sql.ErrNoRows) {
			return errcode.FailedPrecondition.New("media example message handle is no longer available")
		} else if err != nil {
			return err
		}
		if _, err := c.Edit(ref, message.Text(c.T(editedKey))); err != nil {
			return err
		}
		_, err = db.ExecContext(c.Context(), "DELETE FROM messages WHERE button_key = ?", key)
		return err
	})
	return p
}

// Caption 以末位 raw 字段保留用户提供的整段图片说明。
type Caption struct {
	Text string `arg:"text,raw"`
}

func caption(c *sdk.Context, args Caption) (message.Content, error) {
	ref, err := mediaapi.Upload(c.Context(), bytes.NewReader(sample), "image/png")
	if err != nil {
		return nil, err
	}
	return message.ImageSegment{Source: ref.Source(), Caption: args.Text}, nil
}

func main() {
	if err := sdk.Run(New()); err != nil {
		slog.Error("media plugin stopped with an error", "error", err)
		os.Exit(1)
	}
}
