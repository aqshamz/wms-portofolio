package master

import "time"

type CreateOrganizationRequest struct {
	Code         string  `json:"code" binding:"required,max=40"`
	Name         string  `json:"name" binding:"required,max=150"`
	LegalName    *string `json:"legal_name" binding:"omitempty,max=200"`
	TaxNumber    *string `json:"tax_number" binding:"omitempty,max=100"`
	TimezoneName string  `json:"timezone_name" binding:"required,max=50"`
	AddressLine1 *string `json:"address_line_1" binding:"omitempty,max=255"`
	AddressLine2 *string `json:"address_line_2" binding:"omitempty,max=255"`
	City         *string `json:"city" binding:"omitempty,max=100"`
	Province     *string `json:"province" binding:"omitempty,max=100"`
	PostalCode   *string `json:"postal_code" binding:"omitempty,max=20"`
	CountryCode  *string `json:"country_code" binding:"omitempty,len=2"`
}

type UpdateOrganizationRequest struct {
	Name              string    `json:"name" binding:"required,max=150"`
	LegalName         *string   `json:"legal_name" binding:"omitempty,max=200"`
	TaxNumber         *string   `json:"tax_number" binding:"omitempty,max=100"`
	TimezoneName      string    `json:"timezone_name" binding:"required,max=50"`
	AddressLine1      *string   `json:"address_line_1" binding:"omitempty,max=255"`
	AddressLine2      *string   `json:"address_line_2" binding:"omitempty,max=255"`
	City              *string   `json:"city" binding:"omitempty,max=100"`
	Province          *string   `json:"province" binding:"omitempty,max=100"`
	PostalCode        *string   `json:"postal_code" binding:"omitempty,max=20"`
	CountryCode       *string   `json:"country_code" binding:"omitempty,len=2"`
	IsActive          bool      `json:"is_active"`
	ExpectedUpdatedAt time.Time `json:"expected_updated_at" binding:"required"`
}

type DeactivateRequest struct {
	ExpectedUpdatedAt time.Time `json:"expected_updated_at" binding:"required"`
}
