package master

type TaskPriority struct {
	ID            string `gorm:"column:task_priority_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code          string `gorm:"column:code;size:40;not null"`
	Name          string `gorm:"column:name;size:100;not null"`
	PriorityValue int    `gorm:"column:priority_value;not null"`
	IsActive      bool   `gorm:"column:is_active;not null;default:true"`
}

func (TaskPriority) TableName() string { return "task_priority" }
