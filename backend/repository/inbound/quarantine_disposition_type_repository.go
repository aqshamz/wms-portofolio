package inbound

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/inbound"
)

type QuarantineDispositionTypeRepository struct{ db *gorm.DB }

func NewQuarantineDispositionTypeRepository(db *gorm.DB) *QuarantineDispositionTypeRepository {
	return &QuarantineDispositionTypeRepository{db: db}
}
func (r *QuarantineDispositionTypeRepository) ByCode(ctx context.Context, code string) (model.QuarantineDispositionType, error) {
	var value model.QuarantineDispositionType
	err := r.db.WithContext(ctx).Where("code=?", code).Take(&value).Error
	return value, Error(err)
}
func (r *QuarantineDispositionTypeRepository) List(ctx context.Context, active *bool) ([]model.QuarantineDispositionType, error) {
	query := r.db.WithContext(ctx).Order("code")
	if active != nil {
		query = query.Where("is_active=?", *active)
	}
	rows := make([]model.QuarantineDispositionType, 0)
	return rows, Error(query.Find(&rows).Error)
}
func (r *QuarantineDispositionTypeRepository) Seed(ctx context.Context, values []model.QuarantineDispositionType) error {
	return Error(r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoNothing: true}).Create(&values).Error)
}
func (r *QuarantineDispositionTypeRepository) LinkRemovalMovement(ctx context.Context, code, movementCode string) error {
	return Error(r.db.WithContext(ctx).Exec(`UPDATE quarantine_disposition_type SET removal_movement_type_id=(SELECT movement_type_id FROM movement_type WHERE code=?) WHERE code=? AND removal_movement_type_id IS NULL`, movementCode, code).Error)
}
