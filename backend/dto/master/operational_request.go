package master

type OperationalListRequest struct {
	OwnerID     string
	WarehouseID string
	ModuleCode  string
	Search      string
	Active      *bool
	Page        int
	PageSize    int
}
type ReplaceDocumentNumberRuleRequest struct {
	Prefix               string  `json:"prefix" binding:"required,max=20"`
	Separator            *string `json:"separator" binding:"required,max=3"`
	SequenceLength       int16   `json:"sequence_length" binding:"min=3,max=18"`
	IncludePartnerCode   *bool   `json:"include_partner_code" binding:"required"`
	IncludeWarehouseCode *bool   `json:"include_warehouse_code" binding:"required"`
	EffectiveFrom        string  `json:"effective_from" binding:"required"`
}
type GenerateDocumentIDRequest struct {
	BusinessDate string  `json:"business_date" binding:"required"`
	PartnerID    *string `json:"partner_id" binding:"omitempty,uuid"`
	WarehouseID  *string `json:"warehouse_id" binding:"omitempty,uuid"`
}
type GeneratedDocumentIDResponse struct {
	DocumentID           string `json:"document_id"`
	DocumentTypeID       string `json:"document_type_id"`
	DocumentNumberRuleID string `json:"document_number_rule_id"`
	BusinessDate         string `json:"business_date"`
	SequenceNumber       string `json:"sequence_number"`
}
