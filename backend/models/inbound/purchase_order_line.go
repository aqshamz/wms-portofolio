package inbound

import "time"

type PurchaseOrderLine struct {
	ID                       string     `gorm:"column:purchase_order_line_id;size:150;primaryKey"`
	PurchaseOrderID          string     `gorm:"column:purchase_order_id;size:120;not null;uniqueIndex:uq_purchase_order_line_no,priority:1"`
	OwnerID                  string     `gorm:"column:owner_id;type:uuid;not null"`
	LineNo                   int        `gorm:"column:line_no;not null;uniqueIndex:uq_purchase_order_line_no,priority:2"`
	ItemID                   string     `gorm:"column:item_id;type:uuid;not null"`
	OrderedQty               string     `gorm:"column:ordered_qty;type:numeric(20,6);not null"`
	OverReceiptTolerancePct  string     `gorm:"column:over_receipt_tolerance_pct;type:numeric(7,4);not null;default:0"`
	UnderReceiptTolerancePct string     `gorm:"column:under_receipt_tolerance_pct;type:numeric(7,4);not null;default:0"`
	UOMID                    string     `gorm:"column:uom_id;type:uuid;not null"`
	VendorItemCode           *string    `gorm:"column:vendor_item_code;size:100"`
	ExpectedLotNo            *string    `gorm:"column:expected_lot_no;size:100"`
	ExpectedExpiryDate       *time.Time `gorm:"column:expected_expiry_date;type:date"`
	Notes                    *string    `gorm:"column:notes;type:text"`
	CreatedAt                time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy                string     `gorm:"column:created_by;type:uuid;not null"`
}

func (PurchaseOrderLine) TableName() string { return "purchase_order_line" }
