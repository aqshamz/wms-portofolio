package inventory

import "time"

type InventoryMovement struct {
	ID                   string    `gorm:"column:movement_id;size:140;primaryKey"`
	MovementTypeID       string    `gorm:"column:movement_type_id;type:uuid;not null"`
	OwnerID              string    `gorm:"column:owner_id;type:uuid;not null"`
	WarehouseID          string    `gorm:"column:warehouse_id;type:uuid;not null"`
	BusinessDate         time.Time `gorm:"column:business_date;type:date;not null"`
	OccurredAt           time.Time `gorm:"column:occurred_at;not null;default:clock_timestamp()"`
	ItemID               string    `gorm:"column:item_id;type:uuid;not null"`
	LotID                *string   `gorm:"column:lot_id;size:120"`
	SerialID             *string   `gorm:"column:serial_id;size:160"`
	HandlingUnitID       *string   `gorm:"column:handling_unit_id;size:120"`
	FromLocationID       *string   `gorm:"column:from_location_id;type:uuid"`
	ToLocationID         *string   `gorm:"column:to_location_id;type:uuid"`
	FromStatusID         *string   `gorm:"column:from_status_id;type:uuid"`
	ToStatusID           *string   `gorm:"column:to_status_id;type:uuid"`
	Quantity             string    `gorm:"column:quantity;type:numeric(20,6);not null"`
	UOMID                string    `gorm:"column:uom_id;type:uuid;not null"`
	SourceDocumentID     string    `gorm:"column:source_document_id;size:140;not null"`
	SourceLineID         *string   `gorm:"column:source_line_id;size:160"`
	ReasonCodeID         *string   `gorm:"column:reason_code_id;type:uuid"`
	Notes                *string   `gorm:"column:notes;type:text"`
	OperationKey         *string   `gorm:"column:operation_key;size:160"`
	OperationFingerprint *string   `gorm:"column:operation_fingerprint;size:64"`
	CreatedBy            string    `gorm:"column:created_by;type:uuid;not null"`
}

func (InventoryMovement) TableName() string { return "inventory_movement" }
