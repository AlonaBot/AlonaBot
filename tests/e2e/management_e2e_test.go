//go:build unix

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/buttons"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"google.golang.org/protobuf/proto"
)

// TestT09ManagementFlow 通过真实管理入口核对版本、配置与业务数据跨完整安装流程的连续性。
func TestT09ManagementFlow(t *testing.T) {
	// 1. 复用恢复插件的 SQLite 读写，只给临时产物补充程序版本与配置的观测入口。
	const id, source = "recovery", "github.com/example/recovery"
	const schema = `{"label":{"type":"string","default":"initial"}}`
	packageAt := func(version string) string {
		t.Helper()
		dir := t.TempDir()
		pluginSource := strings.Replace(recoveryFixtureSource, "genimpl.Resources{}", fmt.Sprintf("genimpl.Resources{ConfigSchema:[]byte(%q)}", schema), 1)
		pluginSource = strings.Replace(pluginSource, "p:=sdk.NewPlugin()", fmt.Sprintf(`p:=sdk.NewPlugin()
 p.Route("GET","/info").HandleFunc(func(w http.ResponseWriter,r *http.Request){fmt.Fprintf(w,"%%s %%s",%q,p.Config().String("label"))})`, version), 1)
		pluginSource = strings.Replace(pluginSource, "p.OnReady(func(context.Context)error{", `p.OnReady(func(context.Context)error{slog.Info("OnReady");`, 1)
		name := filepath.Join(dir, "main.go")
		if err := os.WriteFile(name, []byte(pluginSource), 0o600); err != nil {
			t.Fatal(err)
		}
		executable := filepath.Join(dir, "plugin")
		command := exec.Command("go", fixtureBuildArgs("-o", executable, name)...)
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("build management fixture: %v\n%s", err, output)
		}
		manifest := fmt.Sprintf("manifest_version: 1\nid: %s\nname: Management Fixture\nversion: %s\nsource: %s\nrole: plugin\nprotocol_min_major: 1\nprotocol_max_major: 1\nrequired_capabilities: []\nentrypoint: bin/component\nmigration_mode: none\nprepare_effects: managed_db\nos: %s\narch: %s\n", id, version, source, runtime.GOOS, runtime.GOARCH)
		archive, err := writeTestPackage(dir, id, "plugin", "", source, version, executable, "", map[string][]byte{
			"plugin.yaml": []byte(manifest), "plugin_config_schema.json": []byte(schema),
		})
		if err != nil {
			t.Fatal(err)
		}
		return archive
	}
	firstPackage, secondPackage, thirdPackage := packageAt("1.0.0"), packageAt("2.0.0"), packageAt("3.0.0")
	dir := t.TempDir()
	c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0", "--log-format", "json")
	expectState := func(version, label, value string) {
		t.Helper()
		for path, want := range map[string]string{"/info": version + " " + label, "/value": value} {
			status, raw := fixtureHTTP(t, c, dir, http.MethodGet, "/plugins/"+id+path)
			if status != http.StatusOK || string(raw) != want {
				t.Fatalf("GET %s: %d %s; want %s", path, status, raw, want)
			}
		}
	}
	nextSession := func(previous string) string {
		t.Helper()
		var session string
		// 管理响应和 stderr 由不同连接交付，等日志读端观察到本次 OnReady。
		recoveryWait(t, func() bool { session = lastSession(t, c); return session != previous })
		return session
	}
	writeValue := func(value string) {
		t.Helper()
		status, body := fixtureHTTP(t, c, dir, http.MethodPut, "/plugins/"+id+"/value?n="+value)
		if status != http.StatusNoContent {
			t.Fatalf("write business database: %d %s", status, body)
		}
	}
	c.management(t, dir, "", "install", firstPackage, "--yes")
	installed := findPackage(t, c, dir, id)
	if installed.Enabled || installed.Active != "" || installed.Version != "1.0.0" || installed.Signed || installed.Source != source {
		t.Fatalf("installed package: %+v", installed)
	}
	c.management(t, dir, "", "enable", id)
	first := findPackage(t, c, dir, id)
	if !first.Enabled || first.Active == "" || first.State != "Activated" {
		t.Fatalf("enabled package: %+v", first)
	}
	session := nextSession("")
	expectState("1.0.0", "initial", "1")
	writeValue("41")

	// 2. restart 保留 Active 和生效配置；编辑文件只有经过 CLI reload 才进入新代次。
	pluginDir := filepath.Join(dir, "plugins", id)
	databaseFile := func(generation string) os.FileInfo {
		t.Helper()
		file, err := os.Stat(filepath.Join(pluginDir, "generations", generation, "db.sqlite"))
		if err != nil {
			t.Fatal(err)
		}
		return file
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "config.json"), []byte(`{"label":"reloaded"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	expectState("1.0.0", "initial", "41")
	c.management(t, dir, "", "restart", id)
	restarted := findPackage(t, c, dir, id)
	if restarted.Active != first.Active || restarted.State != "Activated" {
		t.Fatalf("restart changed generation: %+v", restarted)
	}
	session = nextSession(session)
	expectState("1.0.0", "initial", "41")
	c.management(t, dir, "", "reload", id)
	reloaded := findPackage(t, c, dir, id)
	if reloaded.Active == first.Active || reloaded.State != "Activated" {
		t.Fatalf("reload did not commit a new generation: %+v", reloaded)
	}
	session = nextSession(session)
	expectState("1.0.0", "reloaded", "41")
	if os.SameFile(databaseFile(first.Active), databaseFile(reloaded.Active)) {
		t.Fatal("reload reused the original database")
	}

	// 3. 更新真实程序版本，pin 拒绝候选时保留运行者，解除后同一候选正常提交。
	c.management(t, dir, "", "update", id, "--archive", secondPackage, "--yes")
	updated := findPackage(t, c, dir, id)
	if updated.Active == reloaded.Active || updated.Version != "2.0.0" || updated.State != "Activated" {
		t.Fatalf("updated package: %+v", updated)
	}
	session = nextSession(session)
	expectState("2.0.0", "reloaded", "41")

	// 在之后仍会保留的代次签发按钮，避免把快照淘汰造成的 token 删除当作代次校验。
	stopTestCore(t, c)
	store := recoveryStore(t, dir)
	newButtons := func() *buttons.Service {
		return buttons.New(store, time.Now, time.Hour, func(pluginID string) (string, []*pb.RouteDescriptor, bool) {
			active, err := store.Active(t.Context(), pluginID)
			if err != nil {
				t.Fatal(err)
			}
			return active, []*pb.RouteDescriptor{{RouteId: "button.check", Kind: pb.RouteKind_ROUTE_KIND_EVENT, Spec: &pb.RouteDescriptor_Event{Event: &pb.EventSpec{EventTypeFilter: []string{"button.clicked"}, ActionId: "check", ActionSchemaVersion: 1}}}}, true
		})
	}
	buttonService := newButtons()
	target := &pb.Target{AdapterId: "test", AdapterInstanceId: "instance", AccountId: "account", ChatId: "chat"}
	chain := &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Buttons{Buttons: &pb.ButtonsSegment{Rows: []*pb.ButtonRow{{Buttons: []*pb.Button{{ActionId: "check", Value: proto.String(`{}`)}}}}}}}}}
	if err := buttonService.Issue(t.Context(), router.NewCaller(id, updated.Active), target, chain); err != nil {
		t.Fatal(err)
	}
	token := chain.Segments[0].GetButtons().Rows[0].Buttons[0].GetCoreToken()
	click := &pb.Event{EventType: "button.clicked", Target: target, Actor: &pb.Actor{ExternalUserId: "alice", AdapterInstanceId: "instance"}, EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_ButtonClicked{ButtonClicked: &pb.ButtonClickedPayload{Value: token}}}}
	if _, _, err := buttonService.Resolve(t.Context(), proto.Clone(click).(*pb.Event)); err != nil {
		t.Fatalf("original button rejected: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	c = startEmptyCore(t, dir, "--listen", "127.0.0.1:0", "--log-format", "json")
	session = nextSession(session)
	expectState("2.0.0", "reloaded", "41")
	c.management(t, dir, "", "pin", id, "2.0.0")
	if pinned := findPackage(t, c, dir, id); pinned.Pin != "2.0.0" {
		t.Fatalf("pin was not saved: %+v", pinned)
	}
	output := managementFailure(t, c, dir, "update", id, "--archive", thirdPackage, "--yes")
	if !strings.Contains(output, "candidate version does not match pin: 2.0.0") {
		t.Fatalf("unexpected pinned update failure: %s", output)
	}
	if rejected := findPackage(t, c, dir, id); rejected.Active != updated.Active || rejected.Version != "2.0.0" || lastSession(t, c) != session {
		t.Fatalf("pinned update changed running package: %+v", rejected)
	}
	expectState("2.0.0", "reloaded", "41")
	c.management(t, dir, "", "pin", id)
	c.management(t, dir, "", "update", id, "--archive", thirdPackage, "--yes")
	unpinned := findPackage(t, c, dir, id)
	if unpinned.Pin != "" || unpinned.Active == updated.Active || unpinned.Version != "3.0.0" || unpinned.State != "Activated" {
		t.Fatalf("unpinned update: %+v", unpinned)
	}
	session = nextSession(session)
	expectState("3.0.0", "reloaded", "41")
	writeValue("99")

	// 4. 缺省卸载和同来源重装沿用原代次、配置与业务写入。
	c.management(t, dir, "", "disable", id)
	if disabled := findPackage(t, c, dir, id); disabled.Enabled || disabled.Active != unpinned.Active || disabled.State != "Stopped" {
		t.Fatalf("disabled package: %+v", disabled)
	}
	c.management(t, dir, "", "uninstall", id, "--yes")
	if removed := findPackage(t, c, dir, id); !removed.Uninstalled || removed.Enabled || removed.Active != unpinned.Active {
		t.Fatalf("uninstalled package: %+v", removed)
	}
	c.management(t, dir, "", "install", thirdPackage, "--yes")
	if reinstalled := findPackage(t, c, dir, id); reinstalled.Uninstalled || reinstalled.Enabled || reinstalled.Active != unpinned.Active || reinstalled.Source != source {
		t.Fatalf("reinstalled package: %+v", reinstalled)
	}
	c.management(t, dir, "", "enable", id)
	session = nextSession(session)
	expectState("3.0.0", "reloaded", "99")

	// 5. 从管理 API 的真实保留列表选择旧快照，显式回退得到新代次及独立 DB。
	status, raw := fixtureHTTP(t, c, dir, http.MethodGet, "/api/v1/plugins/"+id+"/generations")
	var generations struct {
		Generations []struct {
			ID      string `json:"id"`
			Version string `json:"version"`
			HasDB   bool   `json:"has_db"`
		} `json:"generations"`
	}
	if status != http.StatusOK {
		t.Fatalf("list retained generations: %d %s", status, raw)
	}
	if err := json.Unmarshal(raw, &generations); err != nil {
		t.Fatal(err)
	}
	var selected string
	for _, generation := range generations.Generations {
		if generation.ID == updated.Active && generation.Version == "2.0.0" && generation.HasDB {
			selected = generation.ID
		}
	}
	if selected == "" {
		t.Fatalf("rollback source was not retained: %s", raw)
	}
	c.management(t, dir, "", "rollback", id, selected, "--yes")
	rolled := findPackage(t, c, dir, id)
	if rolled.Active == selected || rolled.Active == unpinned.Active || rolled.Version != "2.0.0" || !rolled.Enabled || rolled.State != "Activated" {
		t.Fatalf("rollback package: %+v", rolled)
	}
	nextSession(session)
	expectState("2.0.0", "reloaded", "41")
	stopTestCore(t, c)
	store = recoveryStore(t, dir)
	buttonService = newButtons()
	if _, _, err := buttonService.Resolve(t.Context(), click); err == nil || !strings.Contains(err.Error(), "generation is no longer active") {
		t.Fatalf("rollback revived old button: %v", err)
	}
	active, err := store.Active(t.Context(), id)
	if err != nil || active != rolled.Active {
		t.Fatalf("persisted Active: %s, %v; want %s", active, err, rolled.Active)
	}
	records, err := store.Journals(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]struct{ operation, old string }{
		first.Active:    {"first_install", ""},
		reloaded.Active: {"reload", first.Active},
		updated.Active:  {"update", reloaded.Active},
		unpinned.Active: {"update", updated.Active},
		rolled.Active:   {"rollback", unpinned.Active},
	}
	if len(records) != len(expected) {
		t.Fatalf("unexpected candidate operations: %+v", records)
	}
	for _, record := range records {
		want := expected[record.NewGeneration]
		if record.Kind != want.operation || record.ExpectedOld != want.old || record.Phase != "committed" || record.DatabasePath != filepath.Join(pluginDir, "generations", record.NewGeneration, "db.sqlite") {
			t.Fatalf("inconsistent committed operation: %+v; want %+v", record, want)
		}
	}
	if os.SameFile(databaseFile(selected), databaseFile(rolled.Active)) {
		t.Fatal("rollback reused the retained database")
	}
}
