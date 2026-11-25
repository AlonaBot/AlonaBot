//go:build unix

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// telegramSource 复制 tests module 已选定的 Telegram 源码，供开发模式编译。
func telegramSource(t *testing.T) string {
	t.Helper()
	const module = "github.com/AlonaBot/alonabot-adapter-telegram"
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", module)
	output, err := cmd.Output()
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			t.Fatalf("locate %s: %v\n%s", module, err, exit.Stderr)
		}
		t.Fatalf("locate %s: %v", module, err)
	}
	directory := filepath.Join(t.TempDir(), "telegram")
	if err := os.CopyFS(directory, os.DirFS(strings.TrimSpace(string(output)))); err != nil {
		t.Fatalf("copy %s: %v", module, err)
	}
	return directory
}

// CLI 的公开退出码与临时终端实例身份由端到端断言直接核对。
const (
	exitFailure      = 1
	exitUsage        = 2
	consoleAdapterID = "console"
)

type sleepPlugin struct {
	ID      string          `json:"id"`
	Policy  json.RawMessage `json:"policy"`
	Runtime json.RawMessage `json:"runtime"`
}
