package outbound

import "time"

type OutboundWave struct {
	ID                string     `gorm:"column:wave_id;size:120;primaryKey"`
	DocumentTypeID    string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID          string     `gorm:"column:status_id;type:uuid;not null"`
	WaveTypeID        string     `gorm:"column:outbound_wave_type_id;type:uuid;not null"`
	OwnerID           string     `gorm:"column:owner_id;type:uuid;not null"`
	WarehouseID       string     `gorm:"column:warehouse_id;type:uuid;not null"`
	PickingStrategyID *string    `gorm:"column:picking_strategy_id;type:uuid"`
	BusinessDate      time.Time  `gorm:"column:business_date;type:date;not null"`
	PlannedReleaseAt  *time.Time `gorm:"column:planned_release_at"`
	ReleasedAt        *time.Time `gorm:"column:released_at"`
	CompletedAt       *time.Time `gorm:"column:completed_at"`
	Notes             *string    `gorm:"column:notes;type:text"`
	VersionNo         int64      `gorm:"column:version_no;not null;default:1"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy         string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy         string     `gorm:"column:updated_by;type:uuid;not null"`
}

func (OutboundWave) TableName() string { return "outbound_wave" }
