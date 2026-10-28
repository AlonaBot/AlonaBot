//go:build unix

package packages

import (
	"archive/tar"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlonaBot/AlonaBot/internal/tune"
)

// TestPreflightFileModes 确认解压后的权限：带执行位的文件为 0700，其余为 0600，setuid 被去掉。
func TestPreflightFileModes(t *testing.T) {
	p := newPackage(t, validEntries(
		testEntry{name: "data.txt", typeflag: tar.TypeReg, mode: 0o666, body: "data"},
		testEntry{name: "setuid", typeflag: tar.TypeReg, mode: 0o4755, body: "x"},
	), nil)
	c, err := Preflight(p.input, filepath.Join(p.root, "staging"), NewLimits(tune.Defaults()))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Discard()
	for name, want := range map[string]os.FileMode{
		"plugin.yaml": 0o600,
		"data.txt":    0o600,
		"bin/plugin":  0o700,
		"setuid":      0o700,
		"bin":         os.ModeDir | 0o700,
	} {
		info, err := os.Stat(filepath.Join(c.Dir, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode() & (os.ModeDir | os.ModePerm | os.ModeSetuid | os.ModeSetgid | os.ModeSticky); got != want {
			t.Errorf("%s mode = %v, want %v", name, got, want)
		}
	}
}
