package authentication

import "time"

type AppRole struct {
	ID          string    `gorm:"column:role_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code        string    `gorm:"column:code;size:60;not null"`
	Name        string    `gorm:"column:name;size:120;not null"`
	Description *string   `gorm:"column:description;type:text"`
	IsActive    bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy   *string   `gorm:"column:created_by;type:uuid"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy   *string   `gorm:"column:updated_by;type:uuid"`
	VersionNo   int       `gorm:"column:version_no;not null;default:1"`
}

func (AppRole) TableName() string { return "app_role" }
