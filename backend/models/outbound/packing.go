package outbound

import "time"

type Packing struct {
	ID                string     `gorm:"column:packing_id;size:120;primaryKey"`
	DocumentTypeID    string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID          string     `gorm:"column:status_id;type:uuid;not null"`
	OutboundID        string     `gorm:"column:outbound_id;size:120;not null;index"`
	WarehouseID       string     `gorm:"column:warehouse_id;type:uuid;not null"`
	PackingLocationID *string    `gorm:"column:packing_location_id;type:uuid"`
	PackedAt          *time.Time `gorm:"column:packed_at"`
	PackedBy          *string    `gorm:"column:packed_by;type:uuid"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy         string     `gorm:"column:created_by;type:uuid;not null"`
}

func (Packing) TableName() string { return "packing" }
