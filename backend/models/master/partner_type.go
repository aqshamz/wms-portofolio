package master

type PartnerType struct {
	ID          string  `gorm:"column:partner_type_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code        string  `gorm:"column:code;size:40;not null"`
	Name        string  `gorm:"column:name;size:100;not null"`
	Description *string `gorm:"column:description;type:text"`
	IsActive    bool    `gorm:"column:is_active;not null;default:true"`
}

func (PartnerType) TableName() string { return "partner_type" }
