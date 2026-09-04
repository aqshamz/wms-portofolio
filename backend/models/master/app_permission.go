package master

import "time"

type AppPermission struct {
	ID          string    `gorm:"column:permission_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code        string    `gorm:"column:code;size:100;not null"`
	Name        string    `gorm:"column:name;size:150;not null"`
	ModuleCode  string    `gorm:"column:module_code;size:50;not null"`
	Description *string   `gorm:"column:description;type:text"`
	IsActive    bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
}

func (AppPermission) TableName() string { return "app_permission" }
