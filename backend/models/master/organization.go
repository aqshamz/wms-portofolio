package master

import "time"

type Organization struct {
	ID           string    `gorm:"column:organization_id;type:uuid;default:gen_random_uuid();primaryKey" json:"organization_id"`
	Code         string    `gorm:"column:code;size:40;not null;uniqueIndex" json:"code"`
	Name         string    `gorm:"column:name;size:150;not null" json:"name"`
	LegalName    *string   `gorm:"column:legal_name;size:200" json:"legal_name,omitempty"`
	TaxNumber    *string   `gorm:"column:tax_number;size:100" json:"tax_number,omitempty"`
	TimezoneName string    `gorm:"column:timezone_name;size:50;not null" json:"timezone_name"`
	AddressLine1 *string   `gorm:"column:address_line_1;size:255" json:"address_line_1,omitempty"`
	AddressLine2 *string   `gorm:"column:address_line_2;size:255" json:"address_line_2,omitempty"`
	City         *string   `gorm:"column:city;size:100" json:"city,omitempty"`
	Province     *string   `gorm:"column:province;size:100" json:"province,omitempty"`
	PostalCode   *string   `gorm:"column:postal_code;size:20" json:"postal_code,omitempty"`
	CountryCode  *string   `gorm:"column:country_code;size:2" json:"country_code,omitempty"`
	IsActive     bool      `gorm:"column:is_active;not null;default:true" json:"is_active"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:clock_timestamp()" json:"created_at"`
	CreatedBy    *string   `gorm:"column:created_by;type:uuid" json:"created_by,omitempty"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:clock_timestamp()" json:"updated_at"`
	UpdatedBy    *string   `gorm:"column:updated_by;type:uuid" json:"updated_by,omitempty"`
}

func (Organization) TableName() string { return "organization" }
