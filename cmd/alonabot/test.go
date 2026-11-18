package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
	"github.com/AlonaBot/AlonaBot/internal/generate"
	"github.com/AlonaBot/AlonaBot/internal/pluginkit"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/spf13/cobra"
)

const offlineGeneration = "offline-test"

func testCommand(code *int, values *tune.Values) *cobra.Command {
	var dir string
	command := &cobra.Command{Use: "test", Short: "Load, register, activate and stop a component in a temporary offline session", Args: cobra.NoArgs}
	command.Flags().StringVar(&dir, "dir", ".", "component project directory")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		*code = exitFailure
		return testComponent(cmd.Context(), dir, *values, cmd.OutOrStdout(), cmd.ErrOrStderr())
	}
	return command
}

func testComponent(ctx context.Context, dir string, values tune.Values, stdout, stderr io.Writer) (err error) {
	// 1. 测试目录、状态库、配置与 runtime.lock 全部归本次调用所有。
	temporary, err := os.MkdirTemp("", "alonabot-test-*")
	if err != nil {
		return err
	}
	defer func() {
		cleanup := os.RemoveAll(temporary)
		if cleanup == nil {
			_, cleanup = fmt.Fprintln(stdout, "Removed temporary test data:", temporary)
		}
		err = errors.Join(err, cleanup)
	}()
	raw, err := os.ReadFile(filepath.Join(dir, "plugin.yaml"))
	if err != nil {
		return err
	}
	project, err := manifest.ParseProject(raw)
	if err != nil {
		return err
	}
	if project.Manifest.Role == "webui" {
		if err := pluginkit.Check(ctx, generate.Options{Dir: dir}); err != nil {
			return err
		}
		_, err = fmt.Fprintln(stdout, "Validated WebUI static resources")
		return err
	}
	build, err := pluginkit.BuildDev(ctx, dir, filepath.Join(temporary, "bin"))
	if err != nil {
		return err
	}
	resources, err := pluginkit.ReadResources(build)
	if err != nil {
		return err
	}
	packaged, err := manifest.Parse(build.PackageManifest)
	if err != nil {
		return err
	}
	m := packaged.Manifest
	store, err := state.Open(ctx, filepath.Join(temporary, "data"))
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, store.Close()) }()
	role := state.Role(m.Role)
	if err := store.AddInstall(ctx, state.Install{Role: role, PluginID: m.ID, AdapterID: m.AdapterID, Source: m.Source, Version: m.Version, Entrypoint: build.Executable, BuildDigest: build.Digest}); err != nil {
		return err
	}
	if err := store.AddGeneration(ctx, m.ID, offlineGeneration, "{}", *m.ConfigVersion); err != nil {
		return err
	}
	pluginDir := filepath.Join(temporary, "plugins", m.ID)
	if err := os.MkdirAll(pluginDir, 0o700); err != nil {
		return err
	}
	database := &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_NONE, ActivationReason: pb.ActivationReason_ACTIVATION_REASON_FIRST_INSTALL}
	if role != state.RoleAdapter {
		database.Backend = pb.DatabaseBackend_DATABASE_BACKEND_SQLITE
		database.SqlitePath = filepath.Join(pluginDir, "db.sqlite")
		switch m.MigrationMode {
		case manifest.MigrationSelf:
			database.MigrationMode = pb.MigrationMode_MIGRATION_MODE_SELF
		case manifest.MigrationVersioned:
			database.MigrationMode = pb.MigrationMode_MIGRATION_MODE_VERSIONED
		}
	}
	// 与状态库共用受保护的根目录，Windows 子目录继承当前账号的 ACL。
	pluginDataDir, err := filepath.Abs(filepath.Join(temporary, "data", "plugin_data", m.ID))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(pluginDataDir, 0o700); err != nil {
		return err
	}
	boot := &pb.Bootstrap{GenerationId: offlineGeneration, CandidateConfigJson: "{}", ProfileOverridesJson: "{}", PreviousConfigVersion: m.ConfigVersion, TargetConfigVersion: *m.ConfigVersion, DatabaseDescriptor: database, LogLevel: pb.LogLevel_LOG_LEVEL_DEBUG, PluginDataDir: pluginDataDir}
	if len(resources.ConfigSchema) > 0 {
		schema, problems := configschema.Parse(resources.ConfigSchema)
		if len(problems) > 0 {
			return fmt.Errorf("config schema: %v", problems)
		}
		boot.ConfigSchemaHash = schema.Hash()
	}
	output, err := logging.Open(logging.Options{Format: logging.FormatText, Level: slog.LevelDebug, Stderr: stderr, BufferRecords: values.LogBufferRecords})
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, output.Close()) }()
	log := &offlineLog{source: output.Plugin(m.ID, offlineGeneration)}
	// 2. 真实 Core 侧校验握手和 REGISTER，提交临时代次后激活；Adapter 保持零实例。
	running, err := supervisor.Launch(ctx, store, supervisor.Spec{Role: role, AdapterID: m.AdapterID, PluginID: m.ID, PluginDir: pluginDir, Executable: build.Executable, BuildDigest: build.Digest, GenerationID: offlineGeneration, Bootstrap: boot, ConfigSchema: resources.ConfigSchema, I18n: resources.I18n, NewCandidate: true, Limits: host.NewLimits(values), Stop: host.NewStopTiming(values), Log: log, MemorySoftLimit: values.PluginMemorySoftLimit})
	if err != nil {
		return err
	}
	defer func() {
		log.beginStop()
		result := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
		var stopErr error
		if result.Killed {
			stopErr = errors.New("component exceeded its stop budget and was killed")
		}
		if exit := running.Process.ExitState(); exit != nil && !exit.Success() {
			stopErr = errors.Join(stopErr, fmt.Errorf("component exited unsuccessfully: %s", exit))
		}
		err = errors.Join(err, result.Err, stopErr, log.failure())
	}()
	if _, err := fmt.Fprintf(stdout, "Activated %s (%s)\n", m.ID, m.Role); err != nil {
		return err
	}
	for _, route := range running.Instance.Routes {
		if command := route.GetCommand(); command != nil {
			if _, err := fmt.Fprintf(stdout, "Command: %s\n", strings.Join(command.GetPath(), " ")); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(stdout, "Route: %s\n", route.RouteId); err != nil {
				return err
			}
		}
	}
	if running.Instance.Adapter != nil {
		_, err = fmt.Fprintf(stdout, "Adapter registered: %s (no platform instances started)\n", running.Instance.Adapter.AdapterId)
	}
	return err
}

