package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

// Canonical 把 raw 转成键排序、无多余空白的 JSON 对象；数字保持原文，不经过 float64。
// 非对象或语法错误返回 INVALID_ARGUMENT。
func Canonical(raw []byte) ([]byte, error) {
	object, err := decodeObject(raw, "config")
	if err != nil {
		return nil, err
	}
	return json.Marshal(object)
}

// MergePatch 按 RFC 7396 把 patch 合并进 doc：null 删除键，数组整体替换，对象递归合并。
// doc 与 patch 都必须是 JSON 对象，输出为 Canonical 形式。
func MergePatch(doc, patch []byte) ([]byte, error) {
	target, err := decodeObject(doc, "config")
	if err != nil {
		return nil, err
	}
	updates, err := decodeObject(patch, "patch")
	if err != nil {
		return nil, err
	}
	return json.Marshal(apply(target, updates))
}

// apply 在 target 上原地应用 RFC 7396 的合并规则并返回它。
func apply(target, patch map[string]any) map[string]any {
	for key, value := range patch {
		if value == nil {
			delete(target, key)
			continue
		}
		if updates, ok := value.(map[string]any); ok {
			current, ok := target[key].(map[string]any)
			if !ok {
				current = map[string]any{}
			}
			target[key] = apply(current, updates)
			continue
		}
		target[key] = value
	}
	return target
}

// decodeObject 解码一个 JSON 对象，数字保留为 json.Number；name 用于错误说明。
func decodeObject(raw []byte, name string) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, configError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "%s is not valid JSON: %v", name, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, configError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "%s has trailing data after the JSON value", name)
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, configError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "%s must be a JSON object", name)
	}
	return object, nil
}

// configError 构造带协议错误码的 *validate.Error。
func configError(code pb.ErrorCode, format string, args ...any) error {
	return &validate.Error{Code: code, Message: fmt.Sprintf(format, args...)}
}
