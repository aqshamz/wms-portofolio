package master

import "time"

type DocumentNumberRuleResponse struct {
	ID                   string    `json:"document_number_rule_id"`
	DocumentTypeID       string    `json:"document_type_id"`
	Prefix               string    `json:"prefix"`
	Separator            string    `json:"separator"`
	SequenceLength       int16     `json:"sequence_length"`
	IncludePartnerCode   bool      `json:"include_partner_code"`
	IncludeWarehouseCode bool      `json:"include_warehouse_code"`
	IsActive             bool      `json:"is_active"`
	EffectiveFrom        string    `json:"effective_from"`
	EffectiveUntil       *string   `json:"effective_until"`
	CreatedAt            time.Time `json:"created_at"`
	CreatedBy            *string   `json:"created_by"`
}
