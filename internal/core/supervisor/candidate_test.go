//go:build unix

package supervisor

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/candidate"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/secret"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// TestDatabaseCandidateProcesses 验证候选隔离、进程锁与 WAL 写入的完整交接（V-24、V-36）。
func TestDatabaseCandidateProcesses(t *testing.T) {
	binary := buildDatabasePlugins(t)["sql"]
	t.Run("failed reload and restart", func(t *testing.T) {
		ctx := t.Context()
		store, spec := setup(t, binary)
		spec.Stdio = true
		spec.ExtraEnv = map[string]string{"GORACE": "atexit_sleep_ms=0"}
		unlock := store.LockPlugin(spec.PluginID)
		prepared, err := candidate.PrepareDatabase(ctx, candidate.Request{PluginID: spec.PluginID, PluginDir: spec.PluginDir, Reason: pb.ActivationReason_ACTIVATION_REASON_FIRST_INSTALL, NewGeneration: spec.GenerationID})
		unlock()
		if err != nil {
			t.Fatal(err)
		}
		spec.Bootstrap.DatabaseDescriptor = &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_SQLITE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_VERSIONED, ActivationReason: pb.ActivationReason_ACTIVATION_REASON_FIRST_INSTALL, SqlitePath: prepared.DatabasePath}
		run, err := Launch(ctx, store, spec)
		if err != nil {
			t.Fatalf("first launch: %v\n%s", err, spec.Log.(*logBuf).String())
		}
		// 在任何断言失败时也回收真实插件，正常停止后 Stop 可以重复调用。
		t.Cleanup(func() { run.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE) })
		request := candidate.Request{PluginID: spec.PluginID, PluginDir: spec.PluginDir, Reason: pb.ActivationReason_ACTIVATION_REASON_RELOAD, FromGeneration: spec.GenerationID, NewGeneration: "gen-2"}
		unlock = store.LockPlugin(spec.PluginID)
		_, err = candidate.PrepareDatabase(ctx, request)
		unlock()
		var control *validate.Error
		if !errors.As(err, &control) || control.Code != pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION || control.Message != "plugin process still holds runtime.lock" {
			t.Fatalf("running writer: %v", err)
		}
		if _, err := os.Stat(filepath.Join(spec.PluginDir, "generations", request.NewGeneration)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("locked candidate exists: %v", err)
		}
		result := run.Stop(time.Now(), pb.StopReason_STOP_REASON_RELOAD)
		run.Process.WaitLogs()
		if result.Err != nil || result.Killed {
			t.Fatalf("stop old plugin: %+v", result)
		}
		snapshot, err := store.GenerationConfig(ctx, spec.GenerationID)
		if err != nil {
			t.Fatal(err)
		}
		// 真实 WAL 连接保持打开且禁用自动检查点，备份须包含主文件之外的写入。
		source := databaseFile(t, prepared.DatabasePath)
		source.SetMaxOpenConns(1)
		for _, query := range []string{"PRAGMA journal_mode=WAL", "PRAGMA wal_autocheckpoint=0", "INSERT INTO activity VALUES ('wal-one'),('wal-two')"} {
			if _, err := source.ExecContext(ctx, query); err != nil {
				t.Fatal(err)
			}
		}
		wal, err := os.Stat(prepared.DatabasePath + "-wal")
		if err != nil || wal.Size() == 0 {
			t.Fatalf("expected live WAL: %v", err)
		}
		before := candidateActivity(t, source)
		if !slices.Equal(before, []string{"migration", "start", "wal-one", "wal-two"}) {
			t.Fatalf("initial activity: %v", before)
		}
		unlock = store.LockPlugin(spec.PluginID)
		next, err := candidate.PrepareDatabase(ctx, request)
		unlock()
		if err != nil {
			t.Fatal(err)
		}
		originalFile, err := os.Stat(prepared.DatabasePath)
		if err != nil {
			t.Fatal(err)
		}
		candidateFile, err := os.Stat(next.DatabasePath)
		if err != nil {
			t.Fatal(err)
		}
		if !next.Created || os.SameFile(originalFile, candidateFile) {
			t.Fatal("reload did not create an independent file")
		}
		copied := databaseFile(t, next.DatabasePath)
		if got := candidateActivity(t, copied); !slices.Equal(got, before) {
			t.Fatalf("WAL backup activity: %v; want %v", got, before)
		}
		if err := copied.Close(); err != nil {
			t.Fatal(err)
		}
		if err := source.Close(); err != nil {
			t.Fatal(err)
		}
		if err := store.AddGeneration(ctx, spec.PluginID, next.Generation, `{}`, 1); err != nil {
			t.Fatal(err)
		}
		failed := spec
		failed.GenerationID = next.Generation
		failed.ExpectedOld = spec.GenerationID
		failed.ExtraEnv = map[string]string{"GORACE": "atexit_sleep_ms=0", "DB_FAIL_START": "1"}
		failed.Bootstrap = &pb.Bootstrap{GenerationId: next.Generation, CandidateConfigJson: `{"candidate":true}`, TargetConfigVersion: 1, DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_SQLITE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_VERSIONED, ActivationReason: request.Reason, SqlitePath: next.DatabasePath}}
		unexpected, err := Launch(ctx, store, failed)
		if unexpected != nil {
			unexpected.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
			t.Fatal("failing OnStart launched successfully")
		}
		var phase *host.PhaseError
		if !errors.As(err, &phase) || phase.Code != pb.ErrorCode_ERROR_CODE_INTERNAL || phase.Phase != pb.SessionPhase_SESSION_PHASE_ON_START || !strings.Contains(spec.Log.(*logBuf).String(), "start failed after database write") {
			t.Fatalf("OnStart failure: %v\n%s", err, spec.Log.(*logBuf).String())
		}
		changed := databaseFile(t, next.DatabasePath)
		if got := candidateActivity(t, changed); !slices.Equal(got, []string{"migration", "start", "start", "wal-one", "wal-two"}) {
			t.Fatalf("failed OnStart did not write: %v", got)
		}
		if err := changed.Close(); err != nil {
			t.Fatal(err)
		}
		unlock = store.LockPlugin(spec.PluginID)
		err = candidate.Discard(ctx, store, secret.New(store), next)
		unlock()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Dir(next.DatabasePath)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("discarded directory: %v", err)
		}
		if _, err := store.GenerationConfig(ctx, next.Generation); !errors.Is(err, state.ErrNotFound) {
			t.Fatalf("discarded generation: %v", err)
		}
		if active, err := store.Active(ctx, spec.PluginID); err != nil || active != spec.GenerationID {
			t.Fatalf("Active changed: %s, %v", active, err)
		}
		if got, err := store.GenerationConfig(ctx, spec.GenerationID); err != nil || got != snapshot {
			t.Fatalf("Active snapshot changed: %s, %v", got, err)
		}
		original := databaseFile(t, prepared.DatabasePath)
		if got := candidateActivity(t, original); !slices.Equal(got, before) {
			t.Fatalf("Active database changed: %v", got)
		}
		spec.NewCandidate = false
		spec.Bootstrap.DatabaseDescriptor.ActivationReason = pb.ActivationReason_ACTIVATION_REASON_RESUME
		spec.Bootstrap.DatabaseDescriptor.AppliedMigrationIds = []uint32{1, 2}
		restarted, err := Launch(ctx, store, spec)
		if err != nil {
			t.Fatalf("restart old generation: %v", err)
		}
		stop := restarted.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
		restarted.Process.WaitLogs()
		if stop.Err != nil || stop.Killed {
			t.Fatalf("restart stop: %+v", stop)
		}
		if got := candidateActivity(t, original); !slices.Equal(got, []string{"migration", "start", "start", "wal-one", "wal-two"}) {
			t.Fatalf("restart old data: %v", got)
		}
	})
	t.Run("secret protection", func(t *testing.T) {
		ctx := t.Context()
		store, spec := setup(t, binary)
		secrets := secret.New(store)
		references := []string{"11111111111111111111111111111111", "22222222222222222222222222222222", "33333333333333333333333333333333", "44444444444444444444444444444444"}
		for _, ref := range references {
			if err := store.AddSecret(ctx, state.Secret{RefID: ref, PluginID: spec.PluginID, Value: []byte("integration secret")}); err != nil {
				t.Fatal(err)
			}
		}
		snapshot := func(own, shared string) string {
			if shared == "" {
				return fmt.Sprintf(`{"token":{"%s":"%s"}}`, secret.RefKey, own)
			}
			return fmt.Sprintf(`{"token":{"%s":"%s"},"shared":{"%s":"%s"}}`, secret.RefKey, own, secret.RefKey, shared)
		}
		activeSnapshot := snapshot(references[0], references[3])
		if err := store.SetGenerationConfig(ctx, spec.GenerationID, activeSnapshot, []string{"/token", "/shared"}); err != nil {
			t.Fatal(err)
		}
		if err := store.CommitActive(ctx, spec.PluginID, "", spec.GenerationID); err != nil {
			t.Fatal(err)
		}
		pending := state.PendingConfig{PluginID: spec.PluginID, BaseGeneration: spec.GenerationID, ConfigJSON: snapshot(references[1], "")}
		if err := store.PutPendingConfig(ctx, pending); err != nil {
			t.Fatal(err)
		}
		unlock := store.LockPlugin(spec.PluginID)
		defer unlock()
		next, err := candidate.PrepareDatabase(ctx, candidate.Request{PluginID: spec.PluginID, PluginDir: spec.PluginDir, Reason: pb.ActivationReason_ACTIVATION_REASON_FIRST_INSTALL, NewGeneration: "secret-candidate"})
		if err != nil {
			t.Fatal(err)
		}
		if err := store.AddGeneration(ctx, spec.PluginID, next.Generation, snapshot(references[2], references[3]), 1); err != nil {
			t.Fatal(err)
		}
		if err := store.SetGenerationConfig(ctx, next.Generation, snapshot(references[2], references[3]), []string{"/token", "/shared"}); err != nil {
			t.Fatal(err)
		}
		if err := candidate.Discard(ctx, store, secrets, next); err != nil {
			t.Fatal(err)
		}
		for i, ref := range references {
			_, err := store.GetSecret(ctx, spec.PluginID, ref)
			if i == 2 {
				if !errors.Is(err, state.ErrNotFound) {
					t.Fatalf("candidate-only secret survived: %v", err)
				}
			} else if err != nil {
				t.Fatalf("protected secret %d lost: %v", i, err)
			}
		}
		rows, err := store.ListSecrets(ctx, spec.PluginID)
		if err != nil || len(rows) != 3 {
			t.Fatalf("remaining secret count %d, %v", len(rows), err)
		}
		if got, err := store.GenerationConfig(ctx, spec.GenerationID); err != nil || got != activeSnapshot {
			t.Fatalf("Active snapshot changed: %s, %v", got, err)
		}
		if got, err := store.GetPendingConfig(ctx, spec.PluginID); err != nil || got.ConfigJSON != pending.ConfigJSON {
			t.Fatalf("Pending changed: %+v, %v", got, err)
		}
		if _, err := store.GenerationConfig(ctx, next.Generation); !errors.Is(err, state.ErrNotFound) {
			t.Fatalf("candidate record survived: %v", err)
		}
		if paths, err := store.GenerationSecretPaths(ctx, next.Generation); err != nil || len(paths) != 0 {
			t.Fatalf("candidate paths survived: %v, %v", paths, err)
		}
		if _, err := os.Stat(filepath.Dir(next.DatabasePath)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("candidate directory survived: %v", err)
		}
		activeDir := filepath.Join(spec.PluginDir, "generations", spec.GenerationID)
		if err := os.MkdirAll(activeDir, 0o700); err != nil {
			t.Fatal(err)
		}
		activePath := filepath.Join(activeDir, "db.sqlite")
		if err := os.WriteFile(activePath, []byte("protected active file"), 0o600); err != nil {
			t.Fatal(err)
		}
		err = candidate.Discard(ctx, store, secrets, &candidate.Prepared{PluginID: spec.PluginID, Generation: spec.GenerationID, DatabasePath: activePath})
		var detail *validate.Error
		if !errors.As(err, &detail) || detail.Code != pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION {
			t.Fatalf("discard Active: %v", err)
		}
		if contents, err := os.ReadFile(activePath); err != nil || string(contents) != "protected active file" {
			t.Fatalf("Active file changed: %q, %v", contents, err)
		}
		if rows, err := store.ListSecrets(ctx, spec.PluginID); err != nil || len(rows) != 3 {
			t.Fatalf("Active rejection changed secrets: %v, %v", rows, err)
		}
	})
}

func candidateActivity(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.Query("SELECT name FROM activity ORDER BY name")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var activity []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		activity = append(activity, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return activity
}
