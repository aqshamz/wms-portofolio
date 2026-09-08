package inbound

import "time"

type ReceiptInventory struct {
	ID                       string    `gorm:"column:receipt_inventory_id;size:160;primaryKey"`
	ReceiptLineID            string    `gorm:"column:receipt_line_id;size:150;not null"`
	ItemID                   string    `gorm:"column:item_id;type:uuid;not null"`
	SourceQty                string    `gorm:"column:source_qty;type:numeric(20,6);not null"`
	SourceUOMID              string    `gorm:"column:source_uom_id;type:uuid;not null"`
	BaseQty                  string    `gorm:"column:base_qty;type:numeric(20,6);not null"`
	BaseUOMID                string    `gorm:"column:base_uom_id;type:uuid;not null"`
	LotID                    *string   `gorm:"column:lot_id;size:120"`
	HandlingUnitID           *string   `gorm:"column:handling_unit_id;size:120"`
	ReceivedLocationID       string    `gorm:"column:received_location_id;type:uuid;not null"`
	InitialInventoryStatusID string    `gorm:"column:initial_inventory_status_id;type:uuid;not null"`
	InitialBalanceID         *string   `gorm:"column:initial_balance_id;size:160"`
	CreatedAt                time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy                string    `gorm:"column:created_by;type:uuid;not null"`
}

func (ReceiptInventory) TableName() string { return "receipt_inventory" }
