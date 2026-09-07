package inventory

import (
	"gorm.io/gorm"
	model "wms-api/models/inventory"
)

type HandlingUnitRepository struct {
	identityTable[model.HandlingUnit]
}

func NewHandlingUnitRepository(db *gorm.DB) *HandlingUnitRepository {
	return &HandlingUnitRepository{identityTable[model.HandlingUnit]{db: db, key: "handling_unit_id", label: "barcode", handlingUnit: true}}
}
