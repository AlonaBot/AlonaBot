//go:build unix

package supervisor

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	coreconfig "github.com/AlonaBot/AlonaBot/internal/core/config"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/process"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// configPreparePlugin 显式返回自身配置，保证测试进入真实 typed 配置路径。
const configPreparePlugin = `package main
import (
 "context"
 "encoding/json"
 "fmt"
 "os"
 "github.com/AlonaBot/AlonaBot/sdk"
 "github.com/AlonaBot/AlonaBot/sdk/config"
)
type Child struct {
 Score int64 ` + "`json:\"score\" form:\"default=9007199254740993\"`" + `
 Note string ` + "`json:\"note\" form:\"default=\\\"child\\\"\"`" + `
}
type Config struct {
 config.Base
 Name string ` + "`json:\"name\" form:\"required\"`" + `
 Enabled bool ` + "`json:\"enabled\" form:\"default=true\"`" + `
 Count int ` + "`json:\"count\" form:\"default=7\"`" + `
 Text string ` + "`json:\"text\" form:\"default=\\\"hello\\\"\"`" + `
 Optional *string ` + "`json:\"optional\" form:\"default=\\\"fallback\\\"\"`" + `
 RequiredDefault string ` + "`json:\"required_default\" form:\"required;default=\\\"provided\\\"\"`" + `
 Nested Child ` + "`json:\"nested\"`" + `
 Rows []Child ` + "`json:\"rows\"`" + `
 Lookup map[string]Child ` + "`json:\"lookup\"`" + `
}
type Plugin struct {
 sdk.Base
 configuration Config
}
func (p *Plugin) Config() config.Config { return &p.configuration }
func (p *Plugin) OnStart(context.Context) error {
 // 文件记录 OnStart 实际读到的 typed 值，同时作为已经进入回调的证据。
 raw, err := json.Marshal(p.configuration)
 if err != nil { return err }
 return os.WriteFile("started.json", raw, 0600)
}
func main() {
 if err := sdk.Run(&Plugin{configuration:Config{Name:"preloaded",Count:99}}); err != nil {
  fmt.Fprintln(os.Stderr, err)
  os.Exit(1)
 }
}
`

const configPrepareOnLoad = `
func (c *Config) OnLoad(_ context.Context, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error) {
 if c.Name != "" || c.Count != 0 || c.Enabled { return nil, fmt.Errorf("OnLoad receiver must be zero") }
 var document map[string]json.RawMessage
 if err := json.Unmarshal(raw, &document); err != nil { return nil, err }
 if _, reject := document["reject"]; reject { return nil, fmt.Errorf("private OnLoad rejection value") }
 if _, change := document["change"]; change { document["name"] = json.RawMessage("\"changed\"") }
 if fromVersion == 1 {
  if old, present := document["legacy_name"]; present {
   document["name"] = old
   delete(document, "legacy_name")
  }
 }
 return json.Marshal(document)
}
`

type configPrepareChild struct {
	Score int64  `json:"score"`
	Note  string `json:"note"`
}
type configPrepareSnapshot struct {
	Name            string                        `json:"name"`
	Enabled         bool                          `json:"enabled"`
	Count           int                           `json:"count"`
	Text            string                        `json:"text"`
	Optional        *string                       `json:"optional"`
	RequiredDefault string                        `json:"required_default"`
	Nested          configPrepareChild            `json:"nested"`
	Rows            []configPrepareChild          `json:"rows"`
	Lookup          map[string]configPrepareChild `json:"lookup"`
}

