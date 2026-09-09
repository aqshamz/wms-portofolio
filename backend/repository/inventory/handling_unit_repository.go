package inventory

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/inventory"
)

type HandlingUnitRepository struct {
	identityTable[model.HandlingUnit]
}

func (r *HandlingUnitRepository) Lock(ctx context.Context, id string) (model.HandlingUnit, error) {
	var value model.HandlingUnit
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("handling_unit_id=?", id).Take(&value).Error
	return value, Error(err)
}

func (r *HandlingUnitRepository) CountChildren(ctx context.Context, id string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.HandlingUnit{}).Where("parent_handling_unit_id=?", id).Count(&count).Error
	return count, Error(err)
}

func (r *HandlingUnitRepository) Relocate(ctx context.Context, id, sourceLocationID, targetLocationID string) error {
	result := r.db.WithContext(ctx).Model(&model.HandlingUnit{}).Where("handling_unit_id=? AND current_location_id=?", id, sourceLocationID).Update("current_location_id", targetLocationID)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConflict
	}
	return nil
}

func (r *HandlingUnitRepository) Place(ctx context.Context, id, targetLocationID string) error {
	result := r.db.WithContext(ctx).Model(&model.HandlingUnit{}).Where("handling_unit_id=? AND current_location_id IS NULL", id).Update("current_location_id", targetLocationID)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConflict
	}
	return nil
}

func (r *HandlingUnitRepository) ClearLocation(ctx context.Context, id, sourceLocationID string) error {
	result := r.db.WithContext(ctx).Model(&model.HandlingUnit{}).Where("handling_unit_id=? AND current_location_id=?", id, sourceLocationID).Update("current_location_id", nil)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConflict
	}
	return nil
}

func NewHandlingUnitRepository(db *gorm.DB) *HandlingUnitRepository {
	return &HandlingUnitRepository{identityTable[model.HandlingUnit]{db: db, key: "handling_unit_id", label: "barcode", handlingUnit: true}}
}
