package master

import "time"

type DocumentNumberRule struct {
	ID                   string     `gorm:"column:document_number_rule_id;type:uuid;default:gen_random_uuid();primaryKey"`
	DocumentTypeID       string     `gorm:"column:document_type_id;type:uuid;not null"`
	Prefix               string     `gorm:"column:prefix;size:20;not null"`
	Separator            string     `gorm:"column:separator;size:3;not null;default:'-'"`
	SequenceLength       int16      `gorm:"column:sequence_length;not null;default:6"`
	IncludePartnerCode   bool       `gorm:"column:include_partner_code;not null;default:true"`
	IncludeWarehouseCode bool       `gorm:"column:include_warehouse_code;not null;default:true"`
	IsActive             bool       `gorm:"column:is_active;not null;default:true"`
	EffectiveFrom        time.Time  `gorm:"column:effective_from;type:date;not null;default:current_date"`
	EffectiveUntil       *time.Time `gorm:"column:effective_until;type:date"`
	CreatedAt            time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy            *string    `gorm:"column:created_by;type:uuid"`
}

func (DocumentNumberRule) TableName() string { return "document_number_rule" }
