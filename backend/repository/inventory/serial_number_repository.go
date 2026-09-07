package inventory

import (
	"gorm.io/gorm"
	model "wms-api/models/inventory"
)

type SerialNumberRepository struct {
	identityTable[model.SerialNumber]
}

func NewSerialNumberRepository(db *gorm.DB) *SerialNumberRepository {
	return &SerialNumberRepository{identityTable[model.SerialNumber]{db: db, key: "serial_id", label: "serial_no", handlingUnit: false}}
}
