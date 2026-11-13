package state

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm/clause"
)

// Job 保存静态任务的 interval 起点、间隔和暂停状态；cron 的起点与间隔为零。
type Job struct {
	PluginID, JobID string
	Anchor          time.Time
	Interval        time.Duration
	Paused          bool
}

type scheduledJob struct {
	PluginID string        `gorm:"column:plugin_id;primaryKey"`
	JobID    string        `gorm:"column:job_id;primaryKey"`
	Anchor   int64         `gorm:"column:anchor"`
	Interval time.Duration `gorm:"column:interval"`
	Paused   bool          `gorm:"column:paused"`
}

func (scheduledJob) TableName() string { return "scheduled_job" }

// Jobs 读取插件的全部静态调度状态。
func (s *Store) Jobs(ctx context.Context, pluginID string) ([]Job, error) {
	if pluginID == "" {
		return nil, errEmptyPluginID
	}
	var rows []scheduledJob
	if err := s.db.WithContext(ctx).Where(&scheduledJob{PluginID: pluginID}).Find(&rows).Error; err != nil {
		return nil, err
	}
	jobs := make([]Job, 0, len(rows))
	for _, r := range rows {
		jobs = append(jobs, Job{PluginID: r.PluginID, JobID: r.JobID, Anchor: taskTime(r.Anchor), Interval: r.Interval, Paused: r.Paused})
	}
	return jobs, nil
}

// SaveJob 按稳定键整体保存状态，false 和零间隔也会覆盖。
func (s *Store) SaveJob(ctx context.Context, job Job) error {
	if job.PluginID == "" {
		return errEmptyPluginID
	}
	if job.JobID == "" {
		return errors.New("job_id is required")
	}
	row := scheduledJob{PluginID: job.PluginID, JobID: job.JobID, Anchor: storedTime(job.Anchor), Interval: job.Interval, Paused: job.Paused}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error
}
