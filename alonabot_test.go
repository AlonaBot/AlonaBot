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

	"github.com/AlonaBot/AlonaBot/internal/tune"
)

func TestLibraryLifecycle(t *testing.T) {
	const startupTimeout = time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()
	temp := t.TempDir()
	executable := filepath.Join(temp, "devbot")
	build := exec.CommandContext(ctx, "go", "build", "-race", "-o", executable, "./examples/devbot")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build devbot: %v\n%s", err, output)
	}
	logPath := filepath.Join(temp, "output.log")
	output, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	dir := filepath.Join(temp, "data")
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
}
