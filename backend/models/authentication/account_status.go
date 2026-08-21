package authentication

import "time"

type AccountStatus struct {
	ID          string    `gorm:"column:account_status_id;type:uuid;default:gen_random_uuid();primaryKey" json:"account_status_id"`
	Code        string    `gorm:"column:code;size:30;not null;uniqueIndex" json:"code"`
	Name        string    `gorm:"column:name;size:100;not null" json:"name"`
	Description *string   `gorm:"column:description;type:text" json:"description,omitempty"`
	AllowsLogin bool      `gorm:"column:allows_login;not null;default:false" json:"allows_login"`
	IsActive    bool      `gorm:"column:is_active;not null;default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:clock_timestamp()" json:"created_at"`
}

func (AccountStatus) TableName() string { return "account_status" }
