//go:build unix

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestDiagnostics 验证运行中与离线诊断包的内容、组件状态和配置脱敏。
func TestDiagnostics(t *testing.T) {
	fixture := t.TempDir()
	source := filepath.Join(fixture, "main.go")
	if err := os.WriteFile(source, []byte(diagnosticFixtureSource), 0o600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(fixture, "diagnostic-fixture")
	if output, err := exec.Command("go", "build", "-o", binary, source).CombinedOutput(); err != nil {
		t.Fatalf("build diagnostic fixture: %v\n%s", err, output)
	}
	archive, err := writeTestPackage(fixture, "diagnostic-fixture", "plugin", "", "github.com/example/diagnostic-fixture", "1.0.0", binary, "", map[string][]byte{"plugin_config_schema.json": []byte(diagnosticFixtureSchema)})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0", "--log-format", "json")
	c.management(t, dir, "", "install", archive, "--yes")
	developmentWrite(t, filepath.Join(dir, "plugins", "diagnostic-fixture", "config.json"), `{"token":"diagnostic-fixture-secret"}`)
	c.management(t, dir, "", "enable", "diagnostic-fixture")
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
		for _, name := range []string{"manifest.json", "settings/core_config.json", "settings/plugins/diagnostic-fixture.json", "components.json", "environment.json"} {
			if _, found := files[name]; !found {
				t.Fatalf("missing %s", name)
			}
		}
		logs := false
		for name := range files {
			logs = logs || strings.HasPrefix(name, "logs/")
		}
		var config map[string]string
		if err := json.Unmarshal(files["settings/plugins/diagnostic-fixture.json"], &config); err != nil || config["token"] != "[REDACTED]" || !logs {
			t.Fatalf("config=%s logs=%v err=%v", files["settings/plugins/diagnostic-fixture.json"], logs, err)
		}
		var components []struct{ ID, State string }
		if err := json.Unmarshal(files["components.json"], &components); err != nil || len(components) != 1 {
			t.Fatalf("components=%s err=%v", files["components.json"], err)
		}
		want := "Stopped"
		if running {
			want = "Activated"
		}
		if components[0].ID != "diagnostic-fixture" || components[0].State != want {
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
}

const diagnosticFixtureSchema = `{"token":{"type":"string","secret":true}}`

const diagnosticFixtureSource = `package main
import (
 "log/slog"
 "os"
 "github.com/AlonaBot/AlonaBot/sdk"
 "github.com/AlonaBot/AlonaBot/sdk/genimpl"
)
func init(){genimpl.Register("main",genimpl.Resources{ConfigSchema:[]byte("{\"token\":{\"type\":\"string\",\"secret\":true}}")})}
func main(){
 p:=sdk.NewPlugin()
 if err:=sdk.Run(p);err!=nil{slog.Error("fixture failed","error",err);os.Exit(1)}
}
`
