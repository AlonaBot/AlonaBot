//go:build unix

package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/AlonaBot/AlonaBot/internal/core/adminauth"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
)

// TestCrossComponentEndToEnd 通过正式包和管理入口联验真实组件进程。
func TestCrossComponentEndToEnd(t *testing.T) {
	t.Run("tcp", func(t *testing.T) {
		t.Run("components", testCrossComponentPackages)
	})
}

func testCrossComponentPackages(t *testing.T) {
	// 1. pack 读取真实工程，安装时只消费正式产物。
	projects := []struct{ directory, id string }{
		{"echo", "echo"},
		{"configdb", "alonabot-plugin-configdb"},
		{"showcase", "showcase"},
	}
	archives := make([]string, len(projects))
	for i, project := range projects {
		archives[i] = packAgentComponent(t, filepath.Join("../../examples/plugins", project.directory), project.id)
	}
	dir := t.TempDir()
	c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0", "--log-format", "json")
	address := c.address()
	tokenFile := filepath.Join(dir, adminauth.FirstTokenFile)
	info, err := os.Stat(tokenFile)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("first token permissions: %v %v", info, err)
	}
	output, err := exec.CommandContext(t.Context(), filepath.Join(binDir, "alonabot"), "admin-token", "show", "--data-dir", dir).CombinedOutput()
	if err != nil {
		t.Fatalf("show first token: %v", err)
	}
	token := strings.TrimSpace(string(output))
	raw, err := os.ReadFile(tokenFile)
	if err != nil || token == "" || token != strings.TrimSpace(string(raw)) {
		t.Fatalf("show did not preserve the first token file: %v", err)
	}
	base, err := url.Parse(settings.Defaults().BaseURL)
	if err != nil {
		t.Fatal(err)
	}
	var cookie *http.Cookie
	var csrf string
	call := func(method, path, body, origin, csrfToken string, authenticated bool, want int) ([]byte, []*http.Cookie) {
		t.Helper()
		request, err := http.NewRequestWithContext(t.Context(), method, "http://"+address+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		request.Host = base.Host
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Origin", origin)
		request.Header.Set("X-Alonabot-CSRF", csrfToken)
		if authenticated {
			request.AddCookie(cookie)
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		content, readErr := io.ReadAll(response.Body)
		if err := errors.Join(readErr, response.Body.Close()); err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != want {
			t.Fatalf("%s %s: got %d want %d: %s\nCore:\n%s", method, path, response.StatusCode, want, content, c.log())
		}
		return content, response.Cookies()
	}
	// 2. 首次认证领取 Cookie；来源和 CSRF 拒绝均发生在组件动作之前。
	call(http.MethodGet, "/api/v1/packages", "", "", "", false, http.StatusUnauthorized)
	body, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		t.Fatal(err)
	}
	call(http.MethodPost, "/api/auth/login", string(body), "http://untrusted.invalid", "", false, http.StatusForbidden)
	if _, err := os.Stat(tokenFile); err != nil {
		t.Fatalf("rejected login consumed first token: %v", err)
	}
	content, cookies := call(http.MethodPost, "/api/auth/login", string(body), base.String(), "", false, http.StatusOK)
	var login struct {
		CSRFToken string `json:"csrf_token"`
	}
	if err := json.Unmarshal(content, &login); err != nil {
		t.Fatal(err)
	}
	if len(cookies) != 1 || login.CSRFToken == "" {
		t.Fatal("login did not return one cookie and CSRF token")
	}
	cookie, csrf = cookies[0], login.CSRFToken
	if cookie.Name != "alonabot_admin" || cookie.Path != "/" || cookie.Domain != "" || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || cookie.Secure {
		t.Fatal("management cookie attributes do not match the HTTP contract")
	}
	if _, err := os.Stat(tokenFile); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("first token file after successful login: %v", err)
	}
	packages := func() []manage.Package {
		t.Helper()
		content, _ := call(http.MethodGet, "/api/v1/packages", "", "", "", true, http.StatusOK)
		var items []manage.Package
		if err := json.Unmarshal(content, &items); err != nil {
			t.Fatal(err)
		}
		return items
	}
	// 3. CLI Bearer 与浏览器 Cookie 观察同一安装与激活状态。
	for i, project := range projects {
		c.management(t, dir, token, "install", archives[i], "--yes")
		items := packages()
		if len(items) != i+1 || items[i].ID != project.id || items[i].Enabled || items[i].Active != "" || items[i].State != "Installed" {
			t.Fatalf("installed packages: %+v", items)
		}
		action, err := json.Marshal(map[string]string{"id": project.id})
		if err != nil {
			t.Fatal(err)
		}
		call(http.MethodPost, "/api/v1/packages/enable", string(action), base.String(), "", true, http.StatusUnauthorized)
		call(http.MethodPost, "/api/v1/packages/enable", string(action), "http://untrusted.invalid", csrf, true, http.StatusUnauthorized)
		if after := packages(); !reflect.DeepEqual(items, after) {
			t.Fatalf("rejected enable changed packages: before=%+v after=%+v", items, after)
		}
		call(http.MethodPost, "/api/v1/packages/enable", string(action), base.String(), csrf, true, http.StatusOK)
		items = packages()
		if !items[i].Enabled || items[i].Active == "" || items[i].State != "Activated" {
			t.Fatalf("enabled package: %+v", items[i])
		}
	}
	var listed []manage.Package
	if err := json.Unmarshal([]byte(c.management(t, dir, token, "list")), &listed); err != nil {
		t.Fatal(err)
	}
	if browser := packages(); !reflect.DeepEqual(browser, listed) {
		t.Fatalf("package states differ: browser=%+v CLI=%+v", browser, listed)
	}
	stopTestCore(t, c)
	for _, credential := range []string{token, cookie.Value, csrf} {
		if strings.Contains(c.log(), credential) {
			t.Fatal("Core logs contain management credentials")
		}
	}
}
