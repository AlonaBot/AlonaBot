//go:build unix

package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// TestManualMigrationApproval 经真实管理 CLI/API、插件进程与 SQLite 验证单次许可及停机恢复。
func TestManualMigrationApproval(t *testing.T) {
	fixture, dir := t.TempDir(), t.TempDir()
	crashCore, _ := buildRecoveryFixture(t, fixture)
	plugin := buildManualMigrationPlugin(t, fixture)
	pack := func(version, effects string) string {
		t.Helper()
		manifest := fmt.Sprintf("manifest_version: 1\nid: recovery\nname: External Fixture\nversion: %s\nsource: github.com/example/recovery\nrole: plugin\nprotocol_min_major: 1\nprotocol_max_major: 1\nrequired_capabilities: []\nentrypoint: bin/component\nmigration_mode: self\nprepare_effects: %s\nos: %s\narch: %s\n", version, effects, runtime.GOOS, runtime.GOARCH)
		archive, err := writeTestPackage(filepath.Join(fixture, "manual-"+version), "recovery", "plugin", "", "github.com/example/recovery", version, plugin, "", map[string][]byte{"plugin.yaml": []byte(manifest)})
		if err != nil {
			t.Fatal(err)
		}
		return archive
	}
	v1, v2, v3 := pack("1.0.0", "external"), pack("2.0.0", "external"), pack("3.0.0", "managed_db")
	start := func() *core {
		c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0", "--log-format", "json")
		c.address()
		return c
	}
	c := start()
	pluginDir := filepath.Join(dir, "plugins", "recovery")
	call := func(args ...string) {
		t.Helper()
		command := recoveryCommand(t, c, dir, args...)
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s\n%s", args, err, output, c.log())
		}
	}
	reject := func(message string, args ...string) {
		t.Helper()
		output, err := recoveryCommand(t, c, dir, args...).CombinedOutput()
		if err == nil || !strings.Contains(string(output), message) {
			t.Fatalf("%v: %v, %s; want %s", args, err, output, message)
		}
	}
	effects := func() string {
		t.Helper()
		raw, err := os.ReadFile(filepath.Join(pluginDir, "external-effects"))
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	checkStopped := func(active string) {
		t.Helper()
		item := findPackage(t, c, dir, "recovery")
		if item.Enabled || item.State != "Stopped" || item.Active != active {
			t.Fatalf("maintenance state: %+v, want stopped at %s", item, active)
		}
		waitRecoveryRuntime(t, pluginDir)
		if status, body := fixtureHTTP(t, c, dir, http.MethodGet, "/plugins/recovery/value"); status != 503 {
			t.Fatalf("maintenance published HTTP: %d %s", status, body)
		}
	}
	checkBoot := func(reason string, approved bool, generation string) manualBootstrap {
		t.Helper()
		boots := manualBootstraps(t, pluginDir)
		boot := boots[len(boots)-1]
		if boot.Reason != reason || boot.Approved != approved || boot.Generation != generation || !filepath.IsAbs(boot.Path) || filepath.Base(filepath.Dir(boot.Path)) != generation {
			t.Fatalf("Bootstrap: %+v, want %s/%t/%s", boot, reason, approved, generation)
		}
		return boot
	}
	// 1. 首装和普通 resume 始终没有许可，resume 不执行迁移。
	call("install", v1, "--yes")
	call("enable", "recovery")
	first := findPackage(t, c, dir, "recovery").Active
	if err := os.WriteFile(filepath.Join(pluginDir, "config.json"), []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	initial := checkBoot("ACTIVATION_REASON_FIRST_INSTALL", false, first)
	if status, body := fixtureHTTP(t, c, dir, http.MethodGet, "/plugins/recovery/value"); status != 200 || string(body) != "active" {
		t.Fatalf("active HTTP route: %d %s", status, body)
	}
	before := effects()
	call("restart", "recovery")
	resumed := checkBoot("ACTIVATION_REASON_RESUME", false, first)
	if initial.Session == resumed.Session || effects() != before {
		t.Fatal("restart reused Session or reran migration")
	}
	reject("NEEDS_MANUAL_MIGRATION", "reload", "recovery", "--yes")
	reject("NEEDS_MANUAL_MIGRATION", "update", "recovery", "--archive", v2, "--yes")
	reject("disabled external component", "reload", "recovery", "--manual-migration-approved")
	if effects() != before {
		t.Fatal("rejected operation ran migration")
	}
	// 2. 备份许可与普通确认分开，候选提交后保持停用，下一次请求重新许可。
	call("disable", "recovery")
	call("reload", "recovery", "--manual-migration-approved")
	second := findPackage(t, c, dir, "recovery").Active
	if second == first {
		t.Fatal("maintenance reload reused generation")
	}
	approved := checkBoot("ACTIVATION_REASON_RELOAD", true, second)
	if approved.Session == resumed.Session || approved.Path == resumed.Path || len(strings.Fields(effects())) != 2 {
		t.Fatal("maintenance did not isolate Session/Generation/DB or run exactly one migration")
	}
	checkStopped(second)
	reject("NEEDS_MANUAL_MIGRATION", "rollback", "recovery", first, "--yes")
	call("enable", "recovery")
	checkBoot("ACTIVATION_REASON_RESUME", false, second)
	before = effects()
	stopTestCore(t, c)
	c = start()
	c.waitFor(`"msg":"plugin activated: `)
	checkBoot("ACTIVATION_REASON_RESUME", false, second)
	if effects() != before {
		t.Fatal("Core restart reran migration")
	}
	call("disable", "recovery")
	call("update", "recovery", "--archive", v2, "--yes", "--manual-migration-approved")
	updated := findPackage(t, c, dir, "recovery")
	if updated.Version != "2.0.0" || updated.Active == second {
		t.Fatalf("maintenance update: %+v", updated)
	}
	checkBoot("ACTIVATION_REASON_UPDATE", true, updated.Active)
	checkStopped(updated.Active)
	reject("confirmation required", "rollback", "recovery", first, "--manual-migration-approved")
	call("rollback", "recovery", first, "--yes", "--manual-migration-approved")
	rolled := findPackage(t, c, dir, "recovery")
	if rolled.Version != "1.0.0" || rolled.Active == first || rolled.Active == updated.Active {
		t.Fatalf("maintenance rollback: %+v", rolled)
	}
	checkBoot("ACTIVATION_REASON_ROLLBACK", true, rolled.Active)
	checkStopped(rolled.Active)
	// 3. 迁移和激活失败均保留真实外部副作用，提交点决定旧/新 Active，重启保持停机。
	for _, marker := range []string{"fail-migration", "fail-ready"} {
		old := findPackage(t, c, dir, "recovery").Active
		before = effects()
		name := filepath.Join(pluginDir, marker)
		if err := os.WriteFile(name, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		message := "database migration failed"
		if marker == "fail-ready" {
			message = "SESSION_PHASE_"
		}
		reject(message, "reload", "recovery", "--manual-migration-approved")
		active := findPackage(t, c, dir, "recovery").Active
		if (active == old) != (marker == "fail-migration") || len(strings.Fields(effects())) != len(strings.Fields(before))+1 {
			t.Fatalf("%s: Active %s -> %s, effects %q -> %q", marker, old, active, before, effects())
		}
		checkStopped(active)
		before = effects()
		boots := len(manualBootstraps(t, pluginDir))
		stopTestCore(t, c)
		c = start()
		checkStopped(active)
		if effects() != before || len(manualBootstraps(t, pluginDir)) != boots {
			t.Fatalf("%s recovery reran candidate", marker)
		}
		if err := os.Remove(name); err != nil {
			t.Fatal(err)
		}
		call("enable", "recovery")
		checkBoot("ACTIVATION_REASON_RESUME", false, active)
		if effects() != before {
			t.Fatal("explicit resume reran failed migration")
		}
		call("disable", "recovery")
	}
	// 4. PREPARED 与提交后杀死 Core，恢复仅清理候选，不重复未知外部写入。
	for _, phase := range []string{"prepared", "committed"} {
		old := findPackage(t, c, dir, "recovery").Active
		stopTestCore(t, c)
		c = startRecoveryCore(t, crashCore, dir, phase)
		request := recoveryCommand(t, c, dir, "reload", "recovery", "--manual-migration-approved")
		var output strings.Builder
		request.Stdout, request.Stderr = &output, &output
		if err := request.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = request.Process.Kill() })
		done := make(chan error, 1)
		go func() { done <- request.Wait() }()
		recoveryWait(t, func() bool {
			select {
			case err := <-done:
				t.Fatalf("candidate exited before %s: %v %s", phase, err, output.String())
			default:
			}
			raw, err := os.ReadFile(filepath.Join(dir, "checkpoint"))
			return err == nil && string(raw) == phase
		})
		if err := c.cmd.Process.Kill(); err != nil {
			t.Fatal(err)
		}
		<-c.done
		if err := <-done; err == nil {
			t.Fatal("interrupted request succeeded")
		}
		waitRecoveryRuntime(t, pluginDir)
		boots := manualBootstraps(t, pluginDir)
		boot := boots[len(boots)-1]
		st := recoveryStore(t, dir)
		records, err := st.Journals(t.Context(), "recovery")
		if err != nil {
			t.Fatal(err)
		}
		var operation state.Journal
		for _, record := range records {
			if record.NewGeneration == boot.Generation {
				operation = record
			}
		}
		if operation.WasEnabled || operation.WasRunning || operation.Phase != phase || hex.EncodeToString(operation.SessionID) != boot.Session || operation.DatabasePath != boot.Path || !boot.Approved {
			t.Fatalf("candidate identity: %+v, Bootstrap %+v", operation, boot)
		}
		if err := st.Close(); err != nil {
			t.Fatal(err)
		}
		before = effects()
		c = start()
		want, wantPhase := old, "cleaned"
		if phase == "committed" {
			want, wantPhase = boot.Generation, phase
		}
		checkStopped(want)
		if effects() != before || len(manualBootstraps(t, pluginDir)) != len(boots) {
			t.Fatalf("%s recovery repeated external effects", phase)
		}
		stopTestCore(t, c)
		st = recoveryStore(t, dir)
		recovered, err := st.GetJournal(t.Context(), operation.OpID)
		if err != nil || recovered.Phase != wantPhase {
			t.Fatalf("recovered Journal: %+v, %v", recovered, err)
		}
		if phase == "prepared" && (len(recovered.SessionID) != 0 || recovered.DatabasePath != "") {
			t.Fatalf("recovery retained candidate resources: %+v", recovered)
		}
		if err := st.Close(); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(filepath.Join(dir, "checkpoint")); err != nil {
			t.Fatal(err)
		}
		c = start()
	}
	// 5. 从 external 更新到 managed_db 仍需许可，之后普通候选与 resume 保持原有用法。
	call("update", "recovery", "--archive", v3, "--yes", "--manual-migration-approved")
	managed := findPackage(t, c, dir, "recovery")
	checkStopped(managed.Active)
	call("enable", "recovery")
	reject("disabled external component", "reload", "recovery", "--manual-migration-approved")
	call("reload", "recovery")
	managed = findPackage(t, c, dir, "recovery")
	checkBoot("ACTIVATION_REASON_RELOAD", false, managed.Active)
	if !managed.Enabled || managed.State != "Activated" {
		t.Fatalf("ordinary managed candidate: %+v", managed)
	}
	stopTestCore(t, c)
}

