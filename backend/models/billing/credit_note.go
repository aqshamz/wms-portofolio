package billing

import "time"

type CreditNote struct {
	ID             string     `gorm:"column:credit_note_id;size:120;primaryKey"`
	InvoiceID      string     `gorm:"column:invoice_id;size:120;not null"`
	DocumentTypeID string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID       string     `gorm:"column:status_id;type:uuid;not null"`
	BusinessDate   time.Time  `gorm:"column:business_date;type:date;not null"`
	Amount         string     `gorm:"column:amount;type:numeric(20,6);not null"`
	Reason         string     `gorm:"column:reason;type:text;not null"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy      string     `gorm:"column:created_by;type:uuid;not null"`
	IssuedAt       *time.Time `gorm:"column:issued_at"`
	IssuedBy       *string    `gorm:"column:issued_by;type:uuid"`
}

func (CreditNote) TableName() string { return "credit_note" }
