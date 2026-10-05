package outbound

import "time"

type VendorReturnTransaction struct {
	ID                      string     `gorm:"column:vendor_return_id;size:140;primaryKey"`
	DocumentTypeID          string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID                string     `gorm:"column:status_id;type:uuid;not null"`
	QuarantineDispositionID string     `gorm:"column:quarantine_disposition_id;size:150;not null;uniqueIndex"`
	QuarantineCaseID        string     `gorm:"column:quarantine_case_id;size:140;not null"`
	OwnerID                 string     `gorm:"column:owner_id;type:uuid;not null"`
	VendorID                string     `gorm:"column:vendor_id;type:uuid;not null"`
	WarehouseID             string     `gorm:"column:warehouse_id;type:uuid;not null"`
	BusinessDate            time.Time  `gorm:"column:business_date;type:date;not null"`
	SourceBalanceID         string     `gorm:"column:source_balance_id;size:160;not null"`
	PlannedBalanceVersionNo int64      `gorm:"column:planned_balance_version_no;not null"`
	ItemID                  string     `gorm:"column:item_id;type:uuid;not null"`
	LotID                   *string    `gorm:"column:lot_id;size:120"`
	SerialID                *string    `gorm:"column:serial_id;size:120"`
	HandlingUnitID          *string    `gorm:"column:handling_unit_id;size:120"`
	SourceLocationID        string     `gorm:"column:source_location_id;type:uuid;not null"`
	SourceInventoryStatusID string     `gorm:"column:source_inventory_status_id;type:uuid;not null"`
	ReturnDockLocationID    *string    `gorm:"column:return_dock_location_id;type:uuid"`
	ReturnPendingStatusID   *string    `gorm:"column:return_pending_status_id;type:uuid"`
	StagedBalanceID         *string    `gorm:"column:staged_balance_id;size:160"`
	StagingMovementID       *string    `gorm:"column:staging_movement_id;size:140"`
	CancellationMovementID  *string    `gorm:"column:cancellation_movement_id;size:140"`
	Quantity                string     `gorm:"column:quantity;type:numeric(20,6);not null"`
	UOMID                   string     `gorm:"column:uom_id;type:uuid;not null"`
	Notes                   *string    `gorm:"column:notes;type:text"`
	PlannedAt               time.Time  `gorm:"column:planned_at;not null"`
	CompletedAt             *time.Time `gorm:"column:completed_at"`
	CompletedBy             *string    `gorm:"column:completed_by;type:uuid"`
	CancelledAt             *time.Time `gorm:"column:cancelled_at"`
	CancelledBy             *string    `gorm:"column:cancelled_by;type:uuid"`
	CancellationReason      *string    `gorm:"column:cancellation_reason;type:text"`
	InventoryMovementID     *string    `gorm:"column:inventory_movement_id;size:140"`
	CreatedAt               time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy               string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt               time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy               *string    `gorm:"column:updated_by;type:uuid"`
	VersionNo               int64      `gorm:"column:version_no;not null;default:1"`
}

func (VendorReturnTransaction) TableName() string { return "vendor_return_transaction" }
