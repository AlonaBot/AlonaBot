package supervisor

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
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
	previous := status.State
	now := time.Now().UTC()
	if next == state.RuntimeActive && status.State == state.RuntimeWaking {
		status.LastWakeAt, status.LastWakeReason = now, reason
	}
	status.State, status.Reason, status.ChangedAt = next, reason, now
	if next == state.RuntimeSuspended {
		status.LastSuspendAt = status.ChangedAt
	}
	if err := r.key.store.SetRuntimeStatus(ctx, r.key.id, status); err != nil {
		return err
	}
	logRuntimeChange(r.key.id, previous, next, reason)
	return nil
}

func logRuntimeChange(id string, previous, next state.RuntimeState, reason string) {
	if previous != next {
		message := fmt.Sprintf("Component %s changed from %s to %s", id, previous, next)
		level := slog.LevelDebug
		if next == state.RuntimeSuspended {
			level = slog.LevelInfo
			message = fmt.Sprintf("Component %s entered sleep", id)
		}
		if next == state.RuntimeActive && (previous == state.RuntimeSuspended || previous == state.RuntimeWaking) {
			level = slog.LevelInfo
			message = fmt.Sprintf("Component %s is ready", id)
		}
		if next == state.RuntimeFailed {
			level = slog.LevelError
			message = fmt.Sprintf("Component %s failed", id)
		}
		if reason != "" {
			description := reason
			if stopReason, ok := strings.CutPrefix(reason, "STOP_REASON_"); ok {
				description = strings.ToLower(strings.ReplaceAll(stopReason, "_", " "))
			}
			message += ": " + description
		}
		slog.Log(context.Background(), level, message, "plugin_id", id, "previous_state", previous, "state", next, "reason", reason)
	}
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
