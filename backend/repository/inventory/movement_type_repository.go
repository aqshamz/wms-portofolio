package inventory

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/inventory"
)

type MovementTypeRepository struct{ db *gorm.DB }

func NewMovementTypeRepository(db *gorm.DB) *MovementTypeRepository {
	return &MovementTypeRepository{db: db}
}
func (r *MovementTypeRepository) ByCodeShared(ctx context.Context, code string) (model.MovementType, error) {
	var v model.MovementType
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Where("code=?", code).Take(&v).Error
	return v, Error(err)
}
func (r *MovementTypeRepository) Get(ctx context.Context, id string) (model.MovementType, error) {
	var v model.MovementType
	err := r.db.WithContext(ctx).Where("movement_type_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *MovementTypeRepository) List(ctx context.Context, active *bool) ([]model.MovementType, error) {
	q := r.db.WithContext(ctx)
	if active != nil {
		q = q.Where("is_active=?", *active)
	}
	rows := make([]model.MovementType, 0)
	err := q.Order("code").Find(&rows).Error
	return rows, Error(err)
}
func (r *MovementTypeRepository) Seed(ctx context.Context, values []model.MovementType) error {
	return Error(r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoNothing: true}).Create(&values).Error)
}
