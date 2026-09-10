package billing

import "time"

type BillingCharge struct {
	ID               string    `gorm:"column:billing_charge_id;size:140;primaryKey"`
	BillingRunID     string    `gorm:"column:billing_run_id;size:120;not null"`
	BillableEventID  string    `gorm:"column:billable_event_id;size:140;not null"`
	ServiceCode      string    `gorm:"column:service_code;size:40;not null"`
	Description      string    `gorm:"column:description;size:200;not null"`
	SourceDocumentID string    `gorm:"column:source_document_id;size:140;not null"`
	Quantity         string    `gorm:"column:quantity;type:numeric(20,6);not null"`
	UnitRate         string    `gorm:"column:unit_rate;type:numeric(20,6);not null"`
	SubtotalAmount   string    `gorm:"column:subtotal_amount;type:numeric(20,6);not null"`
	TaxPercent       string    `gorm:"column:tax_percent;type:numeric(7,4);not null"`
	TaxAmount        string    `gorm:"column:tax_amount;type:numeric(20,6);not null"`
	TotalAmount      string    `gorm:"column:total_amount;type:numeric(20,6);not null"`
	CreatedAt        time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy        string    `gorm:"column:created_by;type:uuid;not null"`
}

func (BillingCharge) TableName() string { return "billing_charge" }
