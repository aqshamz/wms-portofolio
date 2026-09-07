package inventory

import (
	"gorm.io/gorm"
	model "wms-api/models/inventory"
)

type InventoryLotRepository struct {
	identityTable[model.InventoryLot]
}

func NewInventoryLotRepository(db *gorm.DB) *InventoryLotRepository {
	return &InventoryLotRepository{identityTable[model.InventoryLot]{db: db, key: "lot_id", label: "lot_number", handlingUnit: false}}
}
