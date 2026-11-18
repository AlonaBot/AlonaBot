//go:build unix

package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/settings"
)

// TestLogManagement 通过真实 Core、CLI 和安装进程贯通级别写回、SDK 通知、SSE 与历史文件。
func TestLogManagement(t *testing.T) {
	fixture := t.TempDir()
	source := filepath.Join(fixture, "main.go")
	if err := os.WriteFile(source, []byte(logFixtureSource), 0o600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(fixture, "logger")
	if output, err := exec.Command("go", "build", "-o", binary, source).CombinedOutput(); err != nil {
		t.Fatalf("build log fixture: %v\n%s", err, output)
	}
	archive, err := writeTestPackage(fixture, "log-fixture", "plugin", "", "github.com/example/log-fixture", "1.0.0", binary, "", map[string][]byte{"config_schema.json": []byte(logFixtureSchema)})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0", "--log-format", "json")
	c.management(t, dir, "", "install", archive, "--yes")
	developmentWrite(t, filepath.Join(dir, "plugins", "log-fixture", "config.json"), `{"token":"diagnostic-fixture-secret"}`)
	c.management(t, dir, "", "enable", "log-fixture")
	probe := func(message string, wantDebug, wantInfo bool) {
		t.Helper()
		status, body := fixtureHTTP(t, c, dir, http.MethodGet, "/plugins/log-fixture/probe?message="+url.QueryEscape(message))
		var result struct{ Debug, Info bool }
		if status != http.StatusOK || json.Unmarshal(body, &result) != nil || result.Debug != wantDebug || result.Info != wantInfo {
			t.Fatalf("SDK thresholds: %d %s, want debug=%v info=%v", status, body, wantDebug, wantInfo)
		}
	}
	logs := func(args ...string) string { return c.management(t, dir, "", append([]string{"logs"}, args...)...) }
	waitLog := func(marker string, args ...string) string {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for {
			output := logs(args...)
			if strings.Contains(output, marker) {
				return output
			}
			if time.Now().After(deadline) {
				t.Fatalf("missing log %q: %s\n%s", marker, output, c.log())
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	probe("default-info", false, true)
	c.management(t, dir, "", "log-level", "log-fixture", "debug")
	probe("dynamic-debug", true, true)
	if output := waitLog("dynamic-debug-debug", "--source", "log-fixture", "--level", "debug"); !strings.Contains(output, "[DBUG]") {
		t.Fatal(output)
	}
	c.management(t, dir, "", "log-level", "log-fixture", "off")
	probe("suppressed-off", false, false)
	c.management(t, dir, "", "restart", "log-fixture")
	probe("suppressed-after-restart", false, false)
	if output := logs("--source", "log-fixture"); strings.Contains(output, "suppressed") {
		t.Fatal(output)
	}
	if output := logs("--source", "core"); !strings.Contains(output, "plugin activated") || !strings.Contains(output, "log-fixture") {
		t.Fatal(output)
	}
	c.management(t, dir, "", "log-level", "log-fixture", "inherit")
	c.management(t, dir, "", "log-level", "global", "warn")
	probe("inherit-warn", false, false)
	waitLog("inherit-warn-error", "--source", "log-fixture", "--level", "error")
	c.management(t, dir, "", "log-level", "global", "debug")
	probe("inherit-debug", true, true)
	waitLog("inherit-debug-debug", "--source", "log-fixture")
	// follow 直接取得快照，再观察一次请求触发的新日志。
	token, err := os.ReadFile(filepath.Join(dir, "admin_token"))
	if err != nil {
		t.Fatal(err)
	}
	follow := exec.Command(filepath.Join(binDir, "alonabot"), "logs", "--follow", "--source", "log-fixture", "--admin-url", "http://"+c.address())
	follow.Env = append(os.Environ(), "ALONABOT_ADMIN_TOKEN="+strings.TrimSpace(string(token)))
	stdout, err := follow.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var diagnostics strings.Builder
	follow.Stderr = &diagnostics
	if err := follow.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = follow.Process.Kill() })
	lines := make(chan string, 128)
	go func() {
		reader := bufio.NewReader(stdout)
		for {
			line, err := reader.ReadString('\n')
			if line != "" {
				lines <- line
			}
			if err != nil {
				close(lines)
				return
			}
		}
	}()
	waitFollow := func(marker string) {
		t.Helper()
		for {
			select {
			case line, ok := <-lines:
				if !ok {
					t.Fatalf("follow closed before %s", marker)
				}
				if strings.Contains(line, marker) {
					return
				}
			case <-time.After(5 * time.Second):
				t.Fatalf("follow missed %s", marker)
			}
		}
	}
	waitFollow("inherit-debug-debug")
	probe("following-new", true, true)
	waitFollow("following-new-debug")
	if err := follow.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatal(err)
	}
	if err := follow.Wait(); err != nil {
		t.Fatalf("follow cancellation: %v: %s", err, diagnostics.String())
	}
	for _, endpoint := range []string{"recent", "stream"} {
		response, err := http.Get("http://" + c.address() + "/api/v1/logs/" + endpoint)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusUnauthorized {
			t.Fatalf("unauthenticated %s = %d", endpoint, response.StatusCode)
		}
	}
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	history := filepath.Join(dir, "logs", "alonabot-"+yesterday+".log")
	if err := os.WriteFile(history, []byte("[12:34:56.789] [Plug:log-fixture] [INFO] [fixture/main.go:7]: previous day message trace=12345678\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if output := logs("--date", yesterday, "--source", "log-fixture", "--trace", "12345678abcdef"); !strings.Contains(output, "previous day message") {
		t.Fatal(output)
	}
	readSettings := func() {
		t.Helper()
		raw, err := os.ReadFile(filepath.Join(dir, settings.FileName))
		if err != nil {
			t.Fatal(err)
		}
		var document struct {
			Log struct {
				Level   string            `json:"level"`
				Sources map[string]string `json:"sources"`
			} `json:"log"`
		}
		if err := json.Unmarshal(raw, &document); err != nil {
			t.Fatal(err)
		}
		if document.Log.Level != "debug" || document.Log.Sources["log-fixture"] != "inherit" {
			t.Fatalf("persisted levels: %s", raw)
		}
	}
	readSettings()
	stopTestCore(t, c)
	c = startEmptyCore(t, dir, "--listen", "127.0.0.1:0", "--log-format", "json")
	c.waitFor(`"msg":"plugin activated"[^\n]*"plugin_id":"log-fixture"`)
	if value := strings.TrimSpace(c.management(t, dir, "", "log-level", "global")); value != "debug" {
		t.Fatal(value)
	}
	probe("persisted-debug", true, true)
	waitLog("persisted-debug-debug", "--source", "log-fixture")
	stopTestCore(t, c)
	c = startEmptyCore(t, dir, "--listen", "127.0.0.1:0", "--log-level", "info", "--log-format", "json")
	c.waitFor(`"msg":"plugin activated"[^\n]*"plugin_id":"log-fixture"`)
	if value := strings.TrimSpace(c.management(t, dir, "", "log-level", "global")); value != "info" {
		t.Fatal(value)
	}
	probe("startup-override", false, true)
	readSettings()
	stopTestCore(t, c)
	c = startEmptyCore(t, dir, "--listen", "127.0.0.1:0", "--log-format", "json")
	c.waitFor(`"msg":"plugin activated"[^\n]*"plugin_id":"log-fixture"`)
	probe("override-ended", true, true)
	waitLog("override-ended-debug", "--source", "log-fixture")
	stopTestCore(t, c)
	t.Run("diagnostics", func(t *testing.T) {
		c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0", "--log-format", "json")
		c.waitFor(`"msg":"plugin activated"[^\n]*"plugin_id":"log-fixture"`)
		checkArchive := func(raw []byte, running bool) {
			t.Helper()
			archive, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
			if err != nil {
				t.Fatal(err)
			}
			files := make(map[string][]byte)
			for _, entry := range archive.File {
				reader, err := entry.Open()
				if err != nil {
					t.Fatal(err)
				}
				content, readErr := io.ReadAll(reader)
				closeErr := reader.Close()
				if readErr != nil || closeErr != nil {
					t.Fatalf("read diagnostic entry %s: %v %v", entry.Name, readErr, closeErr)
				}
				files[entry.Name] = content
				if bytes.Contains(content, []byte("diagnostic-fixture-secret")) {
					t.Fatalf("secret leaked in %s", entry.Name)
				}
			}
			for _, name := range []string{"manifest.json", "settings/core_config.json", "settings/plugins/log-fixture.json", "components.json", "environment.json"} {
				if _, found := files[name]; !found {
					t.Fatalf("missing %s", name)
				}
			}
			logs := false
			for name := range files {
				logs = logs || strings.HasPrefix(name, "logs/")
			}
			var config map[string]string
			if err := json.Unmarshal(files["settings/plugins/log-fixture.json"], &config); err != nil || config["token"] != "[REDACTED]" || !logs {
				t.Fatalf("config=%s logs=%v err=%v", files["settings/plugins/log-fixture.json"], logs, err)
			}
			var components []struct{ ID, State string }
			if err := json.Unmarshal(files["components.json"], &components); err != nil || len(components) != 1 {
				t.Fatalf("components=%s err=%v", files["components.json"], err)
			}
			want := "Stopped"
			if running {
				want = "Activated"
			}
			if components[0].ID != "log-fixture" || components[0].State != want {
				t.Fatalf("components=%+v, want state %s", components, want)
			}
		}
		request, err := http.NewRequest(http.MethodPost, "http://"+c.address()+"/api/v1/diagnostics", strings.NewReader(`{}`))
		if err != nil {
			t.Fatal(err)
		}
		token, err := os.ReadFile(filepath.Join(dir, "admin_token"))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))
		request.Header.Set("Content-Type", "application/json")
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		raw, readErr := io.ReadAll(response.Body)
		closeErr := response.Body.Close()
		if readErr != nil || closeErr != nil || response.StatusCode != http.StatusOK || response.Header.Get("Content-Type") != "application/zip" {
			t.Fatalf("diagnostics API: %d %s %v %v", response.StatusCode, raw, readErr, closeErr)
		}
		checkArchive(raw, true)
		output := filepath.Join(t.TempDir(), "diagnostics.zip")
		c.management(t, dir, "", "diagnose", "-o", output)
		checkArchive(developmentRead(t, output), true)
		address := "http://" + c.address()
		stopTestCore(t, c)
		command := exec.Command(filepath.Join(binDir, "alonabot"), "diagnose", "--data-dir", dir, "--admin-url", address, "-o", output)
		if raw, err := command.CombinedOutput(); err != nil {
			t.Fatalf("offline diagnose: %v\n%s", err, raw)
		}
		checkArchive(developmentRead(t, output), false)
	})
}

