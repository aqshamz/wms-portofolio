package inbound

import "time"

// InboundException is the immutable audit record for quantity variances,
// cancellations, and reversals. It deliberately does not replace the source
// document's workflow status.
type InboundException struct {
	ID                string    `gorm:"column:inbound_exception_id;size:140;primaryKey"`
	OwnerID           string    `gorm:"column:owner_id;type:uuid;not null"`
	WarehouseID       string    `gorm:"column:warehouse_id;type:uuid;not null"`
	SourceDocumentID  string    `gorm:"column:source_document_id;size:140;not null"`
	SourceLineID      *string   `gorm:"column:source_line_id;size:160"`
	ExceptionTypeCode string    `gorm:"column:exception_type_code;size:40;not null"`
	ExpectedQty       *string   `gorm:"column:expected_qty;type:numeric(20,6)"`
	ActualQty         *string   `gorm:"column:actual_qty;type:numeric(20,6)"`
	VarianceQty       *string   `gorm:"column:variance_qty;type:numeric(20,6)"`
	Notes             *string   `gorm:"column:notes;type:text"`
	CreatedAt         time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy         string    `gorm:"column:created_by;type:uuid;not null"`
}

func (InboundException) TableName() string { return "inbound_exception" }
