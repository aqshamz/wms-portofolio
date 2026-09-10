package billing

import "time"

type Payment struct {
	ID             string    `gorm:"column:payment_id;size:120;primaryKey"`
	InvoiceID      string    `gorm:"column:invoice_id;size:120;not null"`
	DocumentTypeID string    `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID       string    `gorm:"column:status_id;type:uuid;not null"`
	BusinessDate   time.Time `gorm:"column:business_date;type:date;not null"`
	Amount         string    `gorm:"column:amount;type:numeric(20,6);not null"`
	Reference      string    `gorm:"column:reference;size:120;not null"`
	Notes          *string   `gorm:"column:notes;type:text"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy      string    `gorm:"column:created_by;type:uuid;not null"`
}

func (Payment) TableName() string { return "payment" }
