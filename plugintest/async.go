package plugintest

import (
	"context"

	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// Pending 保留一次异步注入的最终结果；Wait 可取消等待并在之后继续等待。
type Pending struct {
	done   chan struct{}
	result Result
	err    error
}

// Wait 等待 Core 完成或进入会话等待，取消等待保持已提交事件继续运行。
func (p *Pending) Wait(ctx context.Context) (Result, error) {
	select {
	case <-p.done:
		return p.result, p.err
	case <-ctx.Done():
		return Result{}, ctx.Err()
	}
}

// Submit 异步注入标准消息，可并行提交到不同用户或会话。
func (h *Harness) Submit(chat Chat, from User, content message.Content) *Pending {
	h.t.Helper()
	var chain message.Chain
	switch c := content.(type) {
	case message.Chain:
		chain = c
	case message.Segment:
		chain = message.Chain{c}
	default:
		h.t.Fatalf("unsupported message content %T", content)
	}
	config := h.selectInstance(chat)
	toMe := chat.chatType == message.ChatPrivate
	for _, segment := range chain {
		if mention, ok := segment.(message.MentionSegment); ok && mention.ExternalUserID == config.AccountID {
			toMe = true
		}
	}
	return h.submitEvent(from, &message.Event{Type: "message.created", AdapterID: config.AdapterID, AdapterInstanceID: config.InstanceID, AccountID: config.AccountID, Session: message.SID(config.AdapterID, chat.id, ""), ChatType: chat.chatType, Payload: &message.CreatedPayload{Chain: chain, ToMe: toMe}})
}

// WaitBackground 等待至少 count 条未归属事件的发送，随后按顺序消费。
func (h *Harness) WaitBackground(ctx context.Context, count int) ([]Sent, error) {
	for {
		h.mu.Lock()
		var background []Sent
		for _, sent := range h.sends {
			if sent.eventID == "" {
				background = append(background, sent)
			}
		}
		changed := h.changed
		h.mu.Unlock()
		if len(background) >= count {
			return h.Background(), nil
		}
		select {
		case <-changed:
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-h.ctx.Done():
			return nil, h.ctx.Err()
		}
	}
}
