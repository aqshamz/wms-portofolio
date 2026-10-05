package stockcontrol

import "time"

type WarehouseTransfer struct {
	ID                 string     `gorm:"column:warehouse_transfer_id;size:140;primaryKey"`
	DocumentTypeID     string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID           string     `gorm:"column:status_id;type:uuid;not null"`
	OwnerID            string     `gorm:"column:owner_id;type:uuid;not null"`
	SourceWarehouseID  string     `gorm:"column:source_warehouse_id;type:uuid;not null"`
	TargetWarehouseID  string     `gorm:"column:target_warehouse_id;type:uuid;not null"`
	BusinessDate       time.Time  `gorm:"column:business_date;type:date;not null"`
	Notes              *string    `gorm:"column:notes;type:text"`
	ApprovedAt         *time.Time `gorm:"column:approved_at"`
	ApprovedBy         *string    `gorm:"column:approved_by;type:uuid"`
	DispatchedAt       *time.Time `gorm:"column:dispatched_at"`
	DispatchedBy       *string    `gorm:"column:dispatched_by;type:uuid"`
	ReceivedAt         *time.Time `gorm:"column:received_at"`
	ReceivedBy         *string    `gorm:"column:received_by;type:uuid"`
	CancelledAt        *time.Time `gorm:"column:cancelled_at"`
	CancelledBy        *string    `gorm:"column:cancelled_by;type:uuid"`
	CancellationReason *string    `gorm:"column:cancellation_reason;type:text"`
	CreatedAt          time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy          string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy          *string    `gorm:"column:updated_by;type:uuid"`
	VersionNo          int64      `gorm:"column:version_no;not null;default:1"`
}

func (WarehouseTransfer) TableName() string { return "warehouse_transfer" }

type WarehouseTransferLine struct {
	ID                      string     `gorm:"column:warehouse_transfer_line_id;size:180;primaryKey"`
	WarehouseTransferID     string     `gorm:"column:warehouse_transfer_id;size:140;not null"`
	LineNo                  int        `gorm:"column:line_no;not null"`
	SourceBalanceID         string     `gorm:"column:source_balance_id;size:160;not null"`
	SourceBalanceVersionNo  int64      `gorm:"column:source_balance_version_no;not null"`
	ItemID                  string     `gorm:"column:item_id;type:uuid;not null"`
	LotID                   *string    `gorm:"column:lot_id;size:120"`
	SerialID                *string    `gorm:"column:serial_id;size:160"`
	HandlingUnitID          *string    `gorm:"column:handling_unit_id;size:120"`
	SourceLocationID        string     `gorm:"column:source_location_id;type:uuid;not null"`
	SourceInventoryStatusID string     `gorm:"column:source_inventory_status_id;type:uuid;not null"`
	UOMID                   string     `gorm:"column:uom_id;type:uuid;not null"`
	Quantity                string     `gorm:"column:quantity;type:numeric(20,6);not null"`
	DispatchMovementID      *string    `gorm:"column:dispatch_movement_id;size:140"`
	ReceiptLocationID       *string    `gorm:"column:receipt_location_id;type:uuid"`
	ReceiptMovementID       *string    `gorm:"column:receipt_movement_id;size:140"`
	ReceivedBalanceID       *string    `gorm:"column:received_balance_id;size:160"`
	PutawayTargetLocationID *string    `gorm:"column:putaway_target_location_id;type:uuid"`
	PutawayMovementID       *string    `gorm:"column:putaway_movement_id;size:140"`
	PutawayResultBalanceID  *string    `gorm:"column:putaway_result_balance_id;size:160"`
	PutawayCompletedAt      *time.Time `gorm:"column:putaway_completed_at"`
	PutawayCompletedBy      *string    `gorm:"column:putaway_completed_by;type:uuid"`
	CreatedAt               time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy               string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt               time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy               *string    `gorm:"column:updated_by;type:uuid"`
	VersionNo               int64      `gorm:"column:version_no;not null;default:1"`
}

func (WarehouseTransferLine) TableName() string { return "warehouse_transfer_line" }
