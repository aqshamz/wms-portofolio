package stockcontrol

import "time"

type CycleCount struct {
	ID                 string     `gorm:"column:cycle_count_id;size:140;primaryKey"`
	DocumentTypeID     string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID           string     `gorm:"column:status_id;type:uuid;not null"`
	OwnerID            string     `gorm:"column:owner_id;type:uuid;not null"`
	WarehouseID        string     `gorm:"column:warehouse_id;type:uuid;not null"`
	BusinessDate       time.Time  `gorm:"column:business_date;type:date;not null"`
	ToleranceQty       string     `gorm:"column:tolerance_qty;type:numeric(20,6);not null;default:0"`
	BlindCount         bool       `gorm:"column:blind_count;not null;default:true"`
	Notes              *string    `gorm:"column:notes;type:text"`
	CompletedAt        *time.Time `gorm:"column:completed_at"`
	CancelledAt        *time.Time `gorm:"column:cancelled_at"`
	CancelledBy        *string    `gorm:"column:cancelled_by;type:uuid"`
	CancellationReason *string    `gorm:"column:cancellation_reason;type:text"`
	CreatedAt          time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy          string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy          *string    `gorm:"column:updated_by;type:uuid"`
	VersionNo          int64      `gorm:"column:version_no;not null;default:1"`
}

func (CycleCount) TableName() string { return "cycle_count" }

type CycleCountLine struct {
	ID                  string     `gorm:"column:cycle_count_line_id;size:180;primaryKey"`
	CycleCountID        string     `gorm:"column:cycle_count_id;size:140;not null"`
	LineNo              int        `gorm:"column:line_no;not null"`
	BalanceID           string     `gorm:"column:balance_id;size:160;not null"`
	SnapshotVersionNo   int64      `gorm:"column:snapshot_version_no;not null"`
	ItemID              string     `gorm:"column:item_id;type:uuid;not null"`
	LotID               *string    `gorm:"column:lot_id;size:120"`
	HandlingUnitID      *string    `gorm:"column:handling_unit_id;size:120"`
	LocationID          string     `gorm:"column:location_id;type:uuid;not null"`
	InventoryStatusID   string     `gorm:"column:inventory_status_id;type:uuid;not null"`
	UOMID               string     `gorm:"column:uom_id;type:uuid;not null"`
	SystemQty           string     `gorm:"column:system_qty;type:numeric(20,6);not null"`
	CountedQty          *string    `gorm:"column:counted_qty;type:numeric(20,6)"`
	VarianceQty         *string    `gorm:"column:variance_qty;type:numeric(20,6)"`
	CountAttempts       int        `gorm:"column:count_attempts;not null;default:0"`
	RequiresRecount     bool       `gorm:"column:requires_recount;not null;default:false"`
	DecisionCode        string     `gorm:"column:decision_code;size:20;not null;default:OPEN"`
	CountedAt           *time.Time `gorm:"column:counted_at"`
	CountedBy           *string    `gorm:"column:counted_by;type:uuid"`
	CountNotes          *string    `gorm:"column:count_notes;type:text"`
	DecidedAt           *time.Time `gorm:"column:decided_at"`
	DecidedBy           *string    `gorm:"column:decided_by;type:uuid"`
	DecisionReason      *string    `gorm:"column:decision_reason;type:text"`
	InventoryMovementID *string    `gorm:"column:inventory_movement_id;size:140"`
	ResultingBalanceID  *string    `gorm:"column:resulting_balance_id;size:160"`
	CreatedAt           time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy           string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy           *string    `gorm:"column:updated_by;type:uuid"`
	VersionNo           int64      `gorm:"column:version_no;not null;default:1"`
}

func (CycleCountLine) TableName() string { return "cycle_count_line" }

type CycleCountEntry struct {
	ID                string    `gorm:"column:cycle_count_entry_id;size:200;primaryKey"`
	CycleCountLineID  string    `gorm:"column:cycle_count_line_id;size:180;not null"`
	AttemptNo         int       `gorm:"column:attempt_no;not null"`
	SystemQty         string    `gorm:"column:system_qty;type:numeric(20,6);not null"`
	SnapshotVersionNo int64     `gorm:"column:snapshot_version_no;not null"`
	CountedQty        string    `gorm:"column:counted_qty;type:numeric(20,6);not null"`
	VarianceQty       string    `gorm:"column:variance_qty;type:numeric(20,6);not null"`
	Notes             *string   `gorm:"column:notes;type:text"`
	CountedAt         time.Time `gorm:"column:counted_at;not null;default:clock_timestamp()"`
	CountedBy         string    `gorm:"column:counted_by;type:uuid;not null"`
}

func (CycleCountEntry) TableName() string { return "cycle_count_entry" }
