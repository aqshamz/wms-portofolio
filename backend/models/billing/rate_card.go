package billing

import "time"

type RateCard struct {
	ID                string     `gorm:"column:rate_card_id;size:120;primaryKey"`
	BillingContractID string     `gorm:"column:billing_contract_id;size:120;not null"`
	DocumentTypeID    string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID          string     `gorm:"column:status_id;type:uuid;not null"`
	Name              string     `gorm:"column:name;size:150;not null"`
	EffectiveFrom     time.Time  `gorm:"column:effective_from;type:date;not null"`
	EffectiveUntil    *time.Time `gorm:"column:effective_until;type:date"`
	VersionNo         int64      `gorm:"column:version_no;not null;default:1"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy         string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy         *string    `gorm:"column:updated_by;type:uuid"`
}

func (RateCard) TableName() string { return "rate_card" }
