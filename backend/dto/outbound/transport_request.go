package outbound

type CreateCarrierRequest struct {
	BusinessPartnerID *string `json:"business_partner_id" binding:"omitempty,uuid"`
	Code              string  `json:"code" binding:"required,max=40"`
	Name              string  `json:"name" binding:"required,max=150"`
}

type UpdateCarrierRequest struct {
	BusinessPartnerID *string `json:"business_partner_id" binding:"omitempty,uuid"`
	Code              string  `json:"code" binding:"required,max=40"`
	Name              string  `json:"name" binding:"required,max=150"`
	IsActive          *bool   `json:"is_active"`
}

type CreateCarrierServiceRequest struct {
	Code string `json:"code" binding:"required,max=40"`
	Name string `json:"name" binding:"required,max=100"`
}

type UpdateCarrierServiceRequest struct {
	Code     string `json:"code" binding:"required,max=40"`
	Name     string `json:"name" binding:"required,max=100"`
	IsActive *bool  `json:"is_active"`
}

type CreateCarrierDriverRequest struct {
	AccountID     *string `json:"account_id" binding:"omitempty,uuid"`
	Code          string  `json:"code" binding:"required,max=40"`
	Name          string  `json:"name" binding:"required,max=150"`
	PhoneNumber   *string `json:"phone_number" binding:"omitempty,max=50"`
	LicenseNumber *string `json:"license_number" binding:"omitempty,max=80"`
}

type UpdateCarrierDriverRequest struct {
	AccountID     *string `json:"account_id" binding:"omitempty,uuid"`
	Code          string  `json:"code" binding:"required,max=40"`
	Name          string  `json:"name" binding:"required,max=150"`
	PhoneNumber   *string `json:"phone_number" binding:"omitempty,max=50"`
	LicenseNumber *string `json:"license_number" binding:"omitempty,max=80"`
	IsActive      *bool   `json:"is_active"`
}

type AssignShipmentDriverRequest struct {
	DriverID  string `json:"driver_id" binding:"required,uuid"`
	IsPrimary bool   `json:"is_primary"`
}
