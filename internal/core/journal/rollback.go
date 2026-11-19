package journal

import (
	"context"
	"slices"

	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// PrepareRollback 从保留代次复制新候选，返回 PREPARED Session，由调用方 Commit 或 Abort。
func (o *Orchestrator) PrepareRollback(ctx context.Context, pluginID, generation string, manualMigrationApproved bool) (*Result, error) {
	return o.prepare(ctx, pluginID, KindRollback, nil, generation, manualMigrationApproved)
}

func (o *Orchestrator) rollbackSource(ctx context.Context, id, generation, active string, records []state.Journal) (state.Journal, error) {
	retained, err := o.opt.Store.Retained(ctx, id)
	if err != nil {
		return state.Journal{}, err
	}
	if generation == active || !slices.Contains(retained, generation) {
		return state.Journal{}, precondition("rollback requires a retained generation other than active")
	}
	for _, record := range records {
		if record.Phase == phaseCommitted && record.NewGeneration == generation {
			return record, nil
		}
	}
	return state.Journal{}, precondition("retained generation has no committed program snapshot")
}
