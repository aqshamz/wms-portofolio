package billing

import "time"

type RateCardLine struct {
	ID             string    `gorm:"column:rate_card_line_id;type:uuid;default:gen_random_uuid();primaryKey"`
	RateCardID     string    `gorm:"column:rate_card_id;size:120;not null"`
	ServiceCode    string    `gorm:"column:service_code;size:40;not null"`
	Description    string    `gorm:"column:description;size:200;not null"`
	SourceKind     string    `gorm:"column:source_kind;size:20;not null"`
	MovementTypeID *string   `gorm:"column:movement_type_id;type:uuid"`
	BillingBasis   string    `gorm:"column:billing_basis;size:20;not null"`
	UnitRate       string    `gorm:"column:unit_rate;type:numeric(20,6);not null"`
	MinimumCharge  string    `gorm:"column:minimum_charge;type:numeric(20,6);not null;default:0"`
	TaxPercent     string    `gorm:"column:tax_percent;type:numeric(7,4);not null;default:0"`
	IsActive       bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy      string    `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy      *string   `gorm:"column:updated_by;type:uuid"`
}

func (RateCardLine) TableName() string { return "rate_card_line" }
