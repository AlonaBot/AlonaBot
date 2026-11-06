package router

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/proto"
)

var errCommandArguments = errors.New("invalid command arguments")

// argumentCursor 保留原片段位置；offset 为当前 text 中的 UTF-8 字节偏移，实体仍按原 UTF-16 范围解释。
type argumentCursor struct {
	segments []*pb.Segment
	segment  int
	offset   int
}

type argumentToken struct {
	text    string
	textual bool
	userID  string
	image   *pb.ImageSegment
}

func (c *argumentCursor) trimSpace() {
	for c.segment < len(c.segments) {
		text := c.segments[c.segment].GetText()
		if text == nil {
			return
		}
		suffix := text.GetText()[c.offset:]
		trimmed := strings.TrimLeftFunc(suffix, unicode.IsSpace)
		c.offset += len(suffix) - len(trimmed)
		if trimmed != "" {
			return
		}
		c.segment++
		c.offset = 0
	}
}

// skipText 只用于跳过已经匹配的前缀与路径，其长度由连续开头文本给出。
func (c *argumentCursor) skipText(bytes int) {
	for bytes > 0 {
		left := len(c.segments[c.segment].GetText().GetText()) - c.offset
		if bytes < left {
			c.offset += bytes
			return
		}
		bytes -= left
		c.segment++
		c.offset = 0
	}
}

func (c *argumentCursor) next() (argumentToken, bool, error) {
	for {
		c.trimSpace()
		if c.segment == len(c.segments) {
			return argumentToken{}, false, nil
		}
		segment := c.segments[c.segment]
		switch value := segment.GetData().(type) {
		case *pb.Segment_Text:
			if token, end, ok := textMention(value.Text, c.offset); ok {
				c.offset = end
				return token, true, nil
			}
			suffix := value.Text.GetText()[c.offset:]
			text, rest, present, err := nextArgument(suffix)
			c.offset += len(suffix) - len(rest)
			return argumentToken{text: text, textual: true}, present, err
		case *pb.Segment_Mention:
			c.segment++
			return argumentToken{userID: value.Mention.GetExternalUserId()}, true, nil
		case *pb.Segment_Image:
			c.segment++
			return argumentToken{image: value.Image}, true, nil
		default:
			c.segment++
		}
	}
}

func utf16Length(text string) int {
	n := 0
	for _, r := range text {
		n += utf16.RuneLen(r)
	}
	return n
}

func textMention(text *pb.TextSegment, start int) (argumentToken, int, bool) {
	value := text.GetText()
	if start > 0 {
		previous, _ := utf8.DecodeLastRuneInString(value[:start])
		if !unicode.IsSpace(previous) {
			return argumentToken{}, 0, false
		}
	}
	startUTF16 := utf16Length(value[:start])
	for _, entity := range text.GetEntities() {
		if entity.GetType() != pb.EntityType_ENTITY_TYPE_TEXT_MENTION || int(entity.GetOffsetUtf16()) != startUTF16 {
			continue
		}
		end, length := start, 0
		for _, r := range value[start:] {
			if length >= int(entity.GetLengthUtf16()) {
				break
			}
			end += utf8.RuneLen(r)
			length += utf16.RuneLen(r)
		}
		if length != int(entity.GetLengthUtf16()) {
			continue
		}
		if end < len(value) {
			next, _ := utf8.DecodeRuneInString(value[end:])
			if !unicode.IsSpace(next) {
				continue
			}
		}
		return argumentToken{text: value[start:end], textual: true, userID: entity.GetExternalUserId()}, end, true
	}
	return argumentToken{}, 0, false
}

func (c *argumentCursor) rawTail() (string, error) {
	c.trimSpace()
	var text strings.Builder
	for c.segment < len(c.segments) {
		segment := c.segments[c.segment]
		if segment.GetMention() != nil || segment.GetImage() != nil {
			return "", errors.New("non-text token in raw argument")
		}
		if segment.GetText() != nil {
			text.WriteString(segment.GetText().GetText()[c.offset:])
		}
		c.segment++
		c.offset = 0
	}
	return text.String(), nil
}

func (c *argumentCursor) messageTail() *pb.MessageChain {
	c.trimSpace()
	if c.segment == len(c.segments) {
		return nil
	}
	segments := append([]*pb.Segment(nil), c.segments[c.segment:]...)
	if c.offset != 0 {
		segments[0] = proto.CloneOf(segments[0])
		text := segments[0].GetText()
		removed := uint32(utf16Length(text.Text[:c.offset]))
		text.Text = text.Text[c.offset:]
		var entities []*pb.TextEntity
		for _, entity := range text.Entities {
			end := entity.OffsetUtf16 + entity.LengthUtf16
			if end <= removed {
				continue
			}
			if entity.OffsetUtf16 < removed {
				// 被切断的 text_mention 不再代表一个完整用户标签；样式范围可裁到剩余文字。
				if entity.Type == pb.EntityType_ENTITY_TYPE_TEXT_MENTION {
					continue
				}
				entity.OffsetUtf16 = removed
				entity.LengthUtf16 = end - removed
			}
			entity.OffsetUtf16 -= removed
			entities = append(entities, entity)
		}
		text.Entities = entities
	}
	c.segment = len(c.segments)
	c.offset = 0
	return &pb.MessageChain{Segments: segments}
}

