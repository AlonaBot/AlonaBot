// Package config 定义插件配置的可选处理接口。
//
// 配置类型可以实现 ConfigLoader，把旧版本的配置 JSON 转换为当前版本；也可以实现
// ConfigValidator，在解码之后检查业务约束。它们处理配置数据，与 OnStart、OnReady、OnStop
// 这些生命周期 Hook 分开，不需要时不必实现。方法名相同但签名不符时 sdk.Collect 报错，不会静默跳过。
// 执行顺序见 spec/CONTRACTS-v0.1.md §5.2。
package config

import (
	"context"
	"encoding/json"
)

// Loader 在零值 *Config 上运行：只依据 raw 与 fromVersion 把旧 JSON 转成本版本 JSON，
// 必须是可重复的纯变换，不做数据库、文件、网络或发信副作用。
type Loader interface {
	// OnLoad 把 fromVersion 版本的 raw 转换为本版本的 JSON。
	OnLoad(ctx context.Context, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error)
}

// Validator 在 typed decode 之后执行。
type Validator interface {
	// Validate 校验已解码的配置，返回错误时本次准备失败。
	Validate(ctx context.Context) error
}
