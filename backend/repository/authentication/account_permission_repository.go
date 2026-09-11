package authentication

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/authentication"
)

type AccountPermissionRepository struct{ db *gorm.DB }

var ErrLastSecurityAdministrator = errors.New("cannot revoke the last SECURITY.WRITE permission")

type AccountPermissionDetails struct {
	AccountID    string
	PermissionID string
	Code         string
	Name         string
	ModuleCode   string
	GrantedAt    time.Time
	GrantedBy    *string
}

type PermissionDetails struct {
	PermissionID string
	Code         string
	Name         string
	ModuleCode   string
}

func NewAccountPermissionRepository(db *gorm.DB) *AccountPermissionRepository {
	return &AccountPermissionRepository{db: db}
}

func (r *AccountPermissionRepository) Codes(ctx context.Context, accountID string) ([]string, error) {
	codes := make([]string, 0)
	err := r.db.WithContext(ctx).Raw(`
		SELECT code FROM (
			SELECT permission.code
			FROM account_permission apg
			JOIN app_permission permission ON permission.permission_id = apg.permission_id
			WHERE apg.account_id = ? AND permission.is_active
			UNION
			SELECT permission.code
			FROM account_role assignment
			JOIN app_role role ON role.role_id = assignment.role_id AND role.is_active
			JOIN role_permission grant_row ON grant_row.role_id = role.role_id
			JOIN app_permission permission ON permission.permission_id = grant_row.permission_id
			WHERE assignment.account_id = ? AND permission.is_active
		) effective ORDER BY code`, accountID, accountID).Scan(&codes).Error
	return codes, err
}

func (r *AccountPermissionRepository) List(ctx context.Context, accountID string) ([]AccountPermissionDetails, error) {
	rows := make([]AccountPermissionDetails, 0)
	err := r.db.WithContext(ctx).
		Table("account_permission apg").
		Select(`apg.account_id, apg.permission_id, permission.code,
			permission.name, permission.module_code, apg.granted_at, apg.granted_by`).
		Joins("JOIN app_permission permission ON permission.permission_id = apg.permission_id").
		Where("apg.account_id = ?", accountID).
		Order("permission.code").
		Scan(&rows).Error
	return rows, err
}

func (r *AccountPermissionRepository) ActivePermissionExists(ctx context.Context, permissionID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("app_permission").
		Where("permission_id = ? AND is_active", permissionID).Count(&count).Error
	return count == 1, err
}

func (r *AccountPermissionRepository) Available(ctx context.Context) ([]PermissionDetails, error) {
	rows := make([]PermissionDetails, 0)
	err := r.db.WithContext(ctx).Table("app_permission").
		Select("permission_id, code, name, module_code").
		Where("is_active").Order("code").Scan(&rows).Error
	return rows, err
}

func (r *AccountPermissionRepository) GrantAll(ctx context.Context, accountID string) error {
	return r.db.WithContext(ctx).Exec(`INSERT INTO account_permission(account_id,permission_id,granted_by)
		SELECT ?,permission_id,? FROM app_permission WHERE is_active
		ON CONFLICT(account_id,permission_id) DO NOTHING`, accountID, accountID).Error
}

func (r *AccountPermissionRepository) Grant(ctx context.Context, value *model.AccountPermission) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(value).Error
}

func (r *AccountPermissionRepository) Revoke(ctx context.Context, accountID, permissionID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", int64(873246211)).Error; err != nil {
			return err
		}
		var code string
		if err := tx.Table("app_permission").Select("code").
			Where("permission_id = ?", permissionID).Scan(&code).Error; err != nil {
			return err
		}
		if code == "SECURITY.WRITE" {
			// The invariant is checked after deletion so role-derived access is
			// considered as well as direct permission grants.
		}
		result := tx.Where("account_id = ? AND permission_id = ?", accountID, permissionID).
			Delete(&model.AccountPermission{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		if code == "SECURITY.WRITE" {
			return ensureSecurityAdministrator(tx)
		}
		return nil
	})
}

func (r *AccountPermissionRepository) EnsureSecurityAdministrator(ctx context.Context) error {
	db := r.db.WithContext(ctx)
	if err := db.Exec("SELECT pg_advisory_xact_lock(?)", int64(873246211)).Error; err != nil {
		return err
	}
	return ensureSecurityAdministrator(db)
}

func ensureSecurityAdministrator(db *gorm.DB) error {
	var count int64
	err := db.Raw(`
		SELECT count(DISTINCT effective.account_id)
		FROM (
			SELECT direct.account_id
			FROM account_permission direct
			JOIN app_permission permission ON permission.permission_id = direct.permission_id
			WHERE permission.code = 'SECURITY.WRITE' AND permission.is_active
			UNION
			SELECT assignment.account_id
			FROM account_role assignment
			JOIN app_role role ON role.role_id = assignment.role_id AND role.is_active
			JOIN role_permission role_grant ON role_grant.role_id = role.role_id
			JOIN app_permission permission ON permission.permission_id = role_grant.permission_id
			WHERE permission.code = 'SECURITY.WRITE' AND permission.is_active
		) effective
		JOIN app_account account ON account.account_id = effective.account_id
		JOIN account_status status ON status.account_status_id = account.account_status_id
		WHERE status.is_active AND status.allows_login`).Scan(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrLastSecurityAdministrator
	}
	return nil
}
