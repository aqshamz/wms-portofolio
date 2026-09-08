package inbound

import "time"

type Receipt struct {
	ID             string    `gorm:"column:receipt_id;size:120;primaryKey"`
	DocumentTypeID string    `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID       string    `gorm:"column:status_id;type:uuid;not null"`
	InboundID      *string   `gorm:"column:inbound_id;size:120"`
	OwnerID        string    `gorm:"column:owner_id;type:uuid;not null"`
	WarehouseID    string    `gorm:"column:warehouse_id;type:uuid;not null"`
	BusinessDate   time.Time `gorm:"column:business_date;type:date;not null"`
	ReceivedAt     time.Time `gorm:"column:received_at;not null"`
	DockLocationID *string   `gorm:"column:dock_location_id;type:uuid"`
	VehicleNumber  *string   `gorm:"column:vehicle_number;size:60"`
	SealNumber     *string   `gorm:"column:seal_number;size:60"`
	DeliveryNoteNo *string   `gorm:"column:delivery_note_no;size:100"`
	Notes          *string   `gorm:"column:notes;type:text"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy      string    `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy      *string   `gorm:"column:updated_by;type:uuid"`
	VersionNo      int64     `gorm:"column:version_no;not null;default:1"`
}

func (Receipt) TableName() string { return "receipt" }
