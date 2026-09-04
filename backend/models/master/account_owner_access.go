package master

import "time"

type AccountOwnerAccess struct {
	AccountID string    `gorm:"column:account_id;type:uuid;primaryKey" json:"account_id"`
	OwnerID   string    `gorm:"column:owner_id;type:uuid;primaryKey" json:"owner_id"`
	GrantedBy *string   `gorm:"column:granted_by;type:uuid" json:"granted_by,omitempty"`
	GrantedAt time.Time `gorm:"column:granted_at;not null;default:clock_timestamp()" json:"granted_at"`
}

func (AccountOwnerAccess) TableName() string { return "account_owner_access" }
