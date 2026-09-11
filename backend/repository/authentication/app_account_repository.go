package authentication

import (
	"context"
	"time"
	model "wms-api/models/authentication"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

type AccountFilter struct {
	Search         string
	StatusID       string
	Page, PageSize int
}

type AccountDetails struct {
	ID                       string
	Username                 string
	Email                    *string
	DisplayName              string
	AccountStatusID          string
	StatusCode               string
	StatusName               string
	StatusAllowsLogin        bool
	AuthenticationPolicyID   *string
	AuthenticationPolicyCode *string
	AuthenticationPolicyName *string
	PreferredTimezone        *string
	FailedLoginCount         int
	LockedUntil              *time.Time
	LastLoginAt              *time.Time
	CreatedAt                time.Time
	CreatedBy                *string
	UpdatedAt                time.Time
	UpdatedBy                *string
	VersionNo                int
}

func NewAppAccountRepository(db *gorm.DB) *AppAccountRepository {
	return &AppAccountRepository{db: db}
}

func (r *AppAccountRepository) FindForLogin(ctx context.Context, identifier string) (LoginAccount, error) {
	return r.findForLogin(ctx, identifier, false)
}

func (r *AppAccountRepository) FindForLoginForUpdate(ctx context.Context, identifier string) (LoginAccount, error) {
	return r.findForLogin(ctx, identifier, true)
}

func (r *AppAccountRepository) findForLogin(ctx context.Context, identifier string, lock bool) (LoginAccount, error) {
	var account LoginAccount
	query := r.db.WithContext(ctx).
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
		Where("lower(a.username) = lower(?) OR lower(a.email) = lower(?)", identifier, identifier)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "a"}})
	}
	err := query.Take(&account).Error
	return account, err
}

func (r *AppAccountRepository) FindByUsername(ctx context.Context, username string) (model.AppAccount, error) {
	var account model.AppAccount
	err := r.db.WithContext(ctx).Where("lower(username) = lower(?)", username).Take(&account).Error
	return account, err
}

func (r *AppAccountRepository) FindByID(ctx context.Context, id string) (model.AppAccount, error) {
	var account model.AppAccount
	err := r.db.WithContext(ctx).Where("account_id = ?", id).Take(&account).Error
	return account, err
}

func (r *AppAccountRepository) Create(ctx context.Context, account *model.AppAccount) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *AppAccountRepository) Details(ctx context.Context, accountID string) (AccountDetails, error) {
	var value AccountDetails
	err := r.detailsQuery(ctx).Where("account.account_id = ?", accountID).Take(&value).Error
	return value, err
}

func (r *AppAccountRepository) List(ctx context.Context, filter AccountFilter) ([]AccountDetails, int64, error) {
	base := r.db.WithContext(ctx).Table("app_account account")
	if filter.Search != "" {
		term := "%" + filter.Search + "%"
		base = base.Where("account.username ILIKE ? OR account.display_name ILIKE ? OR account.email ILIKE ?", term, term, term)
	}
	if filter.StatusID != "" {
		base = base.Where("account.account_status_id = ?", filter.StatusID)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]AccountDetails, 0)
	query := r.detailsQuery(ctx)
	if filter.Search != "" {
		term := "%" + filter.Search + "%"
		query = query.Where("account.username ILIKE ? OR account.display_name ILIKE ? OR account.email ILIKE ?", term, term, term)
	}
	if filter.StatusID != "" {
		query = query.Where("account.account_status_id = ?", filter.StatusID)
	}
	err := query.Order("account.username").Limit(filter.PageSize).
		Offset((filter.Page - 1) * filter.PageSize).Scan(&rows).Error
	return rows, total, err
}

func (r *AppAccountRepository) IdentityExists(ctx context.Context, username string, email *string, excludeID string) (bool, error) {
	identity := r.db.Where("lower(username) = lower(?)", username)
	if email != nil {
		identity = identity.Or("lower(email) = lower(?)", *email)
	}
	query := r.db.WithContext(ctx).Model(&model.AppAccount{}).Where(identity)
	if excludeID != "" {
		query = query.Where("account_id <> ?", excludeID)
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *AppAccountRepository) Update(ctx context.Context, accountID string, expectedVersion int, changes map[string]interface{}) (model.AppAccount, error) {
	changes["updated_at"] = gorm.Expr("clock_timestamp()")
	changes["version_no"] = gorm.Expr("version_no + 1")
	var value model.AppAccount
	result := r.db.WithContext(ctx).Model(&value).Clauses(clause.Returning{}).
		Where("account_id = ? AND version_no = ?", accountID, expectedVersion).Updates(changes)
	if result.Error != nil {
		return value, result.Error
	}
	if result.RowsAffected == 0 {
		return value, gorm.ErrRecordNotFound
	}
	return value, nil
}

func (r *AppAccountRepository) SetPassword(ctx context.Context, accountID, passwordHash, actorID string, expectedVersion int) error {
	result := r.db.WithContext(ctx).Model(&model.AppAccount{}).
		Where("account_id = ? AND version_no = ?", accountID, expectedVersion).
		Updates(map[string]interface{}{
			"password_hash": passwordHash, "failed_login_count": 0, "locked_until": nil,
			"updated_at": gorm.Expr("clock_timestamp()"), "updated_by": actorID,
			"version_no": gorm.Expr("version_no + 1"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *AppAccountRepository) Unlock(ctx context.Context, accountID, actorID string, expectedVersion int) error {
	result := r.db.WithContext(ctx).Model(&model.AppAccount{}).
		Where("account_id = ? AND version_no = ?", accountID, expectedVersion).
		Updates(map[string]interface{}{
			"failed_login_count": 0, "locked_until": nil,
			"updated_at": gorm.Expr("clock_timestamp()"), "updated_by": actorID,
			"version_no": gorm.Expr("version_no + 1"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *AppAccountRepository) detailsQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table("app_account account").
		Select(`account.account_id AS id, account.username, account.email, account.display_name,
			account.account_status_id, status.code AS status_code, status.name AS status_name,
			status.allows_login AS status_allows_login, account.authentication_policy_id,
			policy.code AS authentication_policy_code, policy.name AS authentication_policy_name,
			account.preferred_timezone, account.failed_login_count, account.locked_until,
			account.last_login_at, account.created_at, account.created_by, account.updated_at,
			account.updated_by, account.version_no`).
		Joins("JOIN account_status status ON status.account_status_id = account.account_status_id").
		Joins("LEFT JOIN authentication_policy policy ON policy.authentication_policy_id = account.authentication_policy_id")
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
