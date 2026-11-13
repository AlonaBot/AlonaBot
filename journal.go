package alonabot

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/AlonaBot/AlonaBot/contracts/i18n"
	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/process"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
)

// installedSpec 在停机前只读构造完整启动输入；调用方决定准备或直接启动。
// 返回的运行上下文与 Handler/Detach 共用 Caller，启动后由调用方附加 Instance。
func (c *Core) installedSpec(ctx context.Context, in state.Install, packageDir string, boot *pb.Bootstrap) (supervisor.Spec, *runningComponent, error) {
	raw, err := os.ReadFile(filepath.Join(packageDir, "plugin.yaml"))
	if err != nil {
		return supervisor.Spec{}, nil, err
	}
	parsed, err := manifest.Parse(raw)
	if err != nil {
		return supervisor.Spec{}, nil, err
	}
	m := parsed.Manifest
	if m.ID != in.PluginID || state.Role(m.Role) != in.Role || m.AdapterID != in.AdapterID || m.Entrypoint != in.Entrypoint {
		return supervisor.Spec{}, nil, fmt.Errorf("installed manifest identity differs from record for %s", in.PluginID)
	}
	for _, warning := range parsed.Warnings {
		slog.Warn("installed manifest warning", "plugin_id", m.ID, "warning", warning)
	}
	executable, err := process.Entrypoint(packageDir, in.Entrypoint)
	if err != nil {
		return supervisor.Spec{}, nil, err
	}
	resources, err := packageResources(packageDir)
	if err != nil {
		return supervisor.Spec{}, nil, err
	}
	schemas, err := optionSchemas(resources)
	if err != nil {
		return supervisor.Spec{}, nil, err
	}
	if _, problems := i18n.Parse(resources.I18n); len(problems) != 0 {
		return supervisor.Spec{}, nil, fmt.Errorf("installed message catalog: %v", problems)
	}
	// Schema 摘要与版本绑定包内声明，所有已安装组件的启动共用同一输入。
	boot.TargetConfigVersion = *m.ConfigVersion
	boot.ConfigSchemaHash = nil
	if (in.Role == state.RolePlugin || in.Role == state.RoleAgent) && schemas.config != nil {
		if uint32(schemas.config.Version()) != boot.TargetConfigVersion {
			return supervisor.Spec{}, nil, fmt.Errorf("config schema version differs from manifest for %s", in.PluginID)
		}
		boot.ConfigSchemaHash = schemas.config.Hash()
	}
	boot.LogLevel = logging.ProtoLevel(c.output.Levels().Source(in.PluginID))
	database := boot.DatabaseDescriptor
	database.Backend = pb.DatabaseBackend_DATABASE_BACKEND_NONE
	database.MigrationMode = pb.MigrationMode_MIGRATION_MODE_NONE
	database.ManualMigrationApproved = false
	switch m.MigrationMode {
	case manifest.MigrationSelf:
		database.MigrationMode = pb.MigrationMode_MIGRATION_MODE_SELF
	case manifest.MigrationVersioned:
		database.MigrationMode = pb.MigrationMode_MIGRATION_MODE_VERSIONED
	}
	if (in.Role == state.RolePlugin || in.Role == state.RoleAgent) && (m.PrepareEffects == manifest.EffectsManagedDB || m.MigrationMode == manifest.MigrationSelf || m.MigrationMode == manifest.MigrationVersioned) {
		database.Backend = pb.DatabaseBackend_DATABASE_BACKEND_SQLITE
	}
	running := &runningComponent{id: in.PluginID, install: in, caller: router.NewCaller(in.PluginID, boot.GenerationId), schemas: schemas}
	spec := supervisor.Spec{Role: in.Role, AdapterID: in.AdapterID, PluginID: in.PluginID, PluginDir: filepath.Join(c.dir, "plugins", in.PluginID),
		Executable: executable, BuildDigest: in.BuildDigest, GenerationID: boot.GenerationId, Bootstrap: boot, ConfigSchema: resources.ConfigSchema, I18n: resources.I18n,
		DefaultLocale: c.assembly.settings.Settings().DefaultLocale, Handler: c.assembly.handler(running, in), Global: c.assembly.global,
		Limits: host.NewLimits(c.values), Stop: host.NewStopTiming(c.values), Log: c.output.Plugin(in.PluginID, boot.GenerationId), Scheduler: c.assembly.scheduler, HTTP: c.assembly.http,
		MemorySoftLimit: c.values.PluginMemorySoftLimit, ExtraEnv: c.environment, Detach: func() { c.assembly.detach(running) }}
	return spec, running, nil
}
