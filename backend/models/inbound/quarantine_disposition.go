package inbound

import "time"

type QuarantineDisposition struct {
	ID                          string     `gorm:"column:quarantine_disposition_id;size:150;primaryKey"`
	QuarantineCaseID            string     `gorm:"column:quarantine_case_id;size:140;not null"`
	DocumentTypeID              string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID                    string     `gorm:"column:status_id;type:uuid;not null"`
	QuarantineDispositionTypeID string     `gorm:"column:quarantine_disposition_type_id;type:uuid;not null"`
	DispositionQty              string     `gorm:"column:disposition_qty;type:numeric(20,6);not null"`
	UOMID                       string     `gorm:"column:uom_id;type:uuid;not null"`
	ClientDecisionReference     *string    `gorm:"column:client_decision_reference;size:120"`
	DecisionNotes               *string    `gorm:"column:decision_notes;type:text"`
	DecidedAt                   time.Time  `gorm:"column:decided_at;not null"`
	DecidedBy                   string     `gorm:"column:decided_by;type:uuid;not null"`
	ProcessedAt                 *time.Time `gorm:"column:processed_at"`
	InventoryMovementID         *string    `gorm:"column:inventory_movement_id;size:140"`
	ResultingBalanceID          *string    `gorm:"column:resulting_balance_id;size:160"`
	TargetLocationID            *string    `gorm:"column:target_location_id;type:uuid"`
	CreatedAt                   time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy                   string     `gorm:"column:created_by;type:uuid;not null"`
}

func (QuarantineDisposition) TableName() string { return "quarantine_disposition" }
