package master

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/master"
)

func (r *ItemUOMRepository) ForUnit(ctx context.Context, itemID, uomID string) (model.ItemUOM, error) {
	var value model.ItemUOM
	err := r.db.WithContext(ctx).Where("item_id = ? AND uom_id = ?", itemID, uomID).Take(&value).Error
	return value, err
}

type ItemUOMRepository struct{ catalogTable[model.ItemUOM] }

// Preserve explicit false flags: GORM otherwise replaces zero bools with a
// default:true tag during Create. Both statements remain atomic.
func (r *ItemUOMRepository) Create(ctx context.Context, value *model.ItemUOM) error {
	receiving, picking := value.IsReceivingUOM, value.IsPickingUOM
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(value).Error; err != nil {
			return err
		}
		value.IsReceivingUOM, value.IsPickingUOM = receiving, picking
		return tx.Model(value).Updates(map[string]interface{}{"is_receiving_uom": receiving, "is_picking_uom": picking}).Error
	})
}

func NewItemUOMRepository(db *gorm.DB) *ItemUOMRepository {
	return &ItemUOMRepository{catalogTable[model.ItemUOM]{db: db, key: "item_uom_id", searchable: false, audited: false}}
}
