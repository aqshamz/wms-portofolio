package master

type ValidationSeverity struct {
	ID               string `gorm:"column:validation_severity_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code             string `gorm:"column:code;size:30;not null;uniqueIndex"`
	Name             string `gorm:"column:name;size:100;not null"`
	BlocksProcessing bool   `gorm:"column:blocks_processing;not null;default:true"`
	IsActive         bool   `gorm:"column:is_active;not null;default:true"`
}

func (ValidationSeverity) TableName() string { return "validation_severity" }
