package master

type DocumentStatusTransition struct {
	ID                   string  `gorm:"column:transition_id;type:uuid;default:gen_random_uuid();primaryKey"`
	DocumentTypeID       string  `gorm:"column:document_type_id;type:uuid;not null"`
	FromStatusID         string  `gorm:"column:from_status_id;type:uuid;not null"`
	ToStatusID           string  `gorm:"column:to_status_id;type:uuid;not null"`
	RequiredPermissionID *string `gorm:"column:required_permission_id;type:uuid"`
	IsActive             bool    `gorm:"column:is_active;not null;default:true"`
}

func (DocumentStatusTransition) TableName() string { return "document_status_transition" }
