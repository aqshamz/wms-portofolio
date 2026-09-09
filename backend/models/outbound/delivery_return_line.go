package outbound

import "time"

type DeliveryReturnLine struct {
	ID                string    `gorm:"column:delivery_return_line_id;size:170;primaryKey"`
	DeliveryID        string    `gorm:"column:delivery_id;size:120;not null"`
	DeliveryLineID    string    `gorm:"column:delivery_line_id;size:160;not null"`
	ReturnLocationID  string    `gorm:"column:return_location_id;type:uuid;not null"`
	ReturnedBalanceID string    `gorm:"column:returned_balance_id;size:160;not null"`
	ReturnedQty       string    `gorm:"column:returned_qty;type:numeric(20,6);not null"`
	UOMID             string    `gorm:"column:uom_id;type:uuid;not null"`
	MovementID        string    `gorm:"column:movement_id;size:140;not null;uniqueIndex"`
	ReturnedAt        time.Time `gorm:"column:returned_at;not null"`
	ReturnedBy        string    `gorm:"column:returned_by;type:uuid;not null"`
}

func (DeliveryReturnLine) TableName() string { return "delivery_return_line" }
