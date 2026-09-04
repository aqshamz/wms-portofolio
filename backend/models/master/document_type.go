package master

type DocumentType struct {
	ID          string  `gorm:"column:document_type_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code        string  `gorm:"column:code;size:40;not null"`
	Name        string  `gorm:"column:name;size:100;not null"`
	ModuleCode  string  `gorm:"column:module_code;size:50;not null"`
	Description *string `gorm:"column:description;type:text"`
	IsActive    bool    `gorm:"column:is_active;not null;default:true"`
}

func (DocumentType) TableName() string { return "document_type" }
