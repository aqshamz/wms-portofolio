package master

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/master"
)

type ValidationSeverityRepository struct{ db *gorm.DB }

func NewValidationSeverityRepository(db *gorm.DB) *ValidationSeverityRepository {
	return &ValidationSeverityRepository{db: db}
}
func (r *ValidationSeverityRepository) ByCode(ctx context.Context, code string) (model.ValidationSeverity, error) {
	var value model.ValidationSeverity
	err := r.db.WithContext(ctx).Where("code=? AND is_active", code).Take(&value).Error
	return value, err
}
func (r *ValidationSeverityRepository) Seed(ctx context.Context, values []model.ValidationSeverity) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "blocks_processing", "is_active"})}).Create(&values).Error
}
