//go:build unix

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/buttons"
	"github.com/AlonaBot/AlonaBot/internal/core/process"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/oslock"
	"google.golang.org/protobuf/proto"
)

// TestRecoveryEndToEnd 覆盖提交两侧崩溃、残留写入者与显式回退；这些是独立的恢复主路径。
func TestRecoveryEndToEnd(t *testing.T) {
	fixture := t.TempDir()
	binary, plugin := buildRecoveryFixture(t, fixture)
	archive := recoveryPackage(t, fixture, plugin, "1.0.0")
	start := func(t *testing.T, dir, phase string) *core {
		return startRecoveryCore(t, binary, dir, phase)
	}
	for _, phase := range []string{"prepared", "committed", "started"} {
		t.Run(phase, func(t *testing.T) {
			dir := t.TempDir()
			c := start(t, dir, "")
			c.management(t, dir, "", "install", archive, "--yes")
			c.management(t, dir, "", "enable", "recovery")
			old := findPackage(t, c, dir, "recovery").Active
			stopTestCore(t, c)
			c = start(t, dir, phase)
			c.waitFor(`"msg":"plugin activated"`)
			pluginDir := filepath.Join(dir, "plugins", "recovery")
			survivor := false
			if phase == "started" {
				starts, err := os.ReadFile(filepath.Join(pluginDir, "starts"))
				if err != nil {
					t.Fatal(err)
				}
				pids := strings.Fields(string(starts))
				pid, err := strconv.Atoi(pids[len(pids)-1])
				if err != nil {
					t.Fatal(err)
				}
				survivor = true
				t.Cleanup(func() {
					if survivor {
						_ = syscall.Kill(pid, syscall.SIGKILL)
					}
				})
				if err := syscall.Mkfifo(filepath.Join(pluginDir, "hold-runtime"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(pluginDir, "config.json"), []byte(`{}`), 0o600); err != nil {
				t.Fatal(err)
			}
			// CLI 阻塞在真实准备操作；钩子只在临时 overlay 二进制中存在。
			request := recoveryCommand(t, c, dir, "reload", "recovery")
			var output strings.Builder
			request.Stdout, request.Stderr = &output, &output
			if err := request.Start(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = request.Process.Kill() })
			requestDone := make(chan error, 1)
			go func() { requestDone <- request.Wait() }()
			recoveryWait(t, func() bool {
				select {
				case err := <-requestDone:
					t.Fatalf("reload exited before checkpoint: %v: %s\n%s", err, output.String(), c.log())
				default:
				}
				_, err := os.Stat(filepath.Join(dir, "checkpoint"))
				return err == nil
			})
			if err := c.cmd.Process.Kill(); err != nil {
				t.Fatal(err)
			}
			<-c.done
			if err := <-requestDone; err == nil {
				t.Fatalf("interrupted reload succeeded: %s", output.String())
			}
			st := recoveryStore(t, dir)
			records, err := st.Journals(t.Context(), "recovery")
			if err != nil || len(records) != 2 {
				t.Fatalf("journals before crash: %+v, %v", records, err)
			}
			var operation state.Journal
			for _, record := range records {
				if record.ExpectedOld == old {
					operation = record
				}
			}
			if operation.Phase != phase {
				t.Fatalf("checkpoint phase = %s, want %s", operation.Phase, phase)
			}
			if err := st.Close(); err != nil {
				t.Fatal(err)
			}
			if phase == "started" {
				recoveryWait(t, func() bool { _, err := os.Stat(filepath.Join(pluginDir, "runtime-held")); return err == nil })
				pidRaw, err := os.ReadFile(filepath.Join(pluginDir, "runtime-held"))
				if err != nil {
					t.Fatal(err)
				}
				pid, err := strconv.Atoi(string(pidRaw))
				if err != nil {
					t.Fatal(err)
				}
				before, err := os.ReadFile(filepath.Join(pluginDir, "starts"))
				if err != nil {
					t.Fatal(err)
				}
				c = start(t, dir, "")
				c.waitFor(`old process may still be running`)
				item := findPackage(t, c, dir, "recovery")
				after, err := os.ReadFile(filepath.Join(pluginDir, "starts"))
				if err != nil || item.State != "Degraded" || item.Active != old || string(before) != string(after) {
					t.Fatalf("surviving writer: %+v, starts %q -> %q, %v", item, before, after, err)
				}
				if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
					t.Fatal(err)
				}
				survivor = false
				waitRecoveryRuntime(t, pluginDir)
				if err := os.Remove(filepath.Join(pluginDir, "hold-runtime")); err != nil {
					t.Fatal(err)
				}
				c.management(t, dir, "", "restart", "recovery")
			} else {
				waitRecoveryRuntime(t, pluginDir)
				c = start(t, dir, "")
				c.waitFor(`"msg":"plugin activated"`)
			}
			item := findPackage(t, c, dir, "recovery")
			wantActive, wantPhase := old, "cleaned"
			if phase == "committed" {
				wantActive, wantPhase = operation.NewGeneration, "committed"
			}
			if item.Active != wantActive || item.State != "Activated" {
				t.Fatalf("recovered package: %+v, want active %s", item, wantActive)
			}
			stopTestCore(t, c)
			st = recoveryStore(t, dir)
			got, err := st.GetJournal(t.Context(), operation.OpID)
			if err != nil || got.Phase != wantPhase {
				t.Fatalf("recovered journal: %+v, %v", got, err)
			}
			candidateDir := filepath.Join(pluginDir, "generations", operation.NewGeneration)
			_, err = os.Stat(candidateDir)
			if phase == "committed" && err != nil || phase != "committed" && !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("candidate directory after %s: %v", phase, err)
			}
			// 新 Session 只激活一次；禁止同时恢复旧版和新代次。
			if count := strings.Count(c.log(), `"msg":"plugin activated"`); count != 1 {
				t.Fatalf("activation count = %d: %s", count, c.log())
			}
		})
	}
	t.Run("rollback", func(t *testing.T) {
		dir := t.TempDir()
		c := start(t, dir, "")
		c.management(t, dir, "", "install", archive, "--yes")
		c.management(t, dir, "", "enable", "recovery")
		old := findPackage(t, c, dir, "recovery").Active
		stopTestCore(t, c)
		st := recoveryStore(t, dir)
		// 真实 token 持久化并经同一解析器检查当前 Active，跨 rollback 保留旧映射。
		newService := func() *buttons.Service {
			return buttons.New(st, time.Now, time.Hour, func(id string) (string, []*pb.RouteDescriptor, bool) {
				active, err := st.Active(t.Context(), id)
				if err != nil {
					t.Fatal(err)
				}
				return active, []*pb.RouteDescriptor{{RouteId: "button.check", Kind: pb.RouteKind_ROUTE_KIND_EVENT, Spec: &pb.RouteDescriptor_Event{Event: &pb.EventSpec{EventTypeFilter: []string{"button.clicked"}, ActionId: "check", ActionSchemaVersion: 1}}}}, true
			})
		}
		service := newService()
		target := &pb.Target{AdapterId: "test", AdapterInstanceId: "instance", AccountId: "account", ChatId: "chat"}
		chain := &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Buttons{Buttons: &pb.ButtonsSegment{Rows: []*pb.ButtonRow{{Buttons: []*pb.Button{{ActionId: "check", Value: proto.String(`{}`)}}}}}}}}}
		if err := service.Issue(t.Context(), router.NewCaller("recovery", old), target, chain); err != nil {
			t.Fatal(err)
		}
		token := chain.Segments[0].GetButtons().Rows[0].Buttons[0].GetCoreToken()
		click := func() error {
			_, _, err := service.Resolve(t.Context(), &pb.Event{EventType: "button.clicked", Target: target, Actor: &pb.Actor{ExternalUserId: "alice", AdapterInstanceId: "instance"}, EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_ButtonClicked{ButtonClicked: &pb.ButtonClickedPayload{Value: token}}}})
			return err
		}
		if err := click(); err != nil {
			t.Fatalf("original token rejected: %v", err)
		}
		if err := st.Close(); err != nil {
			t.Fatal(err)
		}
		c = start(t, dir, "")
		c.waitFor(`"msg":"plugin activated"`)
		newArchive := recoveryPackage(t, fixture, plugin, "2.0.0")
		c.management(t, dir, "", "update", "recovery", "--archive", newArchive, "--yes")
		current := findPackage(t, c, dir, "recovery").Active
		status, body := fixtureHTTP(t, c, dir, http.MethodPut, "/plugins/recovery/value?n=2")
		if status != http.StatusNoContent {
			t.Fatalf("write new database: %d %s", status, body)
		}
		c.management(t, dir, "", "rollback", "recovery", old, "--yes")
		item := findPackage(t, c, dir, "recovery")
		if item.Active == old || item.Active == current || item.Version != "1.0.0" || item.State != "Activated" {
			t.Fatalf("rollback package: %+v", item)
		}
		status, body = fixtureHTTP(t, c, dir, http.MethodGet, "/plugins/recovery/value")
		if status != http.StatusOK || string(body) != "1" {
			t.Fatalf("rollback data: %d %s", status, body)
		}
		oldFile, err := os.Stat(filepath.Join(dir, "plugins", "recovery", "generations", old, "db.sqlite"))
		if err != nil {
			t.Fatal(err)
		}
		newFile, err := os.Stat(filepath.Join(dir, "plugins", "recovery", "generations", item.Active, "db.sqlite"))
		if err != nil || os.SameFile(oldFile, newFile) {
			t.Fatalf("rollback reused source database: %v", err)
		}
		stopTestCore(t, c)
		st = recoveryStore(t, dir)
		service = newService()
		if err := click(); err == nil || !strings.Contains(err.Error(), "generation is no longer active") {
			t.Fatalf("rollback revived old token: %v", err)
		}
	})
}

