package master

// PartnerTypeResponse is the public JSON contract, separate from the GORM entity.
type PartnerTypeResponse struct {
	ID          string  `json:"partner_type_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IsActive    bool    `json:"is_active"`
}
