package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestVersion 通过真实构建与 CLI 进程核对版本输出及提前退出。
func TestVersion(t *testing.T) {
	for _, test := range []struct {
		name, flags, want string
	}{
		{"development", "", "alonabot version dev\nhttps://github.com/AlonaBot/AlonaBot\n"},
		{"release", "-X main.version=0.2.0 -X main.buildDate=2026-11-22", "alonabot version 0.2.0 (2026-11-22)\nhttps://github.com/AlonaBot/AlonaBot/releases/tag/v0.2.0\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			binary := filepath.Join(t.TempDir(), "alonabot")
			if out, err := exec.Command("go", "build", "-ldflags", test.flags, "-o", binary, ".").CombinedOutput(); err != nil {
				t.Fatalf("build failed: %v\n%s", err, out)
			}
			args := []string{"--version"}
			cmd := exec.Command(binary, args...)
			cmd.Dir = t.TempDir()
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			out, err := cmd.Output()
			if err != nil || string(out) != test.want || stderr.Len() != 0 {
				t.Fatalf("%v: error=%v, stdout=%q, stderr=%q; want stdout=%q", args, err, out, stderr.String(), test.want)
			}
			entries, err := os.ReadDir(cmd.Dir)
			if err != nil || len(entries) != 0 {
				t.Fatalf("version working directory: entries=%v, error=%v", entries, err)
			}
			out, err = exec.Command(binary, "--help").CombinedOutput()
			if err != nil || !strings.Contains(string(out), "--version") {
				t.Fatalf("help: error=%v, output=%s", err, out)
			}
		})
	}
}
