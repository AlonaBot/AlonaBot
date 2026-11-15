package adapter

import (
	"context"

	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// Editor 提供实例可选的消息编辑入口，完整保留编辑回执与安全错误。
// 未实现此接口的实例在收到 adapter.edit 时返回 UNSUPPORTED。
type Editor interface {
	Edit(ctx context.Context, to string, platformMessageID string, chain message.Chain) (*message.SendResult, *errcode.Error)
}
