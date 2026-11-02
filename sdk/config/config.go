// Package config 提供插件配置回调及配置 Schema 的规范序列化与摘要。
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
//
// # 配置字段标签
//
// json 标签决定字段名，json:"-" 跳过字段。生成器按 Go 声明顺序构造 Schema，
// 对象保留未知键。字段的翻译路径为 config.<json 路径>，嵌套字段用点连接。
//
// form 标签采用 key=value;key=value;flag 语法，支持以下键；未知键产生构建诊断：
//   - label：显示标签的 i18 键，省略时取字段的翻译路径。
//   - widget：text、password、textarea、number、switch 或 select。
//   - options：用 | 分隔存储值；每个值可用 @ 附包内图标的相对路径，
//     例如 options=openai@assets/icons/openai.svg|custom。图标须存在于模块目录内，
//     扩展名为 .svg、.png 或 .webp；Schema 的 option_icons 保存取值到路径的映射。
//   - min、max：数字边界。
//   - default：JSON 字面量，写入 Schema 的 Default。
//   - required：无值标志，显式声明必需字段；json 的 omitempty 不影响它。
//
// description:"i18 键" 提供描述的翻译键；secret:"true" 标记 string 或 *string
// 秘密字段（包括 slice、map 元素内的字段）；hidden:"true" 隐藏表单字段。
// 这些元数据分别通过 [KeywordSecret]、[KeywordHidden]、[KeywordForm]、[KeywordI18n]
// 写入 Schema 扩展关键字。表单元数据包含 label、widget、options、option_icons、min、max。
//
// 包的组成：[Base] 提供 [Config] 的默认回调；[MarshalSchema] 保留字段顺序并规范序列化，
// [SchemaHash] 计算 REGISTER 使用的 Schema 摘要。字段契约见 spec/CONTRACTS-v0.1.md §5.1、§5.3。
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
