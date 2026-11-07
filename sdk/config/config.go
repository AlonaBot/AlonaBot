// Package config 提供插件配置回调、配置 Schema 的规范序列化与摘要，以及配置的只读取值。
//
// 配置类型嵌入 Base，根插件的 Config 方法返回它的指针：
//
//	type Config struct {
//		config.Base
//		Greeting string `json:"greeting" config:"label:Greeting;description:Text to send"`
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
// 对象保留未知键。config 标签统一承载表单与秘密字段元数据，采用 key:value;key:value;flag
// 语法。键与值按第一个冒号分隔；default 按 JSON 字面量解码，可包含分号与冒号，
// 解码结束后接分号或标签结尾；其他值到下一个分号为止。
//
// 值键：
//   - label、description：英文源文案，分别作为标题和说明；省略 label 时表单显示 JSON 名。
//   - widget：text、password、textarea、number、slider、switch、select 或 code。
//     number、slider 配合数字字段，switch 配合 bool，code 配合 string；select 须有 options。
//   - options：以 | 分隔的存储值，按字段类型解析；slice 字段按元素类型解析并显示为多选。
//   - labels、subtitles：以 | 分隔的选项显示文案与说明，按下标与 options 对齐，数量可少于 options。
//   - min、max、step：数字字段的边界与步长；slider 要求同时声明 min 与 max。
//   - language：code 编辑器的语言，缺省为 json。
//   - when：按兄弟字段的 JSON 名指定显示条件，例如 mode=regex,enabled=true；
//     字符串值原样保留，bool 与数字按 JSON 字面量解析，全部条件满足时显示字段。
//     条件只控制表单显示，字段校验与存储保持原有语义。
//   - default：JSON 字面量，写入 Schema 的 Default，只为缺失键补值。
//
// 标志：required 显式声明必需字段（json 的 omitempty 不影响它）；secret 标记 string 或 *string
// 秘密字段（含嵌套结构、slice 与 map 元素中的字段）；hidden 隐藏表单字段；obvious 醒目显示提示。
// 例如：
//
//	Mode string `json:"mode" config:"label:Mode;options:fast|safe;labels:Fast|Safe;required"`
//	Token *string `json:"token" config:"secret;widget:password"`
//	Ports []uint16 `json:"ports" config:"options:80|443;labels:HTTP|HTTPS"`
//	Volume int `json:"volume" config:"widget:slider;min:0;max:100;step:5"`
//	Script string `json:"script" config:"widget:code;language:python;when:mode=safe;obvious"`
//
// description 写入 Schema 的 Description，其余表单元数据由 [KeywordForm] 承载，
// secret 与 hidden 分别写入 [KeywordSecret]、[KeywordHidden]。翻译路径由 WebUI 按 Schema
// 的嵌套位置推导；i18/<locale>.json 的 config 按 JSON 字段名嵌套，覆盖 label、description、labels、
// subtitles，数组元素与 map 值内的字段直接嵌套在所属字段名下。表单依次取当前语言文案、标签源文案、
// JSON 名；options 保存原始存储值。字段的 JSON 名须避开这四个翻译覆盖键。
// 未知键、重复键、标志带值、值键缺值、类型或数量不符，以及残留的 form、description、secret、hidden
// 独立标签，均产生带源码位置的构建诊断。
//
// 包的组成：[Base] 提供 [Config] 的默认回调；[Prepare] 依次执行 OnLoad、补缺失键默认值、
// typed decode 与 Schema 校验，将保留未知键的完整规范 JSON 和目标版本写入 [Prepared]，供 REGISTER 回传。
// [MarshalSchema] 保留字段顺序并规范序列化，[SchemaHash] 计算 REGISTER 使用的 Schema 摘要。
// [Values] 按字段类型读取已准备的配置，[Template] 提供选中的模板键及其 items。
// 字段与准备流程契约见 spec/CONTRACTS-v0.1.md §5.1～§5.3。
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