// TestConfigPrepare 验证生成 Schema、真实 SDK 进程和 Core 的配置准备协议（V-37、V-21）。
func TestConfigPrepare(t *testing.T) {
	// 1. 经正式 CLI 生成 Schema，分别构建覆盖 OnLoad 和沿用 Base 的插件。
	_, file, _, _ := runtime.Caller(0)
	repository := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(file))))
	cli := filepath.Join(t.TempDir(), "alonabot")
	configPrepareCommand(t, repository, "go", "build", "-o", cli, "./cmd/alonabot")
	binaries := map[bool]string{}
	for _, migrate := range []bool{false, true} {
		dir := t.TempDir()
		module, err := os.ReadFile(filepath.Join(repository, "go.mod"))
		if err != nil {
			t.Fatal(err)
		}
		version := strings.Fields(string(module))[3] // go.mod 的 module 路径之后是 go 指令与版本。
		mod := fmt.Sprintf("module example.com/configprepare\n\ngo %s\nrequire github.com/AlonaBot/AlonaBot v0.0.0\nreplace github.com/AlonaBot/AlonaBot => %s\n", version, filepath.ToSlash(repository))
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0600); err != nil {
			t.Fatal(err)
		}
		source := configPreparePlugin
		if migrate {
			source += configPrepareOnLoad
		}
		if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		configPrepareCommand(t, dir, cli, "plugin", "generate")
		binary := filepath.Join(dir, "plugin")
		configPrepareCommand(t, dir, "go", "build", "-race", "-o", binary, ".")
		binaries[migrate] = binary
	}

	// 2. 每个场景通过私有 Bootstrap 驱动新进程，完整候选 JSON 与 typed 值分别核对。
	from := uint32(1)
	current := uint32(2)
	defaults := `"enabled":true,"count":7,"text":"hello","optional":"fallback","required_default":"provided"`
	for _, test := range []struct {
		name     string
		raw      string
		expected string
		base     bool
		from     *uint32
		resume   bool
		code     pb.ErrorCode
		path     string
	}{
		{name: "missing defaults", raw: `{"name":"demo"}`, expected: `{"name":"demo",` + defaults + `}`},
		{name: "explicit values and unknown keys", raw: `{"name":"demo","enabled":false,"count":0,"text":"","optional":null,"unknown":9007199254740995,"nested":{"extra":true}}`, expected: `{"name":"demo","enabled":false,"count":0,"text":"","optional":null,"required_default":"provided","unknown":9007199254740995,"nested":{"score":9007199254740993,"note":"child","extra":true}}`},
		{name: "migrate before schema", raw: `{"legacy_name":"migrated","unknown":"keep"}`, from: &from, expected: `{"name":"migrated","unknown":"keep",` + defaults + `}`},
		{name: "Base cannot migrate", raw: `{"legacy_name":"migrated","unknown":"keep"}`, base: true, from: &from, code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, path: `$["name"]`},
		{name: "first install uses target version", raw: `{"legacy_name":"old"}`, code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, path: `$["name"]`},
		{name: "required missing", raw: `{"unknown":"private-config-value"}`, code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, path: `$["name"]`},
		{name: "int overflow", raw: `{"name":"demo","count":9223372036854775808}`, code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, path: `$["count"]`},
		{name: "OnLoad rejects", raw: `{"name":"demo","reject":true}`, code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT},
		{name: "restart changes", raw: `{"name":"demo","change":true}`, resume: true, from: &current, code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION},
		{name: "restart unchanged", raw: `{"name":"demo","count":9007199254740993}`, resume: true, from: &current, expected: `{"name":"demo","enabled":true,"count":9007199254740993,"text":"hello","optional":"fallback","required_default":"provided"}`},
		{name: "large integer retains source", raw: `{"name":"demo","count":9007199254740995}`, expected: `{"name":"demo","enabled":true,"count":9007199254740995,"text":"hello","optional":"fallback","required_default":"provided"}`},
		{name: "array and map defaults", raw: `{"name":"demo","rows":[{}, {"score":0,"note":""}],"lookup":{"missing":{},"explicit":{"score":0,"note":""}}}`, expected: `{"name":"demo",` + defaults + `,"rows":[{"score":9007199254740993,"note":"child"},{"score":0,"note":""}],"lookup":{"missing":{"score":9007199254740993,"note":"child"},"explicit":{"score":0,"note":""}}}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			boot := &pb.Bootstrap{GenerationId: "gen-1", CandidateConfigJson: test.raw, TargetConfigVersion: 2, PreviousConfigVersion: test.from,
				DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE, ActivationReason: pb.ActivationReason_ACTIVATION_REASON_UPDATE}}
			if test.from == nil {
				boot.DatabaseDescriptor.ActivationReason = pb.ActivationReason_ACTIVATION_REASON_FIRST_INSTALL
			}
			if test.resume {
				boot.DatabaseDescriptor.ActivationReason = pb.ActivationReason_ACTIVATION_REASON_RESUME
			}
			dir := t.TempDir()
			registration, err := configPrepareProcess(t, binaries[!test.base], dir, boot)
			snapshotPath := filepath.Join(dir, "started.json")
			if test.code != pb.ErrorCode_ERROR_CODE_UNSPECIFIED {
				var phase *host.PhaseError
				if !errors.As(err, &phase) || phase.Phase != pb.SessionPhase_SESSION_PHASE_CONFIG || phase.Code != test.code {
					t.Fatalf("expected CONFIG/%s, got %v", test.code, err)
				}
				if registration != nil {
					t.Fatal("failed config left a registration")
				}
				if _, err := os.Stat(snapshotPath); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("OnStart ran after config failure: %v", err)
				}
				if test.path != "" && !strings.Contains(phase.Message, test.path) {
					t.Fatalf("error lacks JSON path %s: %s", test.path, phase.Message)
				}
				if test.name == "OnLoad rejects" && phase.Message != "config OnLoad rejected the configuration" {
					t.Fatalf("unexpected OnLoad safe_message: %q", phase.Message)
				}
				for _, private := range []string{"private-config-value", "private OnLoad rejection value", "9223372036854775808"} {
					if strings.Contains(err.Error(), private) {
						t.Fatalf("config value in error: %v", err)
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			expected, err := coreconfig.Canonical([]byte(test.expected))
			if err != nil {
				t.Fatal(err)
			}
			if registration.GetCandidateConfigJson() != string(expected) || registration.GetCandidateConfigVersion() != 2 {
				t.Fatalf("REGISTER candidate = %s version %d; want %s version 2", registration.GetCandidateConfigJson(), registration.GetCandidateConfigVersion(), expected)
			}
			if len(registration.GetConfigSchemaHash()) == 0 {
				t.Fatal("generated Config Schema was not reported")
			}
			actual, err := os.ReadFile(snapshotPath)
			if err != nil {
				t.Fatal(err)
			}
			var got, want configPrepareSnapshot
			if err := json.Unmarshal(actual, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(expected, &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("OnStart config = %+v, want %+v", got, want)
			}
		})
	}
}

func configPrepareCommand(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	command := exec.Command(name, args...)
	command.Dir = dir
	command.Env = append(os.Environ(), "GOFLAGS=-mod=mod", "GOWORK=off")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, output)
	}
}

// configPrepareProcess 使用真实 process/host 两端完成认证与准备，并在返回前确认进程退出。
func configPrepareProcess(t *testing.T, binary, dir string, boot *pb.Bootstrap) (*pb.Register, error) {
	t.Helper()
	secret, err := process.NewSecret()
	if err != nil {
		t.Fatal(err)
	}
	secretBytes, err := hex.DecodeString(secret)
	if err != nil {
		t.Fatal(err)
	}
	digest := []byte("config preparation integration plugin")
	logs := &logBuf{}
	child, err := process.Start(process.Spec{PluginID: "demo", Path: binary, Dir: dir, Stdio: true, Env: map[string]string{
		wire.EnvPluginID: "demo", wire.EnvSecret: secret, wire.EnvBuildDigest: hex.EncodeToString(digest), wire.EnvTransport: wire.TransportStdio, wire.EnvRuntimeLock: filepath.Join(dir, wire.RuntimeLockFile),
		"GORACE": "atexit_sleep_ms=0", // 停止预算验证退出本身，跳过 race 运行时的退出后等待。
	}}, logs.Line)
	if err != nil {
		t.Fatal(err)
	}
	conn := wire.NewConn(child.Stdout, child.Stdin, child.Stdout)
	defer func() {
		_ = conn.Close()
		empty, err := child.WaitTree(time.Now())
		if err != nil {
			t.Errorf("query config plugin process tree: %v", err)
		} else if !empty {
			if err := child.KillTree(); err != nil {
				t.Errorf("kill config plugin process tree: %v", err)
			}
		}
		if err := child.Release(); err != nil {
			t.Errorf("release config plugin: %v", err)
		}
		child.WaitLogs()
		if t.Failed() {
			t.Logf("config plugin logs:\n%s", logs.String())
		}
	}()
	ctx := context.Background()
	prepared, err := host.Prepare(ctx, conn, host.Expect{Role: state.RolePlugin, PluginID: "demo", Secret: secretBytes, BuildDigest: digest, Offer: validate.Offer{Major: validate.ProtocolMajor, Minor: validate.ProtocolMinor}}, boot, fastLimits)
	if err != nil {
		// 先确认失败进程已经退出，再由调用方断言 OnStart 没有留下文件。
		select {
		case <-child.Exited():
			if exit := child.ExitState(); exit == nil || exit.ExitCode() != 1 {
				t.Fatalf("config plugin failure exit: %v", exit)
			}
		case <-time.After(fastLimits.OnStart):
			t.Fatal("config plugin did not exit after preparation failure")
		}
		return nil, err
	}
	instance, err := host.Activate(ctx, conn, prepared, host.ActivateOptions{PluginID: "demo", GenerationID: boot.GetGenerationId(), Limits: fastLimits})
	if err != nil {
		t.Fatalf("activate config plugin: %v", err)
	}
	stopped := instance.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE, child, fastStop)
	if stopped.Err != nil || stopped.Killed {
		t.Fatalf("stop config plugin: %+v", stopped)
	}
	if exit := child.ExitState(); exit == nil || !exit.Success() {
		t.Fatalf("config plugin exit: %v", exit)
	}
	return prepared.Register, nil
}
