package inbound

import (
	"context"

	model "wms-api/models/inbound"

	"gorm.io/gorm"
)

type ReceiptLineSerialRepository struct{ db *gorm.DB }

func NewReceiptLineSerialRepository(db *gorm.DB) *ReceiptLineSerialRepository {
	return &ReceiptLineSerialRepository{db: db}
}
func (r *ReceiptLineSerialRepository) Create(ctx context.Context, value *model.ReceiptLineSerial) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}
func (r *ReceiptLineSerialRepository) GetByBatch(ctx context.Context, batchID string) (model.ReceiptLineSerial, error) {
	var value model.ReceiptLineSerial
	err := r.db.WithContext(ctx).Where("receipt_inventory_id=?", batchID).Take(&value).Error
	return value, Error(err)
}
