package outbound

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type ShipmentPackingRepository struct{ db *gorm.DB }

func NewShipmentPackingRepository(db *gorm.DB) *ShipmentPackingRepository {
	return &ShipmentPackingRepository{db: db}
}
func (r *ShipmentPackingRepository) CreateBatch(ctx context.Context, v []model.ShipmentPacking) error {
	return Error(r.db.WithContext(ctx).Create(&v).Error)
}
