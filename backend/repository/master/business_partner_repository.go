package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type BusinessPartnerRepository struct {
	catalogTable[model.BusinessPartner]
}

func NewBusinessPartnerRepository(db *gorm.DB) *BusinessPartnerRepository {
	return &BusinessPartnerRepository{catalogTable[model.BusinessPartner]{db: db, key: "partner_id", searchable: true, audited: true}}
}
