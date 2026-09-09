package outbound

import "time"

type PickExecution struct {
	ID               string    `gorm:"column:pick_execution_id;size:170;primaryKey"`
	PickTaskID       string    `gorm:"column:pick_task_id;size:140;not null"`
	SourceBalanceID  string    `gorm:"column:source_balance_id;size:160;not null"`
	StagingBalanceID string    `gorm:"column:staging_balance_id;size:160;not null"`
	PickedQty        string    `gorm:"column:picked_qty;type:numeric(20,6);not null"`
	UOMID            string    `gorm:"column:uom_id;type:uuid;not null"`
	MovementID       string    `gorm:"column:movement_id;size:140;not null;uniqueIndex"`
	PickedAt         time.Time `gorm:"column:picked_at;not null;default:clock_timestamp()"`
	PickedBy         string    `gorm:"column:picked_by;type:uuid;not null"`
}

func (PickExecution) TableName() string { return "pick_execution" }