func recoveryStore(t *testing.T, dir string) *state.Store {
	t.Helper()
	st, err := state.Open(t.Context(), dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Error(err)
		}
	})
	return st
}

func recoveryWait(t *testing.T, ready func() bool) {
	t.Helper()
	deadline := time.NewTimer(20 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for !ready() {
		select {
		case <-deadline.C:
			t.Fatal("recovery checkpoint timed out")
		case <-tick.C:
		}
	}
}

func waitRecoveryRuntime(t *testing.T, dir string) {
	t.Helper()
	recoveryWait(t, func() bool {
		lock, err := process.LockRuntime(dir)
		if errors.Is(err, oslock.ErrLocked) {
			return false
		}
		if err != nil {
			t.Fatal(err)
		}
		if err := lock.Unlock(); err != nil {
			t.Fatal(err)
		}
		return true
	})
}

func recoveryCommand(t *testing.T, c *core, dir string, args ...string) *exec.Cmd {
	t.Helper()
	token, err := os.ReadFile(filepath.Join(dir, "admin_token"))
	if err != nil {
		t.Fatal(err)
	}
	destination, err := url.Parse("http://" + c.address())
	if err != nil {
		t.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(destination)
	base, err := url.Parse(settings.Defaults().BaseURL)
	if err != nil {
		t.Fatal(err)
	}
	director := proxy.Director
	proxy.Director = func(r *http.Request) { director(r); r.Host = base.Host }
	server := httptest.NewServer(proxy)
	t.Cleanup(server.Close)
	command := exec.Command(filepath.Join(binDir, "alonabot"), append(args, "--admin-url", server.URL)...)
	command.Env = append(os.Environ(), "ALONABOT_ADMIN_TOKEN="+strings.TrimSpace(string(token)))
	return command
}

func startRecoveryCore(t *testing.T, binary, dir, phase string) *core {
	t.Helper()
	c := &core{t: t, cmd: exec.Command(binary, "run", "--data-dir", dir, "--listen", "127.0.0.1:0", "--log-format", "json"), added: make(chan struct{}, 1), done: make(chan struct{})}
	c.cmd.Env = append(os.Environ(), "ALONABOT_TEST_CRASH_PHASE="+phase, "ALONABOT_TEST_CHECKPOINT="+filepath.Join(dir, "checkpoint"))
	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := c.cmd.Start(); err != nil {
		t.Fatal(err)
	}
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			c.mu.Lock()
			c.lines = append(c.lines, scanner.Text())
			c.mu.Unlock()
			select {
			case c.added <- struct{}{}:
			default:
			}
		}
		_ = c.cmd.Wait()
		close(c.done)
	}()
	t.Cleanup(func() { _ = c.cmd.Process.Kill(); <-c.done })
	return c
}

