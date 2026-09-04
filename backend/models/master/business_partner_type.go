package master

type BusinessPartnerType struct {
	PartnerID     string `gorm:"column:partner_id;type:uuid;primaryKey"`
	PartnerTypeID string `gorm:"column:partner_type_id;type:uuid;primaryKey"`
}

func (BusinessPartnerType) TableName() string { return "business_partner_type" }
