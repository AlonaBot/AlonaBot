package generate_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk/config"
	"github.com/AlonaBot/AlonaBot/sdk/generate"
	"github.com/google/jsonschema-go/jsonschema"
	"google.golang.org/protobuf/encoding/protojson"
)

const configPluginMain = `package main
import (
 "fmt"
 "os"
 "github.com/AlonaBot/AlonaBot/sdk"
 "github.com/AlonaBot/AlonaBot/sdk/config"
)
type Plugin struct { sdk.Base }
func main() {
 if err := sdk.Run(&Plugin{}); err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
}
`

const configDeclarations = "type Provider struct {\n" +
	" APIKey *string `json:\"api_key\" config:\"secret;hidden;widget:password;default:null\"`\n" +
	"}\n" +
	"type Config struct {\n" +
	" config.Base\n" +
	" Name string `json:\"name,omitempty\" config:\"label:Provider;description:Provider to use;widget:select;options:openai|custom;labels:OpenAI|Custom;subtitles:Hosted service;required;default:\\\"openai\\\"\"`\n" +
	" Enabled bool `json:\"enabled\" config:\"widget:switch;default:true\"`\n" +
	" Count uint64 `json:\"count\" config:\"widget:slider;min:1;max:10;step:1;options:1|10;default:10\"`\n" +
	" Nested Provider `json:\"nested\"`\n" +
	" Providers []Provider `json:\"providers\" config:\"default:[{\\\"api_key\\\":\\\"x\\\"}]\"`\n" +
	" Lookup map[string]Provider `json:\"lookup\" config:\"default:{\\\"first\\\":{\\\"api_key\\\":null}}\"`\n" +
	" Maybe *int16 `json:\"maybe,omitempty\" config:\"default:null\"`\n" +
	" Note string `json:\"note\" config:\"description:Advanced settings;widget:textarea;obvious;when:name=custom;default:\\\"hello;world:ok\\\"\"`\n" +
	" Modes []string `json:\"modes\" config:\"widget:select;options:fast|safe@literal;labels:Fast|Safe;subtitles:Quick response|Extra checks;default:[\\\"fast\\\"]\"`\n" +
	" Ports []uint16 `json:\"ports\" config:\"options:80|443;default:[80,443]\"`\n" +
	" Script string `json:\"script\" config:\"widget:code;when:enabled=true,count=10\"`\n" +
	" Query string `json:\"query\" config:\"widget:code;language:sql\"`\n" +
	"}\n"

// configCoreMain 在独立临时 module 中装配真实 Core，使 SDK 测试通过进程边界观察 REGISTER。
const configCoreMain = `package main
import (
 "bytes"
 "context"
 "encoding/hex"
 "errors"
 "fmt"
 "os"
 "path/filepath"
 "time"
 pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
 "github.com/AlonaBot/AlonaBot/contracts/validate"
 "github.com/AlonaBot/AlonaBot/internal/core/host"
 "github.com/AlonaBot/AlonaBot/internal/core/process"
 "github.com/AlonaBot/AlonaBot/internal/core/state"
 "github.com/AlonaBot/AlonaBot/internal/tune"
 "github.com/AlonaBot/AlonaBot/internal/wire"
 "google.golang.org/protobuf/encoding/protojson"
)
func run() (runErr error) {
 // 1. 启动真实 SDK 插件，stdio 只承载协议帧，日志转到辅助程序的 stderr。
 secret, err := process.NewSecret()
 if err != nil { return err }
 secretBytes, err := hex.DecodeString(secret)
 if err != nil { return err }
 digest := bytes.Repeat([]byte{1}, 32)
 pluginDir := filepath.Dir(os.Args[1])
 child, err := process.Start(process.Spec{
  PluginID: "schema-test", Path: os.Args[1], Dir: pluginDir, Stdio: true,
  Env: map[string]string{
   wire.EnvPluginID: "schema-test", wire.EnvSecret: secret,
   wire.EnvBuildDigest: hex.EncodeToString(digest), wire.EnvTransport: wire.TransportStdio,
   wire.EnvRuntimeLock: filepath.Join(pluginDir, wire.RuntimeLockFile),
  },
 }, func(id, stream string, line []byte) { fmt.Fprintf(os.Stderr, "%s|%s|%s\n", id, stream, line) })
 if err != nil { return err }
 // 失败时回收仍存活的进程树，正常退出后只释放资源。
 defer func() {
  empty, cleanupErr := child.WaitTree(time.Now())
  if cleanupErr == nil && !empty { cleanupErr = child.KillTree() }
  runErr = errors.Join(runErr, cleanupErr, child.Release())
 }()
 conn := wire.NewConn(child.Stdout, child.Stdin, child.Stdout)
 defer conn.Close()

 // 2. Core 完成认证、准备与激活，取得已经过正式校验的 REGISTER。
 ctx := context.Background()
 values := tune.Defaults()
 limits := host.NewLimits(values)
 boot := &pb.Bootstrap{GenerationId: "gen-1", CandidateConfigJson: os.Args[2], TargetConfigVersion: 7,
  DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_NONE}}
 prepared, err := host.Prepare(ctx, conn, host.Expect{
  Role: state.RolePlugin, PluginID: "schema-test", Secret: secretBytes, BuildDigest: digest,
  Offer: validate.Offer{Major: validate.ProtocolMajor, Minor: validate.ProtocolMinor},
 }, boot, limits)
 if err != nil { return err }
 instance, err := host.Activate(ctx, conn, prepared, host.ActivateOptions{PluginID: "schema-test", GenerationID: "gen-1", Limits: limits})
 if err != nil { return err }

 // 3. 正常停止并确认插件成功退出，再打印完整注册载荷供测试断言。
 stopped := instance.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE, child, host.NewStopTiming(values))
 if stopped.Err != nil { return stopped.Err }
 if stopped.Killed || !child.ExitState().Success() { return fmt.Errorf("plugin did not stop cleanly: %+v; exit=%v", stopped, child.ExitState()) }
 child.WaitLogs()
 encoded, err := protojson.Marshal(prepared.Register)
 if err != nil { return err }
 _, err = fmt.Fprintln(os.Stdout, string(encoded))
 return err
}
func main() {
 if err := run(); err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
}
`

