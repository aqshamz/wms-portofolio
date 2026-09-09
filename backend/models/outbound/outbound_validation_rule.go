package outbound

type OutboundValidationRule struct {
	ID                   string  `gorm:"column:outbound_validation_rule_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code                 string  `gorm:"column:code;size:50;not null;uniqueIndex"`
	Name                 string  `gorm:"column:name;size:120;not null"`
	Description          *string `gorm:"column:description;type:text"`
	HandlerCode          string  `gorm:"column:handler_code;size:60;not null"`
	ValidationSeverityID string  `gorm:"column:validation_severity_id;type:uuid;not null"`
	DisplayOrder         int     `gorm:"column:display_order;not null;default:0"`
	IsActive             bool    `gorm:"column:is_active;not null;default:true"`
}

func (OutboundValidationRule) TableName() string { return "outbound_validation_rule" }
