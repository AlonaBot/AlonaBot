//go:build unix

package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/candidate"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/core/secret"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/generate"
	"github.com/AlonaBot/AlonaBot/internal/oslock"
	"github.com/AlonaBot/AlonaBot/internal/pluginkit"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

// TestCandidateJournalResources 验证 Journal、代次、secret 与真实 SQLite/WAL 的资源交接。
func TestCandidateJournalResources(t *testing.T) {
	t.Run("committed_route_snapshot_after_core_restart", testCommittedRouteSnapshot)
	ctx := context.Background()
	dir := t.TempDir()
	st, err := state.Open(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	const id = "journal-resource"
	if err := st.AddInstall(ctx, state.Install{PluginID: id, Role: state.RolePlugin, Source: "local", BuildDigest: []byte{1}}); err != nil {
		t.Fatal(err)
	}
	secrets := secret.New(st)
	schema := []byte(`{"token":{"type":"string","description":"Token","secret":true}}`)
	oldConfig, paths, err := secrets.Extract(ctx, id, schema, []byte(`{"token":"old-secret","unknown":0}`), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.AddGeneration(ctx, id, "old", string(oldConfig)); err != nil {
		t.Fatal(err)
	}
	if err := st.SetGenerationConfig(ctx, "old", string(oldConfig), paths, nil); err != nil {
		t.Fatal(err)
	}
	if err := st.CommitActive(ctx, id, "", "old"); err != nil {
		t.Fatal(err)
	}
	pluginDir := filepath.Join(dir, "plugins", id)
	oldPath := filepath.Join(pluginDir, "generations", "old", "db.sqlite")
	if err := os.MkdirAll(filepath.Dir(oldPath), 0o700); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", oldPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec("PRAGMA journal_mode=WAL; PRAGMA wal_autocheckpoint=0; CREATE TABLE value (v TEXT); INSERT INTO value VALUES ('old-wal')"); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(oldPath + "-wal"); err != nil || info.Size() == 0 {
		t.Fatalf("WAL missing: %v", err)
	}
	record := state.Journal{OpID: "resource-operation", PluginID: id, Kind: "reload", ExpectedOld: "old", NewGeneration: "new", Phase: "started", PackageDir: filepath.Join(dir, "packages", id, "old")}
	if err := st.CreateJournal(ctx, record); err != nil {
		t.Fatal(err)
	}
	record.Phase = "stopped_old"
	if err := st.UpdateJournal(ctx, record); err != nil {
		t.Fatal(err)
	}
	// 系统锁归 Core 短暂持有，交接后同一路径可再次加锁。
	lock, err := oslock.TryLock(filepath.Join(pluginDir, wire.RuntimeLockFile))
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.Unlock(); err != nil {
		t.Fatal(err)
	}
	prepared, err := candidate.PrepareDatabase(ctx, candidate.Request{PluginID: id, PluginDir: pluginDir, Reason: pb.ActivationReason_ACTIVATION_REASON_RELOAD, FromGeneration: "old", NewGeneration: "new"})
	if err != nil {
		t.Fatal(err)
	}
	if prepared.DatabasePath == oldPath {
		t.Fatal("candidate reused old database")
	}
	copyDB, err := sql.Open("sqlite", prepared.DatabasePath)
	if err != nil {
		t.Fatal(err)
	}
	var value string
	if err := copyDB.QueryRow("SELECT v FROM value").Scan(&value); err != nil || value != "old-wal" {
		t.Fatalf("copied WAL value %q: %v", value, err)
	}
	if _, err := copyDB.Exec("UPDATE value SET v = 'candidate'"); err != nil {
		t.Fatal(err)
	}
	if err := copyDB.Close(); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT v FROM value").Scan(&value); err != nil || value != "old-wal" {
		t.Fatalf("old value %q: %v", value, err)
	}
	nextConfig, nextPaths, err := secrets.Extract(ctx, id, schema, []byte(`{"token":"candidate-secret","unknown":0}`), oldConfig, paths)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.AddGeneration(ctx, id, "new", string(nextConfig)); err != nil {
		t.Fatal(err)
	}
	if err := st.SetGenerationConfig(ctx, "new", string(nextConfig), nextPaths, nil); err != nil {
		t.Fatal(err)
	}
	record.DatabasePath = prepared.DatabasePath
	record.Phase = "candidate_ready"
	if err := st.UpdateJournal(ctx, record); err != nil {
		t.Fatal(err)
	}
	if err := st.PutPendingConfig(ctx, state.PendingConfig{PluginID: id, BaseGeneration: "old", ConfigJSON: string(oldConfig)}); err != nil {
		t.Fatal(err)
	}
	record.Phase = "failed"
	if err := st.UpdateJournal(ctx, record); err != nil {
		t.Fatal(err)
	}
	if err := candidate.Discard(ctx, st, secrets, prepared); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Dir(prepared.DatabasePath)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("candidate directory remains: %v", err)
	}
	record.DatabasePath = ""
	record.PackageDir = ""
	record.Phase = "cleaned"
	if err := st.UpdateJournal(ctx, record); err != nil {
		t.Fatal(err)
	}
	persisted, err := st.GetJournal(ctx, record.OpID)
	if err != nil || persisted.Phase != "cleaned" || persisted.ExpectedOld != "old" || persisted.DatabasePath != "" || persisted.CreatedAt.IsZero() {
		t.Fatalf("persisted journal %+v: %v", persisted, err)
	}
	if active, err := st.Active(ctx, id); err != nil || active != "old" {
		t.Fatalf("active %q: %v", active, err)
	}
	snapshot, err := st.GenerationConfig(ctx, "old")
	if err != nil || snapshot != string(oldConfig) {
		t.Fatalf("old snapshot changed: %v", err)
	}
	resolved, err := secrets.Resolve(ctx, id, []byte(snapshot))
	if err != nil || !bytes.Contains(resolved, []byte("old-secret")) {
		t.Fatalf("old secret lost: %v", err)
	}
	retained, err := st.ListSecrets(ctx, id)
	if err != nil || len(retained) != 1 {
		t.Fatalf("retained secret count %d: %v", len(retained), err)
	}
	lock, err = oslock.TryLock(filepath.Join(pluginDir, wire.RuntimeLockFile))
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.Unlock(); err != nil {
		t.Fatal(err)
	}
}

