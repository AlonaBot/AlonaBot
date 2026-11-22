//go:build unix

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/AlonaBot/AlonaBot/internal/core/adminauth"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/tune"
)

// TestFirstRunDistribution 联验 Core 与整合发行物的凭证交付、组件清单及重启持久化。
func TestFirstRunDistribution(t *testing.T) {
	for _, edition := range []struct {
		name, directory string
		bundle          bool
	}{
		{"core", os.Getenv("ALONABOT_RELEASE_DIR"), false},
		{"bundle", os.Getenv("ALONABOT_BUNDLE_DIR"), true},
	} {
		t.Run(edition.name, func(t *testing.T) {
			if edition.directory == "" {
				t.Skip("set ALONABOT_RELEASE_DIR or ALONABOT_BUNDLE_DIR to verify prebuilt assets")
			}
			verifyFirstRunDistribution(t, edition.directory, edition.bundle)
		})
	}
}

func verifyFirstRunDistribution(t *testing.T, out string, bundle bool) {
	t.Helper()
	dir := t.TempDir()
	archives, err := filepath.Glob(filepath.Join(out, fmt.Sprintf("alonabot-*-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)))
	if err != nil || len(archives) != 1 {
		t.Fatalf("host archives: %v %v", archives, err)
	}
	root := strings.TrimSuffix(filepath.Base(archives[0]), ".tar.gz")
	extracted := filepath.Join(dir, "extracted")
	if err := packages.Extract(filepath.Join(out, root+".tar.gz"), extracted, packages.NewLimits(tune.Defaults())); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(extracted, root, "alonabot")
	data := filepath.Join(dir, "data")
	c := startRecoveryCore(t, binary, data, "")
	c.waitFor(`"msg":"Core started"`)
	if bundle {
		c.waitFor(`"msg":"plugin activated"`)
	}
	tokenFile := filepath.Join(data, adminauth.FirstTokenFile)
	info, err := os.Stat(tokenFile)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("first token permissions: %v %v", info, err)
	}
	output, err := exec.CommandContext(t.Context(), binary, "admin-token", "show", "--data-dir", data).CombinedOutput()
	if err != nil {
		t.Fatalf("show token: %v %s", err, output)
	}
	token := strings.TrimSpace(string(output))
	if token == "" || strings.Contains(c.log(), token) {
		t.Fatal("missing token or token leaked into logs")
	}
	base, err := url.Parse(settings.Defaults().BaseURL)
	if err != nil {
		t.Fatal(err)
	}
	call := func(method, path, body string) []byte {
		t.Helper()
		req, err := http.NewRequestWithContext(t.Context(), method, "http://"+c.address()+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Host = base.Host
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Origin", base.String())
		req.Header.Set("Content-Type", "application/json")
		response, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		raw, readErr := io.ReadAll(response.Body)
		if err := errors.Join(readErr, response.Body.Close()); err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != http.StatusOK {
			t.Fatalf("%s %s: %d %s", method, path, response.StatusCode, raw)
		}
		return raw
	}
	// 1. API 与发行 CLI 比对同一状态；转发器只适配临时监听地址的 Host。
	target, err := url.Parse("http://" + c.address())
	if err != nil {
		t.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	director := proxy.Director
	proxy.Director = func(r *http.Request) { director(r); r.Host = base.Host }
	server := httptest.NewServer(proxy)
	defer server.Close()
	cli := func(args ...string) []byte {
		t.Helper()
		command := exec.CommandContext(t.Context(), binary, append(args, "--data-dir", data, "--admin-url", server.URL)...)
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("CLI %v: %v\n%s\n%s", args, err, output, c.log())
		}
		return output
	}
	var api, listed []manage.Package
	if err := json.Unmarshal(call(http.MethodGet, "/api/v1/packages", ""), &api); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(cli("list"), &listed); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{}
	if bundle {
		want = map[string]bool{"alonabot-webui-admin": true, "alonabot-agent-official": false}
	}
	if len(api) != len(want) || !reflect.DeepEqual(api, listed) {
		t.Fatalf("package states: API=%+v CLI=%+v", api, listed)
	}
	for _, item := range api {
		enabled, exists := want[item.ID]
		if !exists || item.Enabled != enabled || item.Uninstalled || item.Signed || item.AvailableUpdate != "" {
			t.Fatalf("preinstalled: %+v", item)
		}
		if enabled && item.State != "Activated" || !enabled && (item.Active != "" || item.State != "Installed") {
			t.Fatalf("activation: %+v", item)
		}
	}
	if bundle {
		if strings.Count(c.log(), `"msg":"plugin activated"`) != 1 {
			t.Fatalf("unexpected component processes:\n%s", c.log())
		}
		if raw := call(http.MethodGet, "/", ""); !bytes.Contains(raw, []byte("<html")) {
			t.Fatalf("static WebUI: %s", raw)
		}
	} else if strings.Contains(c.log(), `"msg":"plugin activated"`) {
		t.Fatalf("unexpected component process:\n%s", c.log())
	}
	// 2. 登录后保留持久凭据；卸载选择跨真实进程重启保留。
	input, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		t.Fatal(err)
	}
	call(http.MethodPost, "/api/auth/login", string(input))
	if _, err := os.Stat(tokenFile); err != nil {
		t.Fatalf("token file after login: %v", err)
	}
	if bundle {
		cli("uninstall", "alonabot-webui-admin")
	}
	stopTestCore(t, c)
	c = startRecoveryCore(t, binary, data, "")
	c.waitFor(`"msg":"Core started"`)
	// 重启改变实际监听地址，API 直接读取新 Core。
	if err := json.Unmarshal(call(http.MethodGet, "/api/v1/packages", ""), &api); err != nil {
		t.Fatal(err)
	}
	if len(api) != len(want) {
		t.Fatalf("restart packages: %+v", api)
	}
	for _, item := range api {
		if item.ID == "alonabot-webui-admin" && (!item.Uninstalled || item.Enabled) {
			t.Fatalf("uninstall choice lost: %+v", item)
		}
	}
	if strings.Count(c.log(), `"msg":"plugin activated"`) != 0 {
		t.Fatalf("uninstalled WebUI restarted:\n%s", c.log())
	}
	stopTestCore(t, c)
}
