package manage_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adminauth"
	"github.com/AlonaBot/AlonaBot/internal/core/buttons"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/tool"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// TestToolPolicyIdentities 验证管理目录身份与真实状态库中的独立配置文件覆盖。
func TestToolPolicyIdentities(t *testing.T) {
	// 1. 管理 HTTP 与固定确认服务共享目录和临时 SQLite。
	dir := t.TempDir()
	store, err := state.Open(t.Context(), dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := store.AddInstall(t.Context(), state.Install{PluginID: "demo", Role: state.RolePlugin, BuildDigest: []byte{}}); err != nil {
		t.Fatal(err)
	}
	token, err := adminauth.InitToken(t.Context(), store, dir)
	if err != nil {
		t.Fatal(err)
	}
	routes := router.New(router.Options{})
	catalog := tool.New(store, routes)
	auth := adminauth.New(store, time.Now, &url.URL{Scheme: "http", Host: "localhost"}, dir)
	handler := manage.New(manage.Options{Store: store, Router: routes, Tools: catalog}).Handler(auth)
	request := func(method, path, body string) []byte {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
		return w.Body.Bytes()
	}
	const path = "/api/v1/tools/policies"
	if raw := request(http.MethodGet, path, ""); strings.TrimSpace(string(raw)) != `{"policies":[]}` {
		t.Fatalf("empty directory: %s", raw)
	}
	declarations := make([]*pb.RouteDescriptor, 0, 2)
	for _, name := range []string{"echo", "shared"} {
		declarations = append(declarations, &pb.RouteDescriptor{
			Kind: pb.RouteKind_ROUTE_KIND_SERVICE,
			Spec: &pb.RouteDescriptor_Service{Service: &pb.ServiceSpec{ServiceName: name, Tool: &pb.ToolSpec{
				DefaultPolicy: pb.ToolPolicy_TOOL_POLICY_ALLOW, Effect: pb.ToolEffect_TOOL_EFFECT_READ_ONLY,
			}}},
		})
	}
	routes.Attach(router.FromInstance(&host.Instance{PluginID: "demo", Routes: declarations}), router.Caller{PluginID: "demo"})
	mux := router.NewMux()
	catalog.RegisterInvoke(mux, tool.InvokeOptions{Buttons: buttons.New(store, time.Now, time.Minute, nil)})
	for _, operation := range []struct {
		name   string
		effect pb.ToolEffect
	}{{"mcp/fs/write", pb.ToolEffect_TOOL_EFFECT_DESTRUCTIVE}, {"shared", pb.ToolEffect_TOOL_EFFECT_WRITE}} {
		message := &pb.ToolConfirmRequest{Operation: operation.name, Effect: operation.effect}
		encoded, err := proto.Marshal(message)
		if err != nil {
			t.Fatal(err)
		}
		result := mux.For(router.Caller{PluginID: "demo"})(t.Context(), &wire.Inbound{
			Target: wire.ServiceToolConfirm, ParentCallID: 1, ParentAttach: &router.Invocation{},
			Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(message), SchemaVersion: 1, Bytes: encoded},
		})
		if result.Err != nil {
			t.Fatalf("confirmation: %v", result.Err)
		}
	}
	request(http.MethodPut, path, `{"items":[{"plugin_id":"demo","service":"mcp/fs/write","policy":"deny"},{"plugin_id":"demo","service":"shared","policy":"deny"},{"plugin_id":"demo","service":"stale","policy":"allow"}]}`)

	// 2. 四类完整名各占一行；重复读取保持身份顺序与既有合并值。
	type policyRow struct {
		PluginID   string   `json:"plugin_id"`
		Service    string   `json:"service"`
		Identities []string `json:"identities"`
		Default    string   `json:"default"`
		Override   string   `json:"override"`
		Effect     string   `json:"effect"`
	}
	check := func(want []policyRow) {
		t.Helper()
		raw := request(http.MethodGet, path, "")
		var result struct {
			Policies []policyRow `json:"policies"`
		}
		if err := json.Unmarshal(raw, &result); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(result.Policies, want) {
			t.Fatalf("directory: %s, want %+v", raw, want)
		}
	}
	want := []policyRow{
		{"demo", "echo", []string{"tool"}, "allow", "inherit", "read_only"},
		{"demo", "mcp/fs/write", []string{"operation"}, "allow", "deny", "destructive"},
		{"demo", "shared", []string{"tool", "operation"}, "allow", "deny", "write"},
		{"demo", "stale", []string{}, "inherit", "allow", "unspecified"},
	}
	check(want)
	check(want)

	// 3. 目录身份消失后，配置文件仍分别读写同名工具与操作覆盖。
	const profilePath = "/api/v1/profiles/default/tools"
	const overrides = `{"items":[{"tool":"demo/shared","policy":"allow"},{"operation":{"plugin_id":"demo","name":"shared"},"policy":"deny"}]}`
	request(http.MethodPut, profilePath, overrides)
	routes.Detach("demo")
	want = want[1:]
	want[1].Identities = []string{"operation"}
	want[1].Default = "allow"
	check(want)
	handler = manage.New(manage.Options{Store: store, Router: routes, Tools: tool.New(store, routes)}).Handler(auth)
	for index := range want {
		want[index].Identities = []string{}
		want[index].Default = "inherit"
		want[index].Effect = "unspecified"
	}
	check(want)
	if raw := request(http.MethodGet, profilePath, ""); strings.TrimSpace(string(raw)) != `{"items":[{"operation":{"plugin_id":"demo","name":"shared"},"policy":"deny"},{"tool":"demo/shared","policy":"allow"}]}` {
		t.Fatalf("saved profile identities: %s", raw)
	}
	request(http.MethodPut, profilePath, `{"items":[{"tool":"demo/shared","policy":"deny"},{"operation":{"plugin_id":"demo","name":"shared"},"policy":"allow"}]}`)
	request(http.MethodPut, profilePath, `{"items":[{"tool":"demo/shared","policy":"inherit"}]}`)
	if raw := request(http.MethodGet, profilePath, ""); strings.TrimSpace(string(raw)) != `{"items":[{"operation":{"plugin_id":"demo","name":"shared"},"policy":"allow"}]}` {
		t.Fatalf("independent profile update: %s", raw)
	}
}
