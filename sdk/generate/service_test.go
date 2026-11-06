package generate_test

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/AlonaBot/AlonaBot/sdk/generate"
)

// TestServiceSchemaGenerationIntegration 验证真实模块的实例扫描、固定 Schema 字节、编译收集与 CLI 过期检查。
func TestServiceSchemaGenerationIntegration(t *testing.T) {
	dir := pluginModuleDir(t)
	const source = `package main
import (
 "fmt"
 "github.com/AlonaBot/AlonaBot/sdk"
 "example.com/scanplugin/handlers"
)
type Plugin struct { sdk.Base }
type Request struct {
 Name string ` + "`json:\"name\" config:\"invalid-tag-is-ignored\"`" + `
 Count *int32 ` + "`json:\"count,omitempty\"`" + `
}
type Result struct { OK bool ` + "`json:\"ok\"`" + ` }
type BusPayload chan int
func typeReferences() {
 _ = sdk.Subscribe[BusPayload]
 _ = sdk.CallJSON[Result,Result]
 _ = sdk.Publish[BusPayload]
}
func (*Plugin) Services(r *sdk.ServiceRegistry) {
 sdk.JSON[Request,Result](r,"first").Handle(func(*sdk.Context,Request)(Result,error){return Result{OK:true},nil})
 handlers.Register(r)
}
func main(){if _,err:=sdk.Collect(&Plugin{});err!=nil{panic(err)};fmt.Println("registered")}
`
	writePluginFile(t, dir, "main.go", source)
	writePluginFile(t, dir, "handlers/services.go", `package handlers
import "github.com/AlonaBot/AlonaBot/sdk"
type Input struct { Name string `+"`json:\"name\"`"+` }
type Output struct { Values []string `+"`json:\"values\"`"+` }
func Register(r *sdk.ServiceRegistry) {
 sdk.JSON[Input,Output](r,"second").Handle(func(*sdk.Context,Input)(Output,error){return Output{Values:[]string{}},nil})
}
`)
	first := scanPlugin(t, generate.Options{Dir: dir})
	if len(first.Diagnostics) != 0 {
		t.Fatal(first.Diagnostics)
	}
	expected := map[string]string{
		"main.Request":                    `{"$schema":"https://json-schema.org/draft/2020-12/schema","properties":{"count":{"maximum":2147483647,"minimum":-2147483648,"type":["integer","null"]},"name":{"type":"string"}},"required":["name"],"type":"object"}`,
		"main.Result":                     `{"$schema":"https://json-schema.org/draft/2020-12/schema","properties":{"ok":{"type":"boolean"}},"required":["ok"],"type":"object"}`,
		pluginModule + "/handlers.Input":  `{"$schema":"https://json-schema.org/draft/2020-12/schema","properties":{"name":{"type":"string"}},"required":["name"],"type":"object"}`,
		pluginModule + "/handlers.Output": `{"$schema":"https://json-schema.org/draft/2020-12/schema","properties":{"values":{"items":{"type":"string"},"type":"array"}},"required":["values"],"type":"object"}`,
	}
	actual := generatedServiceSchemas(t, first.Source)
	if len(actual) != len(expected) {
		t.Fatalf("generated schemas=%v", actual)
	}
	for id, want := range expected {
		if actual[id] != want {
			t.Errorf("schema %s:\ngot  %s\nwant %s", id, actual[id], want)
		}
	}
	writePluginFile(t, dir, generate.FileName, string(first.Source))
	second := scanPlugin(t, generate.Options{Dir: dir})
	if len(second.Diagnostics) != 0 || !bytes.Equal(first.Source, second.Source) {
		t.Fatalf("unstable generation: %v\n%s", second.Diagnostics, second.Source)
	}
	executable := filepath.Join(dir, "plugin")
	cli := filepath.Join(t.TempDir(), "alonabot")
	if runtime.GOOS == "windows" {
		executable += ".exe"
		cli += ".exe"
	}
	runPluginCommand(t, dir, nil, "go", "build", "-o", executable, ".")
	if got := string(runPluginCommand(t, dir, nil, executable)); got != "registered\n" {
		t.Fatalf("plugin output=%q", got)
	}
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Dir(filepath.Dir(filepath.Dir(file)))
	runPluginCommand(t, root, nil, "go", "build", "-o", cli, "./cmd/alonabot")
	runPluginCommand(t, dir, nil, cli, "plugin", "check", "--dir", dir)
	// 只改 Go 类型字段，保持旧生成项；check 必须发现 Schema 资源已经过期。
	writePluginFile(t, dir, "main.go", strings.Replace(source, "Name string", "Name bool", 1))
	command := exec.Command(cli, "plugin", "check", "--dir", dir)
	command.Dir = dir
	output, err := command.CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "alonabot_gen.go is out of date") {
		t.Fatalf("stale check: %v\n%s", err, output)
	}
	runPluginCommand(t, dir, nil, cli, "plugin", "generate", "--dir", dir)
	runPluginCommand(t, dir, nil, cli, "plugin", "check", "--dir", dir)
}

