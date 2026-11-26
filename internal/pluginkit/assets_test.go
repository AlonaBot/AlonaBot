package pluginkit

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlonaBot/AlonaBot/internal/generate"
)

// TestPackAssets 验证真实构建归档保留随包文件，资源登记与二进制均不嵌入其内容。
func TestPackAssets(t *testing.T) {
	// 1. 独立工程沿用仓库依赖与本地 SDK，随包文件具有可辨识的原始内容。
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	module, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	module = bytes.Replace(module, []byte("module github.com/AlonaBot/AlonaBot"), []byte("module example.invalid/assetsprobe"), 1)
	module = append(module, []byte("\nrequire github.com/AlonaBot/AlonaBot v0.0.0\nreplace github.com/AlonaBot/AlonaBot => "+root+"\n")...)
	sum, err := os.ReadFile(filepath.Join(root, "go.sum"))
	if err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	const content = "archive-only-asset-0123456789-keep-this-content-external\x00\xff"
	for name, raw := range map[string][]byte{
		"go.mod": module, "go.sum": sum,
		"plugin.yaml":            []byte("id: assetsprobe\nname: Assets probe\nversion: 0.1.0\nsource: github.com/example/assetsprobe\nrole: plugin\n"),
		"main.go":                []byte("package assetsprobe\nimport \"github.com/AlonaBot/AlonaBot/sdk\"\nfunc New() *sdk.Plugin { return sdk.NewPlugin() }\n"),
		"assets/help/nested.bin": []byte(content),
	} {
		path := filepath.Join(project, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), privateMode); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, raw, fileMode); err != nil {
			t.Fatal(err)
		}
	}
	// 2. 生成与打包共用资源快照，实际归档中的可执行文件保持资源外置。
	snapshot, err := scanProject(t.Context(), generate.Options{Dir: project})
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(snapshot.generated.Source, []byte(content)) || strings.Contains(string(snapshot.generated.Source), "nested.bin") {
		t.Fatal("asset embedded in generated registration")
	}
	archive, err := Pack(t.Context(), generate.Options{Dir: project}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	compressed, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer compressed.Close()
	reader := tar.NewReader(compressed)
	found, executable := false, false
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		raw, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		if header.Name == "assets/help/nested.bin" {
			found = true
			if string(raw) != content {
				t.Fatalf("asset content changed: %q", raw)
			}
		}
		if header.Name == "component" {
			executable = true
			if bytes.Contains(raw, []byte(content)) {
				t.Fatal("asset embedded in executable")
			}
		}
	}
	if !found || !executable {
		t.Fatalf("archive asset=%v executable=%v", found, executable)
	}
}
