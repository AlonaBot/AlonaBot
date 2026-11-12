package adapter

import (
	"context"

	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// Deleter 提供实例可选的消息撤回入口。
// 结果按输入顺序保留每个 ID；平台明确失败或结果未知后停止，其余记为 PartNotAttempted。
type Deleter interface {
	Delete(ctx context.Context, to message.Session, platformMessageIDs []string) (*message.DeleteResult, *errcode.Error)
}
