// Package trust 定义官方组件的仓库来源。
package trust

import "strings"

// Source 返回官方组件独立仓库的规范来源；未知组件返回空字符串。
func Source(id string) string {
	switch id {
	case "alonabot-agent-official":
		return "github.com/AlonaBot/alonabot-agent"
	case "alonabot-webui-admin":
		return "github.com/AlonaBot/alonabot-webui"
	case "alonabot-adapter-telegram":
		return "github.com/AlonaBot/alonabot-adapter-telegram"
	case "alonabot-adapter-onebot11":
		return "github.com/AlonaBot/alonabot-adapter-onebot11"
	}
	return ""
}

// MatchesSource 同时绑定组件身份与确切仓库来源。
func MatchesSource(id, source string) bool {
	expected := Source(id)
	return expected != "" && strings.TrimPrefix(source, "https://") == expected
}
