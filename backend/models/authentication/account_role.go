package authentication

import "time"

type AccountRole struct {
	AccountID  string    `gorm:"column:account_id;type:uuid;primaryKey"`
	RoleID     string    `gorm:"column:role_id;type:uuid;primaryKey"`
	AssignedAt time.Time `gorm:"column:assigned_at;not null;default:clock_timestamp()"`
	AssignedBy *string   `gorm:"column:assigned_by;type:uuid"`
}

func (AccountRole) TableName() string { return "account_role" }
