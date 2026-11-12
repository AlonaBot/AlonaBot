package sdk

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// Stream 将当前事件的文本增量发送为消息，按实例能力节流编辑并按文本上限换条。
// next 返回的文本先处理，io.EOF 正常结束；失败时按请求顺序保留全部已有回执。
func (c *Context) Stream(next func(context.Context) (string, error)) ([]*message.SendResult, error) {
	// 1. 在读取输入前检查调用生命周期和事件目标。
	if !c.active.Load() {
		return nil, errcode.FailedPrecondition.New("handler context has expired")
	}
	if c.event.GetTarget() == nil {
		return nil, errcode.NoDefaultTarget.New("no default event target")
	}
	if next == nil {
		return nil, errcode.InvalidArgument.New("stream callback is nil")
	}
	var edit bool
	var limit *message.TextLimit
	for _, capability := range c.Event().Capabilities {
		if !capability.Supported {
			continue
		}
		switch capability.Name {
		case "message.edit":
			edit = true
		case "message.send":
			limit = capability.TextLimit
		}
	}
	stream := streamOutput{context: c, limit: limit}
	// 2. 无编辑能力时只在正常结束后发送一次，输入错误时不发布缓存。
	if !edit {
		var text strings.Builder
		for {
			if err := c.ctx.Err(); err != nil {
				return nil, err
			}
			delta, err := next(c.Context())
			text.WriteString(delta)
			if canceled := c.ctx.Err(); canceled != nil {
				return nil, canceled
			}
			if errors.Is(err, io.EOF) {
				if text.Len() == 0 {
					return nil, nil
				}
				result, sendError := c.Send(message.Text(text.String()))
				if result != nil {
					stream.results = append(stream.results, result)
				}
				if canceled := c.ctx.Err(); canceled != nil {
					return stream.results, canceled
				}
				return stream.results, sendError
			}
			if err != nil {
				return nil, err
			}
		}
	}
	// 3. 编辑流逐段填充当前消息，换条前刷新旧消息，结束后刷新余量。
	for {
		if err := c.ctx.Err(); err != nil {
			return stream.results, err
		}
		delta, err := next(c.Context())
		if canceled := c.ctx.Err(); canceled != nil {
			return stream.results, canceled
		}
		if outputError := stream.append(delta); outputError != nil {
			return stream.results, outputError
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return stream.results, err
			}
			flushError := stream.flush(true)
			return stream.results, flushError
		}
	}
}

type streamOutput struct {
	context   *Context
	limit     *message.TextLimit
	text      strings.Builder
	units     uint64
	sent      bool
	displayed int                // 当前消息最后一次成功请求的 UTF-8 字节长度。
	ref       message.MessageRef // 只取首次 Send 的真实回执。
	results   []*message.SendResult
}

func (s *streamOutput) append(delta string) error {
	if delta == "" {
		return nil
	}
	for delta != "" {
		end := len(delta)
		if s.limit != nil {
			end = 0
			for offset, r := range delta {
				_, width := utf8.DecodeRuneInString(delta[offset:])
				units := uint64(1)
				switch s.limit.Unit {
				case message.TextUTF16CodeUnits:
					if r > 0xffff {
						units = 2
					}
				case message.TextBytes:
					units = uint64(width)
				}
				if units > s.limit.Max {
					s.text.WriteString(delta[:end])
					if err := s.flush(true); err != nil {
						return err
					}
					return errcode.InvalidArgument.New("a code point exceeds the text limit")
				}
				if units > s.limit.Max-s.units {
					break
				}
				s.units += units
				end = offset + width
			}
		}
		s.text.WriteString(delta[:end])
		delta = delta[end:]
		if delta != "" {
			if err := s.flush(true); err != nil {
				return err
			}
			s.text.Reset()
			s.units, s.displayed, s.sent, s.ref = 0, 0, false, ""
		}
	}
	return s.flush(false)
}

func (s *streamOutput) flush(wait bool) error {
	if s.text.Len() == s.displayed {
		return nil
	}
	if err := s.context.ctx.Err(); err != nil {
		return err
	}
	var result *message.SendResult
	var err error
	if !s.sent {
		result, err = s.context.Send(message.Text(s.text.String()))
		s.sent = true
		if result != nil {
			s.ref = result.Ref()
		}
	} else {
		if s.ref == "" {
			return errcode.FailedPrecondition.New("stream message has no editable ID")
		}
		session, sessionError := currentSession(s.context.ctx)
		if sessionError != nil {
			return sessionError
		}
		throttle := sessionStreamThrottle(session, s.context.event.Target)
		acquired, acquireError := throttle.acquire(s.context.ctx, session.Done(), wait)
		if acquireError != nil || !acquired {
			return acquireError
		}
		// 令牌跨 RPC 保持同目标的编辑顺序，状态 mutex 不跨网络调用。
		result, err = s.edit()
		throttle.release()
	}
	if result != nil {
		s.results = append(s.results, result)
	}
	if canceled := s.context.ctx.Err(); canceled != nil {
		return canceled
	}
	if err == nil {
		s.displayed = s.text.Len()
	}
	return err
}

func (s *streamOutput) edit() (*message.SendResult, error) {
	c := s.context
	if !c.active.Load() {
		return nil, errcode.FailedPrecondition.New("handler context has expired")
	}
	_, id, err := convert.RefToProto(s.ref)
	if err != nil {
		return nil, err
	}
	// 句柄只提供实际消息 ID，完整目标保留原事件账号，避免同名实例换绑后串用账号。
	target := c.event.Target
	chain, err := convert.ContentToProto(message.Text(s.text.String()), target)
	if err != nil {
		return nil, err
	}
	return callMessage(c.ctx, wire.ServiceEdit, &pb.EditRequest{To: target, PlatformMessageId: id, Chain: chain}, c.inbound)
}

const streamEditInterval = time.Second // MESSAGE-MODEL §4A 的同目标最小编辑间隔。
