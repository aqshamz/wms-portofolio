package outbound

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/outbound"
)

type DeliveryFailureReasonRepository struct{ db *gorm.DB }

func NewDeliveryFailureReasonRepository(db *gorm.DB) *DeliveryFailureReasonRepository {
	return &DeliveryFailureReasonRepository{db: db}
}
func (r *DeliveryFailureReasonRepository) ByCode(ctx context.Context, code string) (model.DeliveryFailureReason, error) {
	var v model.DeliveryFailureReason
	err := r.db.WithContext(ctx).Where("code=? AND is_active", code).Take(&v).Error
	return v, Error(err)
}
func (r *DeliveryFailureReasonRepository) Seed(ctx context.Context, v *model.DeliveryFailureReason) error {
	return Error(r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "description", "is_active"})}).Create(v).Error)
}
