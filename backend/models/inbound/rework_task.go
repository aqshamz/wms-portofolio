package inbound

import "time"

type ReworkTask struct {
	ID                      string     `gorm:"column:rework_task_id;size:140;primaryKey"`
	QuarantineDispositionID string     `gorm:"column:quarantine_disposition_id;size:150;not null;uniqueIndex"`
	TaskTypeID              string     `gorm:"column:task_type_id;type:uuid;not null"`
	TaskStatusID            string     `gorm:"column:task_status_id;type:uuid;not null"`
	TaskPriorityID          string     `gorm:"column:task_priority_id;type:uuid;not null"`
	SourceBalanceID         string     `gorm:"column:source_balance_id;size:160;not null"`
	PlannedQty              string     `gorm:"column:planned_qty;type:numeric(20,6);not null"`
	CompletedQty            string     `gorm:"column:completed_qty;type:numeric(20,6);not null;default:0"`
	UOMID                   string     `gorm:"column:uom_id;type:uuid;not null"`
	AssignedTo              *string    `gorm:"column:assigned_to;type:uuid"`
	WorkInstructions        *string    `gorm:"column:work_instructions;type:text"`
	ResultNotes             *string    `gorm:"column:result_notes;type:text"`
	StartedAt               *time.Time `gorm:"column:started_at"`
	CompletedAt             *time.Time `gorm:"column:completed_at"`
	ReinspectionID          *string    `gorm:"column:reinspection_id;size:120;uniqueIndex"`
	CreatedAt               time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy               string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt               time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy               *string    `gorm:"column:updated_by;type:uuid"`
	VersionNo               int64      `gorm:"column:version_no;not null;default:1"`
}

func (ReworkTask) TableName() string { return "rework_task" }
