package outbound

import "time"

type ShipmentLine struct {
	ID              string    `gorm:"column:shipment_line_id;size:160;primaryKey"`
	ShipmentID      string    `gorm:"column:shipment_id;size:120;not null"`
	PackingLineID   string    `gorm:"column:packing_line_id;size:150;not null;uniqueIndex"`
	SourceBalanceID string    `gorm:"column:source_balance_id;size:160;not null"`
	ShippedQty      string    `gorm:"column:shipped_qty;type:numeric(20,6);not null"`
	UOMID           string    `gorm:"column:uom_id;type:uuid;not null"`
	MovementID      string    `gorm:"column:movement_id;size:140;not null;uniqueIndex"`
	CreatedAt       time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy       string    `gorm:"column:created_by;type:uuid;not null"`
}

func (ShipmentLine) TableName() string { return "shipment_line" }
