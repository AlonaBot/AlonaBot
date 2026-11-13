package state

import (
	"context"
	"errors"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Task 保存安装身份下的动态实例；At、Interval、Cron 恰有一项有效。
type Task struct {
	PluginID, Name, Key string
	At                  time.Time
	Interval            time.Duration
	Cron, Timezone      string
	PayloadJSON         string
	Anchor, CreatedAt   time.Time // interval 起点与创建/替换时刻，重启保持原值。
}

type dynamicTask struct {
	PluginID    string        `gorm:"column:plugin_id;primaryKey"`
	Name        string        `gorm:"column:name;primaryKey"`
	Key         string        `gorm:"column:key;primaryKey"`
	At          int64         `gorm:"column:at"`
	Interval    time.Duration `gorm:"column:interval"`
	Cron        string        `gorm:"column:cron"`
	Timezone    string        `gorm:"column:timezone"`
	PayloadJSON string        `gorm:"column:payload_json"`
	Anchor      int64         `gorm:"column:anchor"`
	CreatedAt   int64         `gorm:"column:created_at;autoCreateTime:false"`
}

func (dynamicTask) TableName() string { return "dynamic_task" }

// 纳秒保留固定起点精度；0 代表该触发类型没有此时间字段。
func storedTime(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixNano()
}
func taskTime(n int64) time.Time {
	if n == 0 {
		return time.Time{}
	}
	return time.Unix(0, n).UTC()
}

// Tasks 读取该插件的全部实例，name 非空时只读取同名实例。
func (s *Store) Tasks(ctx context.Context, pluginID, name string) ([]Task, error) {
	if pluginID == "" {
		return nil, errEmptyPluginID
	}
	var rows []dynamicTask
	if err := s.db.WithContext(ctx).Where(&dynamicTask{PluginID: pluginID, Name: name}).Find(&rows).Error; err != nil {
		return nil, err
	}
	tasks := make([]Task, 0, len(rows))
	for _, r := range rows {
		tasks = append(tasks, Task{PluginID: r.PluginID, Name: r.Name, Key: r.Key, At: time.Time{}, Interval: r.Interval, Cron: r.Cron, Timezone: r.Timezone, PayloadJSON: r.PayloadJSON, Anchor: taskTime(r.Anchor), CreatedAt: taskTime(r.CreatedAt)})
		if r.Interval == 0 && r.Cron == "" {
			tasks[len(tasks)-1].At = time.UnixMilli(r.At).UTC()
		}
	}
	return tasks, nil
}

// SetTask 在同一事务内检查插件总配额并整体替换；替换已有键不占新配额。
func (s *Store) SetTask(ctx context.Context, task Task, maxTasks int) error {
	if task.PluginID == "" {
		return errEmptyPluginID
	}
	if task.Name == "" || task.Key == "" {
		return errors.New("task name and key are required")
	}
	return s.tx(ctx, func(tx *gorm.DB) error {
		var previous dynamicTask
		err := tx.Where(&dynamicTask{PluginID: task.PluginID, Name: task.Name, Key: task.Key}).Take(&previous).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var count int64
			if err := tx.Model(&dynamicTask{}).Where(&dynamicTask{PluginID: task.PluginID}).Count(&count).Error; err != nil {
				return err
			}
			if count >= int64(maxTasks) {
				return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, Message: "plugin task quota exhausted"}
			}
		}
		row := dynamicTask{PluginID: task.PluginID, Name: task.Name, Key: task.Key, At: 0, Interval: task.Interval, Cron: task.Cron, Timezone: task.Timezone, PayloadJSON: task.PayloadJSON, Anchor: storedTime(task.Anchor), CreatedAt: storedTime(task.CreatedAt)}
		if !task.At.IsZero() {
			row.At = task.At.UnixMilli()
		}
		return tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error
	})
}

// DeleteTask 在状态库事务内删除实例；缺失时返回 ErrNotFound。
func (s *Store) DeleteTask(ctx context.Context, pluginID, name, key string) error {
	if pluginID == "" {
		return errEmptyPluginID
	}
	if name == "" || key == "" {
		return errors.New("task name and key are required")
	}
	return s.tx(ctx, func(tx *gorm.DB) error {
		result := tx.Delete(&dynamicTask{PluginID: pluginID, Name: name, Key: key})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}
