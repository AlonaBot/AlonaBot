// Package config 定义插件配置类型嵌入的 Base。
//
// 配置类型嵌入 Base，根插件的 Config 方法返回它的指针：
//
//	type Config struct {
//		config.Base
//		Greeting string `json:"greeting"`
//	}
//
//	func (p *Demo) Config() config.Config { return &p.config }
//
// 需要迁移旧版本配置或拒绝不合法的取值时覆盖 OnLoad；它处理配置数据，与 OnStart、OnReady、OnStop
// 这些生命周期 Hook 分开。执行顺序见 spec/CONTRACTS-v0.1.md §5.2。
package config

import (
	"context"
	"encoding/json"
)

// Config 是根插件 Config 方法返回的配置；嵌入 [Base] 即满足它。
type Config interface {
	// OnLoad 在零值 *Config 上运行：只依据 raw 与 fromVersion 把旧 JSON 转成本版本 JSON，也可以拒绝
	// 不合法的取值，返回错误时本次准备失败。必须是可重复的纯变换，不做数据库、文件、网络或发信副作用。
	OnLoad(ctx context.Context, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error)
}

// Base 由配置类型以值的形式嵌入，提供 [Config] 的默认实现；它没有导出字段，不影响 JSON 解码。
type Base struct{}

// OnLoad 默认原样返回 raw。
func (*Base) OnLoad(_ context.Context, _ uint32, raw json.RawMessage) (json.RawMessage, error) {
	return raw, nil
}
