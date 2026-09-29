package stockcontrol

import "time"

// InventoryAdjustment is a multi-line adjustment document header. The Legacy
// fields keep migration 18 data readable while new documents use line rows.
type InventoryAdjustment struct {
	ID                            string     `gorm:"column:inventory_adjustment_request_id;size:140;primaryKey"`
	DocumentTypeID                string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID                      string     `gorm:"column:status_id;type:uuid;not null"`
	OwnerID                       string     `gorm:"column:owner_id;type:uuid;not null"`
	WarehouseID                   string     `gorm:"column:warehouse_id;type:uuid;not null"`
	BusinessDate                  time.Time  `gorm:"column:business_date;type:date;not null"`
	Direction                     string     `gorm:"column:direction;size:10;not null"`
	ReasonCodeID                  string     `gorm:"column:reason_code_id;type:uuid;not null"`
	Notes                         *string    `gorm:"column:notes;type:text"`
	CompletedAt                   *time.Time `gorm:"column:completed_at"`
	CancelledAt                   *time.Time `gorm:"column:cancelled_at"`
	CancelledBy                   *string    `gorm:"column:cancelled_by;type:uuid"`
	CancellationReason            *string    `gorm:"column:cancellation_reason;type:text"`
	LegacyBalanceID               *string    `gorm:"column:balance_id;size:160"`
	LegacyPlannedBalanceVersionNo *int64     `gorm:"column:planned_balance_version_no"`
	LegacyItemID                  *string    `gorm:"column:item_id;type:uuid"`
	LegacyLotID                   *string    `gorm:"column:lot_id;size:120"`
	LegacySerialID                *string    `gorm:"column:serial_id;size:160"`
	LegacyHandlingUnitID          *string    `gorm:"column:handling_unit_id;size:120"`
	LegacyLocationID              *string    `gorm:"column:location_id;type:uuid"`
	LegacyInventoryStatusID       *string    `gorm:"column:inventory_status_id;type:uuid"`
	LegacyUOMID                   *string    `gorm:"column:uom_id;type:uuid"`
	LegacyQuantity                *string    `gorm:"column:quantity;type:numeric(20,6)"`
	LegacyApprovedAt              *time.Time `gorm:"column:approved_at"`
	LegacyApprovedBy              *string    `gorm:"column:approved_by;type:uuid"`
	LegacyPostedAt                *time.Time `gorm:"column:posted_at"`
	LegacyPostedBy                *string    `gorm:"column:posted_by;type:uuid"`
	LegacyInventoryMovementID     *string    `gorm:"column:inventory_movement_id;size:140"`
	LegacyResultingBalanceID      *string    `gorm:"column:resulting_balance_id;size:160"`
	CreatedAt                     time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy                     string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt                     time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy                     *string    `gorm:"column:updated_by;type:uuid"`
	VersionNo                     int64      `gorm:"column:version_no;not null;default:1"`
}

func (InventoryAdjustment) TableName() string { return "inventory_adjustment_request" }

type InventoryAdjustmentLine struct {
	ID                      string     `gorm:"column:inventory_adjustment_request_line_id;size:180;primaryKey"`
	AdjustmentID            string     `gorm:"column:inventory_adjustment_request_id;size:140;not null"`
	LineNo                  int        `gorm:"column:line_no;not null"`
	BalanceID               string     `gorm:"column:balance_id;size:160;not null"`
	PlannedBalanceVersionNo int64      `gorm:"column:planned_balance_version_no;not null"`
	ItemID                  string     `gorm:"column:item_id;type:uuid;not null"`
	LotID                   *string    `gorm:"column:lot_id;size:120"`
	SerialID                *string    `gorm:"column:serial_id;size:160"`
	HandlingUnitID          *string    `gorm:"column:handling_unit_id;size:120"`
	LocationID              string     `gorm:"column:location_id;type:uuid;not null"`
	InventoryStatusID       string     `gorm:"column:inventory_status_id;type:uuid;not null"`
	UOMID                   string     `gorm:"column:uom_id;type:uuid;not null"`
	Quantity                string     `gorm:"column:quantity;type:numeric(20,6);not null"`
	DecisionCode            string     `gorm:"column:decision_code;size:20;not null;default:PENDING"`
	ApprovedAt              *time.Time `gorm:"column:approved_at"`
	ApprovedBy              *string    `gorm:"column:approved_by;type:uuid"`
	RejectedAt              *time.Time `gorm:"column:rejected_at"`
	RejectedBy              *string    `gorm:"column:rejected_by;type:uuid"`
	RejectionReason         *string    `gorm:"column:rejection_reason;type:text"`
	CancelledAt             *time.Time `gorm:"column:cancelled_at"`
	CancelledBy             *string    `gorm:"column:cancelled_by;type:uuid"`
	CancellationReason      *string    `gorm:"column:cancellation_reason;type:text"`
	InventoryMovementID     *string    `gorm:"column:inventory_movement_id;size:140"`
	ResultingBalanceID      *string    `gorm:"column:resulting_balance_id;size:160"`
	CreatedAt               time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy               string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt               time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy               *string    `gorm:"column:updated_by;type:uuid"`
	VersionNo               int64      `gorm:"column:version_no;not null;default:1"`
}

func (InventoryAdjustmentLine) TableName() string { return "inventory_adjustment_request_line" }
