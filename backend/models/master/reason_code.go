package master

type ReasonCode struct {
	ID           string  `gorm:"column:reason_code_id;type:uuid;default:gen_random_uuid();primaryKey"`
	ModuleCode   string  `gorm:"column:module_code;size:50;not null"`
	Code         string  `gorm:"column:code;size:40;not null"`
	Name         string  `gorm:"column:name;size:100;not null"`
	Description  *string `gorm:"column:description;type:text"`
	RequiresNote bool    `gorm:"column:requires_note;not null;default:false"`
	IsActive     bool    `gorm:"column:is_active;not null;default:true"`
}

func (ReasonCode) TableName() string { return "reason_code" }
