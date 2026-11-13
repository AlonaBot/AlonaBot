package journal

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
)

// Commit 复核持久资源，在元数据事务提交后激活候选，将运行资源交给调用方。
// 提交前失败会 Abort；提交后失败仍返回 Degraded Running，保留新代次权威。
func (o *Orchestrator) Commit(ctx context.Context, r *Result) (*supervisor.Running, error) {
	unlock := o.lock(r.record.PluginID)
	committed := false
	running, err := o.commit(ctx, r, &committed)
	// Abort 自己取得同一插件的编排锁，先归还本次提交锁。
	unlock()
	if err != nil && !committed {
		return nil, o.Abort(ctx, r, err)
	}
	if err != nil {
		slog.Error("committed candidate activation failed", "plugin_id", r.record.PluginID, "op_id", r.OpID, "generation_id", r.Generation, "kind", r.record.Kind, "phase", phaseCommitted, "error", err)
	}
	return running, err
}

func (o *Orchestrator) commit(ctx context.Context, r *Result, committed *bool) (*supervisor.Running, error) {
	// 1. 文件检查和同步先完成；提交回调只访问状态库元数据。
	current, err := o.opt.Store.GetJournal(ctx, r.OpID)
	if err != nil {
		return nil, err
	}
	if current.Phase != phasePrepared {
		return nil, precondition("candidate journal is not prepared")
	}
	if _, err := os.Stat(filepath.Join(current.PackageDir, current.Entrypoint)); err != nil {
		return nil, err
	}
	if err := persistPackage(current.PackageDir); err != nil {
		return nil, err
	}
	if current.DatabasePath != "" {
		if err := persistDatabase(current.DatabasePath, false); err != nil {
			return nil, err
		}
	}
	if _, err := os.Stat(r.pluginDir); err != nil {
		return nil, err
	}
	// 2. Activate 只在回调成功后发送 ACTIVATE，返回值携带激活及失败清理所有权。
	return r.Session.Activate(ctx, func(ctx context.Context) error {
		if err := o.opt.Store.CommitCandidate(ctx, r.OpID, r.record.PluginID, r.record.ExpectedOld, r.Generation); err != nil {
			return err
		}
		*committed = true
		r.record.Phase = phaseCommitted
		return nil
	})
}
