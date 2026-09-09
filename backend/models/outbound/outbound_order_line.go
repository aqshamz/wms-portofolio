package outbound

import "time"

type OutboundOrderLine struct {
	ID                    string    `gorm:"column:outbound_line_id;size:150;primaryKey"`
	OutboundID            string    `gorm:"column:outbound_id;size:120;not null"`
	LineNo                int       `gorm:"column:line_no;not null"`
	ItemID                string    `gorm:"column:item_id;type:uuid;not null"`
	OrderedQty            string    `gorm:"column:ordered_qty;type:numeric(20,6);not null"`
	AllocatedQty          string    `gorm:"column:allocated_qty;type:numeric(20,6);not null;default:0"`
	PickedQty             string    `gorm:"column:picked_qty;type:numeric(20,6);not null;default:0"`
	CheckedQty            string    `gorm:"column:checked_qty;type:numeric(20,6);not null;default:0"`
	PackedQty             string    `gorm:"column:packed_qty;type:numeric(20,6);not null;default:0"`
	ShippedQty            string    `gorm:"column:shipped_qty;type:numeric(20,6);not null;default:0"`
	DeliveredQty          string    `gorm:"column:delivered_qty;type:numeric(20,6);not null;default:0"`
	RejectedQty           string    `gorm:"column:rejected_qty;type:numeric(20,6);not null;default:0"`
	ShortAcceptedQty      string    `gorm:"column:short_accepted_qty;type:numeric(20,6);not null;default:0"`
	UOMID                 string    `gorm:"column:uom_id;type:uuid;not null"`
	RequestedLotNo        *string   `gorm:"column:requested_lot_no;size:100"`
	CustomerLineReference *string   `gorm:"column:customer_line_reference;size:100"`
	Notes                 *string   `gorm:"column:notes;type:text"`
	CreatedAt             time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy             string    `gorm:"column:created_by;type:uuid;not null"`
}

func (OutboundOrderLine) TableName() string { return "outbound_order_line" }
