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
func (r *ShipmentPackingRepository) HasActive(ctx context.Context, packingID string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("shipment_packing sp").Joins("JOIN shipment s ON s.shipment_id=sp.shipment_id").Joins("JOIN document_status ds ON ds.status_id=s.status_id").Where("sp.packing_id=? AND NOT ds.is_cancelled", packingID).Count(&n).Error
	return n > 0, Error(err)
}
