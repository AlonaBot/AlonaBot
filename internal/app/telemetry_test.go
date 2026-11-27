//go:build unix

package app

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/telemetry"
)

// TestCoreTelemetryEndpoint 经真实 Core 与管理认证验证 /api/v1/core/telemetry 的窗口校验与进程采样结果。
func TestCoreTelemetryEndpoint(t *testing.T) {
	h := newWebUIHarness(t)
	h.start(nil)

	h.expect("GET", "/api/v1/core/telemetry", nil, false, http.StatusUnauthorized, "unauthorized")
	h.expect("GET", "/api/v1/core/telemetry?window=99h", nil, true, http.StatusBadRequest, "invalid window")

	for _, query := range []string{"", "?window=5m", "?window=15m", "?window=30m"} {
		_, body := h.expect("GET", "/api/v1/core/telemetry"+query, nil, true, http.StatusOK, `"host_normalized_percent"`)
		var report telemetry.Report
		if err := json.Unmarshal(body, &report); err != nil {
			t.Fatalf("unmarshal telemetry(%q): %v\nbody=%s", query, err, body)
		}
		if report.Core.PID != os.Getpid() {
			t.Fatalf("got core pid=%d, want %d", report.Core.PID, os.Getpid())
		}
		if _, err := time.Parse(time.RFC3339, report.Core.StartedAt); err != nil {
			t.Fatalf("invalid core.started_at=%q: %v", report.Core.StartedAt, err)
		}
		if report.Host.CPUCores <= 0 || report.Core.RSSBytes == 0 || report.Total.RSSBytes < report.Core.RSSBytes {
			t.Fatalf("unexpected telemetry metrics: host=%+v core=%+v total=%+v", report.Host, report.Core, report.Total)
		}
		if len(report.History) == 0 {
			t.Fatalf("expected at least one history sample for query %q", query)
		}
		if report.Components == nil {
			t.Fatalf("expected non-nil components array for query %q, body=%s", query, body)
		}
	}
}
