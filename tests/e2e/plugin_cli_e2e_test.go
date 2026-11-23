//go:build unix

package main

import (
	"bytes"
	"encoding/json"
	"go/parser"
	"go/token"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/generate"
	"github.com/AlonaBot/AlonaBot/internal/pluginkit"
	"github.com/AlonaBot/AlonaBot/internal/tune"
)

// TestDevelopmentCLI 通过真实 CLI、组件进程与预检闭合作者工程到可安装包的主路径。
func TestDevelopmentCLI(t *testing.T) {
	for _, role := range []string{"plugin", "agent"} {
		t.Run(role, func(t *testing.T) { testDevelopmentCLI(t, role) })
	}
}

func testDevelopmentCLI(t *testing.T, role string) {
	id := "alonabot-" + role + "-cli"
	dir := filepath.Join(t.TempDir(), "alonabot-plugin-cli")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	cli := func(args ...string) string {
		t.Helper()
		command := exec.Command(filepath.Join(binDir, "alonabot"), args...)
		command.Dir = dir
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("alonabot %s: %v\n%s", strings.Join(args, " "), err, output)
		}
		return string(output)
	}
	if role == "plugin" {
		cli("init")
	} else {
		cli("init", "--role", role, "--id", id)
	}
	if mod := string(developmentRead(t, filepath.Join(dir, "go.mod"))); !strings.Contains(mod, "module alonabot-plugin-cli\n") {
		t.Fatalf("default module: %s", mod)
	}
	for _, name := range []string{"plugin.go", "plugin.yaml", "plugin_config_schema.json", ".gitignore"} {
		developmentRead(t, filepath.Join(dir, name))
	}
	if _, err := os.Stat(filepath.Join(dir, "main.go")); !os.IsNotExist(err) {
		t.Fatalf("unexpected main.go: %v", err)
	}
	source := string(developmentRead(t, filepath.Join(dir, "plugin.go")))
	for _, required := range []string{"package plugin", "func New() *sdk.Plugin", "sdk.NewPlugin()", `p.Command("hello")`, `.Description("Show a greeting")`, `c.Reply("Hello AlonaBot")`} {
		if !strings.Contains(source, required) {
			t.Fatalf("template missing %q", required)
		}
	}
	if strings.Contains(source, "UseGenerated") {
		t.Fatal("template contains generated API")
	}
	noGenerated := func() {
		t.Helper()
		if err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.Name() == "alonabot_gen.go" {
				t.Errorf("generated source written to project: %s", path)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	noGenerated()
	developmentWrite(t, filepath.Join(dir, "plugin_test.go"), `package plugin
import (
 "testing"
 "github.com/AlonaBot/AlonaBot/plugintest"
 "github.com/AlonaBot/AlonaBot/sdk/message"
)
func TestHello(t *testing.T) {
 h := plugintest.Start(t, plugintest.Plugin(New()))
 result := h.Message(plugintest.Private("alice"), plugintest.User{ID:"alice"}, message.Text("/hello"))
 if len(result.Sends) != 1 || result.Sends[0].Text() != "Hello AlonaBot" { t.Fatalf("unexpected greeting: %+v", result) }
}
`)
	if !strings.Contains(string(developmentRead(t, filepath.Join(dir, ".gitignore"))), "data/") {
		t.Fatal("data ignore missing")
	}
	author := developmentRead(t, filepath.Join(dir, "plugin.yaml"))
	project, err := manifest.ParseProject(author)
	if err != nil {
		t.Fatal(err)
	}
	if project.Manifest.ID != id || len(project.Warnings) != 0 {
		t.Fatalf("default identity: %+v", project)
	}
	if _, problems := configschema.Parse(developmentRead(t, filepath.Join(dir, "plugin_config_schema.json"))); len(problems) > 0 {
		t.Fatal(problems)
	}
	if _, err := os.Stat(filepath.Join(dir, "i18n")); !os.IsNotExist(err) {
		t.Fatalf("unexpected i18n directory: %v", err)
	}
	parsed, err := parser.ParseFile(token.NewFileSet(), "plugin.go", source, parser.ParseComments)
	if err != nil || len(parsed.Comments) != 0 {
		t.Fatalf("template comments: %v", err)
	}
	if bytes.HasPrefix(author, []byte("#")) {
		t.Fatal("manifest contains a header comment")
	}

	// 外部作者工程用明确版本和本 checkout 的 replace，所有 module 写入都在临时目录。
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"mod", "edit", "-require=github.com/AlonaBot/AlonaBot@v0.0.0", "-replace=github.com/AlonaBot/AlonaBot=" + root}, {"mod", "tidy"}} {
		command := exec.Command("go", args...)
		command.Dir = dir
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("go %v: %v\n%s", args, err, output)
		}
	}
	// 与外层 go test -race 使用相同的构建参数，复用已编译的依赖。
	command := exec.Command("go", "test", "-race", "./...")
	command.Dir = dir
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("direct go test: %v\n%s", err, output)
	}
	noGenerated()
	cli("build", dir)
	noGenerated()
	cli("check")
	noGenerated()
	output := cli("test", dir)
	noGenerated()
	for _, required := range []string{"Activated " + id + " (" + role + ")", "Command: hello", "Removed temporary test data:"} {
		if !strings.Contains(output, required) {
			t.Fatalf("missing %q:\n%s", required, output)
		}
	}
	for _, line := range strings.Split(output, "\n") {
		if path, ok := strings.CutPrefix(line, "Removed temporary test data: "); ok {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("test data remains: %v", err)
			}
		}
	}
	out := filepath.Join(dir, "bin")
	other := "linux/amd64"
	if other == runtime.GOOS+"/"+runtime.GOARCH {
		other = "windows/amd64"
	}
	cli("pack", "--target", runtime.GOOS+"/"+runtime.GOARCH, "--target", other)
	noGenerated()
	for _, target := range []string{runtime.GOOS + "/" + runtime.GOARCH, other} {
		goos, goarch, _ := strings.Cut(target, "/")
		developmentRead(t, filepath.Join(out, id+"-0.1.0-"+goos+"-"+goarch+".tar.gz"))
	}
	for _, name := range []string{"release.json", "release.json.sig", "release.pub"} {
		if _, err := os.Stat(filepath.Join(out, name)); !os.IsNotExist(err) {
			t.Fatalf("unexpected release asset %s: %v", name, err)
		}
	}
	archive := filepath.Join(out, id+"-0.1.0-"+runtime.GOOS+"-"+runtime.GOARCH+".tar.gz")
	candidate, err := packages.Preflight(packages.Input{Source: "github.com/CHANGE-ME/CHANGE-ME", OS: runtime.GOOS, Arch: runtime.GOARCH, ArchivePath: archive}, t.TempDir(), packages.NewLimits(tune.Defaults()))
	if err != nil {
		t.Fatal(err)
	}
	m := candidate.Manifest
	if m.Role != role || m.MigrationMode != manifest.MigrationNone || m.PrepareEffects != manifest.EffectsNone || m.OS != runtime.GOOS || m.Arch != runtime.GOARCH {
		t.Fatalf("package manifest: %+v", m)
	}
	for _, name := range []string{m.Entrypoint, "plugin.yaml", "plugin_config_schema.json", "build-info"} {
		developmentRead(t, filepath.Join(candidate.Dir, name))
	}
	info, err := debug.ParseBuildInfo(string(developmentRead(t, filepath.Join(candidate.Dir, "build-info"))))
	if err != nil {
		t.Fatal(err)
	}
	settings := map[string]string{}
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}
	if settings["GOOS"] != m.OS || settings["GOARCH"] != m.Arch {
		t.Fatal("build-info target differs from manifest", settings)
	}
	for _, name := range []string{"i18n", "data", "config.json", "release.json", "release.json.sig", "release.pub", "release.key"} {
		if _, err := os.Stat(filepath.Join(candidate.Dir, name)); !os.IsNotExist(err) {
			t.Fatalf("unexpected package material %s: %v", name, err)
		}
	}
	schema, problems := configschema.Parse(developmentRead(t, filepath.Join(dir, "plugin_config_schema.json")))
	if len(problems) > 0 {
		t.Fatal(problems)
	}
	if !bytes.Equal(schema.Canonical(), developmentRead(t, filepath.Join(candidate.Dir, "plugin_config_schema.json"))) {
		t.Fatal("package schema is not canonical")
	}
	if !bytes.Equal(author, developmentRead(t, filepath.Join(dir, "plugin.yaml"))) {
		t.Fatal("author manifest changed")
	}
	if _, err := candidate.MoveTo(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	coreDir := t.TempDir()
	c := startEmptyCore(t, coreDir, "--listen", "127.0.0.1:0")
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(out)
	var installed manage.Package
	if err := json.Unmarshal([]byte(c.management(t, coreDir, "", "install", "--yes")), &installed); err != nil {
		t.Fatal(err)
	}
	t.Chdir(cwd)
	if installed.ID != id || installed.Source != "github.com/CHANGE-ME/CHANGE-ME" {
		t.Fatalf("installed: %+v", installed)
	}
	var registered []manage.Package
	if err := json.Unmarshal([]byte(c.management(t, coreDir, "", "list")), &registered); err != nil {
		t.Fatal(err)
	}
	if len(registered) != 1 || registered[0].ID != installed.ID || registered[0].Source != installed.Source {
		t.Fatalf("registered: %+v", registered)
	}
	stopTestCore(t, c)
	developmentWrite(t, filepath.Join(dir, "cmd", "bot", "main.go"), `package main
import (
 "context"
 "fmt"
 "os"
 "github.com/AlonaBot/AlonaBot"
 plugin "alonabot-plugin-cli"
)
var component = plugin.New()
func main() {
 ctx, cancel := context.WithCancel(context.Background())
 defer cancel()
 core, err := alonabot.Start(ctx, alonabot.DataDir(os.Args[1]), alonabot.Listen("127.0.0.1:0"), alonabot.Plugin(component))
 if err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
 fmt.Println("Activated library plugin")
 cancel()
 if err := core.Wait(); err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
}
`)
	command = exec.Command("go", "mod", "tidy")
	command.Dir = dir
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("tidy library app: %v\n%s", err, output)
	}
	cli("build", "cmd/bot")
	noGenerated()
	built := filepath.Join(dir, "cmd", "bot", "bin", "bot")
	if output, err := exec.Command(built, t.TempDir()).CombinedOutput(); err != nil || !strings.Contains(string(output), "Activated library plugin") {
		t.Fatalf("library application: %v\n%s", err, output)
	}
	native := filepath.Join(t.TempDir(), "native")
	command = exec.Command("go", "build", "-race", "-o", native, "./cmd/bot")
	command.Dir = dir
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("native go build: %v\n%s", err, output)
	}
	if output, err := exec.Command(native, t.TempDir()).CombinedOutput(); err == nil || !strings.Contains(string(output), "alonabot build") || strings.Contains(string(output), "Activated library plugin") {
		t.Fatalf("missing resource diagnostic: %v\n%s", err, output)
	}

}

