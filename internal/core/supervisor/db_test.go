//go:build unix

package supervisor

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	_ "modernc.org/sqlite"
)

// databaseSQLPlugin 使用标准库入口，二进制应保持与 GORM 隔离。
const databaseSQLPlugin = `package main
import (
 "context"
 "database/sql"
 "fmt"
 "os"
 "github.com/AlonaBot/AlonaBot/sdk"
 "github.com/AlonaBot/AlonaBot/sdk/sqlitedb"
)
type plugin struct { sdk.Base }
func (*plugin) Migrations() []sqlitedb.Migration {
 return []sqlitedb.Migration{
  {ID:1, Up:func(ctx context.Context,tx *sql.Tx) error { _,err:=tx.ExecContext(ctx,"CREATE TABLE activity (name TEXT NOT NULL)"); return err }},
  {ID:2, Up:func(ctx context.Context,tx *sql.Tx) error {
   if _,err:=tx.ExecContext(ctx,"INSERT INTO activity VALUES ('migration')"); err!=nil{return err}
   if os.Getenv("DB_FAIL_SECOND")=="1" { return fmt.Errorf("second migration failed") }; return nil
  }},
 }
}
func (*plugin) OnStart(ctx context.Context) error {
 db,err:=sqlitedb.Open(ctx); if err!=nil{return err}
 if err:=checkDatabase(ctx,db);err!=nil{return err}
 _,err=db.ExecContext(ctx,"INSERT INTO activity VALUES ('start')");return err
}
func databaseRoot() sdk.Root { return &plugin{} }
`

// databaseGORMPlugin 以原生 GORM 迁移，并提供跨包接口混用的故障入口。
const databaseGORMPlugin = `package main
import (
 "context"
 "database/sql"
 "fmt"
 "log/slog"
 "os"
 "github.com/AlonaBot/AlonaBot/sdk"
 "github.com/AlonaBot/AlonaBot/sdk/config"
 "github.com/AlonaBot/AlonaBot/sdk/gormdb"
 "gorm.io/gorm"
)
type plugin struct { sdk.Base }
func (*plugin) Migrations() []gormdb.Migration {
 return []gormdb.Migration{
  {ID:1, Up:func(ctx context.Context,tx *gorm.DB) error { return tx.WithContext(ctx).Exec("CREATE TABLE activity (name TEXT NOT NULL)").Error }},
  {ID:2, Up:func(ctx context.Context,tx *gorm.DB) error {
   if err:=tx.WithContext(ctx).Exec("INSERT INTO activity VALUES ('migration')").Error;err!=nil{return err}
   if os.Getenv("DB_FAIL_SECOND")=="1" { return fmt.Errorf("second migration failed") }; return nil
  }},
 }
}
func (*plugin) OnStart(ctx context.Context) error {
 db,err:=gormdb.Open(ctx); if err!=nil{return err}
 underlying,err:=db.DB();if err!=nil{return err}
 if err:=checkDatabase(ctx,underlying);err!=nil{return err}
 return db.WithContext(ctx).Exec("INSERT INTO activity VALUES ('start')").Error
}
type mixed struct { plugin }
func (*mixed) Migrate(context.Context,*sql.DB) error { return fmt.Errorf("ambiguous migration ran") }
func (*mixed) Config() config.Config { slog.Info("config-prepared"); return nil }
func databaseRoot() sdk.Root { if os.Getenv("DB_MIXED")=="1" {return &mixed{}}; return &plugin{} }
`

