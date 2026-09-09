package outbound

import "time"

type PackingLine struct {
	ID                  string    `gorm:"column:packing_line_id;size:150;primaryKey"`
	PackingID           string    `gorm:"column:packing_id;size:120;not null"`
	PickTaskID          string    `gorm:"column:pick_task_id;size:140;not null"`
	OutboundCheckLineID string    `gorm:"column:outbound_check_line_id;size:160;not null;uniqueIndex"`
	SourceBalanceID     string    `gorm:"column:source_balance_id;size:160;not null"`
	PackingBalanceID    string    `gorm:"column:packing_balance_id;size:160;not null"`
	HandlingUnitID      *string   `gorm:"column:handling_unit_id;size:120"`
	PackedQty           string    `gorm:"column:packed_qty;type:numeric(20,6);not null"`
	UOMID               string    `gorm:"column:uom_id;type:uuid;not null"`
	MovementID          string    `gorm:"column:movement_id;size:140;not null;uniqueIndex"`
	CreatedAt           time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy           string    `gorm:"column:created_by;type:uuid;not null"`
}

func (PackingLine) TableName() string { return "packing_line" }
