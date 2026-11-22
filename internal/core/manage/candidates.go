package manage

import (
	"context"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/trust"
)

// Candidates 接收通过管理前置校验的候选操作，由候选编排器实现。
type Candidates interface {
	Reload(ctx context.Context, id string, manualMigrationApproved bool) error
	Update(ctx context.Context, id string, staged *packages.Candidate, manualMigrationApproved bool) error
	Rollback(ctx context.Context, id, generation string, manualMigrationApproved bool) error
}

type unsupportedCandidates struct{}

func (unsupportedCandidates) Reload(context.Context, string, bool) error { return unsupported() }
func (unsupportedCandidates) Update(context.Context, string, *packages.Candidate, bool) error {
	return unsupported()
}
func (unsupportedCandidates) Rollback(context.Context, string, string, bool) error {
	return unsupported()
}
func unsupported() error {
	return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, Message: "candidate orchestration is not available"}
}

// Reload 校验安装与运行状态，再交给候选编排器；显式维护许可允许已停用的 external 组件。
func (s *Service) Reload(ctx context.Context, id string, manualMigrationApproved bool) (err error) {
	defer finish(&err)
	unlock := s.lock(id)
	defer unlock()
	in, err := s.installed(ctx, id)
	if err != nil {
		return err
	}
	if in.Role == "webui" && in.Entrypoint == "" {
		return precondition("webui has no process to reload")
	}
	if !manualMigrationApproved && s.opt.State(id) != "Activated" {
		return precondition("reload requires an activated component")
	}
	if in.Source == "dev" {
		if manualMigrationApproved {
			return precondition("manual migration requires an installed external candidate")
		}
		pending, found, err := s.configurations().Pending(ctx, id)
		if err != nil {
			return err
		}
		if found {
			// 开发组件从编辑文件启动；保留 Pending 直到重启成功，失败仍可修订后重载。
			name := filepath.Join(filepath.Dir(s.opt.Packages), "plugins", id, configSnapshotFile)
			if err := replaceDevelopmentConfig(name, pending.ConfigJSON); err != nil {
				return err
			}
			if err := s.opt.Candidates.Reload(ctx, id, manualMigrationApproved); err != nil {
				return err
			}
			return s.configurations().Discard(ctx, id)
		}
	}
	return s.opt.Candidates.Reload(ctx, id, manualMigrationApproved)
}

// UpdateResult 返回本次更新的清单警告。
type UpdateResult struct {
	Warnings []string `json:"warnings,omitempty"`
}

// Update 验证来源、公钥、pin 与降版影响，再移交暂存候选。
func (s *Service) Update(ctx context.Context, id string, req InstallRequest, manualMigrationApproved bool) (result UpdateResult, err error) {
	defer finish(&err)
	s.catalog.Lock()
	defer s.catalog.Unlock()
	unlock := s.lock(id)
	defer unlock()
	in, err := s.installed(ctx, id)
	if err != nil {
		return result, err
	}
	if in.Source == "dev" {
		return result, developmentError(id)
	}
	if req.Source == "" && req.Archive == "" {
		bundled, version, err := s.bundledUpdate(ctx, id)
		if err != nil {
			return result, err
		}
		if version != "" {
			bundled.Confirm, bundled.Replace = req.Confirm, req.Replace
			req = bundled
		} else {
			req.Source = strings.TrimPrefix(in.Source, "github.com/")
		}
	}
	candidate, err := s.prepare(ctx, req)
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, candidate.Discard()) }()
	result.Warnings = candidate.Warnings
	m := candidate.Manifest
	if err := checkWebUIVersion(m); err != nil {
		return result, err
	}
	if m.ID != id || m.Role != string(in.Role) || m.AdapterID != in.AdapterID {
		return result, invalid("candidate identity differs from installed component")
	}
	if m.Source != in.Source {
		return result, denied("candidate source differs from installed binding")
	}
	if in.Role == state.RoleWebUI && in.Entrypoint != "" && m.Entrypoint == "" {
		return result, precondition("updating a WebUI with an entrypoint to a static WebUI is not supported")
	}
	key := candidate.Release.PublisherKeyID
	if err := checkPublisher(in.PublisherKeyID, key, req.Confirm); err != nil {
		return result, err
	}
	if in.Pin != "" && m.Version != in.Pin {
		return result, precondition("candidate version does not match pin: " + in.Pin)
	}
	if key == "" && !req.Confirm {
		return result, &ErrNeedsConfirm{Affected: []string{"unsigned package: " + id}}
	}
	affected, err := s.affected(ctx, in, m.Version)
	if err != nil {
		return result, err
	}
	if len(affected) != 0 && !req.Confirm {
		return result, &ErrNeedsConfirm{Affected: affected}
	}
	if err := s.checkCycles(ctx, m, nil); err != nil {
		return result, err
	}
	if err := s.dependencies(ctx, m); err != nil {
		return result, err
	}
	active, err := s.opt.Store.Active(ctx, id)
	if err != nil {
		return result, err
	}
	firstWebUIProcess := in.Role == state.RoleWebUI && in.Entrypoint == "" && m.Entrypoint != ""
	if in.Role == state.RoleWebUI && in.Entrypoint == "" && m.Entrypoint == "" || active == "" && !firstWebUIProcess {
		// 直接替换包时不会进入 Journal，仍须遵守更新的维护门槛。
		if manualMigrationApproved {
			return result, precondition("manual migration requires an active process generation")
		}
		previous, err := s.manifest(ctx, id)
		if err != nil {
			return result, err
		}
		if previous.PrepareEffects == manifest.EffectsExternal || m.PrepareEffects == manifest.EffectsExternal {
			return result, precondition("NEEDS_MANUAL_MIGRATION")
		}
		digest, err := hex.DecodeString(candidate.Asset.SHA256)
		if err != nil {
			return result, err
		}
		if _, err := candidate.MoveTo(s.opt.Packages); err != nil {
			return result, err
		}
		return result, s.opt.Store.ReplaceProgram(ctx, id, m.Version, m.Entrypoint, key, digest)
	}
	return result, s.opt.Candidates.Update(ctx, id, candidate, manualMigrationApproved)
}