// databaseProbe 只观察真实线上 Bootstrap/REGISTER，原帧交给 SDK 和 Core。
const databaseProbe = `package main
import (
 "context"
 "database/sql"
 "fmt"
 "os"
 "github.com/AlonaBot/AlonaBot/sdk"
 "github.com/AlonaBot/AlonaBot/sdk/sqlitedb"
 "github.com/AlonaBot/AlonaBot/contracts/validate"
 pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
 "github.com/AlonaBot/AlonaBot/internal/wire"
)
func checkDatabase(ctx context.Context,db *sql.DB) error {
 values,err:=sqlitedb.ReadPragmas(ctx,db);if err!=nil{return err}
 if values!=(sqlitedb.PragmaValues{JournalMode:"wal",Synchronous:2,ForeignKeys:1,BusyTimeout:5000})||db.Stats().MaxOpenConnections!=1 { return fmt.Errorf("database settings mismatch: %+v",values) }
 return nil
}
func main() {
 input,output:=os.Stdin,os.Stdout
 sdkInput,inputWriter,err:=os.Pipe();if err!=nil{panic(err)}
 outputReader,sdkOutput,err:=os.Pipe();if err!=nil{panic(err)}
 os.Stdin,os.Stdout=sdkInput,sdkOutput
 core:=wire.NewConn(input,output,input)
 pluginWire:=wire.NewConn(outputReader,inputWriter,outputReader)
 forward:=func(in,out *wire.Conn) {
  for {
   frame,err:=in.ReadFrame();if err!=nil{return}
   message,err:=validate.DecodeControl(frame);if err!=nil{return}
   switch value:=message.(type) {
   case *pb.HelloAck: fmt.Fprintln(os.Stderr,"database-path",value.GetBootstrap().GetDatabaseDescriptor().GetSqlitePath())
   case *pb.Register: fmt.Fprintln(os.Stderr,"migration-report",value.GetMigration().GetCompletedMigrationIds())
   }
   if err:=out.WriteFrame(frame);err!=nil{return}
  }
 }
 go forward(core,pluginWire)
 forwarded:=make(chan struct{})
 go func(){defer close(forwarded);forward(pluginWire,core)}()
 err=sdk.Run(databaseRoot())
 // SDK 写入本地管道后才返回，退出前须等待代理把最后的 SESSION_ERROR 交给 Core。
 _=sdkOutput.Close()
 <-forwarded
 if err!=nil { fmt.Fprintln(os.Stderr,err);os.Exit(1) }
}
`

func buildDatabasePlugins(t *testing.T) map[string]string {
	t.Helper()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	module, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	sums, err := os.ReadFile(filepath.Join(root, "go.sum"))
	if err != nil {
		t.Fatal(err)
	}
	binaries := map[string]string{}
	for name, source := range map[string]string{"sql": databaseSQLPlugin, "gorm": databaseGORMPlugin} {
		dir := t.TempDir()
		mod := strings.Replace(string(module), "module github.com/AlonaBot/AlonaBot", "module github.com/AlonaBot/AlonaBot/dbintegration", 1)
		mod += "\nrequire github.com/AlonaBot/AlonaBot v0.0.0\nreplace github.com/AlonaBot/AlonaBot => " + strconv.Quote(filepath.ToSlash(root)) + "\n"
		// 不同文件各自保存 import，真实插件仍使用同一个 main 包。
		for file, contents := range map[string]string{"go.mod": mod, "go.sum": string(sums), "plugin.go": source, "probe.go": databaseProbe} {
			if err := os.WriteFile(filepath.Join(dir, file), []byte(contents), 0600); err != nil {
				t.Fatal(err)
			}
		}
		binary := filepath.Join(dir, "plugin")
		command := exec.Command("go", "build", "-mod=mod", "-race", "-o", binary, ".")
		command.Dir = dir
		if out, err := command.CombinedOutput(); err != nil {
			t.Fatalf("build %s plugin: %v\n%s", name, err, out)
		}
		if name == "sql" {
			command = exec.Command("go", "list", "-mod=mod", "-deps", ".")
			command.Dir = dir
			out, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("list SQL plugin dependencies: %v\n%s", err, out)
			}
			if strings.Contains(string(out), "gorm.io/gorm") {
				t.Fatal("SQL-only plugin links GORM")
			}
		}
		binaries[name] = binary
	}
	command := exec.Command("go", "list", "-deps", "./sdk")
	command.Dir = root
	out, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("list SDK dependencies: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "gorm.io/gorm") || strings.Contains(string(out), "modernc.org/sqlite") {
		t.Fatal("SDK root links a database implementation")
	}
	return binaries
}

func databaseFile(t *testing.T, path string) *sql.DB {
	t.Helper()
	dsn := (&url.URL{Scheme: "file", Path: filepath.ToSlash(path), RawQuery: "mode=rw"}).String()
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	return db
}

