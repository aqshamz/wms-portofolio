package stockcontrol

import "time"

// ReplenishmentTask reserves available reserve stock until it is moved into a
// forward pick face or the task is cancelled.
type ReplenishmentTask struct {
	ID                  string     `gorm:"column:replenishment_task_id;size:120;primaryKey"`
	TaskTypeID          string     `gorm:"column:task_type_id;type:uuid;not null"`
	TaskStatusID        string     `gorm:"column:task_status_id;type:uuid;not null"`
	TaskPriorityID      string     `gorm:"column:task_priority_id;type:uuid;not null"`
	SourceBalanceID     string     `gorm:"column:source_balance_id;size:160;not null"`
	OwnerID             string     `gorm:"column:owner_id;type:uuid;not null"`
	WarehouseID         string     `gorm:"column:warehouse_id;type:uuid;not null"`
	ItemID              string     `gorm:"column:item_id;type:uuid;not null"`
	LotID               *string    `gorm:"column:lot_id;size:120"`
	SerialID            *string    `gorm:"column:serial_id;size:160"`
	HandlingUnitID      *string    `gorm:"column:handling_unit_id;size:120"`
	InventoryStatusID   string     `gorm:"column:inventory_status_id;type:uuid;not null"`
	SourceLocationID    string     `gorm:"column:source_location_id;type:uuid;not null"`
	TargetLocationID    string     `gorm:"column:target_location_id;type:uuid;not null"`
	PlannedQty          string     `gorm:"column:planned_qty;type:numeric(20,6);not null"`
	CompletedQty        string     `gorm:"column:completed_qty;type:numeric(20,6);not null;default:0"`
	UOMID               string     `gorm:"column:uom_id;type:uuid;not null"`
	AssignedTo          *string    `gorm:"column:assigned_to;type:uuid"`
	Notes               *string    `gorm:"column:notes;type:text"`
	StartedAt           *time.Time `gorm:"column:started_at"`
	CompletedAt         *time.Time `gorm:"column:completed_at"`
	CancelledAt         *time.Time `gorm:"column:cancelled_at"`
	CancelledBy         *string    `gorm:"column:cancelled_by;type:uuid"`
	CancellationReason  *string    `gorm:"column:cancellation_reason;type:text"`
	InventoryMovementID *string    `gorm:"column:inventory_movement_id;size:140"`
	ResultingBalanceID  *string    `gorm:"column:resulting_balance_id;size:160"`
	CreatedAt           time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy           string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy           *string    `gorm:"column:updated_by;type:uuid"`
	VersionNo           int64      `gorm:"column:version_no;not null;default:1"`
}

func (ReplenishmentTask) TableName() string { return "replenishment_task" }
