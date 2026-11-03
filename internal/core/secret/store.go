// Package secret 管理配置快照中的不可变秘密引用。
//
// Store 在 Bootstrap 前解析所属插件的引用，在 REGISTER 后按 Schema 与保留路径抽取明文，
// 并为管理读取提供脱敏结果。Collect 的调用方须提供 Active、Pending、保留代次和恢复操作的
// 全部受保护快照，并在逐插件管理锁内协调抽取、持久化与清理。规则见 CONTRACTS §5.3。
//
// 包的组成：[New] 绑定 Core 状态库；[Store.Resolve] 与 [Store.Extract] 转换运行配置和快照；
// [Store.Redact] 隐藏引用，[Store.Collect] 清理未受保护的版本。
package secret

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// RefKey 标记持久化配置中的秘密引用。
const RefKey = "$secret_ref"

const (
	keywordSecret = "x-alonabot-secret" // 与 sdk/config 的关键字一致，Core 保持独立依赖方向。
	setKey        = "$secret_set"
	refBytes      = 16
)

// Store 通过 Core 状态库保存不可变秘密版本；管理调用由调用方持逐插件锁串行化。
type Store struct{ state *state.Store }

// New 绑定已经打开的 Core 状态库。
func New(st *state.Store) *Store { return &Store{state: st} }

// Resolve 把快照引用转换为所属插件的明文；越权和缺失分别返回 PERMISSION_DENIED、NOT_FOUND。
func (s *Store) Resolve(ctx context.Context, pluginID string, snapshot []byte) ([]byte, error) {
	object, err := decodeObject(snapshot)
	if err != nil {
		return nil, err
	}
	result, err := walk(object, nil, func(path []string, value any) (any, bool, error) {
		id, found, err := reference(value, path)
		if err != nil || !found {
			return nil, false, err
		}
		row, err := s.read(ctx, pluginID, id, path)
		return string(row.Value), true, err
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// Extract 按 Schema 与 keptPaths 抽取秘密值，保留未知字段；同位置的明文未变时复用原引用。
// 新值创建独立版本；失败时产生的孤立版本由调用方在安全时机通过 Collect 清理。
func (s *Store) Extract(ctx context.Context, pluginID string, schema, candidate, previous []byte, keptPaths []string) ([]byte, []string, error) {
	// 1. 解码对象并合并路径模式；Schema 来自构建期生成器，仅沿生成器创建的结构边读取。
	object, err := decodeObject(candidate)
	if err != nil {
		return nil, nil, err
	}
	var prior any
	if len(previous) > 0 {
		prior, err = decodeObject(previous)
		if err != nil {
			return nil, nil, err
		}
	}
	patterns := make(map[string][]string)
	for _, path := range keptPaths {
		patterns[path] = strings.Split(path, ".")
	}
	if len(schema) > 0 {
		description, err := decodeObject(schema)
		if err != nil {
			return nil, nil, err
		}
		schemaPaths(description, nil, patterns)
	}
	// 2. 每个具体位置只处理一次，数组下标和 map 键均用于查找上一快照的同位置值。
	retained := make(map[string]bool)
	result, err := walk(object, nil, func(path []string, value any) (any, bool, error) {
		if object, ok := value.(map[string]any); ok {
			if _, found := object[RefKey]; found {
				return nil, false, pathError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, path, "plugin returned a secret reference")
			}
		}
		matched := false
		for pattern, parts := range patterns {
			if matches(parts, path) {
				retained[pattern] = true
				matched = true
			}
		}
		if !matched || value == nil {
			return nil, false, nil
		}
		plaintext, ok := value.(string)
		if !ok {
			return nil, false, pathError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, path, "secret must be a string or null")
		}
		id, found, err := reference(at(prior, path), path)
		if err != nil {
			return nil, false, err
		}
		if found {
			row, err := s.read(ctx, pluginID, id, path)
			if err != nil {
				return nil, false, err
			}
			if string(row.Value) == plaintext {
				return map[string]any{RefKey: id}, true, nil
			}
		}
		random := make([]byte, refBytes)
		if _, err := rand.Read(random); err != nil {
			return nil, false, fmt.Errorf("generate secret reference: %w", err)
		}
		id = hex.EncodeToString(random)
		if err := s.state.AddSecret(ctx, state.Secret{RefID: id, PluginID: pluginID, Value: []byte(plaintext), CreatedAt: time.Now().UnixMilli()}); err != nil {
			return nil, false, fmt.Errorf("store secret at %s: %w", displayPath(path), err)
		}
		return map[string]any{RefKey: id}, true, nil
	})
	if err != nil {
		return nil, nil, err
	}
	// 3. 只保留候选中仍有具体位置的模式，并规范编码快照与路径集合。
	paths := make([]string, 0, len(retained))
	for path := range retained {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	snapshot, err := json.Marshal(result)
	return snapshot, paths, err
}

// Redact 把每个秘密引用转换为设置状态，供管理接口显示。
func (s *Store) Redact(snapshot []byte) ([]byte, error) {
	object, err := decodeObject(snapshot)
	if err != nil {
		return nil, err
	}
	result, err := walk(object, nil, func(path []string, value any) (any, bool, error) {
		_, found, err := reference(value, path)
		return map[string]any{setKey: true}, found, err
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// Collect 删除该插件中未被 protected 快照引用的版本；快照解析失败时保留全部记录。
func (s *Store) Collect(ctx context.Context, pluginID string, protected [][]byte) (int, error) {
	// 1. 先列出所属插件的版本并完整收集保护引用，再执行任何删除。
	rows, err := s.state.ListSecrets(ctx, pluginID)
	if err != nil {
		return 0, err
	}
	retained := make(map[string]bool)
	for _, snapshot := range protected {
		object, err := decodeObject(snapshot)
		if err != nil {
			return 0, err
		}
		_, err = walk(object, nil, func(path []string, value any) (any, bool, error) {
			id, found, err := reference(value, path)
			if found {
				retained[id] = true
			}
			return nil, found, err
		})
		if err != nil {
			return 0, err
		}
	}
	// 2. 在 Go 中计算差集，按主键批量删除；空集合由状态库跳过。
	var unused []string
	for _, row := range rows {
		if !retained[row.RefID] {
			unused = append(unused, row.RefID)
		}
	}
	return s.state.DeleteSecrets(ctx, pluginID, unused)
}

func (s *Store) read(ctx context.Context, pluginID, id string, path []string) (state.Secret, error) {
	row, err := s.state.GetSecret(ctx, pluginID, id)
	switch {
	case errors.Is(err, state.ErrSecretOwnerMismatch):
		return state.Secret{}, pathError(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, path, "secret access denied")
	case errors.Is(err, state.ErrNotFound):
		return state.Secret{}, pathError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, path, "secret not found")
	case err != nil:
		return state.Secret{}, fmt.Errorf("read secret at %s: %w", displayPath(path), err)
	}
	return row, nil
}

func pathError(code pb.ErrorCode, path []string, reason string) error {
	return &validate.Error{Code: code, Message: reason + " at " + displayPath(path)}
}
