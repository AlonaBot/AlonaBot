package release

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 仅用于测试的固定密钥；不用于任何真实发布。
func testKey() (ed25519.PublicKey, ed25519.PrivateKey) {
	priv := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{7}, ed25519.SeedSize))
	return priv.Public().(ed25519.PublicKey), priv
}

const placeholderKeyID = "ed25519:0000000000000000000000000000000000000000000000000000000000000000"

// signedFixture 返回带测试公钥标识的 release.json 原始字节及其签名文件内容。
func signedFixture(t *testing.T, edit func(string) string) (raw, sigFile []byte) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "testdata", "release", "release.json"))
	if err != nil {
		t.Fatal(err)
	}
	pub, priv := testKey()
	s := strings.Replace(string(b), placeholderKeyID, KeyID(pub), 1)
	if edit != nil {
		s = edit(s)
	}
	raw = []byte(s)
	return raw, []byte(base64.StdEncoding.EncodeToString(ed25519.Sign(priv, raw)) + "\n")
}

func TestVerifyOK(t *testing.T) {
	pub, _ := testKey()
	raw, sig := signedFixture(t, nil)
	r, err := Verify(raw, sig, pub)
	if err != nil {
		t.Fatalf("合法签名应通过: %v", err)
	}
	if a, ok := r.Asset("darwin", "arm64"); !ok || a.Size != 1032000 {
		t.Fatalf("资产解析错误: %+v", a)
	}
}

func TestVerifyRejectsBadSignature(t *testing.T) {
	pub, _ := testKey()
	raw, sig := signedFixture(t, nil)

	tampered := bytes.Replace(raw, []byte(`"size": 1048576`), []byte(`"size": 1048577`), 1)
	if _, err := Verify(tampered, sig, pub); err == nil || !strings.Contains(err.Error(), "签名校验失败") {
		t.Fatalf("内容被改动后应验签失败，实际 %v", err)
	}

	otherPub := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{8}, ed25519.SeedSize)).Public().(ed25519.PublicKey)
	if _, err := Verify(raw, sig, otherPub); err == nil {
		t.Fatal("非信任公钥应验签失败")
	}
	if _, err := Verify(raw, []byte("not base64!"), pub); err == nil {
		t.Fatal("损坏的签名文件应失败")
	}
}

// 证明验签对象是文件的确切字节：语义相同但重新序列化后的 JSON 不能通过原签名。
func TestVerifyUsesExactBytes(t *testing.T) {
	pub, _ := testKey()
	raw, sig := signedFixture(t, nil)
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	reserialized, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(reserialized, raw) {
		t.Fatal("测试前提不成立：重新序列化后的字节应与原文件不同")
	}
	if _, err := Verify(reserialized, sig, pub); err == nil {
		t.Fatal("重新序列化的 JSON 不应通过原签名")
	}
}

func TestVerifyRejectsContent(t *testing.T) {
	pub, _ := testKey()
	cases := []struct{ name, old, new, wantErr string }{
		{"未知字段", `"source"`, `"extra": 1, "source"`, "unknown field"},
		{"非 https", "https://github.com/AlonaBot/rss/releases/download/v1.2.0/alonabot_plugin_rss_feed-1.2.0-linux", "http://github.com/AlonaBot/rss/releases/download/v1.2.0/alonabot_plugin_rss_feed-1.2.0-linux", "https"},
		{"大写摘要", "9f86d081884c7d659a2feaa0c55ad015", "9F86D081884C7D659A2FEAA0C55AD015", "sha256"},
		{"不支持的平台", `"os": "linux"`, `"os": "plan9"`, "不支持的平台"},
		{"重复平台", `"os": "darwin", "arch": "arm64"`, `"os": "linux", "arch": "amd64"`, "重复"},
		{"非法插件 ID", `"plugin_id": "alonabot_plugin_rss_feed"`, `"plugin_id": "../x"`, "安全 slug"},
		{"非法版本", `"version": "1.2.0"`, `"version": "latest"`, "SemVer"},
		{"公钥标识不符", `"publisher_key_id": "ed25519:`, `"publisher_key_id": "ed25519:ff`, "不符"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			raw, sig := signedFixture(t, func(s string) string {
				out := strings.Replace(s, c.old, c.new, 1)
				if out == s {
					t.Fatalf("用例没有改动 fixture（old=%q）", c.old)
				}
				return out
			})
			// 签名对改动后的字节有效，失败必须来自内容校验。
			if _, err := Verify(raw, sig, pub); err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("期望包含 %q 的错误，实际 %v", c.wantErr, err)
			}
		})
	}
}
