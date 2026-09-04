package master

import "time"

type WarehouseOwner struct {
	WarehouseID string    `gorm:"column:warehouse_id;type:uuid;primaryKey;uniqueIndex:uq_warehouse_owner_reverse,priority:2" json:"warehouse_id"`
	OwnerID     string    `gorm:"column:owner_id;type:uuid;primaryKey;uniqueIndex:uq_warehouse_owner_reverse,priority:1" json:"owner_id"`
	IsActive    bool      `gorm:"column:is_active;not null;default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:clock_timestamp()" json:"created_at"`
	CreatedBy   *string   `gorm:"column:created_by;type:uuid" json:"created_by,omitempty"`
}

func (WarehouseOwner) TableName() string { return "warehouse_owner" }
