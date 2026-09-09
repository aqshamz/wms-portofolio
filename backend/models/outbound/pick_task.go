package outbound

import "time"

type PickTask struct {
	ID                string     `gorm:"column:pick_task_id;size:140;primaryKey"`
	TaskTypeID        string     `gorm:"column:task_type_id;type:uuid;not null"`
	TaskStatusID      string     `gorm:"column:task_status_id;type:uuid;not null"`
	TaskPriorityID    string     `gorm:"column:task_priority_id;type:uuid;not null"`
	WaveID            string     `gorm:"column:wave_id;size:120;not null"`
	ReservationID     string     `gorm:"column:reservation_id;size:140;not null;uniqueIndex"`
	OutboundLineID    string     `gorm:"column:outbound_line_id;size:150;not null"`
	SourceLocationID  string     `gorm:"column:source_location_id;type:uuid;not null"`
	TargetLocationID  *string    `gorm:"column:target_location_id;type:uuid"`
	PlannedQty        string     `gorm:"column:planned_qty;type:numeric(20,6);not null"`
	PickedQty         string     `gorm:"column:picked_qty;type:numeric(20,6);not null;default:0"`
	UOMID             string     `gorm:"column:uom_id;type:uuid;not null"`
	AssignedTo        *string    `gorm:"column:assigned_to;type:uuid"`
	StartedAt         *time.Time `gorm:"column:started_at"`
	CompletedAt       *time.Time `gorm:"column:completed_at"`
	ShortQty          string     `gorm:"column:short_qty;type:numeric(20,6);not null;default:0"`
	ShortReasonCodeID *string    `gorm:"column:short_reason_code_id;type:uuid"`
	ResultNotes       *string    `gorm:"column:result_notes;type:text"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy         string     `gorm:"column:created_by;type:uuid;not null"`
}

func (PickTask) TableName() string { return "pick_task" }