func testCommittedRouteSnapshot(t *testing.T) {
	ctx := t.Context()
	const id = "demo"
	const configJSON = `{"extra":"route-snapshot-config-marker"}`
	out := t.TempDir()
	archive, err := pluginkit.Pack(ctx, generate.Options{Dir: "../../examples/plugins/lifecycle"}, "https://example.invalid/lifecycle.tar.gz", "", out)
	if err != nil {
		t.Fatal(err)
	}
	releaseJSON, err := os.ReadFile(filepath.Join(out, "release.json"))
	if err != nil {
		t.Fatal(err)
	}
	h := newWebUIHarness(t)
	h.start(nil)
	core := h.core
	manager := manage.New(manage.Options{Store: core.store, Settings: core.assembly.settings, Packages: filepath.Join(h.dir, "packages"),
		Limits: packages.NewLimits(core.values), Adapters: core.assembly.adapters, State: core.assembly.componentState})
	if _, err := manager.Install(ctx, manage.InstallRequest{Archive: archive, ReleaseJSON: string(releaseJSON), Confirm: true}); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(h.dir, "plugins", id)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(configJSON), 0o600); err != nil {
		t.Fatal(err)
	}
	h.action("enable", id, 200)
	var stopped state.RouteSnapshot
	if err := core.operation(ctx, func(ctx context.Context) error {
		instance := core.components[0].instance
		wantRoutes := instance.Routes
		if err := core.stopManaged(id, pb.StopReason_STOP_REASON_RESTART); err != nil {
			return err
		}
		stopped, err = core.store.ActiveRouteSnapshot(ctx, id)
		if err != nil {
			return err
		}
		install, err := core.store.GetInstall(ctx, id)
		if err != nil {
			return err
		}
		operations, err := core.store.Journals(ctx, id)
		if err != nil {
			return err
		}
		if len(operations) != 1 || operations[0].Phase != "committed" || operations[0].NewGeneration != stopped.GenerationID || stopped.GenerationID != instance.GenerationID ||
			stopped.PluginID != id || !bytes.Equal(stopped.BuildDigest, install.BuildDigest) || !bytes.Equal(stopped.BuildDigest, operations[0].BuildDigest) || len(wantRoutes) < 2 || !slices.EqualFunc(stopped.Routes, wantRoutes, equalSnapshotRoute) {
			return errors.New("snapshot differs from the committed program or complete registration")
		}
		generation, routes, ok := core.assembly.router.RouteSnapshot(id)
		if _, live := core.assembly.router.Caller(id); live || !ok || generation != stopped.GenerationID || !slices.EqualFunc(routes, wantRoutes, equalSnapshotRoute) {
			return errors.New("stopping the session did not preserve the committed route index alone")
		}
		var expected []byte
		hash := sha256.New() // WIRE 的路由摘要覆盖按序拼接的确定性描述。
		for _, route := range wantRoutes {
			data, err := (proto.MarshalOptions{Deterministic: true}).Marshal(route)
			if err != nil {
				return err
			}
			hash.Write(data)
			expected = protowire.AppendBytes(expected, data)
		}
		if !bytes.Equal(stopped.RouteHash, hash.Sum(nil)) {
			return errors.New("route digest differs from the WIRE digest")
		}
		config, err := core.store.GenerationConfig(ctx, generation)
		if err != nil {
			return err
		}
		if config != configJSON || instance.HTTPListener == nil || len(instance.HTTPProxyToken) == 0 || len(operations[0].SessionID) == 0 {
			return errors.New("test registration did not contain configuration, a listener and session authentication")
		}
		db, err := sql.Open("sqlite", filepath.Join(h.dir, "state.sqlite"))
		if err != nil {
			return err
		}
		defer db.Close()
		var persisted []byte
		if err := db.QueryRowContext(ctx, "SELECT routes FROM route_snapshot WHERE generation_id = ?", generation).Scan(&persisted); err != nil {
			return err
		}
		if !bytes.Equal(persisted, expected) {
			return errors.New("persisted snapshot contains fields outside the route descriptors")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	h.stop()
	h.start(nil)
	generation, routes, ok := h.core.assembly.router.RouteSnapshot(id)
	plugin, live := h.core.assembly.router.Plugin(id)
	caller, bound := h.core.assembly.router.Caller(id)
	if !ok || !live || !plugin.Activated() || !bound || generation != stopped.GenerationID || caller.GenerationID != generation || !slices.EqualFunc(routes, stopped.Routes, equalSnapshotRoute) {
		t.Fatalf("restored routes or activation mismatch: generation=%q snapshot=%t live=%t caller=%t", generation, ok, live, bound)
	}
}

func equalSnapshotRoute(a, b *pb.RouteDescriptor) bool { return proto.Equal(a, b) }
