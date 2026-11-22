package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlonaBot/AlonaBot/contracts/release"
)

// TestCollectPublishedComponent 经 HTTPS、真实签名与文件系统核对公开资产收录和篡改拒绝。
func TestCollectPublishedComponent(t *testing.T) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, exists := files[filepath.Base(r.URL.Path)]
		if !exists {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(raw)
	}))
	t.Cleanup(server.Close)
	previous := http.DefaultClient
	http.DefaultClient = server.Client()
	t.Cleanup(func() { http.DefaultClient = previous })
	source := strings.TrimPrefix(server.URL, "https://")
	name := "alonabot-agent-official-0.1.0-linux-amd64.tar.gz"
	archive := []byte("published component archive")
	digest := sha256.Sum256(archive)
	publication := release.Release{PublisherKeyID: release.KeyID(public), Source: source, PluginID: "alonabot-agent-official", Version: "0.1.0", Assets: []release.Asset{{OS: "linux", Arch: "amd64", URL: server.URL + "/releases/download/v0.1.0/" + name, Size: int64(len(archive)), SHA256: fmt.Sprintf("%x", digest)}}}
	files["release.json"], err = json.Marshal(publication)
	if err != nil {
		t.Fatal(err)
	}
	files["release.json.sig"] = []byte(base64.StdEncoding.EncodeToString(ed25519.Sign(private, files["release.json"])))
	files["release.pub"] = []byte(base64.StdEncoding.EncodeToString(public))
	files[name] = archive
	root := t.TempDir()
	out := t.TempDir()
	component := component{"agent", "alonabot-agent-official"}
	if err := collectComponent(t.Context(), component, source, "v0.1.0", out, []string{"linux/amd64"}, map[string]string{"linux/amd64": root}, public); err != nil {
		t.Fatal(err)
	}
	for name, expected := range files {
		for _, dir := range []string{root, out} {
			raw, err := os.ReadFile(filepath.Join(dir, "packages", component.id, name))
			if err != nil || !bytes.Equal(raw, expected) {
				t.Fatalf("collected %s: %v", name, err)
			}
		}
	}
	files[name] = []byte("tampered component archive")
	if err := collectComponent(t.Context(), component, source, "v0.1.0", t.TempDir(), []string{"linux/amd64"}, map[string]string{"linux/amd64": t.TempDir()}, public); err == nil || !strings.Contains(err.Error(), "integrity mismatch") {
		t.Fatalf("tampered asset: %v", err)
	}
}
