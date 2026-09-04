package master

type TaskStatusTransition struct {
	ID                   string  `gorm:"column:task_status_transition_id;type:uuid;default:gen_random_uuid();primaryKey"`
	FromStatusID         string  `gorm:"column:from_status_id;type:uuid;not null"`
	ToStatusID           string  `gorm:"column:to_status_id;type:uuid;not null"`
	RequiredPermissionID *string `gorm:"column:required_permission_id;type:uuid"`
	IsActive             bool    `gorm:"column:is_active;not null;default:true"`
}

func (TaskStatusTransition) TableName() string { return "task_status_transition" }
