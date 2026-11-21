// Package i18n 提供 Core 发给聊天用户的内置双语文案目录。
//
// [Messages] 返回嵌入的 en-US 与 zh-CN 文案，供调用方按 Actor 与 Core 默认语言选择。
// 文件格式与语言回落的用法见 docs/plugin-guide.md。
package i18n

import (
	_ "embed"
	"fmt"

	"github.com/AlonaBot/AlonaBot/contracts/i18n"
)

//go:embed en-US.json
var english []byte

//go:embed zh-CN.json
var chinese []byte

// 内置资源随程序发布；解析失败表示构建产物有误，保留全部问题并终止初始化。
var catalog = func() *i18n.Catalog {
	messages, problems := i18n.Parse(map[string][]byte{"en-US.json": english, "zh-CN.json": chinese})
	if len(problems) != 0 {
		panic(fmt.Sprintf("invalid Core message catalog: %v", problems))
	}
	return messages
}()

// Messages 返回 Core 内置的只读文案目录。
func Messages() *i18n.Catalog { return catalog }
