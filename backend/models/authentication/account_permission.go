package authentication

import "time"

type AccountPermission struct {
	AccountID    string    `gorm:"column:account_id;type:uuid;primaryKey"`
	PermissionID string    `gorm:"column:permission_id;type:uuid;primaryKey"`
	GrantedAt    time.Time `gorm:"column:granted_at;not null;default:clock_timestamp()"`
	GrantedBy    *string   `gorm:"column:granted_by;type:uuid"`
}

func (AccountPermission) TableName() string { return "account_permission" }
