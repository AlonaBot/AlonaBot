package journal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/AlonaBot/AlonaBot/internal/core/candidate"
	"github.com/AlonaBot/AlonaBot/internal/core/process"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/oslock"
)

// Recover 在 Core 持锁且尚未启动组件时清理全部未提交操作。
// committed 只由正常启动流程按 Active 恢复；取不到 runtime.lock 时保留操作和资源，
// blocked 返回持锁插件，调用方跳过启动并通过管理服务呈现 Degraded。
func (o *Orchestrator) Recover(ctx context.Context) (blocked []string, err error) {
	records, err := o.opt.Store.Journals(ctx, "")
	if err != nil {
		return nil, err
	}
	var failures []error
	for _, record := range records {
		if record.Phase == phaseCommitted || record.Phase == phaseCleaned {
			continue
		}
		if slices.Contains(blocked, record.PluginID) {
			continue
		}
		if err := o.recoverOperation(ctx, record); err != nil {
			if errors.Is(err, oslock.ErrLocked) {
				blocked = append(blocked, record.PluginID)
				continue
			}
			slog.Error("candidate recovery failed", "plugin_id", record.PluginID, "op_id", record.OpID, "phase", record.Phase, "error", err)
			failures = append(failures, fmt.Errorf("recover plugin %s operation %s: %w", record.PluginID, record.OpID, err))
		}
	}
	return blocked, errors.Join(failures...)
}

func (o *Orchestrator) recoverOperation(ctx context.Context, record state.Journal) (err error) {
	unlock := o.lock(record.PluginID)
	defer unlock()
	unlockState := o.opt.Store.LockPlugin(record.PluginID)
	defer unlockState()
	// 1. 只按状态库身份定位资源，清理全程持跨代次系统锁，防止旧写入者与回收重叠。
	switch record.Phase {
	case phaseStarted, phaseStoppedOld, phaseCandidateReady, phasePrepared, phaseFailed:
	default:
		return precondition("unknown journal recovery phase: " + record.Phase)
	}
	if record.NewGeneration == "." || record.NewGeneration == ".." || filepath.Base(record.NewGeneration) != record.NewGeneration {
		return precondition("journal generation is not a path segment")
	}
	pluginDir, err := filepath.Abs(filepath.Join(o.opt.DataDir, "plugins", record.PluginID))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(pluginDir, 0o700); err != nil {
		return err
	}
	lock, err := process.LockRuntime(pluginDir)
	if errors.Is(err, oslock.ErrLocked) {
		return fmt.Errorf("old process may still be running: %w", err)
	}
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, lock.Unlock()) }()
	active, err := o.opt.Store.Active(ctx, record.PluginID)
	if err != nil {
		return err
	}
	if active == record.NewGeneration {
		return precondition("uncommitted journal points to the active generation")
	}
	pending, err := o.opt.Store.GetPendingConfig(ctx, record.PluginID)
	if err != nil && !errors.Is(err, state.ErrNotFound) {
		return err
	}
	if err == nil && pending.BaseGeneration == record.NewGeneration {
		return precondition("candidate generation is referenced by pending config")
	}

	// 2. started 也可能已创建目录但尚未登记数据库路径；代次 ID 足以确定候选清理范围。
	r := &Result{OpID: record.OpID, Generation: record.NewGeneration, record: record, pluginDir: pluginDir, generationOwned: true, databaseOwned: true,
		Prepared: &candidate.Prepared{PluginID: record.PluginID, Generation: record.NewGeneration,
			DatabasePath: filepath.Join(pluginDir, "generations", record.NewGeneration, "db.sqlite")}}
	if err := o.save(ctx, r, phaseFailed); err != nil {
		return err
	}
	if err := o.discardGeneration(ctx, r); err != nil {
		return err
	}
	if err := o.discardPackage(ctx, r); err != nil {
		return err
	}
	r.record.SessionID = nil
	r.record.PackageDir = ""
	return o.save(ctx, r, phaseCleaned)
}
