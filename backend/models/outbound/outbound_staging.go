package outbound

import "time"

type OutboundStaging struct {
	ID                string     `gorm:"column:staging_id;size:120;primaryKey"`
	DocumentTypeID    string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID          string     `gorm:"column:status_id;type:uuid;not null"`
	OutboundID        string     `gorm:"column:outbound_id;size:120;not null"`
	WaveID            string     `gorm:"column:wave_id;size:120;not null"`
	WarehouseID       string     `gorm:"column:warehouse_id;type:uuid;not null"`
	StagingLocationID string     `gorm:"column:staging_location_id;type:uuid;not null"`
	StagedAt          *time.Time `gorm:"column:staged_at"`
	StagedBy          *string    `gorm:"column:staged_by;type:uuid"`
	Notes             *string    `gorm:"column:notes;type:text"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy         string     `gorm:"column:created_by;type:uuid;not null"`
}

func (OutboundStaging) TableName() string { return "outbound_staging" }
