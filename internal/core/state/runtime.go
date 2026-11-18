package state

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RuntimeState 描述插件或 Agent 的实际运行状态。
type RuntimeState string

const (
	RuntimeStopped   RuntimeState = "STOPPED"
	RuntimeWaking    RuntimeState = "WAKING"
	RuntimeActive    RuntimeState = "ACTIVE"
	RuntimeQuiescing RuntimeState = "QUIESCING"
	RuntimeSuspended RuntimeState = "SUSPENDED"
	RuntimeFailed    RuntimeState = "FAILED"
)

// RuntimeStatus 保存最近一次状态变化与实际唤醒、休眠的记录。
type RuntimeStatus struct {
	State          RuntimeState
	Reason         string
	ChangedAt      time.Time
	LastWakeReason string    // 空串表示尚未唤醒。
	LastWakeAt     time.Time // 零值表示尚未唤醒。
	LastSuspendAt  time.Time // 零值表示尚未休眠。
}

type pluginRuntime struct {
	PluginID string `gorm:"primaryKey"`
	RuntimeStatus
}

func (pluginRuntime) TableName() string { return "plugin_runtime" }

func stoppedRuntime(reason string) RuntimeStatus {
	return RuntimeStatus{State: RuntimeStopped, Reason: reason, ChangedAt: time.Now().UTC()}
}

// SetRuntimeStatus 替换已安装插件或 Agent 的运行状态，独立于启用意图和代次。
func (s *Store) SetRuntimeStatus(ctx context.Context, pluginID string, status RuntimeStatus) error {
	result := s.db.WithContext(ctx).Model(&pluginRuntime{}).Where(clause.Eq{Column: columnPluginID, Value: pluginID}).Select("*").Updates(status)
	if result.Error != nil {
		return fmt.Errorf("write runtime status for %s: %w", pluginID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("runtime status for %s: %w", pluginID, ErrNotFound)
	}
	return nil
}

// RuntimeStatus 读取已安装插件或 Agent 的运行状态，不接触进程或 Session。
func (s *Store) RuntimeStatus(ctx context.Context, pluginID string) (RuntimeStatus, error) {
	var row pluginRuntime
	if err := s.db.WithContext(ctx).Where(clause.Eq{Column: columnPluginID, Value: pluginID}).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = ErrNotFound
		}
		return RuntimeStatus{}, fmt.Errorf("read runtime status for %s: %w", pluginID, err)
	}
	status := row.RuntimeStatus
	status.ChangedAt = status.ChangedAt.UTC()
	status.LastWakeAt = status.LastWakeAt.UTC()
	status.LastSuspendAt = status.LastSuspendAt.UTC()
	return status, nil
}