func databaseIDs(t *testing.T, db *sql.DB) []uint32 {
	t.Helper()
	rows, err := db.Query("SELECT id FROM _alonabot_migrations ORDER BY id")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var ids []uint32
	for rows.Next() {
		var id uint32
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return ids
}

// TestManagedDatabaseProcesses 验证真实 SDK 的两种入口、逐步迁移和代次路径（V-26/V-39）。
func TestManagedDatabaseProcesses(t *testing.T) {
	binaries := buildDatabasePlugins(t)
	for name, binary := range binaries {
		for _, scenario := range []string{"install-restart", "second-fails", "mixed"} {
			if scenario == "mixed" && name == "sql" {
				continue
			}
			t.Run(name+"/"+scenario, func(t *testing.T) {
				store, spec := setup(t, binary)
				spec.Stdio = true
				spec.ExtraEnv = map[string]string{"GORACE": "atexit_sleep_ms=0"}
				spec.Bootstrap.DatabaseDescriptor = &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_SQLITE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_VERSIONED, ActivationReason: pb.ActivationReason_ACTIVATION_REASON_FIRST_INSTALL}
				path := filepath.Join(spec.PluginDir, "generations", spec.GenerationID, "db.sqlite")
				if scenario == "second-fails" {
					spec.ExtraEnv["DB_FAIL_SECOND"] = "1"
				}
				if scenario == "mixed" {
					spec.ExtraEnv["DB_MIXED"] = "1"
				}
				logs := spec.Log.(*logBuf)
				run, err := Launch(context.Background(), store, spec)
				if scenario != "install-restart" {
					var phase *host.PhaseError
					code := pb.ErrorCode_ERROR_CODE_INTERNAL
					if scenario == "mixed" {
						code = pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
					}
					if run != nil || !errors.As(err, &phase) || phase.Phase != pb.SessionPhase_SESSION_PHASE_MIGRATION || phase.Code != code {
						t.Fatalf("migration failure = %v; logs:\n%s", err, logs.String())
					}
					if active, err := store.Active(t.Context(), spec.PluginID); err != nil || active != "" {
						t.Fatalf("failed preparation changed Active: %q, %v", active, err)
					}
					if scenario == "mixed" {
						if !strings.Contains(logs.String(), "config-prepared") {
							t.Fatalf("CONFIG did not precede MIGRATION: %s", logs.String())
						}
						return
					}
					db := databaseFile(t, path)
					if got := databaseIDs(t, db); !slices.Equal(got, []uint32{1}) {
						t.Fatalf("failed migration ledger = %v", got)
					}
					var count int
					if err := db.QueryRow("SELECT count(*) FROM activity").Scan(&count); err != nil || count != 0 {
						t.Fatalf("failed Up or OnStart left data: %d, %v", count, err)
					}
					return
				}
				if err != nil {
					t.Fatalf("first install: %v\n%s", err, logs.String())
				}
				stop := func(r *Running) {
					t.Helper()
					result := r.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
					r.Process.WaitLogs()
					if result.Err != nil || result.Killed {
						t.Fatalf("plugin stop = %+v\n%s", result, logs.String())
					}
				}
				stop(run)
				if !strings.Contains(logs.String(), "migration-report [1 2]") || !strings.Contains(logs.String(), "database-path "+path) {
					t.Fatalf("unexpected bootstrap or REGISTER: %s", logs.String())
				}
				if spec.Bootstrap.DatabaseDescriptor.SqlitePath != "" {
					t.Fatal("caller Bootstrap mutated")
				}
				db := databaseFile(t, path)
				if got := databaseIDs(t, db); !slices.Equal(got, []uint32{1, 2}) {
					t.Fatalf("install ledger = %v", got)
				}
				var journal string
				if err := db.QueryRow("PRAGMA journal_mode").Scan(&journal); err != nil || journal != "wal" {
					t.Fatalf("journal mode = %q, %v", journal, err)
				}
				// 连接级 PRAGMA 由插件通过迁移连接回读，见 fixture 中的启动检查。
				spec.NewCandidate = false
				spec.Bootstrap.DatabaseDescriptor.ActivationReason = pb.ActivationReason_ACTIVATION_REASON_RESUME
				spec.Bootstrap.DatabaseDescriptor.AppliedMigrationIds = []uint32{1, 2}
				run, err = Launch(t.Context(), store, spec)
				if err != nil {
					t.Fatalf("restart: %v\n%s", err, logs.String())
				}
				stop(run)
				if strings.Count(logs.String(), "migration-report [1 2]") != 2 {
					t.Fatalf("restart report missing: %s", logs.String())
				}
				var migrations, starts int
				if err := db.QueryRow("SELECT count(*) FROM activity WHERE name='migration'").Scan(&migrations); err != nil {
					t.Fatal(err)
				}
				if err := db.QueryRow("SELECT count(*) FROM activity WHERE name='start'").Scan(&starts); err != nil {
					t.Fatal(err)
				}
				if migrations != 1 || starts != 2 {
					t.Fatalf("migration/start executions = %d/%d", migrations, starts)
				}
			})
		}
	}
}
