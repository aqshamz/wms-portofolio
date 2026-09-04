package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type UOMRepository struct{ catalogTable[model.UOM] }

func NewUOMRepository(db *gorm.DB) *UOMRepository {
	return &UOMRepository{catalogTable[model.UOM]{db: db, key: "uom_id", searchable: true, audited: false}}
}
