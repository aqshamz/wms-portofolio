package authentication

import "time"

type RolePermission struct {
	RoleID       string    `gorm:"column:role_id;type:uuid;primaryKey"`
	PermissionID string    `gorm:"column:permission_id;type:uuid;primaryKey"`
	GrantedAt    time.Time `gorm:"column:granted_at;not null;default:clock_timestamp()"`
	GrantedBy    *string   `gorm:"column:granted_by;type:uuid"`
}

func (RolePermission) TableName() string { return "role_permission" }
