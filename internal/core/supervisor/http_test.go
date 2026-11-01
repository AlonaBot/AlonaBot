//go:build unix

package supervisor

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/httpproxy"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

type allowHTTP struct{}

func (allowHTTP) Authorize(*http.Request, string) bool { return true }

type httpEcho struct {
	Path, Query string
	Header      http.Header
	Calls       int32
	Environment map[string]string
}

// TestHTTPProcesses 驱动真实SDK插件验证统一代理、凭证隔离、停止排空与代次替换（V-22组件分项）。
func TestHTTPProcesses(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "httpplugin")
	if output, err := exec.Command("go", "build", "-o", binary, "./testdata/httpplugin").CombinedOutput(); err != nil {
		t.Fatalf("build HTTP plugin: %v\n%s", err, output)
	}
	store, spec := setup(t, binary)
	spec.Stop = host.StopTiming{Drain: 100 * time.Millisecond, Total: 2 * time.Second, OnStopAt: 1500 * time.Millisecond}
	spec.ExtraEnv = map[string]string{"HTTP_PROXY": "http://proxy.example:8080", "HTTPS_PROXY": "http://proxy.example:8080", "NO_PROXY": "localhost"}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	proxy := httpproxy.New(log, nil)
	launch := func() *Running {
		t.Helper()
		running, err := Launch(context.Background(), store, spec)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			result := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
			if result.Err != nil {
				t.Error(result.Err)
			}
		})
		return running
	}
	first := launch()
	bind := func(server *httpproxy.Server, r *Running) {
		server.Bind("demo", r.Instance.GenerationID, r.Instance.HTTPListener, r.Instance.HTTPProxyToken, r.Instance.Routes)
	}
	bind(proxy, first)
	send := func(server *httpproxy.Server, method, path string, want int) httpEcho {
		t.Helper()
		request := httptest.NewRequest(method, "http://external/plugins/demo"+path, nil)
		request.Header.Set(wire.HeaderProxyAuth, "forged")
		request.Header.Set("Authorization", "Bearer admin")
		request.Header.Set("Forwarded", "for=evil")
		request.Header.Set("X-Forwarded-For", "evil")
		request.Header.Set("X-Forwarded-Custom", "evil")
		request.Header.Set("Cookie", "alonabot_admin=secret; plugin=keep")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != want {
			t.Fatalf("%s %s: status %d want %d: %s", method, path, response.Code, want, response.Body)
		}
		var reply httpEcho
		if want == 200 {
			if err := json.Unmarshal(response.Body.Bytes(), &reply); err != nil {
				t.Fatal(err)
			}
		}
		return reply
	}
	hook := "/hook?state=test-state&query=%2Fvalue"
	reply := send(proxy, "POST", hook, 200)
	if reply.Path != "/hook" || reply.Query != "state=test-state&query=%2Fvalue" {
		t.Fatalf("rewritten URL: %+v", reply)
	}
	for _, name := range []string{wire.HeaderProxyAuth, "Authorization", "Forwarded", "X-Forwarded-For", "X-Forwarded-Custom"} {
		if reply.Header.Get(name) != "" {
			t.Fatalf("credential leaked: %s", name)
		}
	}
	if reply.Header.Get("Cookie") != "plugin=keep" {
		t.Fatalf("cookie leak: %v", reply.Header)
	}
	if reply.Environment["HTTP_PROXY"] != spec.ExtraEnv["HTTP_PROXY"] || reply.Environment["HTTPS_PROXY"] != spec.ExtraEnv["HTTPS_PROXY"] || reply.Environment["NO_PROXY"] != "localhost" {
		t.Fatalf("proxy environment: %v", reply.Environment)
	}
	send(proxy, "POST", "/hook?state=wrong", 401)
	send(proxy, "GET", "/hello", 401)
	send(proxy, "GET", "/static/private", 401)
	for _, path := range []string{"/static/file", "/static/../hello", "/static/%2e%2e/hello"} {
		send(proxy, "GET", path, 200)
	}
	allowed := httpproxy.New(log, allowHTTP{})
	bind(allowed, first)
	before := send(allowed, "GET", "/hello", 200).Calls
	address := net.JoinHostPort(first.Instance.HTTPListener.Host, fmt.Sprint(first.Instance.HTTPListener.Port))
	for _, credential := range []string{"", "wrong"} {
		request, err := http.NewRequest("GET", "http://"+address+"/hello", nil)
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set(wire.HeaderProxyAuth, credential)
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != 401 {
			t.Fatalf("direct endpoint status %d", response.StatusCode)
		}
	}
	if after := send(allowed, "GET", "/hello", 200).Calls; after != before+1 {
		t.Fatalf("unauthorized direct requests executed Handler: before=%d after=%d", before, after)
	}
	send(proxy, "GET", "/missing", 404)
	send(proxy, "GET", "/hook", 405)
	// 流式Handler先确认进入，再在停止期间写完尾部，验证Shutdown等待真实在途请求。
	request, err := http.NewRequest("GET", "http://"+address+"/static/file?wait=1", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set(wire.HeaderProxyAuth, hex.EncodeToString(first.Instance.HTTPProxyToken))
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewReader(response.Body)
	line, err := reader.ReadString('\n')
	if err != nil || line != "started\n" {
		t.Fatalf("handler did not start: %q %v", line, err)
	}
	proxy.Unbind("demo", first.Instance.GenerationID)
	result := first.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
	if result.Err != nil || result.Killed {
		t.Fatalf("HTTP stop: %+v", result)
	}
	tail, err := io.ReadAll(reader)
	response.Body.Close()
	if err != nil || string(tail) != "finished\n" {
		t.Fatalf("in-flight HTTP did not drain: %q %v", tail, err)
	}
	send(proxy, "POST", hook, 503)
	// 另一个代理保留旧快照，真实端点消失后应报告502。
	send(allowed, "GET", "/hello", 502)
	if connection, err := net.Dial("tcp", address); err == nil {
		connection.Close()
		t.Fatal("stopped plugin HTTP port is open")
	}
	// 占住旧端口，让新端点差异的断言独立于操作系统临时端口复用策略。
	reservation, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	defer reservation.Close()
	if err := store.AddGeneration(context.Background(), "demo", "gen-2", "{}", 1); err != nil {
		t.Fatal(err)
	}
	spec.GenerationID = "gen-2"
	spec.ExpectedOld = "gen-1"
	spec.Bootstrap.GenerationId = "gen-2"
	second := launch()
	bind(proxy, second)
	if second.Instance.HTTPListener.Port == first.Instance.HTTPListener.Port || bytes.Equal(second.Instance.HTTPProxyToken, first.Instance.HTTPProxyToken) {
		t.Fatal("endpoint or token reused")
	}
	proxy.Unbind("demo", "gen-1")
	send(proxy, "POST", hook, 200)
	// 真实断链后SDK退出并关闭端点，装配方撤销映射。
	if err := second.Instance.Session.Close(); err != nil {
		t.Fatal(err)
	}
	proxy.Unbind("demo", "gen-2")
	if _, _, err := second.WaitExit(); err != nil {
		t.Fatal(err)
	}
	send(proxy, "POST", hook, 503)
	secondAddress := net.JoinHostPort(second.Instance.HTTPListener.Host, fmt.Sprint(second.Instance.HTTPListener.Port))
	if connection, err := net.Dial("tcp", secondAddress); err == nil {
		connection.Close()
		t.Fatal("disconnected plugin HTTP port is open")
	}
}
