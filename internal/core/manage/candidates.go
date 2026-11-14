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
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// Candidates 接收通过管理前置校验的候选操作，由候选编排器实现。
type Candidates interface {
	Reload(ctx context.Context, id string) error
	Update(ctx context.Context, id string, staged *packages.Candidate) error
	Rollback(ctx context.Context, id, generation string) error
}

type unsupportedCandidates struct{}

func (unsupportedCandidates) Reload(context.Context, string) error { return unsupported() }
func (unsupportedCandidates) Update(context.Context, string, *packages.Candidate) error {
	return unsupported()
}
func (unsupportedCandidates) Rollback(context.Context, string, string) error { return unsupported() }
func unsupported() error {
	return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, Message: "candidate orchestration is not available"}
}

// Reload 校验安装与运行状态，再交给候选编排器。
func (s *Service) Reload(ctx context.Context, id string) (err error) {
	defer finish(&err)
	unlock := s.lock(id)
	defer unlock()
	in, err := s.installed(ctx, id)
	if err != nil {
		return err
	}
	if in.Role == "webui" {
		return precondition("webui has no process to reload")
	}
	if s.opt.State(id) != "Activated" {
		return precondition("reload requires an activated component")
	}
	return s.opt.Candidates.Reload(ctx, id)
}

// Update 验证来源、公钥、pin 与降版影响，再移交暂存候选。
func (s *Service) Update(ctx context.Context, id string, req InstallRequest) (err error) {
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
	if req.Source == "" && req.Archive == "" {
		req.Source = strings.TrimPrefix(in.Source, "github.com/")
	}
	candidate, err := s.prepare(ctx, req)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, candidate.Discard()) }()
	m := candidate.Manifest
	if err := checkWebUIVersion(m); err != nil {
		return err
	}
	if m.ID != id || m.Role != string(in.Role) || m.AdapterID != in.AdapterID {
		return invalid("candidate identity differs from installed component")
	}
	if m.Source != in.Source {
		return denied("candidate source differs from installed binding")
	}
	key := candidate.Release.PublisherKeyID
	if err := checkPublisher(in.PublisherKeyID, key, req.Confirm); err != nil {
		return err
	}
	if in.Pin != "" && m.Version != in.Pin {
		return precondition("candidate version does not match pin: " + in.Pin)
	}
	if key == "" && !req.Confirm {
		return &ErrNeedsConfirm{Affected: []string{"unsigned package: " + id}}
	}
	affected, err := s.affected(ctx, in, m.Version)
	if err != nil {
		return err
	}
	if len(affected) != 0 && !req.Confirm {
		return &ErrNeedsConfirm{Affected: affected}
	}
	if err := s.checkCycles(ctx, m, nil); err != nil {
		return err
	}
	if err := s.dependencies(ctx, m); err != nil {
		return err
	}
	active, err := s.opt.Store.Active(ctx, id)
	if err != nil {
		return err
	}
	if in.Role == state.RoleWebUI || active == "" {
		digest, err := hex.DecodeString(candidate.Asset.SHA256)
		if err != nil {
			return err
		}
		if _, err := candidate.MoveTo(s.opt.Packages); err != nil {
			return err
		}
		return s.opt.Store.ReplaceProgram(ctx, id, m.Version, m.Entrypoint, key, digest)
	}
	return s.opt.Candidates.Update(ctx, id, candidate)
}

func checkPublisher(bound, candidate string, confirm bool) error {
	if bound == candidate {
		return nil
	}
	if bound != "" {
		return denied("candidate publisher differs from installed binding")
	}
	if !confirm {
		return &ErrNeedsConfirm{Affected: []string{"trust publisher: " + candidate}}
	}
	return nil
}

// Rollback 按目标保留代次检查依赖与丢弃写入的确认，再交给候选编排器。
func (s *Service) Rollback(ctx context.Context, id, generation string, confirm bool) (err error) {
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
	return s.opt.Candidates.Rollback(ctx, id, generation)
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
