//go:build unix

package alonabot_test

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/generate"
	"github.com/AlonaBot/AlonaBot/internal/pluginkit"
	"github.com/AlonaBot/AlonaBot/internal/tune"
)

func TestLibraryLifecycle(t *testing.T) {
	const startupTimeout = time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()
	temp := t.TempDir()
	executable, err := pluginkit.CompileApp(ctx, generate.Options{Dir: "./examples/devbot"}, temp, "-race")
	if err != nil {
		t.Fatalf("build devbot: %v", err)
	}
	logPath := filepath.Join(temp, "output.log")
	output, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	dir := filepath.Join(temp, "data")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "core_config.json"), []byte(`{"listen":"127.0.0.1:0"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, scenario := range []struct{ name, config, want string }{
		{name: "defaults", want: "Hello"},
		{name: "configured", config: `{"greetings":[{"text":"Updated"}]}`, want: "Hello,Updated"},
		{name: "missing_required", config: `{"greetings":[{}]}`},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			if err := output.Truncate(0); err != nil {
				t.Fatal(err)
			}
			if _, err := output.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			if scenario.config != "" {
				path := filepath.Join(dir, "plugins", "alonabot_plugin_devbot", "config.json")
				if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(scenario.config), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			command := exec.CommandContext(ctx, executable, "-data-dir", dir)
			command.Stdout, command.Stderr = output, output
			// 子进程只使用契约缺省预算，避免开发者环境改变本测试的停止期限。
			for _, variable := range os.Environ() {
				if !strings.HasPrefix(variable, "ALONABOT_ARGS=") {
					command.Env = append(command.Env, variable)
				}
			}
			if err := command.Start(); err != nil {
				t.Fatal(err)
			}
			finished := false
			exited := make(chan error, 1)
			go func() { exited <- command.Wait() }()
			defer func() {
				if !finished {
					_ = command.Process.Kill()
					<-exited
				}
			}()
			if scenario.want == "" {
				select {
				case err := <-exited:
					finished = true
					logs, readErr := os.ReadFile(logPath)
					if readErr != nil {
						t.Fatal(readErr)
					}
					if err == nil || !strings.Contains(string(logs), "SESSION_PHASE_CONFIG") || !strings.Contains(string(logs), "required") {
						t.Fatalf("missing required config: %v\n%s", err, logs)
					}
				case <-ctx.Done():
					t.Fatal("invalid config startup timed out")
				}
				return
			}
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()
			ready := false
			for !ready {
				select {
				case err := <-exited:
					finished = true
					logs, _ := os.ReadFile(logPath)
					t.Fatalf("devbot exited before ready: %v\n%s", err, logs)
				case <-ctx.Done():
					t.Fatal("devbot readiness timed out")
				case <-ticker.C:
					logs, err := os.ReadFile(logPath)
					if err != nil {
						t.Fatal(err)
					}
					ready = strings.Contains(string(logs), "devbot ready")
				}
			}
			db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(filepath.Join(dir, "state.sqlite"))+"?mode=ro")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			rows, err := db.QueryContext(ctx, `SELECT i.plugin_id, i.role, i.source, s.active_generation FROM install i JOIN plugin_state s USING (plugin_id) ORDER BY i.install_seq`)
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			want := map[string]string{"alonabot_plugin_devbot": "plugin", "alonabot_adapter_devbot": "adapter"}
			for rows.Next() {
				var id, role, source, generation string
				if err := rows.Scan(&id, &role, &source, &generation); err != nil {
					t.Fatal(err)
				}
				if want[id] != role || source != "dev" || generation != "dev:"+id {
					t.Fatalf("unexpected installation: %s %s %s %s", id, role, source, generation)
				}
				delete(want, id)
			}
			if err := rows.Err(); err != nil {
				t.Fatal(err)
			}
			if err := rows.Close(); err != nil {
				t.Fatal(err)
			}
			if len(want) != 0 {
				t.Fatalf("missing installations: %v", want)
			}

			pluginDir := filepath.Join(dir, "plugins", "alonabot_plugin_devbot")
			pluginDB, err := sql.Open("sqlite", "file:"+filepath.ToSlash(filepath.Join(pluginDir, "dev.sqlite"))+"?mode=ro")
			if err != nil {
				t.Fatal(err)
			}
			defer pluginDB.Close()
			startupRows, err := pluginDB.QueryContext(ctx, "SELECT greeting FROM startup ORDER BY id")
			if err != nil {
				t.Fatal(err)
			}
			var greetings []string
			for startupRows.Next() {
				var greeting string
				if err := startupRows.Scan(&greeting); err != nil {
					t.Fatal(err)
				}
				greetings = append(greetings, greeting)
			}
			if err := startupRows.Err(); err != nil {
				t.Fatal(err)
			}
			if err := startupRows.Close(); err != nil {
				t.Fatal(err)
			}
			if strings.Join(greetings, ",") != scenario.want {
				t.Fatalf("startup greetings = %v, want %s", greetings, scenario.want)
			}
			if _, err := os.Stat(filepath.Join(pluginDir, "generations")); !os.IsNotExist(err) {
				t.Fatalf("unexpected generations directory: %v", err)
			}
			entries, err := os.ReadDir(pluginDir)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if strings.HasSuffix(entry.Name(), ".sqlite") && entry.Name() != "dev.sqlite" {
					t.Fatalf("unexpected database copy: %s", entry.Name())
				}
			}
			configPath := filepath.Join(pluginDir, "config.json")
			if scenario.config == "" {
				if _, err := os.Stat(configPath); !os.IsNotExist(err) {
					t.Fatalf("config was written back: %v", err)
				}
			} else {
				raw, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatal(err)
				}
				if string(raw) != scenario.config {
					t.Fatalf("config was rewritten: %s", raw)
				}
			}

			stoppedAt := time.Now()
			if err := command.Process.Signal(syscall.SIGTERM); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-exited:
				finished = true
				if err != nil {
					logs, _ := os.ReadFile(logPath)
					t.Fatalf("devbot shutdown: %v\n%s", err, logs)
				}
				if elapsed := time.Since(stoppedAt); elapsed > tune.Defaults().StopTotal {
					t.Fatalf("shutdown exceeded STOP_TOTAL: %v", elapsed)
				}
			case <-time.After(tune.Defaults().StopTotal):
				t.Fatal("devbot did not exit within shared STOP_TOTAL")
			}
		})
	}
}
