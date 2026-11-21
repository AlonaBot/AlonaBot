// Package action 调用 Adapter 的平台动作，后台客户端保留任务的取消与期限（CONTRACTS §1.4）。
package action

import (
	"context"
	"encoding/json"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
)

// Caller 为 Adapter 的带类型动作函数提供统一调用入口。
type Caller interface {
	CallAction(to string, name string, version uint32, params any, opts ...sdk.SendOption) (json.RawMessage, error)
}

var (
	_ Caller = (*sdk.Context)(nil)
	_ Caller = (*Client)(nil)
	_ Caller = (*sdk.AdapterClient)(nil)
)

// Client 保存后台任务的取消与期限上下文，每次调用使用当前已激活的 Session。
type Client struct {
	ctx context.Context
}

// NewClient 以后台任务的 context 创建动作客户端。
func NewClient(ctx context.Context) *Client { return &Client{ctx: ctx} }

// CallAction 独立调用平台动作；opts 只接受 Via，账号动作传空会话。
func (c *Client) CallAction(to string, name string, version uint32, params any, opts ...sdk.SendOption) (json.RawMessage, error) {
	return registry.Action(c.ctx, to, name, version, params, opts)
}
