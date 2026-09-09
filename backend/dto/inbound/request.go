package inbound

type PurchaseOrderLineRequest struct {
	ItemID                   string  `json:"item_id" binding:"required,uuid"`
	OrderedQty               string  `json:"ordered_qty" binding:"required,max=30"`
	UOMID                    string  `json:"uom_id" binding:"required,uuid"`
	VendorItemCode           *string `json:"vendor_item_code" binding:"omitempty,max=100"`
	ExpectedLotNo            *string `json:"expected_lot_no" binding:"omitempty,max=100"`
	ExpectedExpiryDate       *string `json:"expected_expiry_date"`
	Notes                    *string `json:"notes" binding:"omitempty,max=4000"`
	OverReceiptTolerancePct  *string `json:"over_receipt_tolerance_pct" binding:"omitempty,max=12"`
	UnderReceiptTolerancePct *string `json:"under_receipt_tolerance_pct" binding:"omitempty,max=12"`
}

type CreatePurchaseOrderRequest struct {
	OwnerID                   string                     `json:"owner_id" binding:"required,uuid"`
	VendorID                  string                     `json:"vendor_id" binding:"required,uuid"`
	WarehouseID               string                     `json:"warehouse_id" binding:"required,uuid"`
	BusinessDate              string                     `json:"business_date" binding:"required"`
	PurchaseOrderNo           string                     `json:"purchase_order_no" binding:"required,max=120"`
	OrderedAt                 string                     `json:"ordered_at" binding:"required"`
	ExpectedArrivalAt         *string                    `json:"expected_arrival_at"`
	Notes                     *string                    `json:"notes" binding:"omitempty,max=4000"`
	SupersedesPurchaseOrderID *string                    `json:"supersedes_purchase_order_id" binding:"omitempty,max=120"`
	Lines                     []PurchaseOrderLineRequest `json:"lines" binding:"required,min=1,max=500,dive"`
}

type UpdatePurchaseOrderRequest struct {
	ExpectedVersion   int64   `json:"expected_version" binding:"required,min=1"`
	PurchaseOrderNo   string  `json:"purchase_order_no" binding:"required,max=120"`
	OrderedAt         string  `json:"ordered_at" binding:"required"`
	ExpectedArrivalAt *string `json:"expected_arrival_at"`
	Notes             *string `json:"notes" binding:"omitempty,max=4000"`
}

type AddPurchaseOrderLineRequest struct {
	ExpectedVersion int64 `json:"expected_version" binding:"required,min=1"`
	PurchaseOrderLineRequest
}

type UpdatePurchaseOrderLineRequest struct {
	ExpectedVersion          int64   `json:"expected_version" binding:"required,min=1"`
	OrderedQty               string  `json:"ordered_qty" binding:"required,max=30"`
	VendorItemCode           *string `json:"vendor_item_code" binding:"omitempty,max=100"`
	ExpectedLotNo            *string `json:"expected_lot_no" binding:"omitempty,max=100"`
	ExpectedExpiryDate       *string `json:"expected_expiry_date"`
	Notes                    *string `json:"notes" binding:"omitempty,max=4000"`
	OverReceiptTolerancePct  *string `json:"over_receipt_tolerance_pct" binding:"omitempty,max=12"`
	UnderReceiptTolerancePct *string `json:"under_receipt_tolerance_pct" binding:"omitempty,max=12"`
}

type DeleteLineRequest struct {
	ExpectedVersion int64 `json:"expected_version" binding:"required,min=1"`
}

type TransitionRequest struct {
	ExpectedVersion int64 `json:"expected_version" binding:"required,min=1"`
}

type ExceptionTransitionRequest struct {
	ExpectedVersion int64  `json:"expected_version" binding:"required,min=1"`
	Reason          string `json:"reason" binding:"required,max=4000"`
}

type InboundOrderLineRequest struct {
	PurchaseOrderLineID   string  `json:"purchase_order_line_id" binding:"required,max=150"`
	ExpectedQty           string  `json:"expected_qty" binding:"required,max=30"`
	CustomerLineReference *string `json:"customer_line_reference" binding:"omitempty,max=100"`
	Notes                 *string `json:"notes" binding:"omitempty,max=4000"`
}

