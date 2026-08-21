package authentication

import (
	"context"
	"time"
	model "wms-api/models/authentication"

	"gorm.io/gorm"
)

type LoginAccount struct {
	ID                string
	Username          string
	Email             *string
	DisplayName       string
	PasswordHash      *string
	StatusIsActive    bool
	AllowsLogin       bool
	FailedLoginCount  int
	LockedUntil       *time.Time
	MaxFailedAttempts int
	LockoutSeconds    int
	SessionTTLSeconds int
}

type AppAccountRepository struct{ db *gorm.DB }

func NewAppAccountRepository(db *gorm.DB) *AppAccountRepository {
	return &AppAccountRepository{db: db}
}

func (r *AppAccountRepository) FindForLogin(ctx context.Context, identifier string) (LoginAccount, error) {
	var account LoginAccount
	err := r.db.WithContext(ctx).
		Table("app_account AS a").
		Select(`a.account_id AS id, a.username, a.email, a.display_name,
			a.password_hash, ast.is_active AS status_is_active, ast.allows_login,
			a.failed_login_count, a.locked_until, p.max_failed_attempts,
			p.lockout_seconds, p.session_ttl_seconds`).
		Joins("JOIN account_status ast ON ast.account_status_id = a.account_status_id").
		Joins(`JOIN LATERAL (
			SELECT candidate.max_failed_attempts, candidate.lockout_seconds,
				candidate.session_ttl_seconds
			FROM authentication_policy candidate
			WHERE candidate.is_active
			  AND (candidate.authentication_policy_id = a.authentication_policy_id
			       OR candidate.is_default)
			ORDER BY
				(candidate.authentication_policy_id = a.authentication_policy_id) DESC,
				candidate.is_default DESC
			LIMIT 1
		) p ON true`).
		Where("lower(a.username) = lower(?) OR lower(a.email) = lower(?)", identifier, identifier).
		Take(&account).Error
	return account, err
}

func (r *AppAccountRepository) FindByUsername(ctx context.Context, username string) (model.AppAccount, error) {
	var account model.AppAccount
	err := r.db.WithContext(ctx).Where("lower(username) = lower(?)", username).Take(&account).Error
	return account, err
}

func (r *AppAccountRepository) Create(ctx context.Context, account *model.AppAccount) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *AppAccountRepository) RecordFailedLogin(ctx context.Context, accountID string) error {
	return r.db.WithContext(ctx).Exec(`
		WITH selected_policy AS (
			SELECT a.account_id, p.max_failed_attempts, p.lockout_seconds
			FROM app_account a
			JOIN LATERAL (
				SELECT candidate.max_failed_attempts, candidate.lockout_seconds
				FROM authentication_policy candidate
				WHERE candidate.is_active
				  AND (candidate.authentication_policy_id = a.authentication_policy_id
				       OR candidate.is_default)
				ORDER BY
					(candidate.authentication_policy_id = a.authentication_policy_id) DESC,
					candidate.is_default DESC
				LIMIT 1
			) p ON true
			WHERE a.account_id = ?
		)
		UPDATE app_account a
		SET failed_login_count = a.failed_login_count + 1,
			locked_until = CASE
				WHEN a.failed_login_count + 1 >= p.max_failed_attempts
				THEN clock_timestamp() + make_interval(secs => p.lockout_seconds)
				ELSE a.locked_until
			END,
			updated_at = clock_timestamp(),
			version_no = a.version_no + 1
		FROM selected_policy p
		WHERE a.account_id = p.account_id
	`, accountID).Error
}

func (r *AppAccountRepository) RecordSuccessfulLogin(ctx context.Context, accountID string) error {
	result := r.db.WithContext(ctx).Model(&model.AppAccount{}).
		Where("account_id = ?", accountID).
		Updates(map[string]interface{}{
			"failed_login_count": 0,
			"locked_until":       nil,
			"last_login_at":      gorm.Expr("clock_timestamp()"),
			"updated_at":         gorm.Expr("clock_timestamp()"),
			"version_no":         gorm.Expr("version_no + 1"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
