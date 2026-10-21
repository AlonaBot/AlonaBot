// Package config 定义可选的配置处理接口（CONTRACTS §5.2）。它们是配置数据回调，
// 与 OnStart/OnReady/OnStop 生命周期 Hook 分开，不需要空实现。
package config

import (
	"context"
	"encoding/json"
)

// Provider 由根插件实现，返回唯一 Config 结构体的指针；每个插件包至多一个 Config。
type Provider interface {
	Config() any
}

// Loader 在零值 *Config 上运行：只依据 raw 与 fromVersion 把旧 JSON 转成本版本 JSON，
// 必须是可重复的纯变换，不做数据库、文件、网络或发信副作用。
type Loader interface {
	OnLoad(ctx context.Context, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error)
}

// Validator 在 typed decode 之后执行。
type Validator interface {
	Validate(ctx context.Context) error
}
