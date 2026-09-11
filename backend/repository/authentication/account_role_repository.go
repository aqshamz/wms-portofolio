package authentication

import (
	"context"
	"time"

	model "wms-api/models/authentication"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AccountRoleDetails struct {
	AccountID  string
	RoleID     string
	Code       string
	Name       string
	IsActive   bool
	AssignedAt time.Time
	AssignedBy *string
}

type AccountRoleRepository struct{ db *gorm.DB }

func NewAccountRoleRepository(db *gorm.DB) *AccountRoleRepository {
	return &AccountRoleRepository{db: db}
}

func (r *AccountRoleRepository) Assign(ctx context.Context, value *model.AccountRole) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "account_id"}, {Name: "role_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"assigned_at": gorm.Expr("clock_timestamp()"), "assigned_by": value.AssignedBy,
		}),
	}).Create(value).Error
}

func (r *AccountRoleRepository) List(ctx context.Context, accountID string) ([]AccountRoleDetails, error) {
	rows := make([]AccountRoleDetails, 0)
	err := r.db.WithContext(ctx).Table("account_role assignment").
		Select(`assignment.account_id, assignment.role_id, role.code, role.name,
			role.is_active, assignment.assigned_at, assignment.assigned_by`).
		Joins("JOIN app_role role ON role.role_id = assignment.role_id").
		Where("assignment.account_id = ?", accountID).Order("role.code").Scan(&rows).Error
	return rows, err
}

func (r *AccountRoleRepository) Revoke(ctx context.Context, accountID, roleID string) error {
	result := r.db.WithContext(ctx).Where("account_id = ? AND role_id = ?", accountID, roleID).
		Delete(&model.AccountRole{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
