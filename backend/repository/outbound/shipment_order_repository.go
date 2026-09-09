package outbound

import (
	"context"
	"gorm.io/gorm"
	"time"
	model "wms-api/models/outbound"
)

type ShipmentOrderRepository struct{ db *gorm.DB }

func NewShipmentOrderRepository(db *gorm.DB) *ShipmentOrderRepository {
	return &ShipmentOrderRepository{db: db}
}
func (r *ShipmentOrderRepository) RemoveAll(ctx context.Context, id, actor string) error {
	return Error(r.db.WithContext(ctx).Model(&model.ShipmentOrder{}).
		Where("shipment_id=? AND removed_at IS NULL", id).
		Updates(map[string]interface{}{"removed_at": time.Now(), "removed_by": actor}).Error)
}
func (r *ShipmentOrderRepository) CreateBatch(ctx context.Context, v []model.ShipmentOrder) error {
	return Error(r.db.WithContext(ctx).Create(&v).Error)
}
func (r *ShipmentOrderRepository) List(ctx context.Context, id string) ([]model.ShipmentOrder, error) {
	var v []model.ShipmentOrder
	err := r.db.WithContext(ctx).Where("shipment_id=? AND removed_at IS NULL", id).Find(&v).Error
	return v, Error(err)
}
