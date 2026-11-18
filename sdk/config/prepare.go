package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

// Loader 在补值与校正之前处理完整原始配置。
// 回调应为可重复的纯变换，通过 ctx 响应准备阶段的取消与预算。
type Loader func(ctx context.Context, raw json.RawMessage) (json.RawMessage, error)

// Prepared 保存准备后的完整候选配置和只读取值，供 REGISTER 与根插件使用。
type Prepared struct {
	JSON        []byte // 规范 JSON：键排序、无多余空白、数字保留原文；包含全部未知键。
	Corrections []configschema.Correction
	Values      Values
}

// Prepare 依次执行 load、补值与校正、required 校验，再构造完整 JSON 与 Values。
// 失败返回 *validate.Error，消息只含路径与错误类别；restart 时完整结果须保持配置不变。
func Prepare(ctx context.Context, load Loader, schema []byte, raw []byte, restart bool) (Prepared, error) {
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

	before, _ := json.Marshal(object) // UseNumber 解码的 JSON 对象可直接编码。
	// 1. 回调只持有独立 JSON 字节；超时返回后不再接纳其结果。
	if load != nil {
		type loadResult struct {
			raw json.RawMessage
			err error
		}
		loaded := make(chan loadResult, 1) // 超时后回调仍能结束，不阻塞发送方。
		if err := ctx.Err(); err != nil {
			return Prepared{}, prepareContextError(err)
		}
		go func() {
			result, err := load(ctx, bytes.Clone(before))
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
				return Prepared{}, prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config OnConfigLoad rejected the configuration")
			}
			transformed, err := prepareObject(result.raw)
			if err != nil {
				return Prepared{}, err
			}
			object = transformed
		}
	}

	// 2. 共享校正保留未知键；普通重启同时核对回调和校正的完整结果。
	candidate, _ := json.Marshal(object)
	canonical, corrections, err := description.Correct(candidate)
	if err != nil {
		return Prepared{}, err
	}
	if restart && !bytes.Equal(before, canonical) {
		return Prepared{}, prepareError(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "config changed during restart")
	}
	values, err := newValues(schema, canonical)
	if err != nil {
		return Prepared{}, prepareError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "config value decoding failed at $")
	}
	if err := ctx.Err(); err != nil {
		return Prepared{}, prepareContextError(err)
	}
	return Prepared{JSON: canonical, Corrections: corrections, Values: values}, nil
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
