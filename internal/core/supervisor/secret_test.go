//go:build unix

package supervisor

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/secret"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// secretPluginSource 在真实 SDK 的 stdio 两端观察控制帧，只在指定故障中修改 REGISTER。
// 传输层故障注入更改 REGISTER，保留实际 SDK 的认证、注册与激活流程。
const secretPluginSource = `package main
import (
 "encoding/hex"
 "encoding/json"
 "fmt"
 "os"
 pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
 "github.com/AlonaBot/AlonaBot/contracts/validate"
 "github.com/AlonaBot/AlonaBot/internal/wire"
 "github.com/AlonaBot/AlonaBot/sdk"
 "github.com/AlonaBot/AlonaBot/sdk/config"
 "google.golang.org/protobuf/proto"
)
type Plugin struct { sdk.Base; settings Config }
func(p *Plugin) Config() config.Config { return &p.settings }
type Provider struct { APIKey string ` + "`json:\"api_key\" secret:\"true\"`" + ` }
type Config struct {
 config.Base
 Token string ` + "`json:\"token\" secret:\"true\"`" + `
 Ordinary string ` + "`json:\"ordinary\"`" + `
 Providers []Provider ` + "`json:\"providers\"`" + `
}
func observe(boot *pb.Bootstrap) error {
 var document map[string]json.RawMessage
 if err := json.Unmarshal([]byte(boot.GetCandidateConfigJson()),&document); err != nil { return fmt.Errorf("decode Bootstrap probe failed") }
 mode := os.Getenv("SECRET_TEST_MODE")
 if mode == "array" || mode == "swapped" {
  var providers []Provider
  if err := json.Unmarshal(document["providers"],&providers); err != nil || len(providers) != 2 { return fmt.Errorf("provider probe failed") }
  first,second := "array-secret-one","array-secret-two"
  if mode == "swapped" { first,second = second,first }
  if providers[0].APIKey != first || providers[1].APIKey != second { return fmt.Errorf("provider plaintext mismatch") }
 } else if mode == "deleted" {
  if _,found := document["token"]; found { return fmt.Errorf("deleted token received") }
 } else {
  want := "integration-secret-one"
  if mode == "two" || mode == "tamper" { want = "integration-secret-two" }
  if mode == "abandoned" { want = "integration-secret-abandoned" }
  var token string
  if err := json.Unmarshal(document["token"],&token); err != nil || token != want { return fmt.Errorf("token plaintext mismatch") }
 }
 fmt.Fprintln(os.Stderr,"bootstrap-plaintext-ok")
 return nil
}
func main() {
 input,output := os.Stdin,os.Stdout
 sdkInput,inputWriter,err := os.Pipe(); if err != nil { panic(err) }
 outputReader,sdkOutput,err := os.Pipe(); if err != nil { panic(err) }
 os.Stdin,os.Stdout = sdkInput,sdkOutput
 core := wire.NewConn(input,output,input)
 plugin := wire.NewConn(outputReader,inputWriter,outputReader)
 // Core→SDK：核对明文输入与实际线上 PREPARED 摘要，原帧交给 SDK 处理。
 go func(){
  defer inputWriter.Close()
  for {
   frame,err := core.ReadFrame(); if err != nil { return }
   message,err := validate.DecodeControl(frame); if err != nil { return }
   switch value := message.(type) {
   case *pb.HelloAck:
    if err := observe(value.GetBootstrap()); err != nil { fmt.Fprintln(os.Stderr,err); return }
   case *pb.Prepared:
    fmt.Fprintln(os.Stderr,"prepared-digest",hex.EncodeToString(value.GetCandidateConfigDigest()))
   }
   if err := plugin.WriteFrame(frame); err != nil { return }
  }
 }()
 // SDK→Core：真实 SDK 构造完整 REGISTER；tamper 模式只替换候选中的一个秘密值。
 go func(){
  defer output.Close()
  for {
   frame,err := plugin.ReadFrame(); if err != nil { return }
   if frame.GetKind() == pb.FrameKind_FRAME_KIND_REGISTER && os.Getenv("SECRET_TEST_MODE") == "tamper" {
    message,err := validate.DecodeControl(frame); if err != nil { return }
    registration := message.(*pb.Register)
    var document map[string]json.RawMessage
    if err := json.Unmarshal([]byte(registration.GetCandidateConfigJson()),&document); err != nil { return }
    document["token"] = json.RawMessage(` + "`\"integration-secret-tampered\"`" + `)
    candidate,err := json.Marshal(document); if err != nil { return }
    registration.CandidateConfigJson = string(candidate)
    frame.Payload,err = proto.Marshal(registration); if err != nil { return }
   }
   if err := core.WriteFrame(frame); err != nil { return }
  }
 }()
 if err := sdk.Run(&Plugin{}); err != nil { fmt.Fprintln(os.Stderr,err); os.Exit(1) }
}
`

