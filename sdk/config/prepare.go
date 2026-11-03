package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/google/jsonschema-go/jsonschema"
)

// Prepared 保存准备后的完整候选配置，供 REGISTER 回传。
type Prepared struct {
	JSON    []byte // 规范 JSON：键排序、无多余空白、数字保留原文；包含全部未知键
	Version uint32 // target_config_version
}

// Prepare 按顺序执行 OnLoad、补缺失键默认值、typed decode 和 Schema 校验。
// cfg 接收解码结果；任何一步失败返回带协议 Code 的错误，消息仅包含路径与错误类别。
func Prepare(ctx context.Context, cfg Config, schema []byte, raw []byte, fromVersion *uint32, targetVersion uint32, restart bool) (Prepared, error) {
	object, err := prepareObject(raw)
	if err != nil {
		return Prepared{}, err
	}
	if cfg == nil {
		canonical, err := json.Marshal(object)
		return Prepared{JSON: canonical, Version: targetVersion}, err
	}

	// 1. 在独立零值实例上迁移；超时返回后，回调仍只持有该实例。
	fresh := reflect.New(reflect.TypeOf(cfg).Elem())
	version := targetVersion
	if fromVersion != nil {
		version = *fromVersion
	}
	if err := ctx.Err(); err != nil {
		return Prepared{}, prepareContextError(err)
	}
	type loadResult struct {
		raw json.RawMessage
		err error
	}
	loaded := make(chan loadResult, 1) // 超时后回调仍能结束，不阻塞发送方。
	go func() {
		result, err := fresh.Interface().(Config).OnLoad(ctx, version, bytes.Clone(raw))
		loaded <- loadResult{result, err}
	}()
	select {
	case <-ctx.Done():
		return Prepared{}, prepareContextError(ctx.Err())
	case result := <-loaded:
		if err := ctx.Err(); err != nil {
			return Prepared{}, prepareContextError(err)
		}
		if result.err != nil {
			slog.ErrorContext(ctx, "config OnLoad rejected the configuration", "err", result.err)
			return Prepared{}, prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config OnLoad rejected the configuration")
		}
		migrated, err := prepareObject(result.raw)
		if err != nil {
			return Prepared{}, err
		}
		if restart {
			before, _ := json.Marshal(object) // UseNumber 解码的 JSON 对象可直接编码。
			after, _ := json.Marshal(migrated)
			if !bytes.Equal(before, after) {
				return Prepared{}, prepareError(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "config OnLoad changed the configuration during restart")
			}
		}
		object = migrated
	}

	// 2. 只给缺失属性补默认值；完整文档始终独立于 typed Config 保存。
	var description jsonschema.Schema
	if err := json.Unmarshal(schema, &description); err != nil {
		return Prepared{}, prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config schema is invalid at $")
	}
	resolved, err := description.Resolve(&jsonschema.ResolveOptions{ValidateDefaults: true})
	if err != nil {
		return Prepared{}, prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config schema defaults are invalid at $")
	}
	if err := prepareDefaults(object, &description, resolved, "$"); err != nil {
		return Prepared{}, err
	}
	canonical, err := json.Marshal(object)
	if err != nil {
		return Prepared{}, prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config cannot be encoded at $")
	}

	// 3. 按 Go 字段类型检查 null 与精确数字范围，再解码到回调使用的实例。
	if err := prepareTyped(object, fresh.Elem().Type(), "$"); err != nil {
		return Prepared{}, err
	}
	if err := json.Unmarshal(canonical, fresh.Interface()); err != nil {
		return Prepared{}, prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config typed decode failed at $")
	}

	// 4. jsonschema-go 将 json.Number 视作 string，仅校验副本转换为数值。
	validation := prepareValidation(object)
	if err := resolved.Validate(validation); err != nil {
		path := prepareSchemaPath(validation, &description, "$")
		return Prepared{}, prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config schema validation failed at "+path)
	}
	if err := ctx.Err(); err != nil {
		return Prepared{}, prepareContextError(err)
	}
	reflect.ValueOf(cfg).Elem().Set(fresh.Elem())
	return Prepared{JSON: canonical, Version: targetVersion}, nil
}

