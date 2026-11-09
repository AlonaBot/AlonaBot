package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

// Loader 在声明式迁移之后处理完整配置；fromVersion 保留迁移前的版本号。
// 回调应为可重复的纯变换，通过 ctx 响应准备阶段的取消与预算。
type Loader func(ctx context.Context, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error)

// Prepared 保存准备后的完整候选配置和只读取值，供 REGISTER 与根插件使用。
type Prepared struct {
	JSON    []byte // 规范 JSON：键排序、无多余空白、数字保留原文；包含全部未知键。
	Version uint32 // target_config_version
	Values  Values
}

// Prepare 依次执行声明式迁移、load、补缺失值和 schema 校验，再构造完整 JSON 与 Values。
// fromVersion 为 nil 时跳过迁移，并把 targetVersion 交给 load；load 为 nil 时跳过回调。
// 失败返回 *validate.Error，消息只含路径与错误类别；restart 时回调须保持配置不变。
func Prepare(ctx context.Context, load Loader, schema []byte, raw []byte, fromVersion *uint32, targetVersion uint32, restart bool) (Prepared, error) {
	object, err := prepareObject(raw)
	if err != nil {
		return Prepared{}, err
	}
	if len(schema) == 0 {
		schema = []byte(`{}`)
	}
	description, problems := configschema.Parse(schema)
	if len(problems) != 0 {
		return Prepared{}, prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config schema is invalid at "+problems[0].Path+": "+problems[0].Message)
	}
	if err := ctx.Err(); err != nil {
		return Prepared{}, prepareContextError(err)
	}

	// 1. 先迁移完整对象，回调接收的版本仍为迁移前的版本；首装不迁移。
	version := targetVersion
	if fromVersion != nil {
		version = *fromVersion
		if err := description.Migrate(object, int(version)); err != nil {
			return Prepared{}, err
		}
	}

	// 2. 回调只持有独立 JSON 字节；超时返回后不再接纳其结果。
	if load != nil {
		before, _ := json.Marshal(object) // 对象来自 UseNumber 解码与已检查的声明式迁移，可直接编码。
		type loadResult struct {
			raw json.RawMessage
			err error
		}
		loaded := make(chan loadResult, 1) // 超时后回调仍能结束，不阻塞发送方。
		if err := ctx.Err(); err != nil {
			return Prepared{}, prepareContextError(err)
		}
		go func() {
			result, err := load(ctx, version, before)
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
				slog.ErrorContext(ctx, "config OnConfigLoad rejected the configuration", "err", result.err)
				return Prepared{}, prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config OnConfigLoad rejected the configuration")
			}
			transformed, err := prepareObject(result.raw)
			if err != nil {
				return Prepared{}, err
			}
			if restart {
				// 回调可以复用输入字节；重新编码迁移结果，避免原地修改绕过比较。
				before, _ := json.Marshal(object)
				after, _ := json.Marshal(transformed)
				if !bytes.Equal(before, after) {
					return Prepared{}, prepareError(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "config OnConfigLoad changed the configuration during restart")
				}
			}
			object = transformed
		}
	}

	// 3. 只补缺失的声明字段，未知键与显式值保存在完整文档中。
	description.Complete(object)
	if err := description.Validate(object); err != nil {
		return Prepared{}, err
	}
	canonical, err := json.Marshal(object)
	if err != nil {
		return Prepared{}, prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config cannot be encoded at $")
	}
	values, err := newValues(schema, canonical)
	if err != nil {
		return Prepared{}, prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config value decoding failed at $")
	}
	if err := ctx.Err(); err != nil {
		return Prepared{}, prepareContextError(err)
	}
	return Prepared{JSON: canonical, Version: targetVersion, Values: values}, nil
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
