package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/i18n"
	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/process"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
)

func hasInstalledRoutes(in state.Install) bool {
	return in.Source != "" && in.Source != devSource && (in.Role == state.RolePlugin || in.Role == state.RoleAgent)
}

func tracksRuntime(role state.Role) bool { return role == state.RolePlugin || role == state.RoleAgent }

// recordRuntime 在 Core 运行循环串行化组件变更的基础上，与进程清理共用逐插件锁。
func (c *Core) recordRuntime(ctx context.Context, in state.Install, next state.RuntimeState, reason string) error {
	if !tracksRuntime(in.Role) {
		return nil
	}
	unlock := c.store.LockPlugin(in.PluginID)
	defer unlock()
	current, err := c.store.RuntimeStatus(ctx, in.PluginID)
	if err != nil {
		return err
	}
	if current.State == next && current.Reason == reason {
		return nil
	}
	previous := current.State
	current.State, current.Reason, current.ChangedAt = next, reason, time.Now().UTC()
	if err := c.store.SetRuntimeStatus(ctx, in.PluginID, current); err != nil {
		return err
	}
	if previous != next {
		level := slog.LevelDebug
		if next == state.RuntimeFailed {
			level = slog.LevelError
		}
		message := fmt.Sprintf("Component %s changed from %s to %s", in.PluginID, previous, next)
		if reason != "" {
			description := reason
			if stopReason, ok := strings.CutPrefix(reason, "STOP_REASON_"); ok {
				description = strings.ToLower(strings.ReplaceAll(stopReason, "_", " "))
			}
			message += ": " + description
		}
		slog.Log(ctx, level, message, "plugin_id", in.PluginID, "previous_state", previous, "state", next, "reason", reason)
	}
	return nil
}

// recoverRuntime 在 Journal 清理后核对旧 SDK 是否仍持有跨代次锁，再发布当前运行状态。
func (c *Core) recoverRuntime(ctx context.Context, installs []state.Install, blocked []string) ([]string, error) {
	for _, in := range installs {
		if !tracksRuntime(in.Role) || in.Uninstalled {
			continue
		}
		if !slices.Contains(blocked, in.PluginID) {
			lock, err := process.LockRuntime(filepath.Join(c.dir, "plugins", in.PluginID))
			if err == nil {
				if err := lock.Unlock(); err != nil {
					return nil, err
				}
			} else if !errors.Is(err, os.ErrNotExist) {
				blocked = append(blocked, in.PluginID)
				slog.Error(fmt.Sprintf("Failed to verify the previous process of component %s: %v", in.PluginID, err), "plugin_id", in.PluginID, "error", err)
			}
		}
		if slices.Contains(blocked, in.PluginID) {
			if err := c.recordRuntime(ctx, in, state.RuntimeFailed, "previous runtime could not be cleared"); err != nil {
				return nil, err
			}
			continue
		}
		status, err := c.store.RuntimeStatus(ctx, in.PluginID)
		if err != nil {
			return nil, err
		}
		switch status.State {
		case state.RuntimeActive, state.RuntimeWaking, state.RuntimeQuiescing:
			if err := c.recordRuntime(ctx, in, state.RuntimeStopped, "core startup"); err != nil {
				return nil, err
			}
		}
	}
	return blocked, nil
}