// generatedServiceSchemas 读取生成源码中的固定资源，不借助 SDK 私有注册状态。
func generatedServiceSchemas(t *testing.T, source []byte) map[string]string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), generate.FileName, source, 0)
	if err != nil {
		t.Fatal(err)
	}
	result := make(map[string]string)
	ast.Inspect(file, func(node ast.Node) bool {
		field, ok := node.(*ast.KeyValueExpr)
		if !ok {
			return true
		}
		name, ok := field.Key.(*ast.Ident)
		if !ok || name.Name != "ServiceSchemas" {
			return true
		}
		for _, element := range field.Value.(*ast.CompositeLit).Elts {
			entry := element.(*ast.KeyValueExpr)
			id, err := strconv.Unquote(entry.Key.(*ast.BasicLit).Value)
			if err != nil {
				t.Fatal(err)
			}
			value, err := strconv.Unquote(entry.Value.(*ast.CallExpr).Args[0].(*ast.BasicLit).Value)
			if err != nil {
				t.Fatal(err)
			}
			result[id] = value
		}
		return false
	})
	return result
}

// TestServiceSchemaDiagnosticsIntegration 验证真实源码中的非法 Service 类型带文件、行、列诊断。
func TestServiceSchemaDiagnosticsIntegration(t *testing.T) {
	for _, test := range []struct{ name, declaration, register, diagnostic string }{
		{"anonymous", "", `sdk.JSON[struct{},Result](r,"bad").Handle(func(*sdk.Context,struct{})(Result,error){return Result{},nil})`, "package-level named non-generic"},
		{"unsupported", `type Input struct { Stream chan int }`, `sdk.JSON[Input,Result](r,"bad").Handle(func(*sdk.Context,Input)(Result,error){return Result{},nil})`, "unsupported service type chan int"},
		{"json_local", "", `type Local string; _ = sdk.JSON[Local,Result]`, "package-level named non-generic"},
		{"subscribe_local", "", `type Local string; _ = sdk.Subscribe[Local]`, "package-level named non-generic"},
		{"call_input_local", "", `type Local string; _ = sdk.CallJSON[Local,Result]`, "package-level named non-generic"},
		{"call_output_local", "", `type Local string; _ = sdk.CallJSON[Result,Local]`, "package-level named non-generic"},
		{"publish_local", "", `type Local string; _ = sdk.Publish[Local]`, "package-level named non-generic"},
		{"subscribe_anonymous", "", `_ = sdk.Subscribe[struct{}]`, "package-level named non-generic"},
		{"call_pointer", "", `_ = sdk.CallJSON[*Result,Result]`, "package-level named non-generic"},
		{"publish_generic", `type Box[T any] struct { Value T }`, `_ = sdk.Publish[Box[string]]`, "package-level named non-generic"},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := pluginModuleDir(t)
			writePluginFile(t, dir, "main.go", `package main
import "github.com/AlonaBot/AlonaBot/sdk"
type Plugin struct { sdk.Base }
type Result struct {}
`+test.declaration+`
func (*Plugin) Services(r *sdk.ServiceRegistry){`+test.register+`}
func main(){}
`)
			result := scanPlugin(t, generate.Options{Dir: dir})
			if len(result.Diagnostics) != 1 || len(result.Source) != 0 {
				t.Fatalf("diagnostics=%v source=%s", result.Diagnostics, result.Source)
			}
			diagnostic := result.Diagnostics[0]
			parts := strings.Split(diagnostic.Pos, ":")
			if len(parts) != 3 || parts[0] != "main.go" || !strings.Contains(diagnostic.Message, test.diagnostic) {
				t.Fatalf("diagnostic=%+v", diagnostic)
			}
			for _, position := range parts[1:] {
				n, err := strconv.Atoi(position)
				if err != nil || n <= 0 {
					t.Fatalf("invalid diagnostic position: %q", diagnostic.Pos)
				}
			}
		})
	}
}
