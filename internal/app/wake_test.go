//go:build unix

package app

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	coreadapter "github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/generate"
	genresources "github.com/AlonaBot/AlonaBot/internal/generated"
	"github.com/AlonaBot/AlonaBot/internal/pluginkit"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// TestWake 验证真实 Core、回环 Adapter、已提交插件进程与 SQLite 的命令/Service/HTTP 唤醒。
func TestWake(t *testing.T) {
	ctx := t.Context()
	entered, finish := make(chan string, 4), make(chan struct{}, 4)
	audit := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entered <- strings.TrimPrefix(r.URL.Path, "/")
		select {
		case <-finish:
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(func() { close(finish); audit.Close() })
	project, out := t.TempDir(), t.TempDir()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	module, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	module = bytes.Replace(module, []byte("module github.com/AlonaBot/AlonaBot"), []byte("module example.invalid/wakeprobe"), 1)
	module = append(module, []byte("\nrequire github.com/AlonaBot/AlonaBot v0.0.0\nreplace github.com/AlonaBot/AlonaBot => "+root+"\n")...)
	sum, err := os.ReadFile(filepath.Join(root, "go.sum"))
	if err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string][]byte{
		"go.mod": module, "go.sum": sum,
		"plugin.yaml": []byte("id: wakeprobe\nname: Wake probe\nversion: 0.1.0\nauthor: AlonaBot\nsource: github.com/AlonaBot/wakeprobe\nrole: plugin\n"),
		"main.go":     []byte(strings.ReplaceAll(wakePluginSource, "AUDIT_URL", audit.URL)),
	} {
		if err := os.WriteFile(filepath.Join(project, name), body, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	archive, err := pluginkit.Pack(ctx, generate.Options{Dir: project}, "https://example.invalid/wakeprobe.tar.gz", "", out)
	if err != nil {
		t.Fatal(err)
	}
	releaseJSON, err := os.ReadFile(filepath.Join(out, "release.json"))
	if err != nil {
		t.Fatal(err)
	}
	h := newWebUIHarness(t)
	h.start(nil)
	core := h.core
	manager := manage.New(manage.Options{Store: core.store, Settings: core.assembly.settings, Packages: filepath.Join(h.dir, "packages"), Limits: packages.NewLimits(core.values), Adapters: core.assembly.adapters, State: core.assembly.componentState})
	if _, err := manager.Install(ctx, manage.InstallRequest{Archive: archive, ReleaseJSON: string(releaseJSON), Confirm: true}); err != nil {
		t.Fatal(err)
	}
	h.stop()
	settingsJSON := `{"listen":"127.0.0.1:0","base_url":"http://127.0.0.1","adapters":{"loop":{"instances":{"pipeline-instance":{}}}}}`
	if err := os.WriteFile(filepath.Join(h.dir, settings.FileName), []byte(settingsJSON), 0o600); err != nil {
		t.Fatal(err)
	}
	loop := &pipelineAdapter{}
	genresources.Bind(loop, genresources.Resources{Manifest: []byte("id: wake-loop\nname: Wake loop\nversion: 0.1.0\nsource: test/wake-loop\nrole: adapter\nadapter_id: loop\n"), AdapterConfigSchema: []byte(`{}`)})
	coreContext, cancelCore := context.WithCancel(ctx)
	core, err = Start(coreContext, Options{DataDir: h.dir, Components: []Component{AdapterComponent(loop)}})
	if err != nil {
		cancelCore()
		t.Fatal(err)
	}
	h.core, h.cancel = core, cancelCore
	const id, config = "wakeprobe", `{"marker":"keep-original"}`
	if err := os.MkdirAll(filepath.Join(h.dir, "plugins", id), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(h.dir, "plugins", id, "config.json"), []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	h.action("enable", id, http.StatusOK)
	original, ok := core.assembly.router.Caller(id)
	if !ok {
		t.Fatal("enabled plugin has no Caller")
	}
	dbPath := filepath.Join(h.dir, "plugins", id, "generations", original.GenerationID, "db.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	stopIdle := func() error {
		return core.operation(ctx, func(context.Context) error {
			for _, r := range core.components {
				if r != nil && r.id == id {
					return core.stopComponent(r, time.Now(), host.NewStopTiming(core.values), pb.StopReason_STOP_REASON_IDLE)
				}
			}
			return errors.New("plugin component missing")
		})
	}
	previous := original
	for _, kind := range []string{"command", "service", "http"} {
		t.Run(kind, func(t *testing.T) {
			if err := stopIdle(); err != nil {
				t.Fatal(err)
			}
			status, err := core.store.RuntimeStatus(ctx, id)
			if err != nil || status.State != state.RuntimeSuspended {
				t.Fatalf("suspended status=%+v error=%v", status, err)
			}
			result := make(chan error, 1)
			go func() {
				if kind == "command" {
					instance, ok := core.assembly.adapters.Get(pipelineInstanceID)
					if !ok {
						result <- errors.New("loop Adapter instance missing")
						return
					}
					data, err := json.Marshal(pipelineInput{Text: "/ping", User: "alice"})
					if err != nil {
						result <- err
						return
					}
					_, failure := instance.(coreadapter.ActionExecutor).Action(ctx, &pb.Target{AdapterId: "loop", AdapterInstanceId: pipelineInstanceID, AccountId: pipelineAccountID, ChatId: "group", ThreadId: proto.String("topic")}, "loop.inject", 1, string(data))
					if failure != nil {
						result <- errors.New(failure.String())
						return
					}
					poll := time.NewTicker(time.Millisecond)
					defer poll.Stop()
					for {
						stats := core.assembly.router.Stats()
						if stats.Running == 0 && stats.Queued == 0 {
							result <- nil
							return
						}
						select {
						case <-poll.C:
						case <-ctx.Done():
							result <- ctx.Err()
							return
						}
					}
				}
				if kind == "service" {
					request := &pb.PublishRequest{EventName: "once"}
					data, err := proto.Marshal(request)
					if err != nil {
						result <- err
						return
					}
					r := core.assembly.broker.Call(ctx, router.Caller{}, &wire.Inbound{Target: id + "/probe", Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: 1, Bytes: data}})
					if r.Err != nil {
						result <- errors.New(r.Err.String())
						return
					}
					out := &pb.PublishResponse{}
					if err := proto.Unmarshal(r.Payload.Bytes, out); err != nil {
						result <- err
						return
					}
					if out.EventId != "once" {
						result <- errors.New("service response changed")
						return
					}
					result <- nil
					return
				}
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+core.http.Addr+"/plugins/"+id+"/probe", nil)
				if err != nil {
					result <- err
					return
				}
				req.Header.Set("Authorization", "Bearer "+h.token)
				response, err := h.client.Do(req)
				if err != nil {
					result <- err
					return
				}
				body, err := io.ReadAll(response.Body)
				err = errors.Join(err, response.Body.Close())
				if response.StatusCode != http.StatusOK || string(body) != "once" {
					err = errors.Join(err, errors.New("HTTP response changed: "+string(body)))
				}
				result <- err
			}()
			select {
			case got := <-entered:
				if got != kind {
					t.Fatalf("delivery=%q want=%q", got, kind)
				}
			case err := <-result:
				t.Fatalf("call ended before delivery: %v", err)
			case <-time.After(pipelineTimeout):
				t.Fatal("wake did not deliver")
			}
			current, ok := core.assembly.router.Caller(id)
			if !ok || current == previous || current.GenerationID != original.GenerationID {
				t.Fatalf("Session/generation: before=%+v after=%+v", previous, current)
			}
			previous = current
			if err := stopIdle(); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION {
				t.Fatalf("idle stop during real activity=%v", err)
			}
			finish <- struct{}{}
			if err := <-result; err != nil {
				t.Fatal(err)
			}
			waitTaskState(t, func() bool {
				err := stopIdle()
				if err != nil && wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION {
					t.Fatal(err)
				}
				return err == nil
			})
			generation, err := core.store.Active(ctx, id)
			if err != nil || generation != original.GenerationID {
				t.Fatalf("generation=%q error=%v", generation, err)
			}
			snapshot, err := core.store.GenerationConfig(ctx, generation)
			if err != nil || snapshot != config {
				t.Fatalf("config=%q error=%v", snapshot, err)
			}
			var value string
			if err := db.QueryRowContext(ctx, "SELECT value FROM wake_marker").Scan(&value); err != nil || value != "original" {
				t.Fatalf("DB=%q error=%v", value, err)
			}
			select {
			case duplicate := <-entered:
				t.Fatalf("duplicate delivery: %q", duplicate)
			default:
			}
		})
	}
}

const wakePluginSource = `package main
import (
 "context"
 "database/sql"
 "fmt"
 "net/http"
 "io"
 pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
 "github.com/AlonaBot/AlonaBot/sdk"
 "github.com/AlonaBot/AlonaBot/sdk/service"
 "github.com/AlonaBot/AlonaBot/sdk/sqlitedb"
)
func observe(kind string) error {
 response, err := http.Get("AUDIT_URL/"+kind)
 if err != nil { return err }
 _, err = io.Copy(io.Discard, response.Body)
 closeErr := response.Body.Close()
 if err != nil { return err }; return closeErr
}
func New() *sdk.Plugin {
 p := sdk.NewPlugin()
 sqlitedb.Migrate(p,func(ctx context.Context,db *sql.DB) error {
  _,err:=db.ExecContext(ctx,"CREATE TABLE wake_marker (value TEXT); INSERT INTO wake_marker VALUES ('original')");return err
 })
 p.OnReady(func(ctx context.Context)error{
  db,err:=sqlitedb.Open(ctx);if err!=nil{return err};var value string
  if err:=db.QueryRowContext(ctx,"SELECT value FROM wake_marker").Scan(&value);err!=nil{return err}
  if value!="original"{return fmt.Errorf("database value changed: %q",value)};return nil
 })
 p.Command("ping").Rule(sdk.ToMe()).Handle(func(c *sdk.Context) error {
  ev:=c.Event();if c.User()!="loop:alice" || c.Session()!="loop:group#topic" || ev.Actor.ExternalUserID!="alice" { return fmt.Errorf("trusted command context changed") };return observe("command")
 })
 service.Serve(p, service.Proto[*pb.PublishRequest,*pb.PublishResponse]("wakeprobe/probe")).Handle(func(_ *sdk.Context, in *pb.PublishRequest) (*pb.PublishResponse,error) {
  if err := observe("service"); err != nil { return nil,err }; return &pb.PublishResponse{EventId:in.EventName},nil
 })
 p.Route(http.MethodGet,"/probe").HandleFunc(func(w http.ResponseWriter,r *http.Request) {
  if err := observe("http"); err != nil { http.Error(w,err.Error(),500); return }; io.WriteString(w,"once")
 })
 return p
}
func main() { if err := sdk.Run(New()); err != nil { panic(err) } }
`
