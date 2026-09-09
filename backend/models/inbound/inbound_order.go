package inbound

import "time"

type InboundOrder struct {
	ID                  string     `gorm:"column:inbound_id;size:120;primaryKey"`
	DocumentTypeID      string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID            string     `gorm:"column:status_id;type:uuid;not null"`
	OwnerID             string     `gorm:"column:owner_id;type:uuid;not null"`
	VendorID            string     `gorm:"column:vendor_id;type:uuid;not null"`
	WarehouseID         string     `gorm:"column:warehouse_id;type:uuid;not null"`
	BusinessDate        time.Time  `gorm:"column:business_date;type:date;not null"`
	ExpectedArrivalAt   *time.Time `gorm:"column:expected_arrival_at"`
	ExternalReference   *string    `gorm:"column:external_reference;size:120"`
	SupplierReference   *string    `gorm:"column:supplier_reference;size:120"`
	Notes               *string    `gorm:"column:notes;type:text"`
	SupersedesInboundID *string    `gorm:"column:supersedes_inbound_id;size:120"`
	CreatedAt           time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy           string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy           *string    `gorm:"column:updated_by;type:uuid"`
	VersionNo           int64      `gorm:"column:version_no;not null;default:1"`
}

func (InboundOrder) TableName() string { return "inbound_order" }
