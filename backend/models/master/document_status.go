package master

type DocumentStatus struct {
	ID             string  `gorm:"column:status_id;type:uuid;default:gen_random_uuid();primaryKey"`
	DocumentTypeID string  `gorm:"column:document_type_id;type:uuid;not null"`
	Code           string  `gorm:"column:code;size:40;not null"`
	Name           string  `gorm:"column:name;size:100;not null"`
	Description    *string `gorm:"column:description;type:text"`
	IsInitial      bool    `gorm:"column:is_initial;not null;default:false"`
	IsFinal        bool    `gorm:"column:is_final;not null;default:false"`
	IsCancelled    bool    `gorm:"column:is_cancelled;not null;default:false"`
	DisplayOrder   int     `gorm:"column:display_order;not null;default:0"`
	IsActive       bool    `gorm:"column:is_active;not null;default:true"`
}

func (DocumentStatus) TableName() string { return "document_status" }
