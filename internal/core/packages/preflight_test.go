package packages

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

const (
	testSource   = "github.com/example/demo"
	testPluginID = "alonabot_plugin_demo"
	testVersion  = "1.2.3"
	testOS       = "darwin"
	testArch     = "arm64"
)

// testEntry 是测试归档中的一个条目。
type testEntry struct {
	name     string
	typeflag byte
	mode     int64
	body     string
}

func manifestYAML(id, version, source, osName, arch, entrypoint string) string {
	return "manifest_version: 1\nid: " + id + "\nversion: " + version +
		"\nprotocol_min_major: 1\nprotocol_max_major: 1\nrequired_capabilities: []\nos: " + osName +
		"\narch: " + arch + "\nentrypoint: " + entrypoint + "\nconfig_version: 1\nsource: " + source +
		"\nmigration_mode: none\nprepare_effects: none\n"
}

// validEntries 返回能通过预检的最小包内容，extra 追加在后面。
func validEntries(extra ...testEntry) []testEntry {
	return append([]testEntry{
		{name: "plugin.yaml", typeflag: tar.TypeReg, mode: 0o644, body: manifestYAML(testPluginID, testVersion, testSource, testOS, testArch, "bin/plugin")},
		{name: "bin/", typeflag: tar.TypeDir, mode: 0o755},
		{name: "bin/plugin", typeflag: tar.TypeReg, mode: 0o755, body: "executable"},
	}, extra...)
}

