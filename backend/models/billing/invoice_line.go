package billing

import "time"

type InvoiceLine struct {
	ID              string    `gorm:"column:invoice_line_id;type:uuid;default:gen_random_uuid();primaryKey"`
	InvoiceID       string    `gorm:"column:invoice_id;size:120;not null"`
	BillingChargeID string    `gorm:"column:billing_charge_id;size:140;not null"`
	LineNo          int       `gorm:"column:line_no;not null"`
	ServiceCode     string    `gorm:"column:service_code;size:40;not null"`
	Description     string    `gorm:"column:description;size:200;not null"`
	Quantity        string    `gorm:"column:quantity;type:numeric(20,6);not null"`
	UnitRate        string    `gorm:"column:unit_rate;type:numeric(20,6);not null"`
	SubtotalAmount  string    `gorm:"column:subtotal_amount;type:numeric(20,6);not null"`
	TaxAmount       string    `gorm:"column:tax_amount;type:numeric(20,6);not null"`
	TotalAmount     string    `gorm:"column:total_amount;type:numeric(20,6);not null"`
	CreatedAt       time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
}

func (InvoiceLine) TableName() string { return "invoice_line" }