func checkPublisher(bound, candidate string, confirm bool) error {
	if bound == candidate {
		return nil
	}
	if bound != "" {
		return denied("candidate publisher differs from installed binding")
	}
	if official, ok := trust.Official(); ok && candidate == release.KeyID(official) {
		return nil
	}
	if !confirm {
		return &ErrNeedsConfirm{Affected: []string{"trust publisher: " + candidate}}
	}
	return nil
}

// Rollback 按目标保留代次检查依赖与丢弃写入的确认，再交给候选编排器。
func (s *Service) Rollback(ctx context.Context, id, generation string, confirm, manualMigrationApproved bool) (err error) {
	defer finish(&err)
	s.catalog.Lock()
	defer s.catalog.Unlock()
	unlock := s.lock(id)
	defer unlock()
	in, err := s.installed(ctx, id)
	if err != nil {
		return err
	}
	if in.Source == "dev" {
		return developmentError(id)
	}
	if generation == "" {
		return invalid("generation is required")
	}
	configs, err := s.opt.Store.PluginGenerationConfigs(ctx, id)
	if err != nil {
		return err
	}
	if _, ok := configs[generation]; !ok {
		return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "generation does not belong to component"}
	}
	m, err := s.rollbackManifest(ctx, in, generation)
	if err != nil {
		return err
	}
	affected, err := s.affected(ctx, in, m.Version)
	if err != nil {
		return err
	}
	if !confirm {
		return &ErrNeedsConfirm{Affected: append([]string{"rollback to generation: " + generation, "discard local writes after generation: " + generation}, affected...)}
	}
	if err := s.checkCycles(ctx, m, nil); err != nil {
		return err
	}
	if err := s.dependencies(ctx, m); err != nil {
		return err
	}
	return s.opt.Candidates.Rollback(ctx, id, generation, manualMigrationApproved)
}

func (s *Service) rollbackManifest(ctx context.Context, in state.Install, generation string) (*manifest.Manifest, error) {
	retained, err := s.opt.Store.Retained(ctx, in.PluginID)
	if err != nil {
		return nil, err
	}
	active, err := s.opt.Store.Active(ctx, in.PluginID)
	if err != nil {
		return nil, err
	}
	if generation == active || !slices.Contains(retained, generation) {
		return nil, precondition("rollback requires a retained generation other than active")
	}
	records, err := s.opt.Store.Journals(ctx, in.PluginID)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		if record.Phase != "committed" || record.NewGeneration != generation {
			continue
		}
		in.Version, in.BuildDigest = record.Version, record.BuildDigest
		raw, err := os.ReadFile(filepath.Join(s.directory(in), "plugin.yaml"))
		if err != nil {
			return nil, err
		}
		parsed, err := manifest.Parse(raw)
		if err != nil {
			return nil, err
		}
		return parsed.Manifest, nil
	}
	return nil, precondition("retained generation has no committed program snapshot")
}