type manualBootstrap struct {
	Session, Generation, Reason, Path string
	Approved                          bool
}

func manualBootstraps(t *testing.T, pluginDir string) []manualBootstrap {
	t.Helper()
	file, err := os.Open(filepath.Join(pluginDir, "bootstraps"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = file.Close() })
	decoder := json.NewDecoder(file)
	var result []manualBootstrap
	for {
		var boot manualBootstrap
		if err := decoder.Decode(&boot); err == io.EOF {
			return result
		} else if err != nil {
			t.Fatal(err)
		}
		result = append(result, boot)
	}
}

// buildManualMigrationPlugin 只在临时 overlay 观察 Core 实际下发的 Bootstrap，不增加 SDK 产品入口。
func buildManualMigrationPlugin(t *testing.T, dir string) string {
	t.Helper()
	original, err := filepath.Abs("../../sdk/session.go")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	const marker = "l.runtime.database.Begin(boot.GetDatabaseDescriptor())"
	if strings.Count(string(raw), marker) != 1 {
		t.Fatal("ambiguous Bootstrap observation point")
	}
	source := strings.Replace(string(raw), marker, `file, openErr := os.OpenFile("bootstraps", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
if openErr != nil { return nil, openErr }
writeErr := json.NewEncoder(file).Encode(map[string]any{"Session": hex.EncodeToString(sid), "Generation": boot.GetGenerationId(), "Reason": boot.GetDatabaseDescriptor().GetActivationReason().String(), "Approved": boot.GetDatabaseDescriptor().GetManualMigrationApproved(), "Path": boot.GetDatabaseDescriptor().GetSqlitePath()})
if err := errors.Join(writeErr, file.Close()); err != nil { return nil, err }
`+marker, 1)
	modified, overlay, input, binary := filepath.Join(dir, "session.go"), filepath.Join(dir, "manual-overlay.json"), filepath.Join(dir, "manual-plugin.go"), filepath.Join(dir, "manual-plugin")
	replacement, err := json.Marshal(map[string]any{"Replace": map[string]string{original: modified}})
	if err != nil {
		t.Fatal(err)
	}
	for name, contents := range map[string][]byte{modified: []byte(source), overlay: replacement, input: []byte(manualMigrationSource)} {
		if err := os.WriteFile(name, contents, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	command := exec.CommandContext(context.Background(), "go", fixtureBuildArgs("-overlay", overlay, "-o", binary, input)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build external plugin: %v\n%s", err, output)
	}
	return binary
}

const manualMigrationSource = `package main
import (
 "context"
 "database/sql"
 "errors"
 "fmt"
 "log/slog"
 "net/http"
 "os"
 "github.com/AlonaBot/AlonaBot/sdk"
 "github.com/AlonaBot/AlonaBot/sdk/genimpl"
 "github.com/AlonaBot/AlonaBot/sdk/sqlitedb"
)
func init(){genimpl.Register("main",genimpl.Resources{})}
func main(){
 p:=sdk.NewPlugin()
 sqlitedb.Migrate(p,func(ctx context.Context,db *sql.DB)error{
  var seq int;var name,path string
  if err:=db.QueryRowContext(ctx,"PRAGMA database_list").Scan(&seq,&name,&path);err!=nil{return err}
  f,err:=os.OpenFile("external-effects",os.O_CREATE|os.O_APPEND|os.O_WRONLY,0600);if err!=nil{return err}
  _,writeErr:=fmt.Fprintln(f,path);if err:=errors.Join(writeErr,f.Close());err!=nil{return err}
  if _,err:=os.Stat("fail-migration");err==nil{return fmt.Errorf("fixture migration failed after external write")}else if !os.IsNotExist(err){return err}
  _,err=db.ExecContext(ctx,"CREATE TABLE IF NOT EXISTS value (n INTEGER)");return err
 })
 p.OnReady(func(context.Context)error{if _,err:=os.Stat("fail-ready");err==nil{return fmt.Errorf("fixture ready failed")}else if !os.IsNotExist(err){return err};return nil})
 p.Route("GET","/value").HandleFunc(func(w http.ResponseWriter,r *http.Request){fmt.Fprint(w,"active")})
 if err:=sdk.Run(p);err!=nil{slog.Error("fixture failed","error",err);os.Exit(1)}
}
`
