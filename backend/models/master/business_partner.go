package master

import "time"

type BusinessPartner struct {
	ID           string    `gorm:"column:partner_id;type:uuid;default:gen_random_uuid();primaryKey"`
	OwnerID      string    `gorm:"column:owner_id;type:uuid;not null"`
	Code         string    `gorm:"column:code;size:40;not null"`
	Name         string    `gorm:"column:name;size:150;not null"`
	LegalName    *string   `gorm:"column:legal_name;size:200"`
	TaxNumber    *string   `gorm:"column:tax_number;size:100"`
	Email        *string   `gorm:"column:email;size:254"`
	Phone        *string   `gorm:"column:phone;size:50"`
	AddressLine1 *string   `gorm:"column:address_line_1;size:255"`
	AddressLine2 *string   `gorm:"column:address_line_2;size:255"`
	City         *string   `gorm:"column:city;size:100"`
	Province     *string   `gorm:"column:province;size:100"`
	PostalCode   *string   `gorm:"column:postal_code;size:20"`
	CountryCode  *string   `gorm:"column:country_code;size:2"`
	IsActive     bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy    *string   `gorm:"column:created_by;type:uuid"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy    *string   `gorm:"column:updated_by;type:uuid"`
}

func (BusinessPartner) TableName() string { return "business_partner" }
