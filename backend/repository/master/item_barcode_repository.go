package master

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/master"
)

func (r *ItemBarcodeRepository) ClearPrimary(ctx context.Context, itemID string) error {
	return r.db.WithContext(ctx).Model(&model.ItemBarcode{}).
		Where("item_id = ? AND is_primary", itemID).Update("is_primary", false).Error
}

func (r *ItemBarcodeRepository) HasActiveUnit(ctx context.Context, itemID, uomID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ItemBarcode{}).
		Where("item_id = ? AND uom_id = ? AND is_active", itemID, uomID).Count(&count).Error
	return count > 0, err
}

type ItemBarcodeRepository struct {
	catalogTable[model.ItemBarcode]
}

func NewItemBarcodeRepository(db *gorm.DB) *ItemBarcodeRepository {
	return &ItemBarcodeRepository{catalogTable[model.ItemBarcode]{db: db, key: "item_barcode_id", searchable: false, audited: false}}
}
