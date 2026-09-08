package inbound

type ReceiptLineSerial struct {
	ReceiptInventoryID string `gorm:"column:receipt_inventory_id;size:160;primaryKey"`
	SerialID           string `gorm:"column:serial_id;size:160;primaryKey;uniqueIndex"`
}

func (ReceiptLineSerial) TableName() string { return "receipt_line_serial" }
