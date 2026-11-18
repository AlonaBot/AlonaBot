package supervisor

import (
	"context"
	"sync"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

type runtimeKey struct {
	store *state.Store
	id    string
}

// runtimeOwners 的每个键由同一 Store 的逐插件锁保护。
var runtimeOwners sync.Map // runtimeKey → *runtimeRecord

// runtimeRecord 跟踪本次进程启动，代次相同的两次启动也有各自的记录。
type runtimeRecord struct {
	key runtimeKey
}

func (r *runtimeRecord) write(ctx context.Context, next state.RuntimeState, reason string) error {
	if r == nil {
		return nil
	}
	status, err := r.key.store.RuntimeStatus(ctx, r.key.id)
	if err != nil {
		return err
	}
	status.State, status.Reason, status.ChangedAt = next, reason, time.Now().UTC()
	return r.key.store.SetRuntimeStatus(ctx, r.key.id, status)
}

func (r *runtimeRecord) writeIfCurrent(ctx context.Context, next state.RuntimeState, reason string) error {
	if r == nil {
		return nil
	}
	if owner, _ := runtimeOwners.Load(r.key); owner != r {
		return nil // 状态已由另一次启动或管理操作发布。
	}
	return r.write(ctx, next, reason)
}

func (r *runtimeRecord) release() {
	if r != nil {
		runtimeOwners.CompareAndDelete(r.key, r)
	}
}