func TestConfigSchemaIntegration(t *testing.T) {
	// 1. 从真实插件源码生成完整 Schema，逐字节核对声明顺序和元数据。
	dir := pluginModuleDir(t)
	source := configPluginMain + configDeclarations
	writePluginFile(t, dir, "main.go", source)
	result := scanPlugin(t, generate.Options{Dir: dir})
	if len(result.Diagnostics) != 0 {
		t.Fatal(result.Diagnostics)
	}
	const expected = `{
 "$schema":"https://json-schema.org/draft/2020-12/schema",
 "properties":{
  "name":{"default":"openai","description":"Provider to use","type":"string","x-alonabot-form":{"label":"Provider","labels":["OpenAI","Custom"],"options":["openai","custom"],"subtitles":["Hosted service"],"widget":"select"}},
  "enabled":{"default":true,"type":"boolean","x-alonabot-form":{"widget":"switch"}},
  "count":{"default":10,"maximum":18446744073709551615,"minimum":0,"type":"integer","x-alonabot-form":{"max":10,"min":1,"options":[1,10],"step":1,"widget":"slider"}},
  "nested":{"properties":{"api_key":{"default":null,"type":["string","null"],"x-alonabot-form":{"widget":"password"},"x-alonabot-hidden":true,"x-alonabot-secret":true}},"type":"object"},
  "providers":{"default":[{"api_key":"x"}],"items":{"properties":{"api_key":{"default":null,"type":["string","null"],"x-alonabot-form":{"widget":"password"},"x-alonabot-hidden":true,"x-alonabot-secret":true}},"type":"object"},"type":"array"},
  "lookup":{"additionalProperties":{"properties":{"api_key":{"default":null,"type":["string","null"],"x-alonabot-form":{"widget":"password"},"x-alonabot-hidden":true,"x-alonabot-secret":true}},"type":"object"},"default":{"first":{"api_key":null}},"type":"object"},
  "maybe":{"default":null,"maximum":32767,"minimum":-32768,"type":["integer","null"]},
  "note":{"default":"hello;world:ok","description":"Advanced settings","type":"string","x-alonabot-form":{"obvious":true,"when":{"name":"custom"},"widget":"textarea"}},
  "modes":{"default":["fast"],"items":{"type":"string"},"type":"array","x-alonabot-form":{"labels":["Fast","Safe"],"options":["fast","safe@literal"],"subtitles":["Quick response","Extra checks"],"widget":"select"}},
  "ports":{"default":[80,443],"items":{"maximum":65535,"minimum":0,"type":"integer"},"type":"array","x-alonabot-form":{"options":[80,443]}},
  "script":{"type":"string","x-alonabot-form":{"language":"json","when":{"count":10,"enabled":true},"widget":"code"}},
  "query":{"type":"string","x-alonabot-form":{"language":"sql","widget":"code"}}
 },
 "required":["name"],
 "type":"object"
}`
	var want bytes.Buffer
	if err := json.Compact(&want, []byte(expected)); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(result.ConfigSchema, want.Bytes()) {
		t.Fatalf("config schema:\n got %s\nwant %s", result.ConfigSchema, &want)
	}
	var schema jsonschema.Schema
	if err := json.Unmarshal(result.ConfigSchema, &schema); err != nil {
		t.Fatal(err)
	}
	resolved, err := schema.Resolve(&jsonschema.ResolveOptions{ValidateDefaults: true})
	if err != nil {
		t.Fatal(err)
	}
	const candidate = `{"name":"custom","unknown":true,"nested":{"extra":1},"providers":[{"extra":2}],"lookup":{"first":{"extra":3}}}`
	const wantCandidate = `{"lookup":{"first":{"extra":3}},"name":"custom","nested":{"extra":1},"providers":[{"extra":2}],"unknown":true}`
	var value any
	if err := json.Unmarshal([]byte(candidate), &value); err != nil {
		t.Fatal(err)
	}
	if err := resolved.Validate(value); err != nil {
		t.Fatalf("schema rejected unknown config keys: %v", err)
	}
	writePluginFile(t, dir, generate.FileName, string(result.Source))
	again := scanPlugin(t, generate.Options{Dir: dir})
	if len(again.Diagnostics) != 0 || !bytes.Equal(result.ConfigSchema, again.ConfigSchema) || !bytes.Equal(result.Source, again.Source) {
		t.Fatalf("unstable schema/source: %+v", again)
	}

	// 2. 辅助程序经真实 Core 与插件 Session 取得 REGISTER，覆盖有配置与无配置。
	coreDir := pluginModuleDir(t)
	module, err := os.ReadFile(filepath.Join(coreDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	writePluginFile(t, coreDir, "go.mod", strings.Replace(string(module), "module "+pluginModule, "module github.com/AlonaBot/AlonaBot/configtest", 1))
	writePluginFile(t, coreDir, "main.go", configCoreMain)
	core := filepath.Join(coreDir, "core")
	plugin := filepath.Join(dir, "plugin")
	if runtime.GOOS == "windows" {
		core += ".exe"
		plugin += ".exe"
	}
	runPluginCommand(t, coreDir, nil, "go", "build", "-o", core, ".")
	checkRegister := func(schema []byte) {
		t.Helper()
		runPluginCommand(t, dir, nil, "go", "build", "-o", plugin, ".")
		command := exec.Command(core, plugin, candidate)
		command.Dir = coreDir
		var stderr bytes.Buffer
		command.Stderr = &stderr
		output, err := command.Output()
		if err != nil {
			t.Fatalf("Core/plugin Session: %v\nstdout=%s\nstderr=%s", err, output, &stderr)
		}
		var registration pb.Register
		if err := protojson.Unmarshal(output, &registration); err != nil {
			t.Fatalf("decode REGISTER: %v\n%s\n%s", err, output, &stderr)
		}
		var hash []byte
		if len(schema) != 0 {
			hash = config.SchemaHash(schema)
		}
		if !bytes.Equal(registration.ConfigSchemaHash, hash) || registration.CandidateConfigJson != wantCandidate || registration.CandidateConfigVersion != 7 || registration.Migration.GetMode() != pb.MigrationMode_MIGRATION_MODE_NONE {
			t.Fatalf("REGISTER = %s, want hash %x, canonical candidate and version 7", output, hash)
		}
	}
	checkRegister(result.ConfigSchema)

	// 3. 只改 Schema 标签就应触发 CLI 的过期检查，重新生成后恢复。
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Dir(filepath.Dir(filepath.Dir(file)))
	cli := filepath.Join(t.TempDir(), "alonabot")
	if runtime.GOOS == "windows" {
		cli += ".exe"
	}
	runPluginCommand(t, root, nil, "go", "build", "-o", cli, "./cmd/alonabot")
	runPluginCommand(t, dir, nil, cli, "plugin", "check")
	writePluginFile(t, dir, "main.go", strings.Replace(source, "label:Provider", "label:Changed provider", 1))
	command := exec.Command(cli, "plugin", "check")
	command.Dir = dir
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	checkErr := command.Run()
	exit, ok := checkErr.(*exec.ExitError)
	if !ok || exit.ExitCode() != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "alonabot_gen.go is out of date; run alonabot plugin generate") {
		t.Fatalf("schema change check: exit=%v stdout=%s stderr=%s", checkErr, &stdout, &stderr)
	}
	preserved, err := os.ReadFile(filepath.Join(dir, generate.FileName))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(preserved, result.Source) {
		t.Fatal("plugin check changed the generated file")
	}
	runPluginCommand(t, dir, nil, cli, "plugin", "generate")
	runPluginCommand(t, dir, nil, cli, "plugin", "check")
	changed := scanPlugin(t, generate.Options{Dir: dir})
	if len(changed.Diagnostics) != 0 || bytes.Equal(changed.ConfigSchema, result.ConfigSchema) || bytes.Equal(changed.Source, result.Source) {
		t.Fatalf("schema change was not generated: %+v", changed)
	}
	checkRegister(changed.ConfigSchema)

	// 4. 无 Config 的生成程序同样走完整 Session，REGISTER 摘要为空。
	writePluginFile(t, dir, "main.go", configPluginMain+"var _ config.Base\n")
	unconfigured := scanPlugin(t, generate.Options{Dir: dir})
	if len(unconfigured.Diagnostics) != 0 || unconfigured.ConfigSchema != nil {
		t.Fatalf("unconfigured schema: %+v", unconfigured)
	}
	writePluginFile(t, dir, generate.FileName, string(unconfigured.Source))
	checkRegister(nil)
}

func TestConfigDiagnosticsIntegration(t *testing.T) {
	dir := pluginModuleDir(t)
	writePluginFile(t, dir, "types.go", `package main
type JSONRead string
func (*JSONRead) UnmarshalJSON([]byte) error { return nil }
type JSONWrite string
func (JSONWrite) MarshalJSON() ([]byte, error) { return nil, nil }
type TextRead string
func (*TextRead) UnmarshalText([]byte) error { return nil }
type TextWrite string
func (TextWrite) MarshalText() ([]byte, error) { return nil, nil }
type Recursive struct { Next *Recursive }
type private struct { Value string }
type Nested struct { *private }
`)
	tests := []struct{ name, field, tag, message string }{
		{"Channel", "chan bool", "", "unsupported config type chan bool"},
		{"Interface", "any", "", "unsupported config type interface{}"},
		{"Function", "func()", "", "unsupported config type func()"},
		{"Complex", "complex64", "", "unsupported config type complex64"},
		{"MapKey", "map[int]string", "", "config map key must be string"},
		{"Time", "time.Time", "", "implements encoding/json.Marshaler"},
		{"JSONRead", "JSONRead", "", "implements encoding/json.Unmarshaler"},
		{"JSONWrite", "JSONWrite", "", "implements encoding/json.Marshaler"},
		{"TextRead", "TextRead", "", "implements encoding.TextUnmarshaler"},
		{"TextWrite", "TextWrite", "", "implements encoding.TextMarshaler"},
		{"Recursive", "*Recursive", "", "recursive config type"},
		{"WrongDefault", "int", `config:"default:\"x\""`, "want int, got string"},
		{"Overflow", "uint64", `config:"default:18446744073709551616"`, "value out of range"},
		{"UnknownConfig", "string", `config:"unknown:value"`, `unknown config key "unknown"`},
		{"FlagValue", "string", `config:"required:true"`, "config required must be a flag"},
		{"MissingValue", "string", `config:"widget"`, "config widget requires a value"},
		{"EmptyValue", "string", `config:"label:"`, "config label requires a value"},
		{"DuplicateKey", "string", `config:"label:First;label:Second"`, `duplicate config key "label"`},
		{"DuplicateFlag", "string", `config:"hidden;hidden"`, `duplicate config key "hidden"`},
		{"DefaultDelimiter", "string", `config:"default:\"x\"label:Title"`, "expected semicolon after config default"},
		{"Duplicate", "string", `json:"duplicate"`, `duplicate config JSON name "duplicate"`},
		{"Choice", "string", `config:"options:a|b;default:\"c\""`, "not in form options"},
		{"Minimum", "int8", `config:"min:2;default:1"`, "below form min"},
		{"Maximum", "int8", `config:"max:2;default:3"`, "exceeds form max"},
		{"Inverted", "int8", `config:"min:2;max:1"`, "form min exceeds max"},
		{"Options", "bool", `config:"options:true|false"`, "options require a string or numeric field"},
		{"InvalidOption", "uint16", `config:"options:bad"`, `invalid form option "bad"`},
		{"Labels", "string", `config:"options:a;labels:A|B"`, "config labels has more entries than options"},
		{"Subtitles", "string", `config:"options:a;subtitles:First|Second"`, "config subtitles has more entries than options"},
		{"Widget", "int", `config:"widget:text"`, "incompatible"},
		{"NumberWidget", "string", `config:"widget:number"`, "incompatible"},
		{"SliderWidget", "string", `config:"widget:slider;min:1;max:2"`, "incompatible"},
		{"SwitchWidget", "string", `config:"widget:switch"`, "incompatible"},
		{"CodeWidget", "int", `config:"widget:code"`, "incompatible"},
		{"SelectOptions", "string", `config:"widget:select"`, "select widget requires options"},
		{"SliderMin", "int", `config:"widget:slider;max:10"`, "slider widget requires min"},
		{"SliderMax", "int", `config:"widget:slider;min:1"`, "slider widget requires max"},
		{"Step", "string", `config:"step:1"`, "form step requires a numeric field"},
		{"Language", "string", `config:"language:sql"`, "language requires the code widget"},
		{"UnknownSibling", "string", `config:"when:missing=x"`, `config when references unknown sibling field "missing"`},
		{"InvalidWhen", "string", `config:"when:condition=wrong"`, `invalid config when value for "condition"`},
		{"ReservedLabel", "string", `json:"label"`, `config JSON name "label" conflicts`},
		{"ReservedDescription", "string", `json:"description"`, `config JSON name "description" conflicts`},
		{"ReservedLabels", "string", `json:"labels"`, `config JSON name "labels" conflicts`},
		{"ReservedSubtitles", "string", `json:"subtitles"`, `config JSON name "subtitles" conflicts`},
		{"LegacyForm", "struct { LegacyValue string `form:\"widget=text\"` }", "", "standalone form tag is unsupported; use the config tag"},
		{"LegacyDescription", "[]struct { LegacyValue string `description:\"hint\"` }", "", "standalone description tag is unsupported; use the config tag"},
		{"LegacySecret", "map[string]struct { LegacyValue string `secret:\"true\"` }", "", "standalone secret tag is unsupported; use the config tag"},
		{"LegacyHidden", "string", `hidden:"true"`, "standalone hidden tag is unsupported; use the config tag"},
		{"Secret", "int", `config:"secret"`, "secret requires string or *string"},
		{"Number", "json.Number", "", "unsupported config type encoding/json.Number"},
		{"Quoted", "int", `json:",string"`, "json string option is unsupported"},
		{"EmbeddedDefault", "Nested", `config:"default:{\"Value\":\"x\"}"`, "cannot set embedded pointer to unexported struct type"},
	}

	var source strings.Builder
	source.WriteString("package main\nimport (\"encoding/json\";\"time\";\"github.com/AlonaBot/AlonaBot/sdk\";\"github.com/AlonaBot/AlonaBot/sdk/config\")\ntype Plugin struct { sdk.Base }\ntype Config struct {\n config.Base\n Existing string `json:\"duplicate\"`\n Condition bool `json:\"condition\"`\n")
	firstLine := strings.Count(source.String(), "\n") + 1
	for _, test := range tests {
		fmt.Fprintf(&source, " %s %s `%s`\n", test.name, test.field, test.tag)
	}
	source.WriteString("}\nfunc main() {}\n")
	writePluginFile(t, dir, "main.go", source.String())
	result := scanPlugin(t, generate.Options{Dir: dir})
	// SliderWidget 同时触发两个非数字边界诊断。
	wantDiagnostics := len(tests) + 2
	if len(result.Diagnostics) != wantDiagnostics || result.ConfigSchema != nil || result.Source != nil {
		t.Fatalf("got %d diagnostics, want %d: %+v", len(result.Diagnostics), wantDiagnostics, result.Diagnostics)
	}
	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			position := fmt.Sprintf("main.go:%d:2", firstLine+index)
			if strings.Contains(test.field, "LegacyValue") {
				line := fmt.Sprintf(" %s %s `%s`", test.name, test.field, test.tag)
				position = fmt.Sprintf("main.go:%d:%d", firstLine+index, strings.Index(line, "LegacyValue")+1)
			}
			if test.name == "Recursive" {
				position = "types.go:10:25"
			}
			for _, diagnostic := range result.Diagnostics {
				if diagnostic.Pos == position && strings.Contains(diagnostic.Message, test.message) {
					return
				}
			}
			t.Fatalf("missing %q at %s: %+v", test.message, position, result.Diagnostics)
		})
	}
}