func buildArchive(t *testing.T, entries []testEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		h := &tar.Header{Name: e.name, Mode: e.mode, Typeflag: e.typeflag, Size: int64(len(e.body))}
		switch e.typeflag {
		case tar.TypeSymlink, tar.TypeLink:
			h.Linkname = "plugin.yaml"
		}
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

// pkg 是一份在临时目录中生成的归档与签名清单。
type pkg struct {
	root  string
	input Input
}

// newPackage 写出归档，并签名 edit 修改后的清单。
func newPackage(t *testing.T, entries []testEntry, edit func(*release.Release)) pkg {
	t.Helper()
	archive := buildArchive(t, entries)
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(archive)
	r := release.Release{
		PublisherKeyID: release.KeyID(pub), Source: testSource, PluginID: testPluginID, Version: testVersion,
		Assets: []release.Asset{{OS: testOS, Arch: testArch, URL: "https://example.com/demo.tar.gz", Size: int64(len(archive)), SHA256: hex.EncodeToString(sum[:])}},
	}
	if edit != nil {
		edit(&r)
	}
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	archivePath := filepath.Join(root, "demo.tar.gz")
	if err := os.WriteFile(archivePath, archive, 0o600); err != nil {
		t.Fatal(err)
	}
	return pkg{root: root, input: Input{
		ReleaseJSON: raw,
		Signature:   []byte(base64.StdEncoding.EncodeToString(ed25519.Sign(priv, raw))),
		PublicKey:   pub, Source: testSource, OS: testOS, Arch: testArch, ArchivePath: archivePath,
	}}
}

// TestPreflightAndMoveTo 用真实归档、签名与文件系统驱动合法包的完整流程（V-27、V-41 的预检分项）。
func TestPreflightAndMoveTo(t *testing.T) {
	p := newPackage(t, validEntries(), nil)
	staging, packages := filepath.Join(p.root, "staging"), filepath.Join(p.root, "packages")

	c, err := Preflight(p.input, staging, DefaultLimits)
	if err != nil {
		t.Fatal(err)
	}
	final, err := c.MoveTo(packages)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(packages, testPluginID, testVersion+"-"+c.Asset.SHA256[:12]); final != want {
		t.Fatalf("MoveTo() = %s, want %s", final, want)
	}
	if got, err := os.ReadFile(filepath.Join(final, "bin", "plugin")); err != nil || string(got) != "executable" {
		t.Fatalf("entrypoint = %q, %v", got, err)
	}
	assertNoStaging(t, staging)

	// 同一包再次预检后移动到同一目标，不能覆盖已有包目录。
	again, err := Preflight(p.input, staging, DefaultLimits)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := again.MoveTo(packages); errorCode(err) != pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS {
		t.Fatalf("second MoveTo() error = %v, want ALREADY_EXISTS", err)
	}
	if err := again.Discard(); err != nil {
		t.Fatal(err)
	}
	assertNoStaging(t, staging)
}

// TestPreflightRejects 覆盖每一种拒绝，并确认暂存目录没有残留（V-27、V-28、V-41 的预检分项）。
func TestPreflightRejects(t *testing.T) {
	const (
		invalid   = pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		exhausted = pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED
	)
	withManifest := func(id, version, source, osName, arch, entrypoint string) []testEntry {
		entries := validEntries()
		entries[0].body = manifestYAML(id, version, source, osName, arch, entrypoint)
		return entries
	}
	file := func(name string) testEntry {
		return testEntry{name: name, typeflag: tar.TypeReg, mode: 0o600, body: "x"}
	}
	special := func(typeflag byte) testEntry { return testEntry{name: "special", typeflag: typeflag, mode: 0o600} }

	for _, tc := range []struct {
		name    string
		entries []testEntry
		release func(*release.Release)
		input   func(*Input)
		limits  Limits
		want    pb.ErrorCode
	}{
		{name: "bad signature", input: func(in *Input) { in.ReleaseJSON = append(in.ReleaseJSON, ' ') }, want: pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED},
		{name: "source mismatch", input: func(in *Input) { in.Source = "github.com/other/demo" }, want: pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED},
		{name: "no asset for platform", input: func(in *Input) { in.Arch = "amd64" }, want: pb.ErrorCode_ERROR_CODE_UNSUPPORTED},
		{name: "size mismatch", release: func(r *release.Release) { r.Assets[0].Size++ }, want: invalid},
		{name: "digest mismatch", release: func(r *release.Release) { r.Assets[0].SHA256 = strings.Repeat("0", 64) }, want: invalid},
		{name: "archive over limit", limits: Limits{MaxArchiveBytes: 1, MaxExpandedBytes: 1 << 20, MaxEntries: 100}, want: exhausted},
		{name: "manifest id", entries: withManifest("alonabot_plugin_other", testVersion, testSource, testOS, testArch, "bin/plugin"), want: invalid},
		{name: "manifest version", entries: withManifest(testPluginID, "1.2.4", testSource, testOS, testArch, "bin/plugin"), want: invalid},
		{name: "manifest source", entries: withManifest(testPluginID, testVersion, "github.com/other/demo", testOS, testArch, "bin/plugin"), want: invalid},
		{name: "manifest platform", entries: withManifest(testPluginID, testVersion, testSource, "linux", testArch, "bin/plugin"), want: invalid},
		{name: "manifest entrypoint missing", entries: withManifest(testPluginID, testVersion, testSource, testOS, testArch, "bin/missing"), want: invalid},
		{name: "parent traversal", entries: validEntries(file("../x")), want: invalid},
		{name: "absolute path", entries: validEntries(file("/abs")), want: invalid},
		{name: "nested traversal", entries: validEntries(file("a/../../x")), want: invalid},
		{name: "backslash", entries: validEntries(file(`a\b`)), want: invalid},
		{name: "drive letter", entries: validEntries(file("C:x")), want: invalid},
		{name: "symlink", entries: validEntries(special(tar.TypeSymlink)), want: invalid},
		{name: "hardlink", entries: validEntries(special(tar.TypeLink)), want: invalid},
		{name: "char device", entries: validEntries(special(tar.TypeChar)), want: invalid},
		{name: "block device", entries: validEntries(special(tar.TypeBlock)), want: invalid},
		{name: "fifo", entries: validEntries(special(tar.TypeFifo)), want: invalid},
		{name: "case collision", entries: validEntries(file("A.txt"), file("a.txt")), want: invalid},
		{name: "duplicate path", entries: validEntries(file("a.txt"), file("a.txt")), want: invalid},
		{name: "too many entries", limits: Limits{MaxArchiveBytes: 1 << 20, MaxExpandedBytes: 1 << 20, MaxEntries: 2}, want: exhausted},
		{name: "expanded size over limit", limits: Limits{MaxArchiveBytes: 1 << 20, MaxExpandedBytes: 16, MaxEntries: 100}, want: exhausted},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entries := tc.entries
			if entries == nil {
				entries = validEntries()
			}
			p := newPackage(t, entries, tc.release)
			if tc.input != nil {
				tc.input(&p.input)
			}
			limits := tc.limits
			if limits == (Limits{}) {
				limits = DefaultLimits
			}
			staging := filepath.Join(p.root, "staging")

			c, err := Preflight(p.input, staging, limits)
			if errorCode(err) != tc.want {
				t.Fatalf("Preflight() = %+v, %v; want %v", c, err, tc.want)
			}
			assertNoStaging(t, staging)
		})
	}
}

func errorCode(err error) pb.ErrorCode {
	var ve *validate.Error
	if errors.As(err, &ve) {
		return ve.Code
	}
	return pb.ErrorCode_ERROR_CODE_UNSPECIFIED
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
