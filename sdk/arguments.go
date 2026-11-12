package sdk

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// checkHandler 在收集边界检查签名，运行时只按已确定的字段和类型赋值。
func checkHandler(h any) (pb.HandlerMode, reflect.Type, []*pb.CommandParam, error) {
	t := reflect.TypeOf(h)
	if t == nil || t.Kind() != reflect.Func || reflect.ValueOf(h).IsNil() {
		return 0, nil, nil, fmt.Errorf("handler must be a non-nil function, got %T", h)
	}
	if t.IsVariadic() || t.NumIn() == 0 || t.In(0) != contextType {
		return 0, nil, nil, fmt.Errorf("handler's first parameter must be *sdk.Context and it must not be variadic: %s", t)
	}
	// 1. 返回值决定手动发送或自动回复，参数形式共用这一约定。
	var mode pb.HandlerMode
	switch {
	case t.NumOut() == 1 && t.Out(0) == errorType:
		mode = pb.HandlerMode_HANDLER_MODE_MANUAL
	case t.NumOut() == 2 && t.Out(0) == contentType && t.Out(1) == errorType:
		mode = pb.HandlerMode_HANDLER_MODE_AUTO_REPLY
	default:
		return 0, nil, nil, fmt.Errorf("handler may only return error or (message.Content, error): %s", t)
	}
	// 2. 结构体独占 Context 之后的位置，其余签名保留位置参数。
	var params []*pb.CommandParam
	for i := 1; i < t.NumIn(); i++ {
		if t.In(i).Kind() == reflect.Struct {
			if t.NumIn() != 2 {
				return 0, nil, nil, fmt.Errorf("struct parameter %d must be the only parameter after *sdk.Context", i)
			}
			params, err := structParameters(t.In(i))
			return mode, t.In(i), params, err
		}
		param, err := parameterType(t.In(i))
		if err != nil {
			return 0, nil, nil, fmt.Errorf("parameter %d: %w", i, err)
		}
		params = append(params, &pb.CommandParam{Name: fmt.Sprintf("arg%d", i), Type: param, Required: true})
	}
	return mode, nil, params, nil
}

func parameterType(t reflect.Type) (pb.ParamType, error) {
	switch t {
	case reflect.TypeFor[UserID]():
		return pb.ParamType_PARAM_TYPE_USER, nil
	case reflect.TypeFor[*message.ImageSegment]():
		return pb.ParamType_PARAM_TYPE_IMAGE, nil
	case reflect.TypeFor[message.Chain]():
		return pb.ParamType_PARAM_TYPE_MESSAGE, nil
	case reflect.TypeFor[time.Duration]():
		return 0, fmt.Errorf("type %s is not supported", t)
	}
	switch t.Kind() {
	case reflect.String:
		return pb.ParamType_PARAM_TYPE_STRING, nil
	case reflect.Int, reflect.Int64:
		return pb.ParamType_PARAM_TYPE_INT, nil
	case reflect.Float64:
		return pb.ParamType_PARAM_TYPE_FLOAT, nil
	case reflect.Bool:
		return pb.ParamType_PARAM_TYPE_BOOL, nil
	default:
		return 0, fmt.Errorf("type %s is not supported (expected string, int64, int, float64, bool, sdk.UserID, *message.ImageSegment or message.Chain)", t)
	}
}

func structParameters(t reflect.Type) ([]*pb.CommandParam, error) {
	var params []*pb.CommandParam
	seen := make(map[string]bool, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() || field.Anonymous {
			return nil, fmt.Errorf("field %s must be exported and not embedded", field.Name)
		}
		name, option, hasOption := strings.Cut(field.Tag.Get("arg"), ",")
		if hasOption && option != "raw" {
			return nil, fmt.Errorf("field %s has invalid arg option %q", field.Name, option)
		}
		if name == "" {
			name = argumentName(field.Name)
		}
		if seen[name] {
			return nil, fmt.Errorf("field %s has duplicate argument name %q", field.Name, name)
		}
		seen[name] = true
		typ := field.Type
		optional := typ.Kind() == reflect.Pointer
		// ImageSegment 本身以指针传入；其他指针字段解引用一次检查参数类型。
		if optional && typ != reflect.TypeFor[*message.ImageSegment]() {
			typ = typ.Elem()
		}
		paramType, err := parameterType(typ)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", field.Name, err)
		}
		param := &pb.CommandParam{Name: name, Type: paramType, Required: !optional, Raw: hasOption}
		if value, ok := field.Tag.Lookup("default"); ok {
			if optional {
				return nil, fmt.Errorf("field %s cannot combine a pointer and a default", field.Name)
			}
			param.Required, param.DefaultValue = false, &value
		}
		params = append(params, param)
	}
	return params, nil
}

func argumentName(name string) string {
	letters := []rune(name)
	end := 0
	for end < len(letters) && unicode.IsUpper(letters[end]) {
		end++
	}
	if end > 1 && end < len(letters) && unicode.IsLower(letters[end]) {
		end--
	}
	for i := 0; i < end; i++ {
		letters[i] = unicode.ToLower(letters[i])
	}
	return string(letters)
}
