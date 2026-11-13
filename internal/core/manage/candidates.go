package manage

import (
	"context"
	"errors"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

// Candidates 接收通过管理前置校验的候选操作，由候选编排器实现。
type Candidates interface {
	Reload(ctx context.Context, id string) error
	Update(ctx context.Context, id, staged string) error
	Rollback(ctx context.Context, id, generation string) error
}

type unsupportedCandidates struct{}

func (unsupportedCandidates) Reload(context.Context, string) error           { return unsupported() }
func (unsupportedCandidates) Update(context.Context, string, string) error   { return unsupported() }
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
	if key != in.PublisherKeyID {
		if err := s.opt.Store.SetPublisherKeyID(ctx, id, key); err != nil {
			return err
		}
	}
	return s.opt.Candidates.Update(ctx, id, candidate.Dir)
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

// Rollback 校验代次属于该安装，再交给候选编排器。
func (s *Service) Rollback(ctx context.Context, id, generation string, confirm bool) (err error) {
	defer finish(&err)
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
	// shortcut: 首装与 restart 的代次共用当前程序包，候选编排接入后按目标代次的程序版本检查。
	affected, err := s.affected(ctx, in, in.Version)
	if err != nil {
		return err
	}
	if len(affected) != 0 && !confirm {
		return &ErrNeedsConfirm{Affected: affected}
	}
	return s.opt.Candidates.Rollback(ctx, id, generation)
}