// buildRecoveryFixture 通过 Go overlay 注入精确崩溃节点与不释放旧进程锁的测试钩子。
// overlay 与二进制只在临时目录，正式 Core 和 SDK 保持原实现。
func buildRecoveryFixture(t *testing.T, dir string) (string, string) {
	t.Helper()
	replacements := map[string]string{}
	for name, edits := range map[string][][2]string{
		"internal/core/journal/commit.go": {
			{"running, err := o.commit(ctx, r, &committed)", "recoveryCheckpoint(\"prepared\")\n\trunning, err := o.commit(ctx, r, &committed)"},
			{"*committed = true", "recoveryCheckpoint(\"committed\")\n\t\t*committed = true"},
		},
		"internal/core/journal/journal.go": {{"// 3. 原停止链", "recoveryCheckpoint(\"started\")\n\t// 3. 原停止链"}},
		"sdk/sdk.go": {{"\"log/slog\"", "\"log/slog\"\n\t\"os\""}, {"defer func() { err = errors.Join(err, lock.Unlock()) }()", `defer func() {
 if err != nil { if _, statErr := os.Stat("hold-runtime"); statErr == nil {
  if writeErr := os.WriteFile("runtime-held", []byte(fmt.Sprint(os.Getpid())), 0600); writeErr != nil { panic(writeErr) }
  file, openErr := os.Open("hold-runtime"); if openErr != nil { panic(openErr) }; file.Close()
 } }
 err = errors.Join(err, lock.Unlock())
}()`}},
	} {
		original, err := filepath.Abs(filepath.Join("../..", name))
		if err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(original)
		if err != nil {
			t.Fatal(err)
		}
		source := string(raw)
		for _, edit := range edits {
			if strings.Count(source, edit[0]) != 1 {
				t.Fatalf("ambiguous recovery hook: %s: %s", name, edit[0])
			}
			source = strings.Replace(source, edit[0], edit[1], 1)
		}
		if name == "internal/core/journal/commit.go" {
			source += `
func recoveryCheckpoint(phase string) {
 if os.Getenv("ALONABOT_TEST_CRASH_PHASE") != phase { return }
 if err := os.WriteFile(os.Getenv("ALONABOT_TEST_CHECKPOINT"), []byte(phase), 0600); err != nil { panic(err) }
 select {}
}
`
		}
		modified := filepath.Join(dir, strings.ReplaceAll(name, "/", "_"))
		if err := os.WriteFile(modified, []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
		replacements[original] = modified
	}
	raw, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(overlay, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	pluginSource := filepath.Join(dir, "plugin.go")
	if err := os.WriteFile(pluginSource, []byte(recoveryFixtureSource), 0o600); err != nil {
		t.Fatal(err)
	}
	binary, plugin := filepath.Join(dir, "core"), filepath.Join(dir, "plugin")
	for target, source := range map[string]string{binary: ".", plugin: pluginSource} {
		cmd := exec.CommandContext(context.Background(), "go", "build", "-overlay", overlay, "-o", target, source)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build recovery fixture: %v\n%s", err, output)
		}
	}
	return binary, plugin
}

func recoveryPackage(t *testing.T, dir, plugin, version string) string {
	t.Helper()
	// 覆盖打包 helper 的默认声明，显式使用受管 SQLite。
	manifest := fmt.Sprintf("manifest_version: 1\nid: recovery\nname: Recovery Fixture\nversion: %s\nsource: github.com/example/recovery\nrole: plugin\nprotocol_min_major: 1\nprotocol_max_major: 1\nrequired_capabilities: []\nentrypoint: bin/component\nconfig_version: 1\nmigration_mode: none\nprepare_effects: managed_db\nos: %s\narch: %s\n", version, runtime.GOOS, runtime.GOARCH)
	archive, err := writeTestPackage(filepath.Join(dir, version), "recovery", "plugin", "", "github.com/example/recovery", version, plugin, "", map[string][]byte{"plugin.yaml": []byte(manifest)})
	if err != nil {
		t.Fatal(err)
	}
	return archive
}

const recoveryFixtureSource = `package main
import (
 "context"
 "fmt"
 "log/slog"
 "net/http"
 "os"
 "github.com/AlonaBot/AlonaBot/sdk"
 "github.com/AlonaBot/AlonaBot/sdk/genimpl"
 "github.com/AlonaBot/AlonaBot/sdk/sqlitedb"
)
func init(){genimpl.Register("main",genimpl.Resources{})}
func main() {
 p:=sdk.NewPlugin()
 p.OnStart(func(ctx context.Context)error{
  if _,err:=os.Stat("fail-start");err==nil{return fmt.Errorf("fixture start failed")}
  db,err:=sqlitedb.Open(ctx);if err!=nil{return err}
  if _,err=db.ExecContext(ctx,"CREATE TABLE IF NOT EXISTS value (id INTEGER PRIMARY KEY, n INTEGER NOT NULL); INSERT OR IGNORE INTO value VALUES (1,1)");err!=nil{return err}
  file,err:=os.OpenFile("starts",os.O_CREATE|os.O_APPEND|os.O_WRONLY,0600);if err!=nil{return err}
  _,writeErr:=fmt.Fprintln(file,os.Getpid());closeErr:=file.Close();if writeErr!=nil{return writeErr};return closeErr
 })
 p.OnReady(func(context.Context)error{if _,err:=os.Stat("fail-ready");err==nil{return fmt.Errorf("fixture ready failed")};return nil})
 p.Route("GET","/value").HandleFunc(func(w http.ResponseWriter,r *http.Request){
  db,err:=sqlitedb.Open(r.Context());if err!=nil{http.Error(w,err.Error(),500);return}
  var n int;if err:=db.QueryRowContext(r.Context(),"SELECT n FROM value WHERE id=1").Scan(&n);err!=nil{http.Error(w,err.Error(),500);return}
  fmt.Fprint(w,n)
 })
 p.Route("PUT","/value").HandleFunc(func(w http.ResponseWriter,r *http.Request){
  db,err:=sqlitedb.Open(r.Context());if err==nil{_,err=db.ExecContext(r.Context(),"UPDATE value SET n=? WHERE id=1",r.URL.Query().Get("n"))};if err!=nil{http.Error(w,err.Error(),500);return};w.WriteHeader(204)
 })
 if err:=sdk.Run(p);err!=nil{slog.Error("fixture failed","error",err);os.Exit(1)}
}
`
