package authentication

import (
	"context"

	model "wms-api/models/authentication"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RoleFilter struct {
	Search         string
	Active         *bool
	Page, PageSize int
}

type AppRoleRepository struct{ db *gorm.DB }

func NewAppRoleRepository(db *gorm.DB) *AppRoleRepository { return &AppRoleRepository{db: db} }

func (r *AppRoleRepository) Create(ctx context.Context, value *model.AppRole) error {
	return r.db.WithContext(ctx).Create(value).Error
}

func (r *AppRoleRepository) Get(ctx context.Context, id string) (model.AppRole, error) {
	var value model.AppRole
	err := r.db.WithContext(ctx).Where("role_id = ?", id).Take(&value).Error
	return value, err
}

func (r *AppRoleRepository) List(ctx context.Context, filter RoleFilter) ([]model.AppRole, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.AppRole{})
	if filter.Search != "" {
		query = query.Where("code ILIKE ? OR name ILIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}
	if filter.Active != nil {
		query = query.Where("is_active = ?", *filter.Active)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]model.AppRole, 0)
	err := query.Order("code").Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize).Find(&rows).Error
	return rows, total, err
}

func (r *AppRoleRepository) Update(ctx context.Context, id string, expectedVersion int, changes map[string]interface{}) (model.AppRole, error) {
	changes["updated_at"] = gorm.Expr("clock_timestamp()")
	changes["version_no"] = gorm.Expr("version_no + 1")
	var value model.AppRole
	result := r.db.WithContext(ctx).Model(&value).Clauses(clause.Returning{}).
		Where("role_id = ? AND version_no = ?", id, expectedVersion).Updates(changes)
	if result.Error != nil {
		return value, result.Error
	}
	if result.RowsAffected == 0 {
		return value, gorm.ErrRecordNotFound
	}
	return value, nil
}

func (r *AppRoleRepository) ActiveExists(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.AppRole{}).Where("role_id = ? AND is_active", id).Count(&count).Error
	return count == 1, err
}

func (r *AppRoleRepository) CodeExists(ctx context.Context, code, excludeID string) (bool, error) {
	query := r.db.WithContext(ctx).Model(&model.AppRole{}).Where("lower(code) = lower(?)", code)
	if excludeID != "" {
		query = query.Where("role_id <> ?", excludeID)
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}
