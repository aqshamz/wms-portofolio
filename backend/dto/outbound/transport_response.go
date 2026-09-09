package outbound

import "time"

type CarrierResponse struct {
	ID                string  `json:"carrier_id"`
	BusinessPartnerID *string `json:"business_partner_id"`
	Code              string  `json:"code"`
	Name              string  `json:"name"`
	IsActive          bool    `json:"is_active"`
}

type CarrierServiceResponse struct {
	ID          string `json:"carrier_service_id"`
	CarrierID   string `json:"carrier_id"`
	CarrierCode string `json:"carrier_code"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	IsActive    bool   `json:"is_active"`
}

type CarrierDriverResponse struct {
	ID            string    `json:"driver_id"`
	CarrierID     string    `json:"carrier_id"`
	CarrierCode   string    `json:"carrier_code"`
	AccountID     *string   `json:"account_id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	PhoneNumber   *string   `json:"phone_number"`
	LicenseNumber *string   `json:"license_number"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

type ShipmentDriverResponse struct {
	DriverID      string    `json:"driver_id"`
	DriverCode    string    `json:"driver_code"`
	DriverName    string    `json:"driver_name"`
	PhoneNumber   string    `json:"phone_number,omitempty"`
	LicenseNumber string    `json:"license_number,omitempty"`
	IsPrimary     bool      `json:"is_primary"`
	AssignedAt    time.Time `json:"assigned_at"`
	AssignedBy    string    `json:"assigned_by"`
}
