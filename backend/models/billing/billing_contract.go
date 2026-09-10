package billing

import "time"

type BillingContract struct {
	ID              string     `gorm:"column:billing_contract_id;size:120;primaryKey"`
	DocumentTypeID  string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID        string     `gorm:"column:status_id;type:uuid;not null"`
	OwnerID         string     `gorm:"column:owner_id;type:uuid;not null"`
	WarehouseID     string     `gorm:"column:warehouse_id;type:uuid;not null"`
	CurrencyCode    string     `gorm:"column:currency_code;size:3;not null"`
	BillingCycle    string     `gorm:"column:billing_cycle;size:20;not null"`
	PaymentTermDays int        `gorm:"column:payment_term_days;not null"`
	EffectiveFrom   time.Time  `gorm:"column:effective_from;type:date;not null"`
	EffectiveUntil  *time.Time `gorm:"column:effective_until;type:date"`
	Notes           *string    `gorm:"column:notes;type:text"`
	VersionNo       int64      `gorm:"column:version_no;not null;default:1"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy       string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy       *string    `gorm:"column:updated_by;type:uuid"`
}

func (BillingContract) TableName() string { return "billing_contract" }
