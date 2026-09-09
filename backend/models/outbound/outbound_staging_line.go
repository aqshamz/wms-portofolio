package outbound

import "time"

type OutboundStagingLine struct {
	ID               string    `gorm:"column:staging_line_id;size:160;primaryKey"`
	StagingID        string    `gorm:"column:staging_id;size:120;not null"`
	PickExecutionID  string    `gorm:"column:pick_execution_id;size:170;not null;uniqueIndex"`
	StagingBalanceID string    `gorm:"column:staging_balance_id;size:160;not null"`
	StagedQty        string    `gorm:"column:staged_qty;type:numeric(20,6);not null"`
	RemovedQty       string    `gorm:"column:removed_qty;type:numeric(20,6);not null;default:0"`
	UOMID            string    `gorm:"column:uom_id;type:uuid;not null"`
	CreatedAt        time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy        string    `gorm:"column:created_by;type:uuid;not null"`
}

func (OutboundStagingLine) TableName() string { return "outbound_staging_line" }
