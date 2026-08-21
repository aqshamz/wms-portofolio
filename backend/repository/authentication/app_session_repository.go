package authentication

import (
	"context"
	"time"
	model "wms-api/models/authentication"

	"gorm.io/gorm"
)

type AuthenticatedAccount struct {
	SessionID         string
	AccountID         string
	Username          string
	Email             *string
	DisplayName       string
	PreferredTimezone *string
	IssuedAt          time.Time
	ExpiresAt         time.Time
}

type AppSessionRepository struct{ db *gorm.DB }

func NewAppSessionRepository(db *gorm.DB) *AppSessionRepository {
	return &AppSessionRepository{db: db}
}

func (r *AppSessionRepository) Create(ctx context.Context, session *model.AppSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *AppSessionRepository) FindValid(ctx context.Context, tokenHash string) (AuthenticatedAccount, error) {
	var account AuthenticatedAccount
	err := r.db.WithContext(ctx).
		Table("app_session AS s").
		Select(`s.session_id, s.account_id, a.username, a.email, a.display_name,
			a.preferred_timezone, s.issued_at, s.expires_at`).
		Joins("JOIN app_account a ON a.account_id = s.account_id").
		Joins("JOIN account_status ast ON ast.account_status_id = a.account_status_id").
		Where(`s.token_hash = ?
			AND s.revoked_at IS NULL
			AND s.expires_at > clock_timestamp()
			AND ast.is_active
			AND ast.allows_login
			AND (a.locked_until IS NULL OR a.locked_until <= clock_timestamp())`, tokenHash).
		Take(&account).Error
	return account, err
}

func (r *AppSessionRepository) Touch(ctx context.Context, tokenHash string) error {
	return r.db.WithContext(ctx).Model(&model.AppSession{}).
		Where("token_hash = ? AND revoked_at IS NULL AND expires_at > clock_timestamp()", tokenHash).
		Update("last_seen_at", gorm.Expr("clock_timestamp()")).Error
}

func (r *AppSessionRepository) Revoke(ctx context.Context, tokenHash, reasonCode string) error {
	reasonID := r.db.Model(&model.SessionRevocationReason{}).
		Select("session_revocation_reason_id").
		Where("code = ? AND is_active", reasonCode)
	result := r.db.WithContext(ctx).Model(&model.AppSession{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Updates(map[string]interface{}{
			"revoked_at":                   gorm.Expr("clock_timestamp()"),
			"session_revocation_reason_id": reasonID,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *AppSessionRepository) RevokeAll(ctx context.Context, tokenHash, reasonCode string) error {
	accountID := r.db.Model(&model.AppSession{}).
		Select("account_id").
		Where("token_hash = ? AND revoked_at IS NULL AND expires_at > clock_timestamp()", tokenHash)
	reasonID := r.db.Model(&model.SessionRevocationReason{}).
		Select("session_revocation_reason_id").
		Where("code = ? AND is_active", reasonCode)
	result := r.db.WithContext(ctx).Model(&model.AppSession{}).
		Where("account_id = (?) AND revoked_at IS NULL", accountID).
		Updates(map[string]interface{}{
			"revoked_at":                   gorm.Expr("clock_timestamp()"),
			"session_revocation_reason_id": reasonID,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
