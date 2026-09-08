package inbound

import "time"

type PutawayTask struct {
	ID                      string     `gorm:"column:putaway_task_id;size:120;primaryKey"`
	TaskTypeID              string     `gorm:"column:task_type_id;type:uuid;not null"`
	TaskStatusID            string     `gorm:"column:task_status_id;type:uuid;not null"`
	TaskPriorityID          string     `gorm:"column:task_priority_id;type:uuid;not null"`
	ReceiptInventoryID      string     `gorm:"column:receipt_inventory_id;size:160;not null"`
	InspectionID            string     `gorm:"column:inspection_id;size:120;not null"`
	SourceBalanceID         string     `gorm:"column:source_balance_id;size:160;not null"`
	OwnerID                 string     `gorm:"column:owner_id;type:uuid;not null"`
	WarehouseID             string     `gorm:"column:warehouse_id;type:uuid;not null"`
	ItemID                  string     `gorm:"column:item_id;type:uuid;not null"`
	LotID                   *string    `gorm:"column:lot_id;size:120"`
	HandlingUnitID          *string    `gorm:"column:handling_unit_id;size:120"`
	SourceLocationID        string     `gorm:"column:source_location_id;type:uuid;not null"`
	TargetLocationID        string     `gorm:"column:target_location_id;type:uuid;not null"`
	PlannedQty              string     `gorm:"column:planned_qty;type:numeric(20,6);not null"`
	CompletedQty            string     `gorm:"column:completed_qty;type:numeric(20,6);not null;default:0"`
	UOMID                   string     `gorm:"column:uom_id;type:uuid;not null"`
	AssignedTo              *string    `gorm:"column:assigned_to;type:uuid"`
	StartedAt               *time.Time `gorm:"column:started_at"`
	CompletedAt             *time.Time `gorm:"column:completed_at"`
	InventoryMovementID     *string    `gorm:"column:inventory_movement_id;size:140"`
	ResultingBalanceID      *string    `gorm:"column:resulting_balance_id;size:160"`
	ReversalMovementID      *string    `gorm:"column:reversal_movement_id;size:140"`
	ReversedAt              *time.Time `gorm:"column:reversed_at"`
	ReversedBy              *string    `gorm:"column:reversed_by;type:uuid"`
	ReversalReason          *string    `gorm:"column:reversal_reason;type:text"`
	ReplacementInspectionID *string    `gorm:"column:replacement_inspection_id;size:120"`
	CreatedAt               time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy               string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt               time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy               *string    `gorm:"column:updated_by;type:uuid"`
	VersionNo               int64      `gorm:"column:version_no;not null;default:1"`
}

func (PutawayTask) TableName() string { return "putaway_task" }
