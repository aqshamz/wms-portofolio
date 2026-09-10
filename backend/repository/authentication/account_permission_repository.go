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
	err := r.db.WithContext(ctx).Table("account_permission apg").Select("permission.code").Joins("JOIN app_permission permission ON permission.permission_id=apg.permission_id AND permission.is_active").Where("apg.account_id=?", accountID).Order("permission.code").Scan(&codes).Error
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
			var administrators int64
			if err := tx.Table("account_permission apg").
				Joins("JOIN app_permission permission ON permission.permission_id = apg.permission_id").
				Joins("JOIN app_account account ON account.account_id = apg.account_id").
				Joins("JOIN account_status status ON status.account_status_id = account.account_status_id").
				Where("permission.code = ? AND apg.account_id <> ? AND status.is_active AND status.allows_login", code, accountID).
				Count(&administrators).Error; err != nil {
				return err
			}
			if administrators == 0 {
				return ErrLastSecurityAdministrator
			}
		}
		result := tx.Where("account_id = ? AND permission_id = ?", accountID, permissionID).
			Delete(&model.AccountPermission{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}