type CreateInboundOrderRequest struct {
	PurchaseOrderID     string                    `json:"purchase_order_id" binding:"required,max=120"`
	BusinessDate        string                    `json:"business_date" binding:"required"`
	ExpectedArrivalAt   *string                   `json:"expected_arrival_at"`
	ExternalReference   *string                   `json:"external_reference" binding:"omitempty,max=120"`
	SupplierReference   *string                   `json:"supplier_reference" binding:"omitempty,max=120"`
	Notes               *string                   `json:"notes" binding:"omitempty,max=4000"`
	SupersedesInboundID *string                   `json:"supersedes_inbound_id" binding:"omitempty,max=120"`
	Lines               []InboundOrderLineRequest `json:"lines" binding:"required,min=1,max=500,dive"`
}

type UpdateInboundOrderRequest struct {
	ExpectedVersion   int64   `json:"expected_version" binding:"required,min=1"`
	ExpectedArrivalAt *string `json:"expected_arrival_at"`
	ExternalReference *string `json:"external_reference" binding:"omitempty,max=120"`
	SupplierReference *string `json:"supplier_reference" binding:"omitempty,max=120"`
	Notes             *string `json:"notes" binding:"omitempty,max=4000"`
}

type AddInboundOrderLineRequest struct {
	ExpectedVersion int64 `json:"expected_version" binding:"required,min=1"`
	InboundOrderLineRequest
}

type UpdateInboundOrderLineRequest struct {
	ExpectedVersion       int64   `json:"expected_version" binding:"required,min=1"`
	ExpectedQty           string  `json:"expected_qty" binding:"required,max=30"`
	CustomerLineReference *string `json:"customer_line_reference" binding:"omitempty,max=100"`
	Notes                 *string `json:"notes" binding:"omitempty,max=4000"`
}

type ReceiptLotRequest struct {
	LotNumber       string  `json:"lot_number" binding:"required,max=100"`
	ManufactureDate *string `json:"manufacture_date"`
	ExpiryDate      *string `json:"expiry_date"`
}

type ReceiptBatchRequest struct {
	SourceQty          string             `json:"source_qty" binding:"required,max=30"`
	ReceivedLocationID string             `json:"received_location_id" binding:"required,uuid"`
	Lot                *ReceiptLotRequest `json:"lot"`
	HandlingUnitID     *string            `json:"handling_unit_id" binding:"omitempty,max=120"`
	SerialNo           *string            `json:"serial_no" binding:"omitempty,max=120"`
}

type ReceiptLineRequest struct {
	InboundLineID     string                `json:"inbound_line_id" binding:"required,max=150"`
	ReceivedQty       string                `json:"received_qty" binding:"required,max=30"`
	RejectedQty       string                `json:"rejected_qty" binding:"required,max=30"`
	Batches           []ReceiptBatchRequest `json:"batches" binding:"max=1000,dive"`
	ExceptionNotes    *string               `json:"exception_notes" binding:"omitempty,max=4000"`
	ExceptionTypeCode *string               `json:"exception_type_code" binding:"omitempty,max=40"`
}

type CreateReceiptRequest struct {
	InboundID           string               `json:"inbound_id" binding:"required,max=120"`
	BusinessDate        string               `json:"business_date" binding:"required"`
	ReceivedAt          string               `json:"received_at" binding:"required"`
	DockLocationID      string               `json:"dock_location_id" binding:"required,uuid"`
	VehicleNumber       *string              `json:"vehicle_number" binding:"omitempty,max=60"`
	SealNumber          *string              `json:"seal_number" binding:"omitempty,max=60"`
	DeliveryNoteNo      *string              `json:"delivery_note_no" binding:"omitempty,max=100"`
	Notes               *string              `json:"notes" binding:"omitempty,max=4000"`
	SupersedesReceiptID *string              `json:"supersedes_receipt_id" binding:"omitempty,max=120"`
	Lines               []ReceiptLineRequest `json:"lines" binding:"required,min=1,max=500,dive"`
}

