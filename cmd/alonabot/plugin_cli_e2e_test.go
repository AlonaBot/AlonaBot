//go:build unix

package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/contracts/i18n"
	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/tune"
)

// TestDevelopmentCLI 通过真实 CLI、组件进程与预检闭合作者工程到可安装包的主路径。
func TestDevelopmentCLI(t *testing.T) {
	dir := t.TempDir()
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
	cli("init", "--id", "alonabot_plugin_cli", "--module", "example.com/cli")
	for _, name := range []string{"plugin.go", "plugin.yaml", "config_schema.json", "i18n/en-US.json", ".gitignore"} {
		developmentRead(t, filepath.Join(dir, name))
	}
	if _, err := os.Stat(filepath.Join(dir, "main.go")); !os.IsNotExist(err) {
		t.Fatalf("unexpected main.go: %v", err)
	}
	source := string(developmentRead(t, filepath.Join(dir, "plugin.go")))
	for _, required := range []string{"package plugin", "func New() *sdk.Plugin", "sdk.NewPlugin()", `p.Command("hello")`, `.Description("Show a greeting")`} {
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
 if len(result.Sends) != 1 || result.Sends[0].Text() != "Hello!" { t.Fatalf("unexpected greeting: %+v", result) }
}
`)
	if !strings.Contains(string(developmentRead(t, filepath.Join(dir, ".gitignore"))), "data/") {
		t.Fatal("data ignore missing")
	}
	author := developmentRead(t, filepath.Join(dir, "plugin.yaml"))
	if _, err := manifest.ParseProject(author); err != nil {
		t.Fatal(err)
	}
	if _, problems := configschema.Parse(developmentRead(t, filepath.Join(dir, "config_schema.json"))); len(problems) > 0 {
		t.Fatal(problems)
	}
	if _, problems := i18n.Parse(map[string][]byte{"en-US.json": developmentRead(t, filepath.Join(dir, "i18n/en-US.json"))}); len(problems) > 0 {
		t.Fatal(problems)
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
	command := exec.Command("go", "test", "./...")
	command.Dir = dir
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("direct go test: %v\n%s", err, output)
	}
	noGenerated()
	cli("build")
	noGenerated()
	cli("check")
	noGenerated()
	output := cli("test")
	noGenerated()
	for _, required := range []string{"Activated alonabot_plugin_cli (plugin)", "Command: hello", "Removed temporary test data:"} {
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
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(t.TempDir(), "release.key")
	developmentWrite(t, keyPath, base64.StdEncoding.EncodeToString(key))
	out := t.TempDir()
	cli("pack", "--asset-url", "https://example.com/cli.tar.gz", "--key", keyPath, "--out", out)
	noGenerated()
	raw := developmentRead(t, filepath.Join(out, "release.json"))
	sig := developmentRead(t, filepath.Join(out, "release.json.sig"))
	publication, err := release.Verify(raw, sig, pub)
	if err != nil {
		t.Fatal(err)
	}
	if publication.PublisherKeyID != release.KeyID(pub) {
		t.Fatal("publisher identity mismatch")
	}
	archive := filepath.Join(out, "alonabot_plugin_cli-0.1.0-"+runtime.GOOS+"-"+runtime.GOARCH+".tar.gz")
	candidate, err := packages.Preflight(packages.Input{ReleaseJSON: raw, Signature: sig, PublicKey: pub, Source: "github.com/CHANGE-ME/CHANGE-ME", OS: runtime.GOOS, Arch: runtime.GOARCH, ArchivePath: archive}, t.TempDir(), packages.NewLimits(tune.Defaults()))
	if err != nil {
		t.Fatal(err)
	}
	m := candidate.Manifest
	if m.Role != "plugin" || m.MigrationMode != manifest.MigrationNone || m.PrepareEffects != manifest.EffectsNone || m.OS != runtime.GOOS || m.Arch != runtime.GOARCH {
		t.Fatalf("package manifest: %+v", m)
	}
	for _, name := range []string{m.Entrypoint, "plugin.yaml", "config_schema.json", "i18n/en-US.json", "build-info"} {
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
	for _, name := range []string{"data", "config.json", "release.json", "release.json.sig", "release.pub", "release.key"} {
		if _, err := os.Stat(filepath.Join(candidate.Dir, name)); !os.IsNotExist(err) {
			t.Fatalf("unexpected package material %s: %v", name, err)
		}
	}
	schema, problems := configschema.Parse(developmentRead(t, filepath.Join(dir, "config_schema.json")))
	if len(problems) > 0 {
		t.Fatal(problems)
	}
	if !bytes.Equal(schema.Canonical(), developmentRead(t, filepath.Join(candidate.Dir, "config_schema.json"))) {
		t.Fatal("package schema is not canonical")
	}
	if !bytes.Equal(author, developmentRead(t, filepath.Join(dir, "plugin.yaml"))) {
		t.Fatal("author manifest changed")
	}
	if _, err := candidate.MoveTo(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	developmentWrite(t, filepath.Join(dir, "cmd", "bot", "main.go"), `package main
import (
 "context"
 "fmt"
 "os"
 "github.com/AlonaBot/AlonaBot"
 plugin "example.com/cli"
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
	cli("build", "--dir", "cmd/bot")
	noGenerated()
	built := filepath.Join(dir, "cmd", "bot", "bin", "bot")
	if output, err := exec.Command(built, t.TempDir()).CombinedOutput(); err != nil || !strings.Contains(string(output), "Activated library plugin") {
		t.Fatalf("library application: %v\n%s", err, output)
	}
	native := filepath.Join(t.TempDir(), "native")
	command = exec.Command("go", "build", "-o", native, "./cmd/bot")
	command.Dir = dir
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("native go build: %v\n%s", err, output)
	}
	if output, err := exec.Command(native, t.TempDir()).CombinedOutput(); err == nil || !strings.Contains(string(output), "alonabot build") || strings.Contains(string(output), "Activated library plugin") {
		t.Fatalf("missing resource diagnostic: %v\n%s", err, output)
	}

}
