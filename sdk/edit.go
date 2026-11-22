package sdk

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// Edit 编辑句柄所指的消息，关联当前 Handler 的父调用。
func (c *Context) Edit(ref string, m message.Content) (*message.SendResult, error) {
	if !c.active.Load() {
		return nil, errcode.FailedPrecondition.New("handler context has expired")
	}
	return callEdit(c.ctx, ref, m, c.inbound)
}

// Edit 通过已激活的 Session 独立编辑句柄所指的消息。
func Edit(ctx context.Context, ref string, m message.Content) (*message.SendResult, error) {
	return callEdit(ctx, ref, m, nil)
}
func callEdit(ctx context.Context, ref string, m message.Content, parent *wire.Inbound) (*message.SendResult, error) {
	target, id, err := convert.RefToProto(ref)
	if err != nil {
		return nil, err
	}
	chain, err := convert.ContentToProto(m, target)
	if err != nil {
		return nil, err
	}
	return callMessage(ctx, wire.ServiceEdit, &pb.EditRequest{To: target, PlatformMessageId: id, Chain: chain}, parent)
}