// TestPluginStorageCLI 通过真实 Core 与 configdb 进程验证插件存储的生命周期。
func TestPluginStorageCLI(t *testing.T) {
	t.Run("kv-and-data-dir", func(t *testing.T) {
		archive, err := pluginkit.Pack(t.Context(), generate.Options{Dir: "../../examples/plugins/configdb"}, t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		dir := t.TempDir()
		c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0", "--log-format", "json")
		adapter := testPackage(t, "alonabot-adapter-fixture", "adapter", "devloop", "github.com/example/adapter", "", map[string][]byte{"adapter_config_schema.json": []byte(`{}`)})
		c.management(t, dir, "", "install", adapter, "--yes")
		c.management(t, dir, "", "adapter", "instance", "create", "alonabot-adapter-fixture", "one")
		c.management(t, dir, "", "enable", "alonabot-adapter-fixture")
		c.waitFor("fixture instance one started")
		const id = "alonabot-plugin-configdb"
		c.management(t, dir, "", "install", archive, "--yes")
		c.management(t, dir, "", "enable", id)

		deliver := func(text string) {
			t.Helper()
			path := "/plugins/alonabot-adapter-fixture/deliver?instance=one&user=alice&text=" + url.QueryEscape(text)
			if status, body := fixtureHTTP(t, c, dir, http.MethodPost, path); status != http.StatusNoContent {
				t.Fatalf("deliver: %d %s", status, body)
			}
		}
		count := 0
		check := func(text, want string) {
			t.Helper()
			deliver(text)
			count++
			deadline := time.Now().Add(5 * time.Second)
			for {
				status, body := fixtureHTTP(t, c, dir, http.MethodGet, "/plugins/alonabot-adapter-fixture/sent")
				var sent []string
				if status != http.StatusOK || json.Unmarshal(body, &sent) != nil {
					t.Fatalf("sent: %d %s", status, body)
				}
				if len(sent) >= count {
					if len(sent) != count || sent[count-1] != want {
						t.Fatalf("reply: %v, want %s", sent, want)
					}
					return
				}
				if time.Now().After(deadline) {
					t.Fatalf("missing reply %s: %s\n%s", want, body, c.log())
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
		check("/note kvset topic retained-kv", "Saved topic")
		check("/note fileset retained-file", "Saved note.txt")
		dataDir := filepath.Join(dir, "plugin_data", id)
		if raw := developmentRead(t, filepath.Join(dataDir, "note.txt")); string(raw) != "retained-file" {
			t.Fatalf("unexpected plugin file: %q", raw)
		}
		c.management(t, dir, "", "restart", id)
		check("/note kvget topic", "topic: retained-kv")
		check("/note fileget", "note.txt: retained-file")
		preview := managementFailure(t, c, dir, "uninstall", id, "--purge-data")
		if !strings.Contains(preview, dataDir) {
			t.Fatalf("plugin data directory missing from confirmation: %s", preview)
		}
		c.management(t, dir, "", "uninstall", id, "--purge-data", "--yes")
		if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
			t.Fatalf("plugin data directory remains: %v", err)
		}
		c.management(t, dir, "", "install", archive, "--yes")
		c.management(t, dir, "", "enable", id)
		deliver("/note kvget topic")
		c.waitFor(`"msg":"command handler returned an error".*"error":"KV key does not exist"`)
		stopTestCore(t, c)
	})
}
