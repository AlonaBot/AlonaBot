package registry

import (
	"context"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

type streamKey struct{}

// WithStreamID 绑定 Core 认证的在途流标识。
func WithStreamID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, streamKey{}, id)
}

// StreamID 读取提供方的在途流标识。
func StreamID(ctx context.Context) string { id, _ := ctx.Value(streamKey{}).(string); return id }

// StreamReceiver 在 service 包链接时接收经 Core 转发的增量请求。
var StreamReceiver func(context.Context, *wire.Inbound) wire.Response
