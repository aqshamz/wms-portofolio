package outbound

import "time"

type OutboundReturnPolicy struct {
	OwnerID                 string    `gorm:"column:owner_id;type:uuid;primaryKey"`
	WarehouseID             string    `gorm:"column:warehouse_id;type:uuid;primaryKey"`
	ReturnLocationID        string    `gorm:"column:return_location_id;type:uuid;not null"`
	ReturnInventoryStatusID string    `gorm:"column:return_inventory_status_id;type:uuid;not null"`
	IsActive                bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt               time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy               string    `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt               time.Time `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy               *string   `gorm:"column:updated_by;type:uuid"`
}

func (OutboundReturnPolicy) TableName() string { return "outbound_return_policy" }
