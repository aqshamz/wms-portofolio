package authentication

import (
	"context"
	"fmt"

	model "wms-api/models/authentication"
	mastermodel "wms-api/models/master"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// EnsureSuperadmin grants the built-in unrestricted role without deleting any
// existing roles, direct permissions, or owner/warehouse access grants.
func (r *AccountPermissionRepository) EnsureSuperadmin(ctx context.Context, accountID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		permission := mastermodel.AppPermission{Code: "*", Name: "Unrestricted system access", ModuleCode: "SECURITY", IsActive: true,
			Description: superadminDescription()}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&permission).Error; err != nil {
			return err
		}
		permission = mastermodel.AppPermission{}
		if err := tx.Where("code=?", "*").Take(&permission).Error; err != nil {
			return err
		}
		if !permission.IsActive {
			return fmt.Errorf("unrestricted permission is disabled; refusing to reactivate it")
		}
		role := model.AppRole{Code: "SUPERADMIN", Name: "Superadmin", Description: superadminDescription(), IsActive: true, CreatedBy: &accountID, UpdatedBy: &accountID}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&role).Error; err != nil {
			return err
		}
		role = model.AppRole{}
		if err := tx.Where("lower(code)=lower(?)", "SUPERADMIN").Take(&role).Error; err != nil {
			return err
		}
		if !role.IsActive {
			return fmt.Errorf("SUPERADMIN role is disabled; refusing to reactivate it")
		}
		if err := tx.Exec(`INSERT INTO role_permission(role_id,permission_id,granted_by)
			SELECT ?,permission_id,? FROM app_permission WHERE is_active
			ON CONFLICT(role_id,permission_id) DO NOTHING`, role.ID, accountID).Error; err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.AccountRole{AccountID: accountID, RoleID: role.ID, AssignedBy: &accountID}).Error
	})
}

func superadminDescription() *string {
	description := "Full access to every menu, action, organization, and warehouse through the * permission. No individual access-scope grants are required. Assign only to trusted system administrators. Workflow and stock validation still apply."
	return &description
}
