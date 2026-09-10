package billing

import "time"

type Invoice struct {
	ID             string    `gorm:"column:invoice_id;size:120;primaryKey"`
	BillingRunID   string    `gorm:"column:billing_run_id;size:120;not null"`
	DocumentTypeID string    `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID       string    `gorm:"column:status_id;type:uuid;not null"`
	OwnerID        string    `gorm:"column:owner_id;type:uuid;not null"`
	WarehouseID    string    `gorm:"column:warehouse_id;type:uuid;not null"`
	IssueDate      time.Time `gorm:"column:issue_date;type:date;not null"`
	DueDate        time.Time `gorm:"column:due_date;type:date;not null"`
	CurrencyCode   string    `gorm:"column:currency_code;size:3;not null"`
	SubtotalAmount string    `gorm:"column:subtotal_amount;type:numeric(20,6);not null"`
	TaxAmount      string    `gorm:"column:tax_amount;type:numeric(20,6);not null"`
	CreditAmount   string    `gorm:"column:credit_amount;type:numeric(20,6);not null;default:0"`
	TotalAmount    string    `gorm:"column:total_amount;type:numeric(20,6);not null"`
	PaidAmount     string    `gorm:"column:paid_amount;type:numeric(20,6);not null;default:0"`
	Notes          *string   `gorm:"column:notes;type:text"`
	VersionNo      int64     `gorm:"column:version_no;not null;default:1"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy      string    `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy      *string   `gorm:"column:updated_by;type:uuid"`
}

func (Invoice) TableName() string { return "invoice" }
