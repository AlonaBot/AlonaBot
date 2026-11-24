package packages

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"debug/buildinfo"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/internal/tune"
)

const (
	testSource   = "github.com/example/demo"
	testPluginID = "alonabot-plugin-demo"
	testVersion  = "1.2.3"
	testOS       = runtime.GOOS
	testArch     = runtime.GOARCH
)

// testEntry 是测试归档中的一个条目。
type testEntry struct {
	name     string
	typeflag byte
	mode     int64
	body     string
}

func manifestYAML(id, version, source, osName, arch, entrypoint string) string {
	return "manifest_version: 1\nname: Demo\nid: " + id + "\nversion: " + version +
		"\nprotocol_min_major: 1\nprotocol_max_major: 1\nrequired_capabilities: []\nos: " + osName +
		"\narch: " + arch + "\nentrypoint: " + entrypoint + "\nsource: " + source +
		"\nmigration_mode: none\nprepare_effects: none\n"
}

// validEntries 返回能通过预检的最小包内容，extra 追加在后面。
func validEntries(extra ...testEntry) []testEntry {
	executable, err := os.ReadFile(os.Args[0])
	if err != nil {
		panic(err)
	}
	info, err := buildinfo.ReadFile(os.Args[0])
	if err != nil {
		panic(err)
	}
	return append([]testEntry{
		{name: "plugin.yaml", typeflag: tar.TypeReg, mode: 0o644, body: manifestYAML(testPluginID, testVersion, testSource, testOS, testArch, "bin/plugin")},
		{name: "bin/", typeflag: tar.TypeDir, mode: 0o755},
		{name: "bin/plugin", typeflag: tar.TypeReg, mode: 0o755, body: string(executable)},
		{name: "build-info", typeflag: tar.TypeReg, mode: 0o644, body: info.String()},
	}, extra...)
}

func buildArchive(t *testing.T, entries []testEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		h := &tar.Header{Name: e.name, Mode: e.mode, Typeflag: e.typeflag, Size: int64(len(e.body))}
		if e.typeflag != tar.TypeReg {
			h.Size = 0
		}
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if e.typeflag == tar.TypeReg {
			if _, err := tw.Write([]byte(e.body)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// pkg 是一份在临时目录中生成的组件归档。
type pkg struct {
	root  string
	input Input
}

// newPackage 写出组件归档及期望身份。
func newPackage(t *testing.T, entries []testEntry) pkg {
	t.Helper()
	archive := buildArchive(t, entries)
	root := t.TempDir()
	archivePath := filepath.Join(root, "demo.tar.gz")
	if err := os.WriteFile(archivePath, archive, 0o600); err != nil {
		t.Fatal(err)
	}
	return pkg{root: root, input: Input{
		ID: testPluginID, Version: testVersion, Source: testSource, OS: testOS, Arch: testArch, ArchivePath: archivePath,
	}}
}

// TestPreflightAndMoveTo 用真实归档与文件系统驱动合法包的完整流程（V-27、V-41 的预检分项）。
func TestPreflightAndMoveTo(t *testing.T) {
	entries := validEntries()
	p := newPackage(t, entries)
	staging, packages := filepath.Join(p.root, "staging"), filepath.Join(p.root, "packages")

	c, err := Preflight(p.input, staging, NewLimits(tune.Defaults()))
	if err != nil {
		t.Fatal(err)
	}
	final, err := c.MoveTo(packages)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(packages, testPluginID, testVersion+"-"+hex.EncodeToString(c.Digest)[:12]); final != want {
		t.Fatalf("MoveTo() = %s, want %s", final, want)
	}
	if got, err := os.ReadFile(filepath.Join(final, "bin", "plugin")); err != nil || string(got) != entries[2].body {
		t.Fatalf("entrypoint differs from the test binary: %v", err)
	}
	assertNoStaging(t, staging)
}

// assertNoStaging 确认 staging 下没有残留的暂存目录；staging 本身可以不存在。
func assertNoStaging(t *testing.T, staging string) {
	t.Helper()
	left, err := os.ReadDir(staging)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Fatalf("staging has leftovers: %v", left)
	}
}

// TestProjectManifestPreflight 通过真实归档验证作者信息合成。
func TestProjectManifestPreflight(t *testing.T) {
	project := "id: " + testPluginID + "\nname: Demo\ndescription: Example component\nauthor: Example author\nversion: " + testVersion + "\nsource: " + testSource + "\n"
	parsed, err := manifest.ParseProject([]byte(project))
	if err != nil {
		t.Fatal(err)
	}
	one := uint32(1)
	capabilities := []manifest.Capability{}
	data, err := manifest.Generate(parsed.Manifest, manifest.BuildInfo{
		ManifestVersion: 1, ProtocolMinMajor: &one, ProtocolMaxMajor: &one,
		RequiredCapabilities: &capabilities, OS: testOS, Arch: testArch,
		Entrypoint: "bin/plugin", MigrationMode: manifest.MigrationNone, PrepareEffects: manifest.EffectsNone,
	})
	if err != nil {
		t.Fatal(err)
	}
	entries := validEntries()
	entries[0].body = string(data)
	p := newPackage(t, entries)
	candidate, err := Preflight(p.input, filepath.Join(p.root, "staging"), NewLimits(tune.Defaults()))
	if err != nil {
		t.Fatal(err)
	}
	m := candidate.Manifest
	if m.Name != parsed.Manifest.Name || m.Description != parsed.Manifest.Description || m.Author != parsed.Manifest.Author {
		t.Errorf("candidate author fields = %q, %q, %q", m.Name, m.Description, m.Author)
	}
	if err := candidate.Discard(); err != nil {
		t.Fatal(err)
	}
}