// offlineLog 保留 SDK 只通过结构化日志报告的停止失败；supervisor 读完日志后再取结果。
type offlineLog struct {
	source   *logging.Source
	mu       sync.Mutex // 停止开始与 stdout/stderr 日志回调可并发。
	stopping bool
	pending  []byte // stderr 的不完整 JSON 记录；进程日志可能按缓冲区分块。
	err      error
}

func (l *offlineLog) SetSession(id []byte) { l.source.SetSession(id) }
func (l *offlineLog) Line(id, stream string, line []byte) {
	if stream != "stderr" {
		l.source.Line(id, stream, line)
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pending = append(l.pending, line...)
	var record struct{ Level, Msg string }
	// 以完整 JSON 记录识别停止错误，跨分块保留原文而不截断错误内容。
	err := json.NewDecoder(bytes.NewReader(l.pending)).Decode(&record)
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return
	}
	line, l.pending = l.pending, nil
	l.source.Line(id, stream, line)
	if err != nil {
		return
	}
	if l.stopping && (record.Level == "ERROR" || record.Msg == "OnStop exceeded the stop deadline; giving up waiting") {
		l.err = errors.Join(l.err, fmt.Errorf("component stop: %s", line))
	}
}
func (l *offlineLog) beginStop()     { l.mu.Lock(); defer l.mu.Unlock(); l.stopping = true }
func (l *offlineLog) failure() error { l.mu.Lock(); defer l.mu.Unlock(); return l.err }
