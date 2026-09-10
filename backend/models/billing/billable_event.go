package billing

import "time"

type BillableEvent struct {
	ID                  string    `gorm:"column:billable_event_id;size:140;primaryKey"`
	DocumentTypeID      string    `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID            string    `gorm:"column:status_id;type:uuid;not null"`
	OwnerID             string    `gorm:"column:owner_id;type:uuid;not null"`
	WarehouseID         string    `gorm:"column:warehouse_id;type:uuid;not null"`
	RateCardLineID      string    `gorm:"column:rate_card_line_id;type:uuid;not null"`
	InventoryMovementID *string   `gorm:"column:inventory_movement_id;size:140"`
	EventKey            string    `gorm:"column:event_key;size:190;not null"`
	BusinessDate        time.Time `gorm:"column:business_date;type:date;not null"`
	SourceDocumentID    string    `gorm:"column:source_document_id;size:140;not null"`
	SourceLineID        *string   `gorm:"column:source_line_id;size:160"`
	Quantity            string    `gorm:"column:quantity;type:numeric(20,6);not null"`
	UOMID               *string   `gorm:"column:uom_id;type:uuid"`
	Notes               *string   `gorm:"column:notes;type:text"`
	CreatedAt           time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy           string    `gorm:"column:created_by;type:uuid;not null"`
}

func (BillableEvent) TableName() string { return "billable_event" }