const secretGeneratorSource = `package main
import (
 "context"
 "fmt"
 "os"
 "path/filepath"
 "github.com/AlonaBot/AlonaBot/sdk/generate"
)
func main() {
 result,err := generate.Scan(context.Background(),generate.Options{Dir:os.Args[1]})
 if err != nil { fmt.Fprintln(os.Stderr,err); os.Exit(1) }
 if len(result.Diagnostics) != 0 { fmt.Fprintln(os.Stderr,result.Diagnostics); os.Exit(1) }
 if err := os.WriteFile(filepath.Join(os.Args[1],generate.FileName),result.Source,0600); err != nil { panic(err) }
 if err := os.WriteFile(filepath.Join(os.Args[1],"schema.json"),result.ConfigSchema,0600); err != nil { panic(err) }
}
`

// buildSecretPlugins 将源码交给现有生成器；Core 测试通过进程使用 SDK，保持静态依赖方向。
func buildSecretPlugins(t *testing.T) (binary string, schema []byte, withoutBinary string, withoutSchema []byte) {
	t.Helper()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	mod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	module := strings.Replace(string(mod), "module github.com/AlonaBot/AlonaBot", "module github.com/AlonaBot/AlonaBot/secretintegration", 1)
	module += "\nrequire github.com/AlonaBot/AlonaBot v0.0.0\nreplace github.com/AlonaBot/AlonaBot => " + strconv.Quote(filepath.ToSlash(root)) + "\n"
	write := func(name, contents string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", module)
	sums, err := os.ReadFile(filepath.Join(root, "go.sum"))
	if err != nil {
		t.Fatal(err)
	}
	write("go.sum", string(sums))
	write("generator/main.go", secretGeneratorSource)
	write("original/main.go", secretPluginSource)
	without := strings.Replace(secretPluginSource, " Token string `json:\"token\" secret:\"true\"`\n", "", 1)
	if without == secretPluginSource {
		t.Fatal("secret field was not removed from fixture")
	}
	write("without/main.go", without)
	run := func(args ...string) {
		t.Helper()
		command := exec.Command("go", args...)
		command.Dir = dir
		command.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=mod")
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("go %v: %v\n%s", args, err, output)
		}
	}
	var binaries []string
	var schemas [][]byte
	for _, variant := range []string{"original", "without"} {
		pluginDir := filepath.Join(dir, variant)
		run("run", "./generator", pluginDir)
		binary := filepath.Join(dir, variant+"-plugin")
		run("build", "-o", binary, "./"+variant)
		raw, err := os.ReadFile(filepath.Join(pluginDir, "schema.json"))
		if err != nil {
			t.Fatal(err)
		}
		binaries = append(binaries, binary)
		schemas = append(schemas, raw)
	}
	return binaries[0], schemas[0], binaries[1], schemas[1]
}

func secretSnapshot(t *testing.T, store *state.Store, generation string) []byte {
	t.Helper()
	snapshot, err := store.GenerationConfig(t.Context(), generation)
	if err != nil {
		t.Fatal(err)
	}
	return []byte(snapshot)
}

func secretReference(t *testing.T, snapshot []byte, path ...any) string {
	t.Helper()
	var value any
	if err := json.Unmarshal(snapshot, &value); err != nil {
		t.Fatal(err)
	}
	for _, part := range path {
		switch part := part.(type) {
		case string:
			value = value.(map[string]any)[part]
		case int:
			value = value.([]any)[part]
		}
	}
	id, ok := value.(map[string]any)[secret.RefKey].(string)
	if !ok {
		t.Fatal("snapshot field is not a secret reference")
	}
	decoded, err := hex.DecodeString(id)
	if err != nil || len(decoded) != 16 {
		t.Fatal("reference is not a 16-byte hexadecimal value")
	}
	return id
}

