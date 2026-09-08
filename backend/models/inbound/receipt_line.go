package inbound

import "time"

type ReceiptLine struct {
	ID                string    `gorm:"column:receipt_line_id;size:150;primaryKey"`
	ReceiptID         string    `gorm:"column:receipt_id;size:120;not null;uniqueIndex:uq_receipt_line_no,priority:1"`
	InboundLineID     *string   `gorm:"column:inbound_line_id;size:150"`
	LineNo            int       `gorm:"column:line_no;not null;uniqueIndex:uq_receipt_line_no,priority:2"`
	ItemID            string    `gorm:"column:item_id;type:uuid;not null"`
	ReceivedQty       string    `gorm:"column:received_qty;type:numeric(20,6);not null"`
	RejectedQty       string    `gorm:"column:rejected_qty;type:numeric(20,6);not null;default:0"`
	ExceptionNotes    *string   `gorm:"column:exception_notes;type:text"`
	ExceptionTypeCode *string   `gorm:"column:exception_type_code;size:40"`
	UOMID             string    `gorm:"column:uom_id;type:uuid;not null"`
	CreatedAt         time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy         string    `gorm:"column:created_by;type:uuid;not null"`
}

func (ReceiptLine) TableName() string { return "receipt_line" }
