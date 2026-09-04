package master

import "time"

type CreateBusinessPartnerRequest struct {
	OwnerID      string  `json:"owner_id" binding:"required,uuid"`
	Code         string  `json:"code" binding:"required,max=40"`
	Name         string  `json:"name" binding:"required,max=150"`
	LegalName    *string `json:"legal_name" binding:"omitempty,max=200"`
	TaxNumber    *string `json:"tax_number" binding:"omitempty,max=100"`
	Email        *string `json:"email" binding:"omitempty,max=254,email"`
	Phone        *string `json:"phone" binding:"omitempty,max=50"`
	AddressLine1 *string `json:"address_line_1" binding:"omitempty,max=255"`
	AddressLine2 *string `json:"address_line_2" binding:"omitempty,max=255"`
	City         *string `json:"city" binding:"omitempty,max=100"`
	Province     *string `json:"province" binding:"omitempty,max=100"`
	PostalCode   *string `json:"postal_code" binding:"omitempty,max=20"`
	CountryCode  *string `json:"country_code" binding:"omitempty,max=2,len=2"`
}

type UpdateBusinessPartnerRequest struct {
	Name              string    `json:"name" binding:"required,max=150"`
	LegalName         *string   `json:"legal_name" binding:"omitempty,max=200"`
	TaxNumber         *string   `json:"tax_number" binding:"omitempty,max=100"`
	Email             *string   `json:"email" binding:"omitempty,max=254,email"`
	Phone             *string   `json:"phone" binding:"omitempty,max=50"`
	AddressLine1      *string   `json:"address_line_1" binding:"omitempty,max=255"`
	AddressLine2      *string   `json:"address_line_2" binding:"omitempty,max=255"`
	City              *string   `json:"city" binding:"omitempty,max=100"`
	Province          *string   `json:"province" binding:"omitempty,max=100"`
	PostalCode        *string   `json:"postal_code" binding:"omitempty,max=20"`
	CountryCode       *string   `json:"country_code" binding:"omitempty,max=2,len=2"`
	IsActive          *bool     `json:"is_active" binding:"required"`
	ExpectedUpdatedAt time.Time `json:"expected_updated_at" binding:"required"`
}
