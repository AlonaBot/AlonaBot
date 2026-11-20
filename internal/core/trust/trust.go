// Package trust 读取发行构建内置的官方发布者公钥。
//
// 公钥按 CONTRACTS §15.2 通过链接器注入；开发构建没有预先信任的发布者。
package trust

import (
	"crypto/ed25519"
	"encoding/base64"
	"strings"
)

var officialKey string

// Official 返回发行构建内置的 Ed25519 公钥；未注入有效公钥时返回 false。
func Official() (ed25519.PublicKey, bool) {
	key, err := base64.StdEncoding.DecodeString(officialKey)
	if err != nil || len(key) != ed25519.PublicKeySize {
		return nil, false
	}
	return ed25519.PublicKey(key), true
}

// Source 返回官方组件独立仓库的规范来源；未知组件返回空字符串。
func Source(id string) string {
	switch id {
	case "alonabot-agent-official":
		return "github.com/AlonaBot/alonabot-agent"
	case "alonabot-webui-admin":
		return "github.com/AlonaBot/alonabot-webui"
	case "alonabot-plugin-help-official":
		return "github.com/AlonaBot/alonabot-plugin-help"
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
