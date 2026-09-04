package master

import "time"

type OrganizationResponse struct {
	ID           string    `json:"organization_id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	LegalName    *string   `json:"legal_name,omitempty"`
	TaxNumber    *string   `json:"tax_number,omitempty"`
	TimezoneName string    `json:"timezone_name"`
	AddressLine1 *string   `json:"address_line_1,omitempty"`
	AddressLine2 *string   `json:"address_line_2,omitempty"`
	City         *string   `json:"city,omitempty"`
	Province     *string   `json:"province,omitempty"`
	PostalCode   *string   `json:"postal_code,omitempty"`
	CountryCode  *string   `json:"country_code,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type WarehouseResponse struct {
	ID           string    `json:"warehouse_id"`
	OperatorID   string    `json:"operator_id"`
	OperatorCode string    `json:"operator_code,omitempty"`
	OperatorName string    `json:"operator_name,omitempty"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	TimezoneName string    `json:"timezone_name"`
	AddressLine1 *string   `json:"address_line_1,omitempty"`
	AddressLine2 *string   `json:"address_line_2,omitempty"`
	City         *string   `json:"city,omitempty"`
	Province     *string   `json:"province,omitempty"`
	PostalCode   *string   `json:"postal_code,omitempty"`
	CountryCode  *string   `json:"country_code,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PageResponse[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}
