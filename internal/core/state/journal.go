package state

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Journal 保存候选操作的身份、阶段和资源引用，供准备、清理与恢复交接。
// NewGeneration 在 started 时分配，generation 行在数据库准备完成后创建。
type Journal struct {
	OpID           string    `gorm:"column:op_id;primaryKey"`
	PluginID       string    `gorm:"column:plugin_id"`
	Kind           string    `gorm:"column:kind"`
	ExpectedOld    string    `gorm:"column:expected_old"`
	NewGeneration  string    `gorm:"column:new_generation"`
	Phase          string    `gorm:"column:phase"`
	PackageDir     string    `gorm:"column:package_dir"`
	DatabasePath   string    `gorm:"column:database_path"` // NONE 后端为空串。
	Entrypoint     string    `gorm:"column:entrypoint"`
	Version        string    `gorm:"column:version"`
	PublisherKeyID string    `gorm:"column:publisher_key_id"`
	BuildDigest    []byte    `gorm:"column:build_digest"`
	OwnsPackage    bool      `gorm:"column:owns_package"` // 只清理由本操作移动到不可变目录的程序包。
	WasRunning     bool      `gorm:"column:was_running"`
	WasEnabled     bool      `gorm:"column:was_enabled"`
	SessionID      []byte    `gorm:"column:session_id"` // 握手前为空 BLOB。
	CreatedAt      time.Time `gorm:"-"`
	UpdatedAt      time.Time `gorm:"-"`
}

const (
	journalCreatedAtColumn = "created_at"
	journalOpIDColumn      = "op_id"
)

type journalRecord struct {
	Journal         `gorm:"embedded"`
	CreatedAtMillis int64 `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAtMillis int64 `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (journalRecord) TableName() string { return "journal" }

func journalRow(record Journal) journalRecord {
	if record.SessionID == nil {
		record.SessionID = []byte{}
	}
	if record.BuildDigest == nil {
		record.BuildDigest = []byte{}
	}
	return journalRecord{Journal: record, CreatedAtMillis: record.CreatedAt.UnixMilli(), UpdatedAtMillis: record.UpdatedAt.UnixMilli()}
}

func (row journalRecord) record() Journal {
	row.CreatedAt = time.UnixMilli(row.CreatedAtMillis)
	row.UpdatedAt = time.UnixMilli(row.UpdatedAtMillis)
	return row.Journal
}

func journalPrecondition(message string) error {
	return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: message}
}

// CreateJournal 登记 started 操作；创建时间由状态库生成，资源尚未获得时保留空引用。
func (s *Store) CreateJournal(ctx context.Context, record Journal) error {
	if record.OpID == "" || record.PluginID == "" || record.NewGeneration == "" || record.Phase != "started" {
		return journalPrecondition("journal requires an operation, plugin, generation and started phase")
	}
	switch record.Kind {
	case "first_install", "reload", "update", "rollback":
	default:
		return journalPrecondition("invalid journal kind")
	}
	record.CreatedAt = time.Now()
	record.UpdatedAt = record.CreatedAt
	row := journalRow(record)
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	logJournal(record)
	return nil
}

// UpdateJournal 在事务中核对身份和合法阶段，连同时间写入实际持有的资源引用。
// committed 只由提交 Active 的事务写入，本入口拒绝该阶段。
func (s *Store) UpdateJournal(ctx context.Context, record Journal) error {
	changed := false
	err := s.tx(ctx, func(tx *gorm.DB) error {
		var previous journalRecord
		err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: record.OpID}).Take(&previous).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		old := previous.Journal
		if old.PluginID != record.PluginID || old.Kind != record.Kind || old.ExpectedOld != record.ExpectedOld || old.NewGeneration != record.NewGeneration || old.WasRunning != record.WasRunning || old.WasEnabled != record.WasEnabled || old.Entrypoint != record.Entrypoint || old.Version != record.Version || old.PublisherKeyID != record.PublisherKeyID || !bytes.Equal(old.BuildDigest, record.BuildDigest) {
			return journalPrecondition("journal operation identity changed")
		}
		if !journalTransition(old.Phase, record.Phase) {
			return journalPrecondition(fmt.Sprintf("invalid journal phase transition %s to %s", old.Phase, record.Phase))
		}
		record.CreatedAt = time.UnixMilli(previous.CreatedAtMillis)
		record.UpdatedAt = time.Now()
		row := journalRow(record)
		changed = old.Phase != record.Phase
		return tx.Save(&row).Error
	})
	if err == nil && changed {
		logJournal(record)
	}
	return err
}

func journalTransition(from, to string) bool {
	if to == "committed" || from == "committed" || from == "cleaned" {
		return false
	}
	if to == from {
		return true
	} // 同阶段可登记已取得或已释放的引用。
	if to == "failed" {
		return true
	}
	switch from {
	case "started":
		return to == "stopped_old"
	case "stopped_old":
		return to == "candidate_ready"
	case "candidate_ready":
		return to == "prepared"
	case "failed":
		return to == "cleaned"
	}
	return false
}

func logJournal(record Journal) {
	slog.Info("candidate journal phase changed", "op_id", record.OpID, "plugin_id", record.PluginID, "kind", record.Kind, "phase", record.Phase)
}

// GetJournal 读取指定操作；缺记录返回 ErrNotFound。
func (s *Store) GetJournal(ctx context.Context, opID string) (Journal, error) {
	var row journalRecord
	err := s.db.WithContext(ctx).Where(clause.Eq{Column: clause.PrimaryColumn, Value: opID}).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Journal{}, ErrNotFound
	}
	return row.record(), err
}

// Journals 按创建时间与操作 ID 排序；pluginID 为空时列出全部操作。
func (s *Store) Journals(ctx context.Context, pluginID string) ([]Journal, error) {
	db := s.db.WithContext(ctx)
	if pluginID != "" {
		db = db.Where(&journalRecord{Journal: Journal{PluginID: pluginID}})
	}
	var rows []journalRecord
	if err := db.Order(clause.OrderByColumn{Column: clause.Column{Name: journalCreatedAtColumn}}).Order(clause.OrderByColumn{Column: clause.Column{Name: journalOpIDColumn}}).Find(&rows).Error; err != nil {
		return nil, err
	}
	records := make([]Journal, 0, len(rows))
	for _, row := range rows {
		records = append(records, row.record())
	}
	return records, nil
}
