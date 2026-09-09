package outbound

type Carrier struct {
	ID                string  `gorm:"column:carrier_id;type:uuid;primaryKey;default:gen_random_uuid()"`
	BusinessPartnerID *string `gorm:"column:business_partner_id;type:uuid;uniqueIndex"`
	Code              string  `gorm:"column:code;size:40;not null;uniqueIndex"`
	Name              string  `gorm:"column:name;size:150;not null"`
	IsActive          bool    `gorm:"column:is_active;not null;default:true"`
}

func (Carrier) TableName() string { return "carrier" }
