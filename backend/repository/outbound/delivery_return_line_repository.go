package outbound

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type DeliveryReturnLineRepository struct{ db *gorm.DB }

func NewDeliveryReturnLineRepository(db *gorm.DB) *DeliveryReturnLineRepository {
	return &DeliveryReturnLineRepository{db: db}
}
func (r *DeliveryReturnLineRepository) Create(ctx context.Context, v *model.DeliveryReturnLine) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *DeliveryReturnLineRepository) GetByMovement(ctx context.Context, id string) (model.DeliveryReturnLine, error) {
	var v model.DeliveryReturnLine
	err := r.db.WithContext(ctx).Where("movement_id=?", id).Take(&v).Error
	return v, Error(err)
}