// installedSpec 在停机前构造完整启动输入并准备数据目录；调用方决定准备或直接启动。
// 返回的运行上下文与 Handler/Detach 共用 Caller，启动后由调用方附加 Instance。
func (c *Core) installedSpec(ctx context.Context, in state.Install, packageDir string, boot *pb.Bootstrap) (supervisor.Spec, *runningComponent, error) {
	parsed, err := installedManifest(in, packageDir)
	if err != nil {
		return supervisor.Spec{}, nil, err
	}
	m := parsed.Manifest
	for _, warning := range parsed.Warnings {
		slog.Warn(fmt.Sprintf("Component %s has an installed manifest warning: %s", m.ID, warning), "plugin_id", m.ID, "warning", warning)
	}
	executable, err := process.Entrypoint(packageDir, in.Entrypoint)
	if err != nil {
		return supervisor.Spec{}, nil, err
	}
	resources, err := packageResources(packageDir)
	if err != nil {
		if hasInstalledRoutes(in) {
			return supervisor.Spec{}, nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: err.Error()}
		}
		return supervisor.Spec{}, nil, err
	}
	schemas, err := optionSchemas(resources)
	if err != nil {
		return supervisor.Spec{}, nil, err
	}
	if _, problems := i18n.Parse(resources.I18n); len(problems) != 0 {
		return supervisor.Spec{}, nil, fmt.Errorf("installed message catalog: %v", problems)
	}
	// Schema 摘要绑定包内声明，所有已安装组件的启动共用同一输入。
	boot.ConfigSchemaHash = nil
	if (hasInstalledRoutes(in) || in.Role == state.RoleWebUI) && schemas.config != nil {
		boot.ConfigSchemaHash = schemas.config.Hash()
	}
	if err := c.assembly.bootstrapProfiles(ctx, in.PluginID, schemas.config, boot); err != nil {
		return supervisor.Spec{}, nil, err
	}
	boot.LogLevel = logging.ProtoLevel(c.output.Levels().Source(in.PluginID))
	database := boot.DatabaseDescriptor
	database.Backend = pb.DatabaseBackend_DATABASE_BACKEND_NONE
	database.MigrationMode = pb.MigrationMode_MIGRATION_MODE_NONE
	switch m.MigrationMode {
	case manifest.MigrationSelf:
		database.MigrationMode = pb.MigrationMode_MIGRATION_MODE_SELF
	case manifest.MigrationVersioned:
		database.MigrationMode = pb.MigrationMode_MIGRATION_MODE_VERSIONED
	}
	if (in.Role == state.RolePlugin || in.Role == state.RoleAgent || in.Role == state.RoleWebUI) && (m.PrepareEffects == manifest.EffectsManagedDB || m.MigrationMode == manifest.MigrationSelf || m.MigrationMode == manifest.MigrationVersioned) {
		database.Backend = pb.DatabaseBackend_DATABASE_BACKEND_SQLITE
	}
	boot.PluginDataDir, err = preparePluginDataDir(c.dir, in.PluginID)
	if err != nil {
		return supervisor.Spec{}, nil, err
	}
	running := &runningComponent{id: in.PluginID, install: in, caller: router.NewCaller(in.PluginID, boot.GenerationId), schemas: schemas}
	spec := supervisor.Spec{Role: in.Role, AdapterID: in.AdapterID, PluginID: in.PluginID, PluginDir: filepath.Join(c.dir, "plugins", in.PluginID),
		Executable: executable, Stdio: c.stdio, BuildDigest: in.BuildDigest, GenerationID: boot.GenerationId, Bootstrap: boot, ConfigSchema: resources.ConfigSchema, I18n: resources.I18n,
		DefaultLocale: c.assembly.settings.Settings().DefaultLocale, Handler: c.assembly.handler(running, in), Global: c.assembly.global,
		Limits: host.NewLimits(c.values), Stop: host.NewStopTiming(c.values), Log: c.output.Plugin(in.PluginID, boot.GenerationId, componentLogAttrs(in)...), Scheduler: c.assembly.scheduler, HTTP: c.assembly.http,
		MemorySoftLimit: c.values.PluginMemorySoftLimit, ExtraEnv: c.environment, Detach: func() { c.assembly.detach(running) }}
	if hasInstalledRoutes(in) {
		spec.Waker = c.assembly.waker
		running.activity = true
	}
	return spec, running, nil
}

// installedManifest 与启动共用安装身份核对，策略读取不会准备进程或业务资源。
func installedManifest(in state.Install, packageDir string) (*manifest.Result, error) {
	raw, err := os.ReadFile(filepath.Join(packageDir, "plugin.yaml"))
	if err != nil {
		return nil, err
	}
	parsed, err := manifest.Parse(raw)
	if err != nil {
		return nil, err
	}
	m := parsed.Manifest
	if m.ID != in.PluginID || state.Role(m.Role) != in.Role || m.AdapterID != in.AdapterID || m.Entrypoint != in.Entrypoint {
		return nil, fmt.Errorf("installed manifest identity differs from record for %s", in.PluginID)
	}
	if hasInstalledRoutes(in) && (m.Source != in.Source || m.Version != in.Version) {
		return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: fmt.Sprintf("installed manifest source or version differs for %s", in.PluginID)}
	}
	return parsed, nil
}
