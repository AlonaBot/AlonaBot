//go:build unix

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/generate"
	"github.com/AlonaBot/AlonaBot/internal/pluginkit"
)

// TestSleepPolicyManagement 通过真实 CLI、认证 API、插件进程与 SQLite 验证策略即时生效及原代次恢复。
func TestSleepPolicyManagement(t *testing.T) {
	const id = "sleepprobe"
	archive := sleepTestPackage(t, id, sleepTestSource)
	dir := t.TempDir()
	t.Setenv("SLEEP_TEST_LOCALE", "en-US")
	original := `{"default_locale":"${SLEEP_TEST_LOCALE}","unknown":{"keep":true},"sleep":{"plugins":{"future":"never"}}}`
	if err := os.WriteFile(filepath.Join(dir, settings.FileName), []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0", "--log-format", "json", "--args", "SLEEP_MIN_UPTIME=50ms,SLEEP_COLD_START_MAX=100")
	c.address()
	api := func(method, path, body string) []byte {
		t.Helper()
		status, raw := fixtureHTTP(t, c, dir, method, path, body)
		if status != http.StatusOK {
			t.Fatalf("%s %s: %d %s\n%s", method, path, status, raw, c.log())
		}
		return raw
	}
	path := "/api/v1/plugins/" + id
	var sleep struct {
		Mode        string            `json:"mode"`
		IdleTimeout string            `json:"idle_timeout"`
		Plugins     map[string]string `json:"plugins"`
	}
	if raw := api("GET", "/api/v1/core/sleep", ""); json.Unmarshal(raw, &sleep) != nil || sleep.Mode != "always_on" || sleep.IdleTimeout != "15m0s" || sleep.Plugins["future"] != "never" {
		t.Fatalf("default sleep: %s", raw)
	}
	c.management(t, dir, "", "install", archive, "--yes")
	api("PUT", "/api/v1/core/sleep", `{"mode":"component_idle","idle_timeout":"300ms"}`)
	c.management(t, dir, "", "enable", id)
	first := findPackage(t, c, dir, id)
	if first.Active == "" || !first.Enabled {
		t.Fatalf("initial package: %+v", first)
	}
	firstSession := lastSession(t, c)
	var runtime struct {
		Enabled      bool                         `json:"enabled"`
		GenerationID string                       `json:"generation_id"`
		State        string                       `json:"state"`
		Reason       string                       `json:"reason"`
		ChangedAt    string                       `json:"changed_at"`
		LastWake     *struct{ Reason, At string } `json:"last_wake"`
		LastSuspend  *struct{ Reason, At string } `json:"last_suspend"`
	}
	readRuntime := func() {
		t.Helper()
		if err := json.Unmarshal(api("GET", path+"/runtime", ""), &runtime); err != nil {
			t.Fatal(err)
		}
	}
	waitRuntime := func(want state.RuntimeState) {
		t.Helper()
		deadline := time.Now().Add(10 * time.Second)
		for {
			readRuntime()
			if runtime.State == string(want) {
				return
			}
			if time.Now().After(deadline) {
				t.Fatalf("waiting for %s: %+v\n%s", want, runtime, c.log())
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
	// 1. 自动退出后，管理读取保留 Session 次数、代次和真实历史。
	waitRuntime(state.RuntimeSuspended)
	if !runtime.Enabled || runtime.GenerationID != first.Active || runtime.Reason != "STOP_REASON_IDLE" || runtime.LastWake == nil || runtime.LastSuspend == nil || runtime.LastSuspend.Reason != "idle" {
		t.Fatalf("suspended runtime: %+v", runtime)
	}
	for _, stamp := range []string{runtime.ChangedAt, runtime.LastWake.At, runtime.LastSuspend.At} {
		parsed, err := time.Parse(time.RFC3339, stamp)
		if err != nil || parsed.IsZero() || !strings.HasSuffix(stamp, "Z") || parsed.After(time.Now()) {
			t.Fatalf("history timestamp: %q %v", stamp, err)
		}
	}
	suspendedAt := runtime.LastSuspend.At
	var policy manage.SleepPolicy
	if raw := api("GET", path+"/sleep-policy", ""); json.Unmarshal(raw, &policy) != nil || policy.Policy != nil || !policy.SleepCapable || !policy.Eligible || policy.EffectivePolicy != "300ms" || policy.Reason != "" {
		t.Fatalf("sleep policy: %s", raw)
	}
	api("GET", "/api/v1/core/sleep", "")
	var listed struct {
		Sleep   json.RawMessage `json:"sleep"`
		Plugins []sleepPlugin   `json:"plugins"`
	}
	if raw := c.management(t, dir, "", "sleep-policy"); json.Unmarshal([]byte(raw), &listed) != nil || len(listed.Plugins) != 1 || listed.Plugins[0].ID != id {
		t.Fatalf("CLI list: %s", raw)
	}
	var single sleepPlugin
	if raw := c.management(t, dir, "", "sleep-policy", id); json.Unmarshal([]byte(raw), &single) != nil || single.ID != id || !bytes.Equal(single.Policy, listed.Plugins[0].Policy) || !bytes.Equal(single.Runtime, listed.Plugins[0].Runtime) {
		t.Fatalf("CLI single: %s", raw)
	}
	item := findPackage(t, c, dir, id)
	if item.State != "Suspended" || item.RuntimeState != "SUSPENDED" || lastSession(t, c) != firstSession || strings.Count(c.log(), `"msg":"OnReady"`) != 1 {
		t.Fatalf("read woke plugin: %+v\n%s", item, c.log())
	}
	jobs := api("GET", path+"/jobs", "")
	if !bytes.Contains(jobs, []byte(`"job_id":"pulse"`)) {
		t.Fatalf("suspended job missing: %s", jobs)
	}
	// 2. never、auto 与全局模式复用原代次，在线时长编辑保持原 Session。
	c.management(t, dir, "", "sleep-policy", id, "never")
	waitRuntime(state.RuntimeActive)
	secondSession := lastSession(t, c)
	if secondSession == firstSession || runtime.GenerationID != first.Active || runtime.LastWake == nil || runtime.LastWake.Reason == "" || runtime.LastSuspend.At != suspendedAt {
		t.Fatalf("never restore: %+v\n%s", runtime, c.log())
	}
	c.management(t, dir, "", "sleep-policy", id, "10s")
	if lastSession(t, c) != secondSession {
		t.Fatal("duration change reloaded plugin")
	}
	c.management(t, dir, "", "sleep-policy", id, "auto")
	if raw := api("GET", path+"/sleep-policy", ""); json.Unmarshal(raw, &policy) != nil || policy.Policy != nil {
		t.Fatalf("auto kept override: %s", raw)
	}
	waitRuntime(state.RuntimeSuspended)
	api("PUT", "/api/v1/core/sleep", `{"mode":"always_on","idle_timeout":"300ms"}`)
	waitRuntime(state.RuntimeActive)
	if lastSession(t, c) == secondSession || runtime.GenerationID != first.Active || strings.Count(c.log(), `"msg":"OnReady"`) != 3 {
		t.Fatalf("always_on restore: %+v\n%s", runtime, c.log())
	}
	if got := api("GET", path+"/jobs", ""); !bytes.Equal(got, jobs) {
		t.Fatalf("jobs changed: before=%s after=%s", jobs, got)
	}
	// 3. 登录 Cookie/CSRF 执行一次实际策略写入；其后仍沿用保存的 Bearer。
	token, err := os.ReadFile(filepath.Join(dir, "admin_token"))
	if err != nil {
		t.Fatal(err)
	}
	cookieCall := func(method, endpoint, body, csrf string, cookie *http.Cookie) *http.Response {
		t.Helper()
		request, err := http.NewRequestWithContext(t.Context(), method, "http://"+c.address()+endpoint, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		request.Host = "127.0.0.1:7650"
		request.Header.Set("Origin", settings.Defaults().BaseURL)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Alonabot-CSRF", csrf)
		if cookie != nil {
			request.AddCookie(cookie)
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		return response
	}
	login := cookieCall("POST", "/api/auth/login", `{"token":"`+strings.TrimSpace(string(token))+`"}`, "", nil)
	var credentials struct {
		CSRF string `json:"csrf_token"`
	}
	decodeErr := json.NewDecoder(login.Body).Decode(&credentials)
	if err := errors.Join(decodeErr, login.Body.Close()); err != nil || login.StatusCode != http.StatusOK || len(login.Cookies()) != 1 {
		t.Fatalf("login: %d %v", login.StatusCode, err)
	}
	response := cookieCall("PUT", path+"/sleep-policy", `{"policy":"never"}`, credentials.CSRF, login.Cookies()[0])
	raw, readErr := io.ReadAll(response.Body)
	if err := errors.Join(readErr, response.Body.Close()); err != nil || response.StatusCode != http.StatusOK || strings.TrimSpace(string(raw)) != "{}" {
		t.Fatalf("Cookie write: %d %s %v", response.StatusCode, raw, err)
	}
	if raw := c.management(t, dir, strings.TrimSpace(string(token)), "sleep-policy", id); !strings.Contains(raw, `"policy":"never"`) {
		t.Fatalf("Cookie policy missing: %s", raw)
	}
	stopTestCore(t, c)
	// 4. 进程结束后直接核对同一业务库、迁移次数与配置文件原文。
	db, err := sql.Open("sqlite", filepath.Join(dir, "plugins", id, "generations", first.Active, "db.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var marker string
	var migrations, starts int
	if err := db.QueryRow("SELECT marker,migrations,starts FROM sleep_probe").Scan(&marker, &migrations, &starts); err != nil || marker != "original" || migrations != 1 || starts != 3 {
		t.Fatalf("business database: %q %d %d %v", marker, migrations, starts, err)
	}
	raw, err = os.ReadFile(filepath.Join(dir, settings.FileName))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`${SLEEP_TEST_LOCALE}`, `"unknown"`, `"keep": true`, `"future": "never"`} {
		if !bytes.Contains(raw, []byte(want)) {
			t.Fatalf("settings lost %q: %s", want, raw)
		}
	}
	store, err := state.Open(t.Context(), dir)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	generation, err := store.Active(t.Context(), id)
	if err != nil || generation != first.Active {
		t.Fatalf("generation changed: %s %v", generation, err)
	}
	configuration, err := store.GenerationConfig(t.Context(), generation)
	if err != nil || configuration != `{"marker":"keep-original"}` {
		t.Fatalf("configuration changed: %s %v", configuration, err)
	}
}

func sleepTestPackage(t *testing.T, id, source string) string {
	t.Helper()
	project, out := t.TempDir(), t.TempDir()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	module, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	module = bytes.Replace(module, []byte("module github.com/AlonaBot/AlonaBot"), []byte("module example.invalid/sleepprobe"), 1)
	module = append(module, []byte("\nrequire github.com/AlonaBot/AlonaBot v0.0.0\nreplace github.com/AlonaBot/AlonaBot => "+root+"\n")...)
	sum, err := os.ReadFile(filepath.Join(root, "go.sum"))
	if err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string][]byte{
		"go.mod": module, "go.sum": sum, "main.go": []byte(source),
		"plugin.yaml":        []byte("id: " + id + "\nname: Sleep probe\nversion: 1.0.0\nsource: github.com/example/" + id + "\nrole: plugin\nsleep_capable: true\n"),
		"plugin_config_schema.json": []byte(`{"marker":{"type":"string","default":"keep-original"}}`),
	} {
		if err := os.WriteFile(filepath.Join(project, name), body, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	archive, err := pluginkit.Pack(t.Context(), generate.Options{Dir: project}, out)
	if err != nil {
		t.Fatal(err)
	}
	return archive
}

const sleepTestSource = `package main
import (
 "context"
 "database/sql"
 "fmt"
 "log/slog"
 "time"
 "github.com/AlonaBot/AlonaBot/sdk"
 "github.com/AlonaBot/AlonaBot/sdk/sqlitedb"
)
func New()*sdk.Plugin{
 p:=sdk.NewPlugin()
 sqlitedb.Migrate(p,func(ctx context.Context,db *sql.DB)error{
  _,err:=db.ExecContext(ctx,"CREATE TABLE IF NOT EXISTS sleep_probe(marker TEXT,migrations INTEGER,starts INTEGER); INSERT INTO sleep_probe SELECT 'original',0,0 WHERE NOT EXISTS (SELECT 1 FROM sleep_probe); UPDATE sleep_probe SET migrations=migrations+1");return err
 })
 p.OnReady(func(ctx context.Context)error{
  db,err:=sqlitedb.Open(ctx);if err!=nil{return err}
  if _,err:=db.ExecContext(ctx,"UPDATE sleep_probe SET starts=starts+1");err!=nil{return err}
  slog.Info("OnReady");return nil
 })
 p.Command("ping").Handle(func(c *sdk.Context)error{if c.Config().String("marker")!="keep-original"{return fmt.Errorf("configuration changed")};return nil})
 p.Every("pulse",time.Hour).Handle(func(*sdk.Context)error{return nil})
 return p
}
func main(){if err:=sdk.Run(New());err!=nil{panic(err)}}
`