// UpdateReceiptRequest replaces the editable header and complete line/batch draft.
// Business date and inbound linkage stay immutable because they drive numbering
// and document ownership.
type UpdateReceiptRequest struct {
	ExpectedVersion int64                `json:"expected_version" binding:"required,min=1"`
	ReceivedAt      string               `json:"received_at" binding:"required"`
	DockLocationID  string               `json:"dock_location_id" binding:"required,uuid"`
	VehicleNumber   *string              `json:"vehicle_number" binding:"omitempty,max=60"`
	SealNumber      *string              `json:"seal_number" binding:"omitempty,max=60"`
	DeliveryNoteNo  *string              `json:"delivery_note_no" binding:"omitempty,max=100"`
	Notes           *string              `json:"notes" binding:"omitempty,max=4000"`
	Lines           []ReceiptLineRequest `json:"lines" binding:"required,min=1,max=500,dive"`
}

type CreateQualityInspectionRequest struct {
	ReceiptInventoryID string  `json:"receipt_inventory_id" binding:"required,max=160"`
	Notes              *string `json:"notes" binding:"omitempty,max=4000"`
}

type CompleteQualityInspectionRequest struct {
	ExpectedVersion         int64   `json:"expected_version" binding:"required,min=1"`
	ExpectedBalanceVersion  int64   `json:"expected_balance_version" binding:"required,min=1"`
	PassedQty               string  `json:"passed_qty" binding:"required,max=30"`
	FailedQty               string  `json:"failed_qty" binding:"required,max=30"`
	PutawayTargetLocationID *string `json:"putaway_target_location_id" binding:"omitempty,uuid"`
	Notes                   *string `json:"notes" binding:"omitempty,max=4000"`
}

type PutawayTransitionRequest struct {
	ExpectedVersion int64 `json:"expected_version" binding:"required,min=1"`
}

type CompletePutawayRequest struct {
	ExpectedVersion        int64  `json:"expected_version" binding:"required,min=1"`
	ExpectedBalanceVersion int64  `json:"expected_balance_version" binding:"required,min=1"`
	BusinessDate           string `json:"business_date" binding:"required"`
}

type AssignPutawayRequest struct {
	ExpectedVersion int64  `json:"expected_version" binding:"required,min=1"`
	AccountID       string `json:"account_id" binding:"required,uuid"`
}

type RetargetPutawayRequest struct {
	ExpectedVersion  int64  `json:"expected_version" binding:"required,min=1"`
	TargetLocationID string `json:"target_location_id" binding:"required,uuid"`
}

type CancelPutawayRequest struct {
	ExpectedVersion        int64  `json:"expected_version" binding:"required,min=1"`
	ExpectedBalanceVersion int64  `json:"expected_balance_version" binding:"required,min=1"`
	BusinessDate           string `json:"business_date" binding:"required"`
	Reason                 string `json:"reason" binding:"required,max=4000"`
}

type BalanceVersionRequest struct {
	BalanceID       string `json:"balance_id" binding:"required,max=160"`
	ExpectedVersion int64  `json:"expected_version" binding:"required,min=1"`
}

type ReverseReceiptRequest struct {
	ExpectedVersion int64                   `json:"expected_version" binding:"required,min=1"`
	BusinessDate    string                  `json:"business_date" binding:"required"`
	Reason          string                  `json:"reason" binding:"required,max=4000"`
	Balances        []BalanceVersionRequest `json:"balances" binding:"max=1000,dive"`
}

type CreateQuarantineDispositionRequest struct {
	ExpectedCaseVersion     int64   `json:"expected_case_version" binding:"required,min=1"`
	ExpectedBalanceVersion  int64   `json:"expected_balance_version" binding:"required,min=1"`
	DispositionTypeCode     string  `json:"disposition_type_code" binding:"required,max=40"`
	DispositionQty          string  `json:"disposition_qty" binding:"required,max=30"`
	BusinessDate            string  `json:"business_date" binding:"required"`
	DecidedAt               string  `json:"decided_at" binding:"required"`
	TargetLocationID        *string `json:"target_location_id" binding:"omitempty,uuid"`
	ClientDecisionReference *string `json:"client_decision_reference" binding:"omitempty,max=120"`
	DecisionNotes           *string `json:"decision_notes" binding:"omitempty,max=4000"`
	WorkInstructions        *string `json:"work_instructions" binding:"omitempty,max=4000"`
}

type ReworkTransitionRequest struct {
	ExpectedVersion int64 `json:"expected_version" binding:"required,min=1"`
}

type CompleteReworkRequest struct {
	ExpectedVersion int64   `json:"expected_version" binding:"required,min=1"`
	ResultNotes     *string `json:"result_notes" binding:"omitempty,max=4000"`
}
