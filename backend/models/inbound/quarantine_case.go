package inbound

import "time"

type QuarantineCase struct {
	ID                     string     `gorm:"column:quarantine_case_id;size:140;primaryKey"`
	ParentQuarantineCaseID *string    `gorm:"column:parent_quarantine_case_id;size:140"`
	DocumentTypeID         string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID               string     `gorm:"column:status_id;type:uuid;not null"`
	ReceiptInventoryID     string     `gorm:"column:receipt_inventory_id;size:160;not null"`
	InspectionID           string     `gorm:"column:inspection_id;size:120;not null"`
	QuarantineBalanceID    string     `gorm:"column:quarantine_balance_id;size:160;not null"`
	OwnerID                string     `gorm:"column:owner_id;type:uuid;not null"`
	WarehouseID            string     `gorm:"column:warehouse_id;type:uuid;not null"`
	QuarantineQty          string     `gorm:"column:quarantine_qty;type:numeric(20,6);not null"`
	UOMID                  string     `gorm:"column:uom_id;type:uuid;not null"`
	OpenedAt               time.Time  `gorm:"column:opened_at;not null;default:clock_timestamp()"`
	ClosedAt               *time.Time `gorm:"column:closed_at"`
	Notes                  *string    `gorm:"column:notes;type:text"`
	CreatedBy              string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt              time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy              *string    `gorm:"column:updated_by;type:uuid"`
	VersionNo              int64      `gorm:"column:version_no;not null;default:1"`
}

func (QuarantineCase) TableName() string { return "quarantine_case" }
