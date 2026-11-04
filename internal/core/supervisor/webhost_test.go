//go:build unix

package supervisor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/httpproxy"
	"github.com/AlonaBot/AlonaBot/internal/core/webhost"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

type webAdminSessions struct {
	t         *testing.T
	expiresAt time.Time
	valid     bool
}

func (s *webAdminSessions) Lookup(r *http.Request) (string, time.Time, bool) {
	if r.Host != "admin.test" {
		s.t.Errorf("admin Lookup called on Host %q", r.Host)
		return "", time.Time{}, false
	}
	cookie, err := r.Cookie("alonabot_admin")
	if err != nil || cookie.Value != "admin-session" || !s.valid {
		return "", time.Time{}, false
	}
	return "admin-session", s.expiresAt, true
}

func (s *webAdminSessions) Valid(sessionID string) bool {
	return s.valid && sessionID == "admin-session"
}

type webAuthorizer func(*http.Request, string) bool

func (a webAuthorizer) Authorize(r *http.Request, pluginID string) bool { return a(r, pluginID) }

// TestWebHostProcesses 验证真实 SDK 插件进程的双 Host、票据、派生撤销与响应策略（V-47 组件分项）。
func TestWebHostProcesses(t *testing.T) {
	// 1. 启动真实插件并将已激活的 Listener 绑定到每个受测 Host 的代理。
	binary := filepath.Join(t.TempDir(), "httpplugin")
	if output, err := exec.Command("go", "build", "-o", binary, "./testdata/httpplugin").CombinedOutput(); err != nil {
		t.Fatalf("build HTTP plugin: %v\n%s", err, output)
	}
	store, spec := setup(t, binary)
	running, err := Launch(context.Background(), store, spec)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		result := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
		if result.Err != nil || result.Killed {
			t.Errorf("stop HTTP plugin: %+v", result)
		}
	})
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	sessions := &webAdminSessions{t: t, expiresAt: now.Add(8 * time.Hour), valid: true}
	var logs bytes.Buffer
	setDefaultLogger(t, slog.New(slog.NewJSONHandler(&logs, nil)))
	adminURL, err := url.Parse("http://admin.test:7650")
	if err != nil {
		t.Fatal(err)
	}
	pluginURL, err := url.Parse("http://plugin.test:7650")
	if err != nil {
		t.Fatal(err)
	}
	newHost := func(second bool) *webhost.Host {
		var host *webhost.Host
		proxy := httpproxy.New(webAuthorizer(func(r *http.Request, id string) bool { return host.Authorize(r, id) }))
		proxy.Bind("demo", running.Instance.GenerationID, running.Instance.HTTPListener, running.Instance.HTTPProxyToken, running.Instance.Routes)
		host = webhost.New(webhost.Options{
			AdminBaseURL: adminURL, PluginBaseURL: pluginURL, SecondHost: second, Proxy: proxy, Sessions: sessions, Now: func() time.Time { return now },
			Admin: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/open" {
					ticket, err := host.IssueTicket("admin-session", "demo")
					if err != nil {
						t.Errorf("issue plugin ticket: %v", err)
						http.Error(w, "issue ticket failed", 500)
						return
					}
					if _, err := fmt.Fprint(w, ticket); err != nil {
						t.Errorf("write ticket: %v", err)
					}
					return
				}
				w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.WriteHeader(http.StatusNoContent)
			}),
		})
		return host
	}
	send := func(host http.Handler, method, target string, headers http.Header, want int) *httptest.ResponseRecorder {
		t.Helper()
		request := httptest.NewRequest(method, target, nil)
		for name, values := range headers {
			request.Header[name] = append([]string(nil), values...)
		}
		response := httptest.NewRecorder()
		host.ServeHTTP(response, request)
		if response.Code != want {
			t.Fatalf("%s %s: status %d want %d: %s", method, target, response.Code, want, response.Body)
		}
		if strings.Contains(request.URL.Path, "/plugins/") && response.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Fatalf("missing nosniff: %v", response.Header())
		}
		return response
	}
	host := newHost(true)
	pluginPath := "http://plugin.test/plugins/demo"
	adminCookie := http.Header{"Cookie": {"alonabot_admin=admin-session"}}
	issue := func() string {
		t.Helper()
		return send(host, "GET", "http://admin.test/open", adminCookie, 200).Body.String()
	}
	decode := func(response *httptest.ResponseRecorder) httpEcho {
		t.Helper()
		var reply httpEcho
		if err := json.Unmarshal(response.Body.Bytes(), &reply); err != nil {
			t.Fatal(err)
		}
		return reply
	}
	// 2. Host 与命名空间分派先于代理，插件主机始终没有管理 API。
	for _, target := range []string{"http://plugin.test/api/x", "http://plugin.test/", "http://admin.test/plugins/demo/hello"} {
		send(host, "GET", target, nil, 404)
	}
	send(host, "GET", "http://unknown.test/plugins/demo/hello", http.Header{"X-Forwarded-Host": {"plugin.test"}}, 421)
	send(host, "GET", pluginPath+"/hello", nil, 401)
	cors := send(host, "GET", "http://admin.test/api/x", http.Header{"Origin": {"http://plugin.test:7650"}}, 204)
	if cors.Header().Get("Access-Control-Allow-Origin") != "" || cors.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Fatalf("admin credential CORS leak: %v", cors.Header())
	}
	// 3. 票据兑换只建立 Cookie，不执行插件；私有路由授权后也不会收到框架凭证。
	before := decode(send(host, "POST", pluginPath+"/hook?state=test-state", nil, 200)).Calls
	ticket := issue()
	exchanged := send(host, "GET", ticket, nil, 303)
	if exchanged.Header().Get("Location") != "/plugins/demo/" || exchanged.Header().Get("Referrer-Policy") != "no-referrer" {
		t.Fatalf("exchange headers: %v", exchanged.Header())
	}
	cookies := exchanged.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("exchange cookies: %v", cookies)
	}
	cookie := cookies[0]
	if cookie.Name != "alonabot_plugin" || !cookie.HttpOnly || cookie.Secure || cookie.Path != "/" || cookie.Domain != "" || cookie.SameSite != http.SameSiteLaxMode || !cookie.Expires.Equal(sessions.expiresAt) {
		t.Fatalf("plugin cookie: %+v", cookie)
	}
	headers := http.Header{
		"Cookie":        {cookie.Name + "=" + cookie.Value + "; alonabot_admin=admin-session; __Host-alonabot_admin=admin-session; plugin=keep"},
		"Authorization": {"Bearer admin-secret"}, wire.HeaderProxyAuth: {"forged"}, "Forwarded": {"host=admin.test"}, "X-Forwarded-Host": {"admin.test"},
	}
	reply := decode(send(host, "GET", pluginPath+"/hello", headers, 200))
	if reply.Calls != before+1 || reply.Path != "/hello" || reply.Query != "" {
		t.Fatalf("ticket reached plugin or URL changed: %+v", reply)
	}
	if reply.Header.Get("Cookie") != "plugin=keep" {
		t.Fatalf("framework Cookie reached plugin: %v", reply.Header)
	}
	for _, name := range []string{"Authorization", wire.HeaderProxyAuth, "Forwarded", "X-Forwarded-Host"} {
		if reply.Header.Get(name) != "" {
			t.Fatalf("framework identity reached plugin: %s", name)
		}
	}
	send(host, "GET", ticket, nil, 403)
	expired := issue()
	now = now.Add(61 * time.Second)
	send(host, "GET", expired, nil, 403)
	outstanding := issue()
	sessions.valid = false
	send(host, "GET", pluginPath+"/hello", headers, 401)
	send(host, "GET", outstanding, nil, 403)
	sessions.valid = true
	// 4. 插件仅能设置自己的 host-only Cookie；主动内容只有第二 Host 能提供。
	response := send(host, "GET", pluginPath+"/static/cookies", nil, 200)
	if got := response.Header().Values("Set-Cookie"); len(got) != 1 || got[0] != "plugin=keep; HttpOnly; Path=/" {
		t.Fatalf("unfiltered plugin cookies: %v", got)
	}
	if response.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Fatal("plugin credential CORS header survived")
	}
	send(host, "GET", pluginPath+"/static/page", nil, 200)
	single := newHost(false)
	base := "http://admin.test/plugins/demo"
	jsonResponse := send(single, "POST", base+"/hook?state=test-state", nil, 200)
	decode(jsonResponse)
	if jsonResponse.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("JSON content type: %v", jsonResponse.Header())
	}
	for _, tc := range []struct {
		method, path string
		headers      http.Header
	}{
		{"GET", "/static/page", nil},
		{"POST", "/hook?state=test-state", http.Header{"Accept": {"text/html"}}},
	} {
		blocked := send(single, tc.method, base+tc.path, tc.headers, 409)
		if !strings.Contains(blocked.Body.String(), "plugin_base_url") || strings.Contains(blocked.Body.String(), "<html>") {
			t.Fatalf("active response body: %s", blocked.Body)
		}
		if blocked.Header().Get("Content-Security-Policy") != "default-src 'none'; frame-ancestors 'none'" {
			t.Fatalf("missing single-host CSP: %v", blocked.Header())
		}
	}
	// 通过真实公开 Listener 检查 net/http 的缺省嗅探，Recorder 无法覆盖这个分支。
	public := httptest.NewServer(single)
	defer public.Close()
	request, err := http.NewRequest(http.MethodGet, public.URL+"/plugins/demo/static/untyped", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Host = "admin.test"
	untyped, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(untyped.Body)
	untyped.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if untyped.StatusCode != 200 || untyped.Header.Get("Content-Type") != "application/octet-stream" || string(body) != "<html><body>plugin page</body></html>" {
		t.Fatalf("untyped plugin response: status %d headers %v body %q", untyped.StatusCode, untyped.Header, body)
	}
	if untyped.Header.Get("X-Content-Type-Options") != "nosniff" || untyped.Header.Get("Content-Security-Policy") != "default-src 'none'; frame-ancestors 'none'" {
		t.Fatalf("untyped plugin response policy: %v", untyped.Header)
	}
	filtered := send(single, "GET", base+"/static/cookies", nil, 200)
	if filtered.Header().Get("Content-Security-Policy") != "default-src 'none'; frame-ancestors 'none'" || len(filtered.Header().Values("Set-Cookie")) != 1 {
		t.Fatalf("plugin overrode response policy: %v", filtered.Header())
	}
	if jsonResponse.Header().Get("Content-Security-Policy") != "default-src 'none'; frame-ancestors 'none'" {
		t.Fatal("JSON response missing CSP")
	}
	if !strings.Contains(logs.String(), `"level":"WARN"`) || !strings.Contains(logs.String(), `"plugin_id":"demo"`) || !strings.Contains(logs.String(), `"cookie_name":"parent"`) {
		t.Fatalf("missing cookie warning: %s", &logs)
	}
	for _, target := range []string{ticket, expired, outstanding} {
		parsed, err := url.Parse(target)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(logs.String(), parsed.Query().Get("t")) {
			t.Fatal("ticket leaked in logs")
		}
	}
	if strings.Contains(logs.String(), cookie.Value) {
		t.Fatal("plugin session leaked in logs")
	}
}
