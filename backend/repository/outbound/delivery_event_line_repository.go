package outbound

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type DeliveryEventLineRepository struct{ db *gorm.DB }

func NewDeliveryEventLineRepository(db *gorm.DB) *DeliveryEventLineRepository {
	return &DeliveryEventLineRepository{db: db}
}
func (r *DeliveryEventLineRepository) Create(ctx context.Context, v *model.DeliveryEventLine) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
