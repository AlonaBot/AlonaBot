//go:build unix

package app

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/candidate"
	"github.com/AlonaBot/AlonaBot/internal/core/secret"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/oslock"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// TestCandidateJournalResources 验证 Journal、代次、secret 与真实 SQLite/WAL 的资源交接。
func TestCandidateJournalResources(t *testing.T) {
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
