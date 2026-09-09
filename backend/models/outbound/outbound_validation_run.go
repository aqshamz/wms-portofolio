package outbound

import "time"

type OutboundValidationRun struct {
	ID             string     `gorm:"column:validation_run_id;size:140;primaryKey"`
	OutboundID     string     `gorm:"column:outbound_id;size:120;not null"`
	DocumentTypeID string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID       string     `gorm:"column:status_id;type:uuid;not null"`
	ValidatedAt    *time.Time `gorm:"column:validated_at"`
	ValidatedBy    *string    `gorm:"column:validated_by;type:uuid"`
	Notes          *string    `gorm:"column:notes;type:text"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy      string     `gorm:"column:created_by;type:uuid;not null"`
}

func (OutboundValidationRun) TableName() string { return "outbound_validation_run" }