// TestSecretStoreProcesses 验证真实生成器、SDK 插件和 SQLite 协作时的不可变快照与清理（V-25）。
func TestSecretStoreProcesses(t *testing.T) {
	binary, schema, withoutBinary, withoutSchema := buildSecretPlugins(t)
	t.Run("generations", func(t *testing.T) {
		store, spec := setup(t, binary)
		secrets := secret.New(store)
		logs := &logBuf{}
		spec.Secrets, spec.ConfigSchema, spec.Stdio, spec.Log = secrets, schema, true, logs
		start := func(current Spec) error {
			t.Helper()
			before := len(logs.String())
			running, err := Launch(t.Context(), store, current)
			if err != nil {
				return err
			}
			stopped := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
			running.Process.WaitLogs()
			if stopped.Err != nil || stopped.Killed {
				t.Fatalf("secret plugin stop failed: %+v", stopped)
			}
			trace := logs.String()[before:]
			if strings.Count(trace, "bootstrap-plaintext-ok") != 1 {
				t.Fatal("plugin did not observe expected plaintext Bootstrap")
			}
			snapshot := secretSnapshot(t, store, current.GenerationID)
			digest := sha256.Sum256(snapshot)
			if !strings.Contains(trace, "prepared-digest "+hex.EncodeToString(digest[:])) {
				t.Fatal("PREPARED digest differs from persisted snapshot")
			}
			return nil
		}
		candidate := func(generation, previous string, executable string, description []byte, config, mode string) Spec {
			t.Helper()
			next := spec
			next.GenerationID, next.ExpectedOld, next.Executable, next.ConfigSchema = generation, previous, executable, description
			next.Bootstrap = &pb.Bootstrap{GenerationId: generation, CandidateConfigJson: config, TargetConfigVersion: 1}
			next.ExtraEnv = map[string]string{"SECRET_TEST_MODE": mode}
			if generation != "gen-1" {
				if err := store.AddGeneration(t.Context(), "demo", generation, `{}`, 1); err != nil {
					t.Fatal(err)
				}
			}
			return next
		}
		firstSpec := candidate("gen-1", "", binary, schema, `{"ordinary":"public","token":"integration-secret-one"}`, "one")
		if err := start(firstSpec); err != nil {
			t.Fatal(err)
		}
		first := secretSnapshot(t, store, "gen-1")
		firstRef := secretReference(t, first, "token")
		rows, err := store.ListSecrets(t.Context(), "demo")
		if err != nil || len(rows) != 1 {
			t.Fatalf("first secret count = %d, %v", len(rows), err)
		}
		if bytes.Contains(first, []byte("integration-secret-one")) {
			t.Fatal("snapshot leaked plaintext")
		}
		if err := start(candidate("gen-2", "gen-1", binary, schema, string(first), "one")); err != nil {
			t.Fatal(err)
		}
		second := secretSnapshot(t, store, "gen-2")
		if secretReference(t, second, "token") != firstRef {
			t.Fatal("unchanged value did not reuse reference")
		}
		if err := start(candidate("gen-3", "gen-2", binary, schema, `{"ordinary":"public","token":"integration-secret-two"}`, "two")); err != nil {
			t.Fatal(err)
		}
		third := secretSnapshot(t, store, "gen-3")
		if secretReference(t, third, "token") == firstRef || !bytes.Equal(first, secretSnapshot(t, store, "gen-1")) {
			t.Fatal("new value overwrote old secret version")
		}
		oldSecret, err := store.GetSecret(t.Context(), "demo", firstRef)
		if err != nil || string(oldSecret.Value) != "integration-secret-one" {
			t.Fatal("old secret value changed after new candidate")
		}
		fourthSpec := candidate("gen-4", "gen-3", withoutBinary, withoutSchema, string(third), "two")
		if err := start(fourthSpec); err != nil {
			t.Fatal(err)
		}
		fourth := secretSnapshot(t, store, "gen-4")
		if !bytes.Equal(third, fourth) {
			t.Fatal("schema removal lost unknown secret protection")
		}
		restart := fourthSpec
		restart.NewCandidate = false
		if err := start(restart); err != nil {
			t.Fatal(err)
		}
		restart.ExtraEnv = map[string]string{"SECRET_TEST_MODE": "tamper"}
		var phase *host.PhaseError
		if err := start(restart); !errors.As(err, &phase) || phase.Code != pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION {
			t.Fatalf("changed restart error = %v", err)
		}
		if active, err := store.Active(t.Context(), "demo"); err != nil || active != "gen-4" || !bytes.Equal(fourth, secretSnapshot(t, store, "gen-4")) {
			t.Fatal("failed restart changed Active")
		}
		abandonedSpec := candidate("abandoned", "gen-1", binary, schema, `{"token":"integration-secret-abandoned"}`, "abandoned")
		if err := start(abandonedSpec); !errors.Is(err, state.ErrStaleGeneration) {
			t.Fatalf("stale candidate error = %v", err)
		}
		abandoned := secretReference(t, secretSnapshot(t, store, "abandoned"), "token")
		removed, err := secrets.Collect(t.Context(), "demo", [][]byte{first, second, third, fourth})
		if err != nil || removed != 2 {
			t.Fatalf("collect discarded versions = %d, %v", removed, err)
		}
		if _, err := store.GetSecret(t.Context(), "demo", abandoned); !errors.Is(err, state.ErrNotFound) {
			t.Fatal("discarded candidate reference survived collection")
		}
		if _, err := secrets.Resolve(t.Context(), "demo", fourth); err != nil {
			t.Fatalf("Active secret was collected: %v", err)
		}
		if err := start(candidate("gen-5", "gen-4", withoutBinary, withoutSchema, `{"ordinary":"public"}`, "deleted")); err != nil {
			t.Fatal(err)
		}
		paths, err := store.GenerationSecretPaths(t.Context(), "gen-5")
		if err != nil || len(paths) != 0 {
			t.Fatalf("deleted field kept secret paths: %v %v", paths, err)
		}
		for _, plaintext := range []string{"integration-secret-one", "integration-secret-two", "integration-secret-tampered", "integration-secret-abandoned"} {
			if strings.Contains(logs.String(), plaintext) {
				t.Fatal("logs contain secret plaintext")
			}
		}
	})
	t.Run("array-positions", func(t *testing.T) {
		store, spec := setup(t, binary)
		logs := &logBuf{}
		spec.ConfigSchema, spec.Secrets, spec.Stdio, spec.Log = schema, secret.New(store), true, logs
		launch := func(mode string) {
			t.Helper()
			spec.ExtraEnv = map[string]string{"SECRET_TEST_MODE": mode}
			running, err := Launch(t.Context(), store, spec)
			if err != nil {
				t.Fatal(err)
			}
			result := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
			running.Process.WaitLogs()
			if result.Err != nil || result.Killed {
				t.Fatalf("array plugin stop: %+v", result)
			}
		}
		spec.Bootstrap.CandidateConfigJson = `{"providers":[{"api_key":"array-secret-one"},{"api_key":"array-secret-two"}]}`
		launch("array")
		first := secretSnapshot(t, store, "gen-1")
		refs := []string{secretReference(t, first, "providers", 0, "api_key"), secretReference(t, first, "providers", 1, "api_key")}
		if refs[0] == refs[1] {
			t.Fatal("array elements share a reference")
		}
		spec.GenerationID, spec.ExpectedOld = "gen-2", "gen-1"
		spec.Bootstrap = &pb.Bootstrap{GenerationId: "gen-2", CandidateConfigJson: `{"providers":[{"api_key":"array-secret-two"},{"api_key":"array-secret-one"}]}`, TargetConfigVersion: 1}
		if err := store.AddGeneration(t.Context(), "demo", "gen-2", `{}`, 1); err != nil {
			t.Fatal(err)
		}
		launch("swapped")
		second := secretSnapshot(t, store, "gen-2")
		for i, ref := range refs {
			if secretReference(t, second, "providers", i, "api_key") == ref {
				t.Fatal("swapped value reused reference from same position")
			}
		}
		rows, err := store.ListSecrets(t.Context(), "demo")
		if err != nil || len(rows) != 4 || !bytes.Equal(first, secretSnapshot(t, store, "gen-1")) {
			t.Fatal("array secret versions are not immutable")
		}
		if strings.Contains(logs.String(), "array-secret-one") || strings.Contains(logs.String(), "array-secret-two") {
			t.Fatal("array secret leaked to logs")
		}
	})
}