func prepareError(code pb.ErrorCode, message string) error {
	return &validate.Error{Code: code, Message: message}
}

func prepareContextError(err error) error {
	code := pb.ErrorCode_ERROR_CODE_CANCELLED
	if err == context.DeadlineExceeded {
		code = pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED
	}
	return prepareError(code, "config preparation context ended")
}

// prepareObject 与 Core Canonical 使用相同的对象解码及尾随检查，错误不带原文。
func prepareObject(raw []byte) (map[string]any, error) {
	value, err := prepareJSON(raw)
	if err != nil {
		return nil, prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config is not valid JSON at $")
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config must be a JSON object at $")
	}
	return object, nil
}

func prepareJSON(raw []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("trailing JSON data")
	}
	return value, nil
}

func prepareDefaults(value any, schema *jsonschema.Schema, resolved *jsonschema.Resolved, path string) error {
	switch value := value.(type) {
	case map[string]any:
		// RawMessage 使 ApplyDefaults 的 json.Unmarshal 保留默认数字的原文。
		// 每层单独调用，再自行进入 items/additionalProperties；库只递归 properties。
		fields := make(map[string]json.RawMessage, len(value))
		for key, child := range value {
			raw, err := json.Marshal(child)
			if err != nil {
				return prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config cannot be encoded at "+preparePath(path, key))
			}
			fields[key] = raw
		}
		if err := resolved.ApplyDefaults(&fields); err != nil {
			return prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config defaults failed at "+path)
		}
		// ApplyDefaults 跳过 required 属性；契约允许它们通过默认值满足必需条件。
		for _, key := range schema.Required {
			child := schema.Properties[key]
			if _, present := fields[key]; !present && child != nil && child.Default != nil {
				fields[key] = child.Default
			}
		}
		for key, raw := range fields {
			if _, present := value[key]; !present {
				child, err := prepareJSON(raw)
				if err != nil {
					return prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config default is invalid at "+preparePath(path, key))
				}
				value[key] = child
			}
		}
		for key, child := range value {
			sub := schema.Properties[key]
			if sub == nil {
				sub = schema.AdditionalProperties
			}
			if sub != nil {
				resolved, err := sub.Resolve(&jsonschema.ResolveOptions{ValidateDefaults: true})
				if err != nil {
					return prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config schema is invalid at "+preparePath(path, key))
				}
				if err := prepareDefaults(child, sub, resolved, preparePath(path, key)); err != nil {
					return err
				}
			}
		}
	case []any:
		if schema.Items != nil {
			resolved, err := schema.Items.Resolve(&jsonschema.ResolveOptions{ValidateDefaults: true})
			if err != nil {
				return prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config items schema is invalid at "+path)
			}
			for i, child := range value {
				if err := prepareDefaults(child, schema.Items, resolved, fmt.Sprintf("%s[%d]", path, i)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// prepareTyped 逐层检查，使数组下标、map 键及嵌套字段出现在安全错误路径中。
// 生成器已拒绝自定义编解码和递归类型；此处只处理其支持的 Go 类型。
func prepareTyped(value any, target reflect.Type, path string) error {
	if value == nil {
		if target.Kind() != reflect.Pointer {
			return prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config null requires a pointer at "+path)
		}
		return nil
	}
	for target.Kind() == reflect.Pointer {
		target = target.Elem()
	}
	switch value := value.(type) {
	case map[string]any:
		switch target.Kind() {
		case reflect.Struct:
			fields := prepareFields(target)
			for key, child := range value {
				var field reflect.Type
				for _, candidate := range fields {
					if candidate.Name == key {
						field = candidate.Type
						break
					}
				}
				if field == nil {
					for _, candidate := range fields {
						if strings.EqualFold(candidate.Name, key) {
							field = candidate.Type
							break
						}
					}
				}
				if field != nil {
					if err := prepareTyped(child, field, preparePath(path, key)); err != nil {
						return err
					}
				}
			}
			return nil
		case reflect.Map:
			for key, child := range value {
				if err := prepareTyped(child, target.Elem(), preparePath(path, key)); err != nil {
					return err
				}
			}
			return nil
		}
	case []any:
		if target.Kind() == reflect.Slice {
			for i, child := range value {
				if err := prepareTyped(child, target.Elem(), fmt.Sprintf("%s[%d]", path, i)); err != nil {
					return err
				}
			}
			return nil
		}
	}
	raw, err := json.Marshal(value)
	if err == nil {
		err = json.Unmarshal(raw, reflect.New(target).Interface())
	}
	if err != nil {
		return prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config type or numeric range mismatch at "+path)
	}
	return nil
}

func prepareFields(target reflect.Type) []reflect.StructField {
	var fields []reflect.StructField
	for i := range target.NumField() {
		field := target.Field(i)
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "-" {
			continue
		}
		base := field.Type
		for base.Kind() == reflect.Pointer {
			base = base.Elem()
		}
		if !prepareJSONName(name) {
			name = ""
		}
		if field.Anonymous && name == "" && base.Kind() == reflect.Struct {
			fields = append(fields, prepareFields(base)...)
			continue
		}
		if !field.IsExported() {
			continue
		}
		if name != "" {
			field.Name = name
		}
		fields = append(fields, field)
	}
	return fields
}

// prepareJSONName 与 encoding/json 及生成器采用同一组合法标签字符。
func prepareJSONName(name string) bool {
	for _, character := range name {
		if !unicode.IsLetter(character) && !unicode.IsDigit(character) && !strings.ContainsRune("!#$%&()*+-./:;<=>?@[]^_{|}~ ", character) {
			return false
		}
	}
	return name != ""
}

func preparePath(parent, key string) string { return parent + "[" + strconv.Quote(key) + "]" }

// prepareValidation 只转换副本，持久化文档与 typed decode 均保留精确数字。
func prepareValidation(value any) any {
	switch value := value.(type) {
	case json.Number:
		number, err := value.Float64()
		if err != nil {
			// 已声明的数字字段已通过 Go 范围检查；超出 float64 的未知数字保留原文。
			return value
		}
		return number
	case map[string]any:
		object := make(map[string]any, len(value))
		for key, child := range value {
			object[key] = prepareValidation(child)
		}
		return object
	case []any:
		array := make([]any, len(value))
		for i, child := range value {
			array[i] = prepareValidation(child)
		}
		return array
	default:
		return value
	}
}

// prepareSchemaPath 仅在校验失败时定位最深的失败子节点，库错误可能含取值，直接丢弃。
func prepareSchemaPath(value any, schema *jsonschema.Schema, path string) string {
	switch value := value.(type) {
	case map[string]any:
		for _, key := range schema.Required {
			if _, present := value[key]; !present {
				return preparePath(path, key)
			}
		}
		for key, child := range value {
			sub := schema.Properties[key]
			if sub == nil {
				sub = schema.AdditionalProperties
			}
			if sub != nil {
				resolved, err := sub.Resolve(nil)
				if err == nil && resolved.Validate(child) != nil {
					return prepareSchemaPath(child, sub, preparePath(path, key))
				}
			}
		}
	case []any:
		if schema.Items != nil {
			resolved, err := schema.Items.Resolve(nil)
			if err == nil {
				for i, child := range value {
					if resolved.Validate(child) != nil {
						return prepareSchemaPath(child, schema.Items, fmt.Sprintf("%s[%d]", path, i))
					}
				}
			}
		}
	}
	return path
}
