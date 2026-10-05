package authentication

import (
	"context"

	model "wms-api/models/authentication"

	"gorm.io/gorm"
)

type RolePermissionRepository struct{ db *gorm.DB }

func NewRolePermissionRepository(db *gorm.DB) *RolePermissionRepository {
	return &RolePermissionRepository{db: db}
}

func (r *RolePermissionRepository) List(ctx context.Context, roleID string) ([]PermissionDetails, error) {
	rows := make([]PermissionDetails, 0)
	err := r.db.WithContext(ctx).Table("role_permission role_grant").
		Select("permission.permission_id, permission.code, permission.name, permission.module_code").
		Joins("JOIN app_permission permission ON permission.permission_id = role_grant.permission_id").
		Where("role_grant.role_id = ?", roleID).Order("permission.code").Scan(&rows).Error
	return rows, err
}

func (r *RolePermissionRepository) Replace(ctx context.Context, roleID string, permissionIDs []string, actorID string) error {
	if err := r.db.WithContext(ctx).Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
		return err
	}
	if len(permissionIDs) == 0 {
		return nil
	}
	rows := make([]model.RolePermission, 0, len(permissionIDs))
	for _, permissionID := range permissionIDs {
		actor := actorID
		rows = append(rows, model.RolePermission{RoleID: roleID, PermissionID: permissionID, GrantedBy: &actor})
	}
	return r.db.WithContext(ctx).Create(&rows).Error
}

// EnsureByCode adds a permission to an existing built-in role without replacing
// any administrator-managed grants. Missing roles are intentionally ignored so
// environments that do not install the optional role remain valid.
func (r *RolePermissionRepository) EnsureByCode(ctx context.Context, roleCode, permissionCode string) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO role_permission(role_id, permission_id)
		SELECT role.role_id, permission.permission_id
		FROM app_role role
		JOIN app_permission permission ON permission.code = ?
		WHERE role.code = ?
		ON CONFLICT(role_id, permission_id) DO NOTHING
	`, permissionCode, roleCode).Error
}
