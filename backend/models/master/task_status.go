package master

type TaskStatus struct {
	ID          string `gorm:"column:task_status_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code        string `gorm:"column:code;size:40;not null"`
	Name        string `gorm:"column:name;size:100;not null"`
	IsInitial   bool   `gorm:"column:is_initial;not null;default:false"`
	IsFinal     bool   `gorm:"column:is_final;not null;default:false"`
	IsCancelled bool   `gorm:"column:is_cancelled;not null;default:false"`
	IsActive    bool   `gorm:"column:is_active;not null;default:true"`
}

func (TaskStatus) TableName() string { return "task_status" }
