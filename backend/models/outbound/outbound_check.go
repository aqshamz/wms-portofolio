package outbound

import "time"

type OutboundCheck struct {
	ID             string     `gorm:"column:outbound_check_id;size:120;primaryKey"`
	ParentCheckID  *string    `gorm:"column:parent_check_id;size:120"`
	DocumentTypeID string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID       string     `gorm:"column:status_id;type:uuid;not null"`
	StagingID      string     `gorm:"column:staging_id;size:120;not null"`
	CheckedAt      *time.Time `gorm:"column:checked_at"`
	CheckedBy      *string    `gorm:"column:checked_by;type:uuid"`
	Notes          *string    `gorm:"column:notes;type:text"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy      string     `gorm:"column:created_by;type:uuid;not null"`
}

func (OutboundCheck) TableName() string { return "outbound_check" }
