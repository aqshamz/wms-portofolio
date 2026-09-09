package outbound

type OutboundCheckResult struct {
	ID           string  `gorm:"column:outbound_check_result_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code         string  `gorm:"column:code;size:40;not null;uniqueIndex"`
	Name         string  `gorm:"column:name;size:100;not null"`
	Description  *string `gorm:"column:description;type:text"`
	IsPass       bool    `gorm:"column:is_pass;not null;default:false"`
	RequiresNote bool    `gorm:"column:requires_note;not null;default:false"`
	IsActive     bool    `gorm:"column:is_active;not null;default:true"`
}

func (OutboundCheckResult) TableName() string { return "outbound_check_result" }
