package inventory

import "time"

// SerialInventory is the current-state pointer. Its balance supplies location,
// lot, handling-unit, status and UOM dimensions without duplicating them here.
type SerialInventory struct {
	SerialID  string    `gorm:"column:serial_id;size:160;primaryKey"`
	BalanceID string    `gorm:"column:balance_id;size:160;not null"`
	OwnerID   string    `gorm:"column:owner_id;type:uuid;not null"`
	ItemID    string    `gorm:"column:item_id;type:uuid;not null"`
	VersionNo int64     `gorm:"column:version_no;not null;default:1"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:clock_timestamp()"`
}

func (SerialInventory) TableName() string { return "serial_inventory" }
