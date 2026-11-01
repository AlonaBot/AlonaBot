package router

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// commandArgs 接收已经过注册校验的参数描述，以及命令路径之后的原始文本。
func commandArgs(params []*pb.CommandParam, text string) (*pb.CommandInvocation, error) {
	invocation := &pb.CommandInvocation{}
	for _, param := range params {
		// 1. raw 直接取得剩余原文，其他参数每次仅解码一个 token。
		var value string
		var present bool
		if param.GetRaw() {
			value = strings.TrimLeftFunc(text, unicode.IsSpace)
			present = value != ""
			if present {
				invocation.RawTail = &value
			}
			text = ""
		} else {
			var err error
			value, text, present, err = nextArgument(text)
			if err != nil {
				return nil, fmt.Errorf("parameter %q: %w", param.GetName(), err)
			}
		}

		// 2. 显式空串仍是已提供的值，缺失可选参数按协议保留无 value 的参数项。
		if !present {
			if param.GetRequired() {
				return nil, fmt.Errorf("missing required parameter %q", param.GetName())
			}
			if param.DefaultValue == nil {
				invocation.Args = append(invocation.Args, &pb.ArgValue{Name: param.GetName()})
				continue
			}
			value = param.GetDefaultValue()
		}
		arg, err := argumentValue(param, value)
		if err != nil {
			return nil, fmt.Errorf("parameter %q: %w", param.GetName(), err)
		}
		invocation.Args = append(invocation.Args, arg)
	}
	if strings.TrimLeftFunc(text, unicode.IsSpace) != "" {
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
