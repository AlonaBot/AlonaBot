//go:build unix

package app

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/internal/core/adminauth"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
)

// TestWebUIComponents 通过真实 Core、管理服务与单一公开 Listener 验证静态组件和图标。
func TestWebUIComponents(t *testing.T) {
	parent := t
	h := newWebUIHarness(t)
	h.start(nil)
	h.expect("GET", "/", nil, false, 503, "alonabot install")
	h.expect("GET", "/api/v1/packages", nil, false, 401, "unauthorized")
	h.expect("GET", "/api/v1/packages", nil, true, 200, "[]")

	t.Run("installed-static-and-routing", func(t *testing.T) {
		h.t = t
		defer func() { h.t = parent }()
		h.install("ui", "webui", "", 1, false)
		h.expect("GET", "/", nil, false, 503, "alonabot install")
		h.action("enable", "ui", 200)
		for _, route := range []string{"/", "/index.html", "/front/route"} {
			h.expect("GET", route, nil, false, 200, "installed-ui")
		}
		h.expect("GET", "/app.js", nil, false, 200, "static-asset")
		h.expect("GET", "/missing.js", nil, false, 404, "")
		response, _ := h.expect("GET", "/api/unknown", nil, true, 404, "NOT_FOUND")
		if response.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("API MIME = %q", response.Header.Get("Content-Type"))
		}
		for _, route := range []string{"/healthz", "/internal/auth", "/media/missing"} {
			h.expect("GET", route, nil, false, 404, "")
		}
		h.expect("GET", "/plugins/missing/page", nil, false, 503, "Service Unavailable")
		h.action("disable", "ui", 200)
		h.expect("GET", "/", nil, false, 503, "alonabot install")
		h.action("enable", "ui", 200)
	})

	t.Run("version-gates-preserve-old-webui", func(t *testing.T) {
		h.t = t
		defer func() { h.t = parent }()
		candidate := h.bundle("incompatible", "webui", "", 2)
		candidate.Replace = true
		h.expect("POST", "/api/v1/packages/install", candidate, true, 412, "FAILED_PRECONDITION")
		h.expect("GET", "/", nil, false, 200, "installed-ui")
		// 改写临时包的声明，模拟 Core 升级后已安装包的范围失配。
		name := h.manifestPath("ui")
		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		incompatible := bytes.ReplaceAll(raw, []byte("major: 1"), []byte("major: 2"))
		if err := os.WriteFile(name, incompatible, 0o600); err != nil {
			t.Fatal(err)
		}
		h.expect("GET", "/", nil, false, 503, "Core Admin API major 1")
		h.expect("GET", "/api/v1/packages", nil, true, 200, "ui")
		h.action("disable", "ui", 200)
		h.action("enable", "ui", 412)
		enabled, err := h.core.store.Enabled(context.Background(), "ui")
		if err != nil || enabled {
			t.Fatalf("failed Enable changed state: enabled=%v err=%v", enabled, err)
		}
		if err := os.WriteFile(name, raw, 0o600); err != nil {
			t.Fatal(err)
		}
		h.action("enable", "ui", 200)
		h.expect("GET", "/", nil, false, 200, "installed-ui")
	})

	t.Run("authenticated-declared-icons", func(t *testing.T) {
		h.t = t
		defer func() { h.t = parent }()
		for _, tc := range []struct{ ext, mime string }{{"svg", "image/svg+xml"}, {"png", "image/png"}, {"webp", "image/webp"}} {
			id := "icon_" + tc.ext
			h.install(id, "plugin", "icon."+tc.ext, 1, false)
			route := "/api/v1/packages/" + id + "/icon"
			h.expect("GET", route, nil, false, 401, "unauthorized")
			response, body := h.expect("GET", route+"?path=private.txt", nil, true, 200, "")
			if string(body) != webUIIcon(tc.ext) {
				t.Fatalf("icon bytes = %q", body)
			}
			if response.Header.Get("Content-Type") != tc.mime || response.Header.Get("Content-Security-Policy") != "sandbox; default-src 'none'; style-src 'unsafe-inline'" || response.Header.Get("X-Content-Type-Options") != "nosniff" {
				t.Fatalf("icon headers = %v", response.Header)
			}
			for _, suffix := range []string{"/private.txt", "/../private.txt", "/%2e%2e/private.txt", "/icon.html"} {
				h.expect("GET", route+suffix, nil, true, 404, "")
			}
		}
		h.expect("GET", "/api/v1/packages/ui/icon", nil, true, 404, "")
		h.expect("GET", "/api/v1/packages/missing/icon", nil, true, 404, "")
		bad := h.bundle("bad_icon", "plugin", "icon.html", 1)
		h.expect("POST", "/api/v1/packages/install", bad, true, 400, "INVALID_ARGUMENT")
	})

	t.Run("development-override-through-start", func(t *testing.T) {
		h.t = t
		defer func() { h.t = parent }()
		h.stop()
		h.start(fstest.MapFS{"index.html": {Data: []byte("development-ui")}, "app.js": {Data: []byte("development-asset")}})
		h.expect("GET", "/", nil, false, 200, "development-ui")
		h.expect("GET", "/front/route", nil, false, 200, "development-ui")
		h.expect("GET", "/app.js", nil, false, 200, "development-asset")
		h.expect("GET", "/api/v1/packages", nil, true, 200, "ui")
		installs, err := h.core.store.Installs(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		for _, in := range installs {
			if in.Source == "dev" {
				t.Fatalf("development WebUI wrote installation: %+v", in)
			}
		}
		enabled, err := h.core.store.Enabled(context.Background(), "ui")
		if err != nil || !enabled {
			t.Fatalf("override changed enabled: %v %v", enabled, err)
		}
		h.stop()
		h.start(nil)
		h.expect("GET", "/", nil, false, 200, "installed-ui")
	})
}

