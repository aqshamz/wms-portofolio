package inbound

import "time"

type InboundOrderLine struct {
	ID                    string     `gorm:"column:inbound_line_id;size:150;primaryKey"`
	InboundID             string     `gorm:"column:inbound_id;size:120;not null;uniqueIndex:uq_inbound_line_no,priority:1"`
	PurchaseOrderLineID   *string    `gorm:"column:purchase_order_line_id;size:150"`
	LineNo                int        `gorm:"column:line_no;not null;uniqueIndex:uq_inbound_line_no,priority:2"`
	ItemID                string     `gorm:"column:item_id;type:uuid;not null"`
	ExpectedQty           string     `gorm:"column:expected_qty;type:numeric(20,6);not null"`
	UOMID                 string     `gorm:"column:uom_id;type:uuid;not null"`
	ExpectedLotNo         *string    `gorm:"column:expected_lot_no;size:100"`
	ExpectedExpiryDate    *time.Time `gorm:"column:expected_expiry_date;type:date"`
	CustomerLineReference *string    `gorm:"column:customer_line_reference;size:100"`
	Notes                 *string    `gorm:"column:notes;type:text"`
	CreatedAt             time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy             string     `gorm:"column:created_by;type:uuid;not null"`
}

func (InboundOrderLine) TableName() string { return "inbound_order_line" }
