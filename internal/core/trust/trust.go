// Package trust 读取发行构建内置的官方发布者公钥。
//
// 公钥按 CONTRACTS §15.2 通过链接器注入；开发构建没有预先信任的发布者。
package trust

import (
	"crypto/ed25519"
	"encoding/base64"
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
