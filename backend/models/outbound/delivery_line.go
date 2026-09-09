package outbound

import "time"

type DeliveryLine struct {
	ID             string    `gorm:"column:delivery_line_id;size:160;primaryKey"`
	DeliveryID     string    `gorm:"column:delivery_id;size:120;not null"`
	ShipmentLineID string    `gorm:"column:shipment_line_id;size:160;not null;uniqueIndex"`
	PlannedQty     string    `gorm:"column:planned_qty;type:numeric(20,6);not null"`
	DeliveredQty   string    `gorm:"column:delivered_qty;type:numeric(20,6);not null;default:0"`
	ReturnedQty    string    `gorm:"column:returned_qty;type:numeric(20,6);not null;default:0"`
	UOMID          string    `gorm:"column:uom_id;type:uuid;not null"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy      string    `gorm:"column:created_by;type:uuid;not null"`
}

func (DeliveryLine) TableName() string { return "delivery_line" }
