package outbound

import "time"

type OutboundCheckException struct {
	ID                  string     `gorm:"column:outbound_check_exception_id;size:170;primaryKey"`
	OutboundCheckLineID string     `gorm:"column:outbound_check_line_id;size:160;not null;uniqueIndex"`
	StatusID            string     `gorm:"column:status_id;type:uuid;not null"`
	ExceptionQty        string     `gorm:"column:exception_qty;type:numeric(20,6);not null"`
	OpenedAt            time.Time  `gorm:"column:opened_at;not null;default:clock_timestamp()"`
	ResolvedAt          *time.Time `gorm:"column:resolved_at"`
	ResolvedBy          *string    `gorm:"column:resolved_by;type:uuid"`
	Notes               *string    `gorm:"column:notes;type:text"`
	CreatedBy           string     `gorm:"column:created_by;type:uuid;not null"`
}

func (OutboundCheckException) TableName() string { return "outbound_check_exception" }
