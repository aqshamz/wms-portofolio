package outbound

import "time"

type OutboundCheckLine struct {
	ID                    string     `gorm:"column:outbound_check_line_id;size:160;primaryKey"`
	OutboundCheckID       string     `gorm:"column:outbound_check_id;size:120;not null"`
	StagingLineID         string     `gorm:"column:staging_line_id;size:160;not null"`
	LineNo                int        `gorm:"column:line_no;not null"`
	ExpectedQty           string     `gorm:"column:expected_qty;type:numeric(20,6);not null"`
	CheckedQty            *string    `gorm:"column:checked_qty;type:numeric(20,6)"`
	ExceptionQty          string     `gorm:"column:exception_qty;type:numeric(20,6);not null;default:0"`
	UOMID                 string     `gorm:"column:uom_id;type:uuid;not null"`
	OutboundCheckResultID *string    `gorm:"column:outbound_check_result_id;type:uuid"`
	Notes                 *string    `gorm:"column:notes;type:text"`
	CheckedAt             *time.Time `gorm:"column:checked_at"`
	CheckedBy             *string    `gorm:"column:checked_by;type:uuid"`
	CreatedAt             time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy             string     `gorm:"column:created_by;type:uuid;not null"`
}

func (OutboundCheckLine) TableName() string { return "outbound_check_line" }
