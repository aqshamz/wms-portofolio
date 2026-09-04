package master

type PartnerDetailResponse struct {
	BusinessPartnerResponse
	PartnerTypes []PartnerTypeResponse `json:"partner_types"`
}

type ItemDetailResponse struct {
	ItemResponse
	UOMs     []ItemUOMResponse     `json:"uoms"`
	Barcodes []ItemBarcodeResponse `json:"barcodes"`
}

type AssignPartnerTypeRequest struct {
	PartnerTypeID string `json:"partner_type_id" binding:"required,uuid"`
}
