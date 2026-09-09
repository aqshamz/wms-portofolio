package outbound

import "time"

type InventoryReservation struct {
	ID                string     `gorm:"column:reservation_id;size:140;primaryKey"`
	DocumentTypeID    string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID          string     `gorm:"column:status_id;type:uuid;not null"`
	PickingStrategyID *string    `gorm:"column:picking_strategy_id;type:uuid"`
	OutboundLineID    string     `gorm:"column:outbound_line_id;size:150;not null"`
	BalanceID         string     `gorm:"column:balance_id;size:160;not null"`
	ReservedQty       string     `gorm:"column:reserved_qty;type:numeric(20,6);not null"`
	PickedQty         string     `gorm:"column:picked_qty;type:numeric(20,6);not null;default:0"`
	UOMID             string     `gorm:"column:uom_id;type:uuid;not null"`
	ReservedAt        time.Time  `gorm:"column:reserved_at;not null;default:clock_timestamp()"`
	ReleasedAt        *time.Time `gorm:"column:released_at"`
	CreatedBy         string     `gorm:"column:created_by;type:uuid;not null"`
}

func (InventoryReservation) TableName() string { return "inventory_reservation" }
