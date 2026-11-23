package state

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProfileSession 保存实例和完整会话字符串组成的身份。
type ProfileSession struct {
	Instance string `json:"instance"`
	Session  string `json:"session"`
}

// ProfileSessionAccess 保存配置文件的一份完整会话访问设置。
type ProfileSessionAccess struct {
	ProfileID string           `gorm:"column:profile_id;primaryKey"`
	Mode      string           `gorm:"column:mode"`
	Sessions  []ProfileSession `gorm:"column:sessions;serializer:json"`
}

// TableName 指定会话访问设置表。
func (ProfileSessionAccess) TableName() string { return "profile_session_access" }

// ProfileSessionAccess 返回指定配置文件的设置；无记录时返回零值。
func (s *Store) ProfileSessionAccess(ctx context.Context, id string) (ProfileSessionAccess, error) {
	return readProfileSessionAccess(s.db.WithContext(ctx), id)
}

func readProfileSessionAccess(db *gorm.DB, id string) (ProfileSessionAccess, error) {
	var row ProfileSessionAccess
	err := db.Where(clause.Eq{Column: clause.PrimaryColumn, Value: id}).Find(&row).Error
	return row, err
}

// SessionProfileAccess 在同一快照内按会话指定、实例缺省、基础配置文件的顺序读取设置。
func (s *Store) SessionProfileAccess(ctx context.Context, instance, session, defaultID string) (ProfileSessionAccess, error) {
	var access ProfileSessionAccess
	err := s.tx(ctx, func(tx *gorm.DB) error {
		var assignments []ProfileAssignment
		if err := tx.Where(clause.Eq{Column: columnProfileInstance, Value: instance}).
			Where(clause.IN{Column: columnProfileSession, Values: []any{session, ""}}).Find(&assignments).Error; err != nil {
			return err
		}
		id := ResolveProfileAssignment(assignments, instance, session)
		if id == "" {
			id = defaultID
		}
		var err error
		access, err = readProfileSessionAccess(tx, id)
		return err
	}, &sql.TxOptions{ReadOnly: true})
	return access, err
}

// PutProfileSessionAccess 原子替换访问模式和列表。
func (s *Store) PutProfileSessionAccess(ctx context.Context, row ProfileSessionAccess) error {
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error
}