// commandArgs 从已准备的接收方事件片段中绑定参数；各 text 的引号与转义独立。
func commandArgs(params []*pb.CommandParam, cursor argumentCursor) (*pb.CommandInvocation, error) {
	invocation := &pb.CommandInvocation{}
	for _, param := range params {
		// 1. message/raw 保留原始尾部；其他类型消费一个文本或片段 token。
		arg := &pb.ArgValue{Name: param.GetName()}
		var token argumentToken
		var present bool
		var err error
		switch {
		case param.GetType() == pb.ParamType_PARAM_TYPE_MESSAGE:
			chain := cursor.messageTail()
			present = chain != nil
			if present {
				arg.Value = &pb.ArgValue_Message{Message: chain}
			}
		case param.GetRaw():
			token.text, err = cursor.rawTail()
			token.textual, present = true, token.text != ""
			if present {
				invocation.RawTail = proto.String(token.text)
			}
		default:
			token, present, err = cursor.next()
		}
		if err != nil {
			return nil, fmt.Errorf("parameter %q: %w", param.GetName(), err)
		}
		// 2. 空字符串 token 与缺失位置分开，可选参数按原契约绑定零值或默认值。
		if !present {
			if param.GetRequired() {
				return nil, fmt.Errorf("missing required parameter %q", param.GetName())
			}
			if param.DefaultValue == nil {
				invocation.Args = append(invocation.Args, arg)
				continue
			}
			token = argumentToken{text: param.GetDefaultValue(), textual: true}
		}
		switch param.GetType() {
		case pb.ParamType_PARAM_TYPE_MESSAGE:
		case pb.ParamType_PARAM_TYPE_USER:
			if token.userID == "" {
				return nil, fmt.Errorf("parameter %q requires a user token", param.GetName())
			}
			arg.Value = &pb.ArgValue_UserId{UserId: token.userID}
		case pb.ParamType_PARAM_TYPE_IMAGE:
			if token.image == nil {
				return nil, fmt.Errorf("parameter %q requires an image token", param.GetName())
			}
			arg.Value = &pb.ArgValue_Image{Image: token.image}
		default:
			if !token.textual {
				return nil, fmt.Errorf("parameter %q requires a text token", param.GetName())
			}
			arg, err = argumentValue(param, token.text)
			if err != nil {
				return nil, fmt.Errorf("parameter %q: %w", param.GetName(), err)
			}
		}
		invocation.Args = append(invocation.Args, arg)
	}
	if _, present, err := cursor.next(); err != nil {
		return nil, err
	} else if present {
		return nil, errors.New("unexpected extra arguments")
	}
	return invocation, nil
}

// nextArgument 保留未消费部分的原文，以便末位 raw 接收原始引号与空白。
func nextArgument(text string) (value, rest string, present bool, err error) {
	text = strings.TrimLeftFunc(text, unicode.IsSpace)
	if text == "" {
		return "", "", false, nil
	}
	var decoded strings.Builder
	var quote rune
	escaped := false
	for i, r := range text {
		switch {
		case escaped:
			decoded.WriteRune(r)
			escaped = false
		case quote == '\'':
			if r == '\'' {
				quote = 0
			} else {
				decoded.WriteRune(r)
			}
		case r == '\\':
			escaped = true
		case quote == '"':
			if r == '"' {
				quote = 0
			} else {
				decoded.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
		case unicode.IsSpace(r):
			return decoded.String(), text[i:], true, nil
		default:
			decoded.WriteRune(r)
		}
	}
	if escaped {
		return "", "", false, errors.New("trailing escape in argument")
	}
	if quote != 0 {
		return "", "", false, errors.New("unterminated quoted argument")
	}
	return decoded.String(), "", true, nil
}

func argumentValue(param *pb.CommandParam, text string) (*pb.ArgValue, error) {
	arg := &pb.ArgValue{Name: param.GetName()}
	switch param.GetType() {
	case pb.ParamType_PARAM_TYPE_STRING:
		arg.Value = &pb.ArgValue_StringValue{StringValue: text}
	case pb.ParamType_PARAM_TYPE_INT:
		value, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return nil, err
		}
		arg.Value = &pb.ArgValue_IntValue{IntValue: value}
	case pb.ParamType_PARAM_TYPE_FLOAT:
		value, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return nil, err
		}
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, errors.New("float must be finite")
		}
		arg.Value = &pb.ArgValue_FloatValue{FloatValue: value}
	case pb.ParamType_PARAM_TYPE_BOOL:
		if text != "true" && text != "false" {
			return nil, errors.New("bool must be true or false")
		}
		arg.Value = &pb.ArgValue_BoolValue{BoolValue: text == "true"}
	default:
		panic("unvalidated command parameter type")
	}
	return arg, nil
}
