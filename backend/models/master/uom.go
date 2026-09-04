package master

type UOM struct {
	ID           string `gorm:"column:uom_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code         string `gorm:"column:code;size:20;not null"`
	Name         string `gorm:"column:name;size:100;not null"`
	DecimalScale int16  `gorm:"column:decimal_scale;type:smallint;not null;default:0"`
	IsActive     bool   `gorm:"column:is_active;not null;default:true"`
}

func (UOM) TableName() string { return "uom" }
