package master

// BusinessPartnerTypeResponse is the public JSON contract, separate from the GORM entity.
type BusinessPartnerTypeResponse struct {
	PartnerID     string `json:"partner_id"`
	PartnerTypeID string `json:"partner_type_id"`
}