type webUIHarness struct {
	t                  *testing.T
	dir, token, origin string
	core               *Core
	cancel             context.CancelFunc
	client             *http.Client
}

func newWebUIHarness(t *testing.T) *webUIHarness {
	t.Helper()
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })
	h := &webUIHarness{t: t, dir: t.TempDir(), origin: "http://127.0.0.1", client: &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
	if err := os.WriteFile(filepath.Join(h.dir, settings.FileName), []byte(`{"listen":"127.0.0.1:0","base_url":"`+h.origin+`"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(h.stop)
	return h
}

func (h *webUIHarness) start(fsys fs.FS) {
	h.t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	opts := Options{DataDir: h.dir}
	if fsys != nil {
		opts.WebUIs = append(opts.WebUIs, fsys)
	}
	core, err := Start(ctx, opts)
	if err != nil {
		cancel()
		h.t.Fatal(err)
	}
	h.core, h.cancel = core, cancel
	if h.token == "" {
		raw, err := os.ReadFile(filepath.Join(h.dir, adminauth.FirstTokenFile))
		if err != nil {
			h.t.Fatal(err)
		}
		h.token = strings.TrimSpace(string(raw))
	}
}

func (h *webUIHarness) stop() {
	if h.cancel == nil {
		return
	}
	h.cancel()
	if err := h.core.Wait(); err != nil {
		h.t.Errorf("stop Core: %v", err)
	}
	h.cancel = nil
	h.client.CloseIdleConnections()
}

func (h *webUIHarness) expect(method, route string, body any, authenticated bool, status int, contains string) (*http.Response, []byte) {
	h.t.Helper()
	var data []byte
	if body != nil {
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			h.t.Fatal(err)
		}
	}
	req, err := http.NewRequest(method, "http://"+h.core.http.Addr+route, bytes.NewReader(data))
	if err != nil {
		h.t.Fatal(err)
	}
	if authenticated {
		req.Header.Set("Authorization", "Bearer "+h.token)
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := h.client.Do(req)
	if err != nil {
		h.t.Fatal(err)
	}
	raw, err := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if err != nil || closeErr != nil {
		h.t.Fatalf("read response: %v %v", err, closeErr)
	}
	if response.StatusCode != status || !strings.Contains(string(raw), contains) {
		h.t.Fatalf("%s %s = %d %q, want %d containing %q", method, route, response.StatusCode, raw, status, contains)
	}
	return response, raw
}

func (h *webUIHarness) action(action, id string, status int) {
	h.t.Helper()
	h.expect("POST", "/api/v1/packages/"+action, map[string]any{"id": id, "confirm": true}, true, status, "")
}

func (h *webUIHarness) install(id, role, icon string, major uint32, replace bool) {
	h.t.Helper()
	req := h.bundle(id, role, icon, major)
	req.Replace = replace
	h.expect("POST", "/api/v1/packages/install", req, true, 200, id)
}

func (h *webUIHarness) bundle(id, role, icon string, major uint32) manage.InstallRequest {
	h.t.Helper()
	platformOS, arch := runtime.GOOS, runtime.GOARCH
	m := fmt.Sprintf("manifest_version: 1\nid: %s\nname: WebUI Fixture\nversion: 1.0.0\nsource: github.com/example/%s\nrole: %s\n", id, id, role)
	if role == "webui" {
		platformOS, arch = "any", "any"
		m += fmt.Sprintf("assets: public\nadmin_api_min_major: %d\nadmin_api_max_major: %d\n", major, major)
	} else {
		m += "protocol_min_major: 1\nprotocol_max_major: 1\nrequired_capabilities: []\nentrypoint: bin/component\nconfig_version: 1\nmigration_mode: none\nprepare_effects: none\n"
	}
	m += fmt.Sprintf("os: %s\narch: %s\n", platformOS, arch)
	files := map[string]string{"plugin.yaml": m, "private.txt": "private-package-file"}
	if role == "webui" {
		for name, body := range map[string]string{"index.html": "installed-ui", "app.js": "static-asset", "api/unknown": "shadow-api", "internal/auth": "shadow-internal", "plugins/missing/page": "shadow-plugin", "media/missing": "shadow-media"} {
			files["public/"+name] = body
		}
	} else {
		files["bin/component"] = "unused executable"
	}
	if icon != "" {
		files["plugin.yaml"] += "icon: " + icon + "\n"
		files[icon] = webUIIcon(strings.TrimPrefix(filepath.Ext(icon), "."))
	}
	var archive bytes.Buffer
	gz := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gz)
	for name, body := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o600, Size: int64(len(body))}); err != nil {
			h.t.Fatal(err)
		}
		if _, err := io.WriteString(tw, body); err != nil {
			h.t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		h.t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		h.t.Fatal(err)
	}
	// 安装清单要求整个归档的摘要，夹具只计算这一份必要值。
	digest := sha256.Sum256(archive.Bytes())
	r := release.Release{PluginID: id, Version: "1.0.0", Source: "github.com/example/" + id, Assets: []release.Asset{{OS: platformOS, Arch: arch, URL: "https://example.invalid/package.tar.gz", Size: int64(archive.Len()), SHA256: hex.EncodeToString(digest[:])}}}
	raw, err := json.Marshal(r)
	if err != nil {
		h.t.Fatal(err)
	}
	name := filepath.Join(h.t.TempDir(), "package.tar.gz")
	if err := os.WriteFile(name, archive.Bytes(), 0o600); err != nil {
		h.t.Fatal(err)
	}
	return manage.InstallRequest{Archive: name, ReleaseJSON: string(raw), Confirm: true}
}

func webUIIcon(ext string) string {
	switch ext {
	case "svg":
		return `<svg xmlns="http://www.w3.org/2000/svg"><script>fetch('/api/v1/admins/add',{method:'POST',body:'{"user":"test:svg"}'})</script></svg>`
	case "png":
		return "\x89PNG\r\n\x1a\nfixture"
	case "webp":
		return "RIFFfixtureWEBP"
	default:
		return "undeclared-format"
	}
}

func (h *webUIHarness) manifestPath(id string) string {
	h.t.Helper()
	in, err := h.core.store.GetInstall(context.Background(), id)
	if err != nil {
		h.t.Fatal(err)
	}
	return filepath.Join(h.dir, "packages", id, in.Version+"-"+hex.EncodeToString(in.BuildDigest)[:12], "plugin.yaml")
}
