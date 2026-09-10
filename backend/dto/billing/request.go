package billing

type CreateContractRequest struct {
	OwnerID         string  `json:"owner_id" binding:"required,uuid"`
	WarehouseID     string  `json:"warehouse_id" binding:"required,uuid"`
	BusinessDate    string  `json:"business_date" binding:"required"`
	CurrencyCode    string  `json:"currency_code" binding:"required,len=3"`
	BillingCycle    string  `json:"billing_cycle" binding:"required,max=20"`
	PaymentTermDays int     `json:"payment_term_days" binding:"min=0,max=365"`
	EffectiveFrom   string  `json:"effective_from" binding:"required"`
	EffectiveUntil  *string `json:"effective_until"`
	Notes           *string `json:"notes" binding:"omitempty,max=4000"`
}

type UpdateContractRequest struct {
	ExpectedVersion int64   `json:"expected_version" binding:"required,min=1"`
	CurrencyCode    string  `json:"currency_code" binding:"required,len=3"`
	BillingCycle    string  `json:"billing_cycle" binding:"required,max=20"`
	PaymentTermDays int     `json:"payment_term_days" binding:"min=0,max=365"`
	EffectiveFrom   string  `json:"effective_from" binding:"required"`
	EffectiveUntil  *string `json:"effective_until"`
	Notes           *string `json:"notes" binding:"omitempty,max=4000"`
}

type TransitionRequest struct {
	ExpectedVersion int64 `json:"expected_version" binding:"required,min=1"`
}

type CreateRateCardLineRequest struct {
	ServiceCode    string  `json:"service_code" binding:"required,max=40"`
	Description    string  `json:"description" binding:"required,max=200"`
	SourceKind     string  `json:"source_kind" binding:"required,max=20"`
	MovementTypeID *string `json:"movement_type_id" binding:"omitempty,uuid"`
	BillingBasis   string  `json:"billing_basis" binding:"required,max=20"`
	UnitRate       string  `json:"unit_rate" binding:"required,max=30"`
	MinimumCharge  string  `json:"minimum_charge" binding:"omitempty,max=30"`
	TaxPercent     string  `json:"tax_percent" binding:"omitempty,max=20"`
}

type CreateRateCardRequest struct {
	BillingContractID string                      `json:"billing_contract_id" binding:"required,max=120"`
	BusinessDate      string                      `json:"business_date" binding:"required"`
	Name              string                      `json:"name" binding:"required,max=150"`
	EffectiveFrom     string                      `json:"effective_from" binding:"required"`
	EffectiveUntil    *string                     `json:"effective_until"`
	Lines             []CreateRateCardLineRequest `json:"lines" binding:"required,min=1,max=200,dive"`
}

type UpdateRateCardRequest struct {
	ExpectedVersion int64   `json:"expected_version" binding:"required,min=1"`
	Name            string  `json:"name" binding:"required,max=150"`
	EffectiveFrom   string  `json:"effective_from" binding:"required"`
	EffectiveUntil  *string `json:"effective_until"`
}

type AddRateCardLineRequest struct {
	ExpectedVersion int64 `json:"expected_version" binding:"required,min=1"`
	CreateRateCardLineRequest
}

type UpdateRateCardLineRequest struct {
	ExpectedVersion int64 `json:"expected_version" binding:"required,min=1"`
	CreateRateCardLineRequest
	IsActive bool `json:"is_active"`
}

type DeleteRateCardLineRequest struct {
	ExpectedVersion int64 `json:"expected_version" binding:"required,min=1"`
}

type CollectEventsRequest struct {
	OwnerID     string `json:"owner_id" binding:"required,uuid"`
	WarehouseID string `json:"warehouse_id" binding:"required,uuid"`
	DateFrom    string `json:"date_from" binding:"required"`
	DateUntil   string `json:"date_until" binding:"required"`
}

type StorageSnapshotRequest struct {
	OwnerID      string `json:"owner_id" binding:"required,uuid"`
	WarehouseID  string `json:"warehouse_id" binding:"required,uuid"`
	BusinessDate string `json:"business_date" binding:"required"`
}

type ExcludeEventRequest struct {
	Reason string `json:"reason" binding:"required,max=4000"`
}

type CreateBillingRunRequest struct {
	BillingContractID string  `json:"billing_contract_id" binding:"required,max=120"`
	BusinessDate      string  `json:"business_date" binding:"required"`
	PeriodFrom        string  `json:"period_from" binding:"required"`
	PeriodUntil       string  `json:"period_until" binding:"required"`
	Notes             *string `json:"notes" binding:"omitempty,max=4000"`
}

type CreateManualEventRequest struct {
	RateCardLineID   string  `json:"rate_card_line_id" binding:"required,uuid"`
	BusinessDate     string  `json:"business_date" binding:"required"`
	SourceDocumentID string  `json:"source_document_id" binding:"required,max=140"`
	Quantity         string  `json:"quantity" binding:"required,max=30"`
	UOMID            *string `json:"uom_id" binding:"omitempty,uuid"`
	Notes            *string `json:"notes" binding:"omitempty,max=4000"`
}

type CreateInvoiceRequest struct {
	ExpectedVersion int64   `json:"expected_version" binding:"required,min=1"`
	IssueDate       string  `json:"issue_date" binding:"required"`
	DueDate         *string `json:"due_date"`
	Notes           *string `json:"notes" binding:"omitempty,max=4000"`
}

type UpdateInvoiceRequest struct {
	ExpectedVersion int64   `json:"expected_version" binding:"required,min=1"`
	IssueDate       string  `json:"issue_date" binding:"required"`
	DueDate         string  `json:"due_date" binding:"required"`
	Notes           *string `json:"notes" binding:"omitempty,max=4000"`
}

type CreditNoteRequest struct {
	ExpectedVersion int64  `json:"expected_version" binding:"required,min=1"`
	BusinessDate    string `json:"business_date" binding:"required"`
	Amount          string `json:"amount" binding:"required,max=30"`
	Reason          string `json:"reason" binding:"required,max=4000"`
}

type PaymentRequest struct {
	ExpectedVersion int64   `json:"expected_version" binding:"required,min=1"`
	BusinessDate    string  `json:"business_date" binding:"required"`
	Amount          string  `json:"amount" binding:"required,max=30"`
	Reference       string  `json:"reference" binding:"required,max=120"`
	Notes           *string `json:"notes" binding:"omitempty,max=4000"`
}
