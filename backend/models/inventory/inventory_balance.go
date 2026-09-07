package inventory

import "time"

type InventoryBalance struct {
	ID                string    `gorm:"column:balance_id;size:160;primaryKey"`
	OwnerID           string    `gorm:"column:owner_id;type:uuid;not null"`
	WarehouseID       string    `gorm:"column:warehouse_id;type:uuid;not null"`
	LocationID        string    `gorm:"column:location_id;type:uuid;not null"`
	ItemID            string    `gorm:"column:item_id;type:uuid;not null"`
	LotID             *string   `gorm:"column:lot_id;size:120"`
	HandlingUnitID    *string   `gorm:"column:handling_unit_id;size:120"`
	InventoryStatusID string    `gorm:"column:inventory_status_id;type:uuid;not null"`
	OnHandQty         string    `gorm:"column:on_hand_qty;type:numeric(20,6);not null;default:0"`
	ReservedQty       string    `gorm:"column:reserved_qty;type:numeric(20,6);not null;default:0"`
	UOMID             string    `gorm:"column:uom_id;type:uuid;not null"`
	VersionNo         int64     `gorm:"column:version_no;not null;default:1"`
	UpdatedAt         time.Time `gorm:"column:updated_at;not null;default:clock_timestamp()"`
}

func (InventoryBalance) TableName() string { return "inventory_balance" }
