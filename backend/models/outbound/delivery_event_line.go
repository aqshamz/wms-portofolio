package outbound

type DeliveryEventLine struct {
	DeliveryEventID string `gorm:"column:delivery_event_id;size:160;primaryKey"`
	DeliveryLineID  string `gorm:"column:delivery_line_id;size:160;primaryKey"`
	DeliveredQty    string `gorm:"column:delivered_qty;type:numeric(20,6);not null"`
}

func (DeliveryEventLine) TableName() string { return "delivery_event_line" }
