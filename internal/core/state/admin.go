package state

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	adminCredentialID     = 1
	initialAuthGeneration = 1
	columnExpiresAt       = "expires_at"
)

var errEmptyAdminSessionID = errors.New("admin session ID is required")

// AdminCredential 保存唯一管理凭证的摘要与认证代次。
type AdminCredential struct {
	ID             int    `gorm:"column:id;primaryKey;autoIncrement:false"`
	TokenSHA256    []byte `gorm:"column:token_sha256"`
	AuthGeneration int64  `gorm:"column:auth_generation"`
	UpdatedAt      int64  `gorm:"column:updated_at;autoUpdateTime:false"` // Unix 毫秒，由调用方赋值
}

// TableName 指定管理凭证表。
func (AdminCredential) TableName() string { return "admin_credential" }

// AdminSession 保存管理会话的 CSRF 凭证、认证代次与固定到期时刻。
type AdminSession struct {
	SessionID      string `gorm:"column:session_id;primaryKey"`
	CSRFToken      string `gorm:"column:csrf_token"`
	AuthGeneration int64  `gorm:"column:auth_generation"`
	ExpiresAt      int64  `gorm:"column:expires_at;autoUpdateTime:false"` // Unix 毫秒，由调用方赋值
}

// TableName 指定管理会话表。
func (AdminSession) TableName() string { return "admin_session" }

// AdminCredential 读取唯一管理凭证；尚未初始化时返回 ErrNotFound。
func (s *Store) AdminCredential(ctx context.Context) (AdminCredential, error) {
	return readAdminCredential(s.db.WithContext(ctx))
}

func readAdminCredential(db *gorm.DB) (AdminCredential, error) {
	var credential AdminCredential
	err := db.Where(clause.Eq{Column: clause.PrimaryColumn, Value: adminCredentialID}).Take(&credential).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return credential, ErrNotFound
	}
	return credential, err
}

// InitAdminToken 在事务内初始化唯一凭证；已有凭证时失败。updatedAt 使用 Unix 毫秒。
func (s *Store) InitAdminToken(ctx context.Context, digest []byte, updatedAt int64) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		_, err := readAdminCredential(tx)
		if err == nil {
			return errors.New("admin token is already initialized")
		}
		if !errors.Is(err, ErrNotFound) {
			return err
		}
		return tx.Create(&AdminCredential{
			ID: adminCredentialID, TokenSHA256: digest, AuthGeneration: initialAuthGeneration, UpdatedAt: updatedAt,
		}).Error
	})
}

// ResetAdminToken 替换摘要、递增认证代次，并在同一事务内撤销全部管理会话。
// 尚未初始化时返回 ErrNotFound；updatedAt 使用 Unix 毫秒。
func (s *Store) ResetAdminToken(ctx context.Context, digest []byte, updatedAt int64) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		// 1. 读出当前代次，再保存新凭证。
		credential, err := readAdminCredential(tx)
		if err != nil {
			return err
		}
		credential.TokenSHA256 = digest
		credential.AuthGeneration++
		credential.UpdatedAt = updatedAt
		if err := tx.Save(&credential).Error; err != nil {
			return err
		}
		// 2. 和凭证替换一起提交，避免遗留旧管理会话。
		return tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&AdminSession{}).Error
	})
}

// CreateAdminSession 在同一事务内清理 expires_at <= now 的旧会话并创建新会话。
// now 使用 Unix 毫秒；是否有效及代次是否匹配由认证方检查。
func (s *Store) CreateAdminSession(ctx context.Context, session AdminSession, now int64) error {
	if session.SessionID == "" {
		return errEmptyAdminSessionID
	}
	return s.tx(ctx, func(tx *gorm.DB) error {
		if err := tx.Where(clause.Lte{Column: clause.Column{Name: columnExpiresAt}, Value: now}).Delete(&AdminSession{}).Error; err != nil {
			return err
		}
		return tx.Create(&session).Error
	})
}

// AdminSession 按主键读取会话；没有记录时返回 ErrNotFound。
func (s *Store) AdminSession(ctx context.Context, sessionID string) (AdminSession, error) {
	if sessionID == "" {
		return AdminSession{}, errEmptyAdminSessionID
	}
	var session AdminSession
	err := s.db.WithContext(ctx).Where(clause.Eq{Column: clause.PrimaryColumn, Value: sessionID}).Take(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return session, ErrNotFound
	}
	return session, err
}

// DeleteAdminSession 按主键删除会话；没有记录时也成功。
func (s *Store) DeleteAdminSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return errEmptyAdminSessionID
	}
	return s.db.WithContext(ctx).Where(clause.Eq{Column: clause.PrimaryColumn, Value: sessionID}).Delete(&AdminSession{}).Error
}
