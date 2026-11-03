//go:build unix

package supervisor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/AlonaBot/AlonaBot/internal/core/adminauth"
	"github.com/AlonaBot/AlonaBot/internal/core/httpproxy"
	"github.com/AlonaBot/AlonaBot/internal/core/webhost"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// TestAdminAuthProcesses 联验真实管理认证、插件票据、凭证隔离与派生会话撤销（V-47）。
func TestAdminAuthProcesses(t *testing.T) {
	// 1. 复用正式插件启动入口，认证、代理和票据都使用实际实现。
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
	now := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)
	var logs bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&logs, nil))
	adminURL, err := url.Parse("http://admin.test:7650")
	if err != nil {
		t.Fatal(err)
	}
	pluginURL, err := url.Parse("http://plugin.test:7650")
	if err != nil {
		t.Fatal(err)
	}
	auth := adminauth.New(store, func() time.Time { return now }, adminURL, log)
	mux := http.NewServeMux()
	auth.Routes(mux)
	mux.Handle("POST /api/change", auth.Require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })))
	var host *webhost.Host
	mux.Handle("POST /api/open", auth.Require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _, ok := auth.Lookup(r)
		if !ok {
			t.Error("ticket handler received an invalid admin session")
			http.Error(w, "invalid session", 500)
			return
		}
		ticket, err := host.IssueTicket(id, "demo")
		if err != nil {
			t.Errorf("issue plugin ticket: %v", err)
			http.Error(w, "issue ticket failed", 500)
			return
		}
		if _, err := fmt.Fprint(w, ticket); err != nil {
			t.Errorf("write ticket: %v", err)
		}
	})))
	proxy := httpproxy.New(log, webAuthorizer(func(r *http.Request, id string) bool { return host.Authorize(r, id) }))
	proxy.Bind("demo", running.Instance.GenerationID, running.Instance.HTTPListener, running.Instance.HTTPProxyToken, running.Instance.Routes)
	host = webhost.New(log, webhost.Options{AdminBaseURL: adminURL, PluginBaseURL: pluginURL, SecondHost: true, Admin: mux, Proxy: proxy, Sessions: auth, Now: func() time.Time { return now }})
	send := func(t *testing.T, method, target, body string, headers http.Header, want int) *httptest.ResponseRecorder {
		t.Helper()
		request := httptest.NewRequest(method, target, strings.NewReader(body))
		for name, values := range headers {
			for _, value := range values {
				request.Header.Add(name, value)
			}
		}
		response := httptest.NewRecorder()
		host.ServeHTTP(response, request)
		if response.Code != want {
			t.Fatalf("%s %s: status %d want %d", method, request.URL.Path, response.Code, want)
		}
		return response
	}
	const origin = "http://admin.test:7650"
	login := func(t *testing.T, token string, want int) *httptest.ResponseRecorder {
		t.Helper()
		body, err := json.Marshal(map[string]string{"token": token})
		if err != nil {
			t.Fatal(err)
		}
		return send(t, "POST", origin+"/api/auth/login", string(body), http.Header{"Origin": {origin}, "Content-Type": {"application/json"}}, want)
	}
	login(t, "unused", 503)
	token, err := adminauth.InitToken(context.Background(), store)
	if err != nil {
		t.Fatal(err)
	}
	login(t, "wrong-token", 401)
	send(t, "POST", origin+"/api/auth/login", `{"token":"wrong"}`, http.Header{"Origin": {"http://plugin.test:7650"}, "Content-Type": {"application/json"}}, 403)
	send(t, "POST", "http://plugin.test:7650/api/auth/login", `{"token":"wrong"}`, http.Header{"Origin": {origin}, "Content-Type": {"application/json"}}, 404)
	const privatePage = "http://plugin.test:7650/plugins/demo/static/private"
	send(t, "GET", privatePage, "", nil, 401)
	secrets := []string{token}
	var successfulPluginCalls int32

	// 2. 每种撤销方式都从真实登录、签发及兑换开始，保持实际授权链完整。
	for _, revoke := range []string{"reset", "logout", "expiry"} {
		t.Run(revoke, func(t *testing.T) {
			response := login(t, token, 200)
			cookies := response.Result().Cookies()
			var reply struct {
				CSRFToken string `json:"csrf_token"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &reply); err != nil {
				t.Fatal(err)
			}
			if len(cookies) != 1 || cookies[0].Value == "" || reply.CSRFToken == "" {
				t.Fatal("login did not supply a cookie and CSRF token")
			}
			adminCookie := cookies[0]
			if !adminCookie.Expires.Equal(now.Add(8 * time.Hour)) {
				t.Fatal("admin cookie did not expire after 8h")
			}
			adminHeaders := http.Header{"Cookie": {adminCookie.Name + "=" + adminCookie.Value}, "Origin": {origin}, "X-Alonabot-CSRF": {reply.CSRFToken}}
			send(t, "POST", origin+"/api/change", "", http.Header{"Cookie": adminHeaders["Cookie"], "Origin": {origin}}, 401)
			send(t, "POST", origin+"/api/change", "", adminHeaders, 204)
			send(t, "POST", origin+"/api/change", "", http.Header{"Authorization": {"Bearer " + token}}, 204)
			// Bearer 和管理 Cookie 不直接授予插件路由权限。
			send(t, "GET", privatePage, "", http.Header{"Cookie": adminHeaders["Cookie"], "Authorization": {"Bearer " + token}}, 401)
			issue := func() string {
				t.Helper()
				return send(t, "POST", origin+"/api/open", "", adminHeaders, 200).Body.String()
			}
			ticket := issue()
			exchanged := send(t, "GET", ticket, "", nil, 303)
			if exchanged.Header().Get("Location") != "/plugins/demo/" || exchanged.Header().Get("Referrer-Policy") != "no-referrer" {
				t.Fatal("ticket exchange did not clear the URL safely")
			}
			pluginCookies := exchanged.Result().Cookies()
			if len(pluginCookies) != 1 || !pluginCookies[0].Expires.Equal(adminCookie.Expires) {
				t.Fatal("plugin session did not inherit the admin expiration")
			}
			pluginCookie := pluginCookies[0]
			headers := http.Header{"Cookie": {pluginCookie.Name + "=" + pluginCookie.Value + "; " + adminCookie.Name + "=" + adminCookie.Value + "; __Host-alonabot_admin=" + adminCookie.Value + "; plugin=keep"}, "Authorization": {"Bearer " + token}, wire.HeaderProxyAuth: {"forged"}}
			page := send(t, "GET", privatePage, "", headers, 200)
			var echoed httpEcho
			if err := json.Unmarshal(page.Body.Bytes(), &echoed); err != nil {
				t.Fatal(err)
			}
			successfulPluginCalls++
			if echoed.Path != "/static/private" || echoed.Calls != successfulPluginCalls || echoed.Header.Get("Cookie") != "plugin=keep" || echoed.Header.Get("Authorization") != "" || echoed.Header.Get(wire.HeaderProxyAuth) != "" {
				t.Fatal("plugin received management credentials or ticket traffic")
			}
			send(t, "GET", ticket, "", nil, 403)
			pending := issue()
			secrets = append(secrets, adminCookie.Value, reply.CSRFToken, pluginCookie.Value)
			for _, value := range []string{ticket, pending} {
				parsed, err := url.Parse(value)
				if err != nil {
					t.Fatal(err)
				}
				credential := parsed.Query().Get("t")
				if credential == "" {
					t.Fatal("ticket has no credential")
				}
				secrets = append(secrets, credential)
			}
			switch revoke {
			case "reset":
				previous := token
				token, err = adminauth.ResetToken(context.Background(), store)
				if err != nil {
					t.Fatal(err)
				}
				secrets = append(secrets, token)
				send(t, "POST", origin+"/api/change", "", http.Header{"Authorization": {"Bearer " + previous}}, 401)
				send(t, "POST", origin+"/api/change", "", http.Header{"Authorization": {"Bearer " + token}}, 204)
			case "logout":
				send(t, "POST", origin+"/api/auth/logout", "", adminHeaders, 204)
			case "expiry":
				now = now.Add(8 * time.Hour)
			}
			if auth.Valid(adminCookie.Value) {
				t.Fatal("revoked admin session remained valid")
			}
			send(t, "POST", origin+"/api/change", "", adminHeaders, 401)
			send(t, "GET", privatePage, "", headers, 401)
			send(t, "GET", pending, "", nil, 403)
		})
	}
	for _, secret := range secrets {
		if strings.Contains(logs.String(), secret) {
			t.Fatal("authentication credential leaked in Core HTTP logs")
		}
	}
}
