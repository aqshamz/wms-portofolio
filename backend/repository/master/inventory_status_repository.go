package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type InventoryStatusRepository struct {
	catalogTable[model.InventoryStatus]
}

func NewInventoryStatusRepository(db *gorm.DB) *InventoryStatusRepository {
	return &InventoryStatusRepository{catalogTable[model.InventoryStatus]{db: db, key: "inventory_status_id", searchable: true, audited: false}}
}
