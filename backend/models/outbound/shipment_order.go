package outbound

import "time"

type ShipmentOrder struct {
	ShipmentID string     `gorm:"column:shipment_id;size:120;primaryKey"`
	OutboundID string     `gorm:"column:outbound_id;size:120;primaryKey"`
	AddedAt    time.Time  `gorm:"column:added_at;not null;default:clock_timestamp()"`
	AddedBy    string     `gorm:"column:added_by;type:uuid;not null"`
	RemovedAt  *time.Time `gorm:"column:removed_at"`
	RemovedBy  *string    `gorm:"column:removed_by;type:uuid"`
}

func (ShipmentOrder) TableName() string { return "shipment_order" }
