package master

import (
	"context"

	model "wms-api/models/master"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LocationTypeRepository struct{ db *gorm.DB }

func NewLocationTypeRepository(db *gorm.DB) *LocationTypeRepository {
	return &LocationTypeRepository{db: db}
}

func (r *LocationTypeRepository) Seed(ctx context.Context, values []model.LocationType) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}},
		DoNothing: true,
	}).Create(&values).Error
}

func (r *LocationTypeRepository) Create(ctx context.Context, value *model.LocationType) error {
	return r.db.WithContext(ctx).Create(value).Error
}

func (r *LocationTypeRepository) FindByID(ctx context.Context, id string) (model.LocationType, error) {
	var value model.LocationType
	err := r.db.WithContext(ctx).Where("location_type_id = ?", id).Take(&value).Error
	return value, err
}

func (r *LocationTypeRepository) List(ctx context.Context, active *bool) ([]model.LocationType, error) {
	var values []model.LocationType
	query := r.db.WithContext(ctx).Order("name")
	if active != nil {
		query = query.Where("is_active = ?", *active)
	}
	err := query.Find(&values).Error
	return values, err
}

func (r *LocationTypeRepository) Update(
	ctx context.Context, id string, changes map[string]interface{},
) (model.LocationType, error) {
	var value model.LocationType
	result := r.db.WithContext(ctx).Model(&value).Clauses(clause.Returning{}).
		Where("location_type_id = ?", id).Updates(changes)
	if result.Error != nil {
		return value, result.Error
	}
	if result.RowsAffected == 0 {
		return value, gorm.ErrRecordNotFound
	}
	return value, nil
}
