package journal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/AlonaBot/AlonaBot/internal/core/candidate"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// Abort 记录失败，确认候选退出后清理本操作的资源，再按原运行状态恢复旧版。
// 请求取消后仍完成清理与恢复；返回值保留 cause 及本次清理、恢复的全部错误。
func (o *Orchestrator) Abort(ctx context.Context, r *Result, cause error) error {
	unlock := o.lock(r.record.PluginID)
	defer unlock()
	ctx = context.WithoutCancel(ctx)
	if r.cause == nil {
		r.cause = cause
	} else if cause != nil {
		cause = errors.Join(r.cause, cause)
	} else {
		cause = r.cause
	}
	current, err := o.opt.Store.GetJournal(ctx, r.OpID)
	if err != nil {
		return errors.Join(cause, err)
	}
	if current.Phase == phaseCleaned {
		return cause
	}
	if current.Phase == phaseCommitted {
		return errors.Join(cause, precondition("cannot abort a committed candidate"))
	}
	// 1. 元数据失败也先关闭候选，防止未提交进程继续占有写入资源。
	markErr := o.save(ctx, r, phaseFailed)
	if cause != nil {
		slog.Error("candidate preparation aborted", "op_id", r.OpID, "plugin_id", r.record.PluginID, "kind", r.record.Kind, "phase", phaseFailed, "error", cause)
	}
	if r.Session != nil {
		if err := r.Session.Close(); err != nil {
			return errors.Join(cause, markErr, err)
		}
		r.Session = nil
		r.record.SessionID = nil
	}
	if markErr != nil {
		return errors.Join(cause, markErr)
	}
	if err := o.save(ctx, r, phaseFailed); err != nil {
		return errors.Join(cause, err)
	}
	// 2. 停止失败的旧树仍须交给原 Stop 路径确认；取锁失败保留全部恢复材料。
	if r.stopAttempted && !r.oldStopped {
		if err := o.opt.Stop(ctx, r.record.PluginID, stopReason(Kind(r.record.Kind))); err != nil {
			return errors.Join(cause, err)
		}
		r.oldStopped = true
	}
	if r.oldStopped && !r.resumed {
		if err := checkRuntime(r.pluginDir); err != nil {
			return errors.Join(cause, err)
		}
	}
	active, err := o.opt.Store.Active(ctx, r.record.PluginID)
	if err != nil {
		return errors.Join(cause, err)
	}
	if active == r.Generation {
		return errors.Join(cause, state.ErrStaleGeneration)
	}
	// 并发提交可能已切换权威；只丢弃本候选，Resume 按当前 Active 建立新 Session。
	// 3. 清理不越过其他操作、已提交代次或 Pending 的有效引用。
	cleanupErr := o.discardGeneration(ctx, r)
	if cleanupErr == nil {
		cleanupErr = o.discardPackage(ctx, r)
	}
	// 旧版恢复即使遇到文件清理失败也继续；此时两棵树和系统锁均已核对。
	var resumeErr error
	if r.oldStopped && r.record.WasRunning && !r.resumed && r.record.Kind != string(KindFirstInstall) {
		resumeErr = o.opt.Resume(ctx, r.record.PluginID)
		if resumeErr == nil {
			r.resumed = true
		}
	}
	if err := errors.Join(cleanupErr, resumeErr); err != nil {
		slog.Error("candidate cleanup or recovery failed", "op_id", r.OpID, "plugin_id", r.record.PluginID, "kind", r.record.Kind, "phase", phaseFailed, "error", err)
		return errors.Join(cause, err)
	}
	r.record.PackageDir = ""
	if err := o.save(ctx, r, phaseCleaned); err != nil {
		return errors.Join(cause, err)
	}
	return cause
}

func (o *Orchestrator) discardGeneration(ctx context.Context, r *Result) error {
	if !r.generationOwned && !r.databaseOwned {
		return nil
	}
	records, err := o.opt.Store.Journals(ctx, "")
	if err != nil {
		return err
	}
	for _, record := range records {
		if record.OpID == r.OpID || record.Phase == phaseCleaned {
			continue
		}
		if record.NewGeneration == r.Generation || record.ExpectedOld == r.Generation || r.record.DatabasePath != "" && record.DatabasePath == r.record.DatabasePath {
			return precondition("candidate resources are referenced by journal " + record.OpID)
		}
	}
	if r.Prepared != nil {
		if err := candidate.Discard(ctx, o.opt.Store, o.opt.Secrets, r.Prepared); err != nil {
			return err
		}
	} else {
		// NONE 与备份中途失败复用相同的代次保护和 secret 收集规则。
		if err := o.opt.Store.DeleteCandidateGeneration(ctx, r.record.PluginID, r.Generation); err != nil {
			return err
		}
		if r.databaseOwned {
			if err := os.RemoveAll(filepath.Dir(r.record.DatabasePath)); err != nil {
				return err
			}
		}
		if err := o.collectSecrets(ctx, r.record.PluginID); err != nil {
			return err
		}
	}
	r.generationOwned, r.databaseOwned = false, false
	r.Prepared = nil
	r.record.DatabasePath = ""
	return o.save(ctx, r, phaseFailed)
}

func (o *Orchestrator) collectSecrets(ctx context.Context, id string) error {
	snapshots, err := o.opt.Store.PluginGenerationConfigs(ctx, id)
	if err != nil {
		return err
	}
	protected := make([][]byte, 0, len(snapshots)+1)
	for _, snapshot := range snapshots {
		protected = append(protected, []byte(snapshot))
	}
	pending, err := o.opt.Store.GetPendingConfig(ctx, id)
	if err != nil && !errors.Is(err, state.ErrNotFound) {
		return err
	}
	if err == nil {
		protected = append(protected, []byte(pending.ConfigJSON))
	}
	_, err = o.opt.Secrets.Collect(ctx, id, protected)
	return err
}

func (o *Orchestrator) discardPackage(ctx context.Context, r *Result) error {
	if !r.record.OwnsPackage {
		return nil
	}
	installs, err := o.opt.Store.Installs(ctx)
	if err != nil {
		return err
	}
	protected := false
	for _, in := range installs {
		if in.Source == "dev" {
			continue
		}
		directory, err := o.installedPackage(in)
		if err != nil {
			return err
		}
		if directory == r.record.PackageDir {
			protected = true
		}
	}
	records, err := o.opt.Store.Journals(ctx, "")
	if err != nil {
		return err
	}
	for _, record := range records {
		if record.OpID != r.OpID && record.Phase != phaseCleaned && record.PackageDir == r.record.PackageDir {
			protected = true
		}
	}
	if !protected {
		if err := os.RemoveAll(r.record.PackageDir); err != nil {
			return fmt.Errorf("remove candidate package: %w", err)
		}
	}
	r.record.OwnsPackage = false
	r.record.PackageDir = ""
	return o.save(ctx, r, phaseFailed)
}
