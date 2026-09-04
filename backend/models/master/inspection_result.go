package master

type InspectionResult struct {
	ID          string  `gorm:"column:inspection_result_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code        string  `gorm:"column:code;size:40;not null"`
	Name        string  `gorm:"column:name;size:100;not null"`
	Description *string `gorm:"column:description;type:text"`
	IsAccepted  bool    `gorm:"column:is_accepted;not null;default:false"`
	IsActive    bool    `gorm:"column:is_active;not null;default:true"`
}

func (InspectionResult) TableName() string { return "inspection_result" }
