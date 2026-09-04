package master

import "time"

type AccountWarehouseAccess struct {
	AccountID   string    `gorm:"column:account_id;type:uuid;primaryKey" json:"account_id"`
	WarehouseID string    `gorm:"column:warehouse_id;type:uuid;primaryKey" json:"warehouse_id"`
	GrantedBy   *string   `gorm:"column:granted_by;type:uuid" json:"granted_by,omitempty"`
	GrantedAt   time.Time `gorm:"column:granted_at;not null;default:clock_timestamp()" json:"granted_at"`
}

func (AccountWarehouseAccess) TableName() string { return "account_warehouse_access" }
