package outbound

import "time"

type OutboundCheckResolution struct {
	ID                            string     `gorm:"column:outbound_check_resolution_id;size:190;primaryKey"`
	OutboundCheckExceptionID      string     `gorm:"column:outbound_check_exception_id;size:170;not null"`
	OutboundCheckResolutionTypeID string     `gorm:"column:outbound_check_resolution_type_id;type:uuid;not null"`
	ResolvedQty                   string     `gorm:"column:resolved_qty;type:numeric(20,6);not null"`
	Notes                         *string    `gorm:"column:notes;type:text"`
	ApprovedAt                    *time.Time `gorm:"column:approved_at"`
	ApprovedBy                    *string    `gorm:"column:approved_by;type:uuid"`
	CreatedAt                     time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy                     string     `gorm:"column:created_by;type:uuid;not null"`
}

func (OutboundCheckResolution) TableName() string { return "outbound_check_resolution" }
