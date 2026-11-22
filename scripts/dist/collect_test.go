package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"debug/buildinfo"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestCollectPublishedComponent 经 HTTPS 与真实文件系统核对组件归档收录及坏包拒绝。
func TestCollectPublishedComponent(t *testing.T) {
	var archive []byte
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, ".tar.gz") {
			t.Errorf("unexpected metadata request: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(archive)
	}))
	t.Cleanup(server.Close)
	previous := http.DefaultClient
	http.DefaultClient = server.Client()
	t.Cleanup(func() { http.DefaultClient = previous })
	source := strings.TrimPrefix(server.URL, "https://")
	binary, err := os.ReadFile(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	info, err := buildinfo.ReadFile(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	var content bytes.Buffer
	gz := gzip.NewWriter(&content)
	writer := tar.NewWriter(gz)
	manifest := fmt.Sprintf("manifest_version: 1\nid: alonabot-agent-official\nname: Agent\nversion: 0.1.0\nrole: agent\nsource: %s\nos: %s\narch: %s\nprotocol_min_major: 1\nprotocol_max_major: 1\nrequired_capabilities: []\nentrypoint: component\nmigration_mode: none\nprepare_effects: none\n", source, runtime.GOOS, runtime.GOARCH)
	for name, raw := range map[string][]byte{"plugin.yaml": []byte(manifest), "component": binary, "build-info": []byte(info.String())} {
		if err := writer.WriteHeader(&tar.Header{Name: name, Mode: 0700, Size: int64(len(raw))}); err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write(raw); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	archive = content.Bytes()
	root, out := t.TempDir(), t.TempDir()
	component := component{"agent", "alonabot-agent-official"}
	target := runtime.GOOS + "/" + runtime.GOARCH
	if err := collectComponent(t.Context(), component, source, "v0.1.0", out, []string{target}, map[string]string{target: root}); err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("%s-0.1.0-%s-%s.tar.gz", component.id, runtime.GOOS, runtime.GOARCH)
	for _, dir := range []string{root, out} {
		directory := filepath.Join(dir, "packages", component.id)
		entries, err := os.ReadDir(directory)
		if err != nil || len(entries) != 1 {
			t.Fatalf("asset inventory: %v %v", entries, err)
		}
		raw, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil || !bytes.Equal(raw, archive) {
			t.Fatalf("collected archive: %v", err)
		}
	}
	archive = []byte("invalid archive")
	if err := collectComponent(t.Context(), component, source, "v0.1.0", t.TempDir(), []string{target}, map[string]string{target: t.TempDir()}); err == nil || !strings.Contains(err.Error(), "gzip") {
		t.Fatalf("invalid archive: %v", err)
	}
}
