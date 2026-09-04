package master

import "time"

// BusinessPartnerResponse is the public JSON contract, separate from the GORM entity.
type BusinessPartnerResponse struct {
	ID           string    `json:"partner_id"`
	OwnerID      string    `json:"owner_id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	LegalName    *string   `json:"legal_name"`
	TaxNumber    *string   `json:"tax_number"`
	Email        *string   `json:"email"`
	Phone        *string   `json:"phone"`
	AddressLine1 *string   `json:"address_line_1"`
	AddressLine2 *string   `json:"address_line_2"`
	City         *string   `json:"city"`
	Province     *string   `json:"province"`
	PostalCode   *string   `json:"postal_code"`
	CountryCode  *string   `json:"country_code"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    *string   `json:"created_by"`
	UpdatedAt    time.Time `json:"updated_at"`
	UpdatedBy    *string   `json:"updated_by"`
}
