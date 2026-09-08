package inbound

import "time"

type PurchaseOrder struct {
	ID                string     `gorm:"column:purchase_order_id;size:120;primaryKey"`
	DocumentTypeID    string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID          string     `gorm:"column:status_id;type:uuid;not null"`
	OwnerID           string     `gorm:"column:owner_id;type:uuid;not null;uniqueIndex:uq_purchase_order_owner_no,priority:1"`
	VendorID          string     `gorm:"column:vendor_id;type:uuid;not null"`
	WarehouseID       string     `gorm:"column:warehouse_id;type:uuid;not null"`
	BusinessDate      time.Time  `gorm:"column:business_date;type:date;not null"`
	PurchaseOrderNo   string     `gorm:"column:purchase_order_no;size:120;not null;uniqueIndex:uq_purchase_order_owner_no,priority:2"`
	OrderedAt         time.Time  `gorm:"column:ordered_at;not null"`
	ExpectedArrivalAt *time.Time `gorm:"column:expected_arrival_at"`
	Notes             *string    `gorm:"column:notes;type:text"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy         string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy         *string    `gorm:"column:updated_by;type:uuid"`
	VersionNo         int64      `gorm:"column:version_no;not null;default:1"`
}

func (PurchaseOrder) TableName() string { return "purchase_order" }
