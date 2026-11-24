//go:build unix

package agent_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// checkAgentSummaries 检查真实模型及工具循环的日志保留追踪，并省略正文、参数与凭据。
func checkAgentSummaries(t *testing.T, dir string) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(dir, "logs", "*.log"))
	if err != nil {
		t.Fatal(err)
	}
	var model, tool bool
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
			var record struct {
				Level, Message string
				Trace          string `json:"trace_id"`
			}
			if err := json.Unmarshal([]byte(line), &record); err != nil {
				t.Fatal(err)
			}
			if record.Level != "INFO" {
				continue
			}
			for _, secret := range []string{"actual parameters", "/ask tools", "test-key"} {
				if strings.Contains(line, secret) {
					t.Fatalf("INFO contains request data or credentials: %s", line)
				}
			}
			if strings.HasPrefix(record.Message, "Agent model ") || strings.HasPrefix(record.Message, "Agent is calling tool ") {
				if len(record.Trace) != 32 {
					t.Fatalf("summary missing complete Trace: %s", line)
				}
				model = model || strings.HasPrefix(record.Message, "Agent model ")
				tool = tool || strings.HasPrefix(record.Message, "Agent is calling tool ")
			}
		}
	}
	if !model || !tool {
		t.Fatalf("missing Agent summaries: model=%t tool=%t", model, tool)
	}
}
