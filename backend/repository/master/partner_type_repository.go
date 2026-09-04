package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type PartnerTypeRepository struct {
	catalogTable[model.PartnerType]
}

func NewPartnerTypeRepository(db *gorm.DB) *PartnerTypeRepository {
	return &PartnerTypeRepository{catalogTable[model.PartnerType]{db: db, key: "partner_type_id", searchable: true, audited: false}}
}