const logFixtureSchema = `{"token":{"type":"string","secret":true}}`

const logFixtureSource = `package main
import (
 "context"
 "encoding/json"
 "log/slog"
 "net/http"
 "os"
 "github.com/AlonaBot/AlonaBot/sdk"
 "github.com/AlonaBot/AlonaBot/sdk/genimpl"
)
func init(){genimpl.Register("main",genimpl.Resources{ConfigSchema:[]byte("{\"token\":{\"type\":\"string\",\"secret\":true}}")})}
func main(){
 p:=sdk.NewPlugin()
 p.Route(http.MethodGet,"/probe").HandleFunc(func(w http.ResponseWriter,r *http.Request){
  marker:=r.URL.Query().Get("message")
  slog.Debug(marker+"-debug","trace_id","123456789abcdef0123456789abcdef0")
  slog.Info(marker+"-info")
  slog.Error(marker+"-error")
  if err:=json.NewEncoder(w).Encode(struct{Debug,Info bool}{slog.Default().Enabled(context.Background(),slog.LevelDebug),slog.Default().Enabled(context.Background(),slog.LevelInfo)});err!=nil{slog.Error("write probe response","error",err)}
 })
 if err:=sdk.Run(p);err!=nil{slog.Error("fixture failed","error",err);os.Exit(1)}
}
`
