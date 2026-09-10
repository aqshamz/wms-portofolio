package billing

import "time"

type PageResponse[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int64 `json:"total_pages"`
}

type ContractResponse struct {
	ID              string    `json:"billing_contract_id"`
	OwnerID         string    `json:"owner_id"`
	OwnerCode       string    `json:"owner_code"`
	WarehouseID     string    `json:"warehouse_id"`
	WarehouseCode   string    `json:"warehouse_code"`
	CurrencyCode    string    `json:"currency_code"`
	BillingCycle    string    `json:"billing_cycle"`
	StatusCode      string    `json:"status_code"`
	PaymentTermDays int       `json:"payment_term_days"`
	EffectiveFrom   string    `json:"effective_from"`
	EffectiveUntil  *string   `json:"effective_until"`
	Notes           *string   `json:"notes"`
	VersionNo       int64     `json:"version_no"`
	CreatedAt       time.Time `json:"created_at"`
}

type RateCardLineResponse struct {
	ID               string  `json:"rate_card_line_id"`
	ServiceCode      string  `json:"service_code"`
	Description      string  `json:"description"`
	SourceKind       string  `json:"source_kind"`
	MovementTypeID   *string `json:"movement_type_id"`
	MovementTypeCode *string `json:"movement_type_code"`
	BillingBasis     string  `json:"billing_basis"`
	UnitRate         string  `json:"unit_rate"`
	MinimumCharge    string  `json:"minimum_charge"`
	TaxPercent       string  `json:"tax_percent"`
	IsActive         bool    `json:"is_active"`
}
type RateCardResponse struct {
	ID                string                 `json:"rate_card_id"`
	BillingContractID string                 `json:"billing_contract_id"`
	OwnerID           string                 `json:"owner_id"`
	OwnerCode         string                 `json:"owner_code"`
	WarehouseID       string                 `json:"warehouse_id"`
	WarehouseCode     string                 `json:"warehouse_code"`
	Name              string                 `json:"name"`
	CurrencyCode      string                 `json:"currency_code"`
	StatusCode        string                 `json:"status_code"`
	EffectiveFrom     string                 `json:"effective_from"`
	EffectiveUntil    *string                `json:"effective_until"`
	VersionNo         int64                  `json:"version_no"`
	CreatedAt         time.Time              `json:"created_at"`
	Lines             []RateCardLineResponse `json:"lines,omitempty"`
}
type EventResponse struct {
	ID                  string    `json:"billable_event_id"`
	OwnerID             string    `json:"owner_id"`
	WarehouseID         string    `json:"warehouse_id"`
	RateCardLineID      string    `json:"rate_card_line_id"`
	ServiceCode         string    `json:"service_code"`
	MovementTypeCode    *string   `json:"movement_type_code"`
	InventoryMovementID *string   `json:"inventory_movement_id"`
	EventKey            string    `json:"event_key"`
	BusinessDate        string    `json:"business_date"`
	SourceDocumentID    string    `json:"source_document_id"`
	SourceLineID        *string   `json:"source_line_id"`
	Quantity            string    `json:"quantity"`
	UOMID               *string   `json:"uom_id"`
	UOMCode             *string   `json:"uom_code"`
	StatusCode          string    `json:"status_code"`
	Notes               *string   `json:"notes"`
	CreatedAt           time.Time `json:"created_at"`
}
type CollectEventsResponse struct {
	Created  int `json:"created"`
	Existing int `json:"existing"`
}
type ChargeResponse struct {
	ID               string `json:"billing_charge_id"`
	BillableEventID  string `json:"billable_event_id"`
	ServiceCode      string `json:"service_code"`
	Description      string `json:"description"`
	SourceDocumentID string `json:"source_document_id"`
	Quantity         string `json:"quantity"`
	UnitRate         string `json:"unit_rate"`
	SubtotalAmount   string `json:"subtotal_amount"`
	TaxPercent       string `json:"tax_percent"`
	TaxAmount        string `json:"tax_amount"`
	TotalAmount      string `json:"total_amount"`
}
type BillingRunResponse struct {
	ID                string           `json:"billing_run_id"`
	BillingContractID string           `json:"billing_contract_id"`
	OwnerID           string           `json:"owner_id"`
	OwnerCode         string           `json:"owner_code"`
	WarehouseID       string           `json:"warehouse_id"`
	WarehouseCode     string           `json:"warehouse_code"`
	PeriodFrom        string           `json:"period_from"`
	PeriodUntil       string           `json:"period_until"`
	BusinessDate      string           `json:"business_date"`
	CurrencyCode      string           `json:"currency_code"`
	StatusCode        string           `json:"status_code"`
	SubtotalAmount    string           `json:"subtotal_amount"`
	TaxAmount         string           `json:"tax_amount"`
	TotalAmount       string           `json:"total_amount"`
	Notes             *string          `json:"notes"`
	VersionNo         int64            `json:"version_no"`
	CreatedAt         time.Time        `json:"created_at"`
	Charges           []ChargeResponse `json:"charges,omitempty"`
}
type InvoiceLineResponse struct {
	ID              string `json:"invoice_line_id"`
	LineNo          int    `json:"line_no"`
	BillingChargeID string `json:"billing_charge_id"`
	ServiceCode     string `json:"service_code"`
	Description     string `json:"description"`
	Quantity        string `json:"quantity"`
	UnitRate        string `json:"unit_rate"`
	SubtotalAmount  string `json:"subtotal_amount"`
	TaxAmount       string `json:"tax_amount"`
	TotalAmount     string `json:"total_amount"`
}
type InvoiceResponse struct {
	ID                string                `json:"invoice_id"`
	BillingRunID      string                `json:"billing_run_id"`
	OwnerID           string                `json:"owner_id"`
	OwnerCode         string                `json:"owner_code"`
	WarehouseID       string                `json:"warehouse_id"`
	WarehouseCode     string                `json:"warehouse_code"`
	IssueDate         string                `json:"issue_date"`
	DueDate           string                `json:"due_date"`
	CurrencyCode      string                `json:"currency_code"`
	StatusCode        string                `json:"status_code"`
	SubtotalAmount    string                `json:"subtotal_amount"`
	TaxAmount         string                `json:"tax_amount"`
	CreditAmount      string                `json:"credit_amount"`
	TotalAmount       string                `json:"total_amount"`
	PaidAmount        string                `json:"paid_amount"`
	OutstandingAmount string                `json:"outstanding_amount"`
	Notes             *string               `json:"notes"`
	VersionNo         int64                 `json:"version_no"`
	CreatedAt         time.Time             `json:"created_at"`
	Lines             []InvoiceLineResponse `json:"lines,omitempty"`
}
type CreditNoteResponse struct {
	ID           string     `json:"credit_note_id"`
	InvoiceID    string     `json:"invoice_id"`
	BusinessDate string     `json:"business_date"`
	Amount       string     `json:"amount"`
	Reason       string     `json:"reason"`
	StatusCode   string     `json:"status_code"`
	CreatedAt    time.Time  `json:"created_at"`
	IssuedAt     *time.Time `json:"issued_at"`
}
type PaymentResponse struct {
	ID           string    `json:"payment_id"`
	InvoiceID    string    `json:"invoice_id"`
	BusinessDate string    `json:"business_date"`
	Amount       string    `json:"amount"`
	Reference    string    `json:"reference"`
	Notes        *string   `json:"notes"`
	StatusCode   string    `json:"status_code"`
	CreatedAt    time.Time `json:"created_at"`
}
