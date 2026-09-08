package inbound

import "time"

type PageResponse[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int64 `json:"total_pages"`
}

type PurchaseOrderLineResponse struct {
	ID                       string  `json:"purchase_order_line_id"`
	LineNo                   int     `json:"line_no"`
	ItemID                   string  `json:"item_id"`
	ItemCode                 string  `json:"item_code"`
	ItemName                 string  `json:"item_name"`
	OrderedQty               string  `json:"ordered_qty"`
	ScheduledQty             string  `json:"scheduled_qty"`
	CompletedReceiptQty      string  `json:"completed_receipt_qty"`
	OverReceiptTolerancePct  string  `json:"over_receipt_tolerance_pct"`
	UnderReceiptTolerancePct string  `json:"under_receipt_tolerance_pct"`
	UOMID                    string  `json:"uom_id"`
	UOMCode                  string  `json:"uom_code"`
	VendorItemCode           *string `json:"vendor_item_code"`
	ExpectedLotNo            *string `json:"expected_lot_no"`
	ExpectedExpiryDate       *string `json:"expected_expiry_date"`
	Notes                    *string `json:"notes"`
}

type PurchaseOrderResponse struct {
	ID                string                      `json:"purchase_order_id"`
	OwnerID           string                      `json:"owner_id"`
	OwnerCode         string                      `json:"owner_code"`
	VendorID          string                      `json:"vendor_id"`
	VendorCode        string                      `json:"vendor_code"`
	VendorName        string                      `json:"vendor_name"`
	WarehouseID       string                      `json:"warehouse_id"`
	WarehouseCode     string                      `json:"warehouse_code"`
	BusinessDate      string                      `json:"business_date"`
	PurchaseOrderNo   string                      `json:"purchase_order_no"`
	OrderedAt         time.Time                   `json:"ordered_at"`
	ExpectedArrivalAt *time.Time                  `json:"expected_arrival_at"`
	StatusCode        string                      `json:"status_code"`
	Notes             *string                     `json:"notes"`
	VersionNo         int64                       `json:"version_no"`
	CreatedAt         time.Time                   `json:"created_at"`
	Lines             []PurchaseOrderLineResponse `json:"lines,omitempty"`
}

type InboundOrderLineResponse struct {
	ID                    string  `json:"inbound_line_id"`
	PurchaseOrderLineID   *string `json:"purchase_order_line_id"`
	LineNo                int     `json:"line_no"`
	ItemID                string  `json:"item_id"`
	ItemCode              string  `json:"item_code"`
	ItemName              string  `json:"item_name"`
	ExpectedQty           string  `json:"expected_qty"`
	CompletedReceiptQty   string  `json:"completed_receipt_qty"`
	UOMID                 string  `json:"uom_id"`
	UOMCode               string  `json:"uom_code"`
	ExpectedLotNo         *string `json:"expected_lot_no"`
	ExpectedExpiryDate    *string `json:"expected_expiry_date"`
	CustomerLineReference *string `json:"customer_line_reference"`
	Notes                 *string `json:"notes"`
}

type InboundOrderResponse struct {
	ID                string                     `json:"inbound_id"`
	OwnerID           string                     `json:"owner_id"`
	OwnerCode         string                     `json:"owner_code"`
	VendorID          string                     `json:"vendor_id"`
	VendorCode        string                     `json:"vendor_code"`
	VendorName        string                     `json:"vendor_name"`
	WarehouseID       string                     `json:"warehouse_id"`
	WarehouseCode     string                     `json:"warehouse_code"`
	BusinessDate      string                     `json:"business_date"`
	ExpectedArrivalAt *time.Time                 `json:"expected_arrival_at"`
	ExternalReference *string                    `json:"external_reference"`
	SupplierReference *string                    `json:"supplier_reference"`
	StatusCode        string                     `json:"status_code"`
	Notes             *string                    `json:"notes"`
	VersionNo         int64                      `json:"version_no"`
	CreatedAt         time.Time                  `json:"created_at"`
	Lines             []InboundOrderLineResponse `json:"lines,omitempty"`
}

type ReceiptBatchResponse struct {
	ID                         string  `json:"receipt_inventory_id"`
	ItemID                     string  `json:"item_id"`
	SourceQty                  string  `json:"source_qty"`
	SourceUOMID                string  `json:"source_uom_id"`
	SourceUOMCode              string  `json:"source_uom_code"`
	BaseQty                    string  `json:"base_qty"`
	BaseUOMID                  string  `json:"base_uom_id"`
	BaseUOMCode                string  `json:"base_uom_code"`
	LotID                      *string `json:"lot_id"`
	LotNumber                  *string `json:"lot_number"`
	HandlingUnitID             *string `json:"handling_unit_id"`
	HandlingUnitBarcode        *string `json:"handling_unit_barcode"`
	SerialID                   *string `json:"serial_id"`
	SerialNo                   *string `json:"serial_no"`
	ReceivedLocationID         string  `json:"received_location_id"`
	ReceivedLocationCode       string  `json:"received_location_code"`
	InitialInventoryStatusID   string  `json:"initial_inventory_status_id"`
	InitialInventoryStatusCode string  `json:"initial_inventory_status_code"`
	InitialBalanceID           *string `json:"initial_balance_id"`
}

type ReceiptLineResponse struct {
	ID                string                 `json:"receipt_line_id"`
	InboundLineID     *string                `json:"inbound_line_id"`
	LineNo            int                    `json:"line_no"`
	ItemID            string                 `json:"item_id"`
	ItemCode          string                 `json:"item_code"`
	ItemName          string                 `json:"item_name"`
	ReceivedQty       string                 `json:"received_qty"`
	RejectedQty       string                 `json:"rejected_qty"`
	ExceptionNotes    *string                `json:"exception_notes"`
	ExceptionTypeCode *string                `json:"exception_type_code"`
	AcceptedQty       string                 `json:"accepted_qty"`
	BatchedQty        string                 `json:"batched_qty"`
	UOMID             string                 `json:"uom_id"`
	UOMCode           string                 `json:"uom_code"`
	Batches           []ReceiptBatchResponse `json:"batches"`
}

type ReceiptResponse struct {
	ID             string                `json:"receipt_id"`
	InboundID      *string               `json:"inbound_id"`
	OwnerID        string                `json:"owner_id"`
	OwnerCode      string                `json:"owner_code"`
	WarehouseID    string                `json:"warehouse_id"`
	WarehouseCode  string                `json:"warehouse_code"`
	BusinessDate   string                `json:"business_date"`
	ReceivedAt     time.Time             `json:"received_at"`
	DockLocationID *string               `json:"dock_location_id"`
	VehicleNumber  *string               `json:"vehicle_number"`
	SealNumber     *string               `json:"seal_number"`
	DeliveryNoteNo *string               `json:"delivery_note_no"`
	StatusCode     string                `json:"status_code"`
	Notes          *string               `json:"notes"`
	VersionNo      int64                 `json:"version_no"`
	CreatedAt      time.Time             `json:"created_at"`
	Lines          []ReceiptLineResponse `json:"lines,omitempty"`
}

type QualityInspectionResponse struct {
	ID                      string                  `json:"inspection_id"`
	ReceiptInventoryID      string                  `json:"receipt_inventory_id"`
	ParentInspectionID      *string                 `json:"parent_inspection_id"`
	SourceBalanceID         *string                 `json:"source_balance_id"`
	ReceiptID               string                  `json:"receipt_id"`
	OwnerID                 string                  `json:"owner_id"`
	WarehouseID             string                  `json:"warehouse_id"`
	ItemID                  string                  `json:"item_id"`
	ItemCode                string                  `json:"item_code"`
	LotNumber               string                  `json:"lot_number,omitempty"`
	LocationCode            string                  `json:"location_code"`
	QualityStatusCode       string                  `json:"quality_status_code"`
	InspectionResultCode    string                  `json:"inspection_result_code,omitempty"`
	InspectedQty            string                  `json:"inspected_qty"`
	PassedQty               string                  `json:"passed_qty"`
	FailedQty               string                  `json:"failed_qty"`
	InspectedAt             *time.Time              `json:"inspected_at"`
	InspectedBy             *string                 `json:"inspected_by"`
	CancelledAt             *time.Time              `json:"cancelled_at"`
	CancelledBy             *string                 `json:"cancelled_by"`
	CancellationReason      *string                 `json:"cancellation_reason"`
	ReplacementInspectionID *string                 `json:"replacement_inspection_id"`
	Notes                   *string                 `json:"notes"`
	VersionNo               int64                   `json:"version_no"`
	CreatedAt               time.Time               `json:"created_at"`
	PutawayTask             *PutawayTaskResponse    `json:"putaway_task,omitempty"`
	QuarantineCase          *QuarantineCaseResponse `json:"quarantine_case,omitempty"`
}

type PutawayTaskResponse struct {
	ID                      string     `json:"putaway_task_id"`
	InspectionID            string     `json:"inspection_id"`
	ReceiptInventoryID      string     `json:"receipt_inventory_id"`
	SourceBalanceID         string     `json:"source_balance_id"`
	OwnerID                 string     `json:"owner_id"`
	WarehouseID             string     `json:"warehouse_id"`
	ItemID                  string     `json:"item_id"`
	ItemCode                string     `json:"item_code"`
	LotID                   *string    `json:"lot_id"`
	LotNumber               string     `json:"lot_number,omitempty"`
	HandlingUnitID          *string    `json:"handling_unit_id"`
	SourceLocationID        string     `json:"source_location_id"`
	SourceLocationCode      string     `json:"source_location_code"`
	TargetLocationID        string     `json:"target_location_id"`
	TargetLocationCode      string     `json:"target_location_code"`
	PlannedQty              string     `json:"planned_qty"`
	CompletedQty            string     `json:"completed_qty"`
	UOMID                   string     `json:"uom_id"`
	TaskStatusCode          string     `json:"task_status_code"`
	TaskPriorityCode        string     `json:"task_priority_code"`
	AssignedTo              *string    `json:"assigned_to"`
	StartedAt               *time.Time `json:"started_at"`
	CompletedAt             *time.Time `json:"completed_at"`
	InventoryMovementID     *string    `json:"inventory_movement_id"`
	ResultingBalanceID      *string    `json:"resulting_balance_id"`
	ReversalMovementID      *string    `json:"reversal_movement_id"`
	ReversedAt              *time.Time `json:"reversed_at"`
	ReversedBy              *string    `json:"reversed_by"`
	ReversalReason          *string    `json:"reversal_reason"`
	ReplacementInspectionID *string    `json:"replacement_inspection_id"`
	VersionNo               int64      `json:"version_no"`
	CreatedAt               time.Time  `json:"created_at"`
}

type QuarantineDispositionTypeResponse struct {
	ID                   string  `json:"quarantine_disposition_type_id"`
	Code                 string  `json:"code"`
	Name                 string  `json:"name"`
	Description          *string `json:"description"`
	ReleasesToAvailable  bool    `json:"releases_to_available"`
	RequiresReinspection bool    `json:"requires_reinspection"`
	RemovesInventory     bool    `json:"removes_inventory"`
	IsActive             bool    `json:"is_active"`
}

type QuarantineDispositionResponse struct {
	ID                      string              `json:"quarantine_disposition_id"`
	QuarantineCaseID        string              `json:"quarantine_case_id"`
	DispositionTypeCode     string              `json:"disposition_type_code"`
	StatusCode              string              `json:"status_code"`
	DispositionQty          string              `json:"disposition_qty"`
	UOMID                   string              `json:"uom_id"`
	ClientDecisionReference *string             `json:"client_decision_reference"`
	DecisionNotes           *string             `json:"decision_notes"`
	DecidedAt               time.Time           `json:"decided_at"`
	DecidedBy               string              `json:"decided_by"`
	ProcessedAt             *time.Time          `json:"processed_at"`
	InventoryMovementID     *string             `json:"inventory_movement_id"`
	ResultingBalanceID      *string             `json:"resulting_balance_id"`
	TargetLocationID        *string             `json:"target_location_id"`
	CreatedAt               time.Time           `json:"created_at"`
	ReworkTask              *ReworkTaskResponse `json:"rework_task,omitempty"`
}

type QuarantineCaseResponse struct {
	ID                     string                          `json:"quarantine_case_id"`
	ParentQuarantineCaseID *string                         `json:"parent_quarantine_case_id"`
	InspectionID           string                          `json:"inspection_id"`
	ReceiptInventoryID     string                          `json:"receipt_inventory_id"`
	QuarantineBalanceID    string                          `json:"quarantine_balance_id"`
	OwnerID                string                          `json:"owner_id"`
	WarehouseID            string                          `json:"warehouse_id"`
	ItemID                 string                          `json:"item_id"`
	ItemCode               string                          `json:"item_code"`
	LotNumber              string                          `json:"lot_number,omitempty"`
	LocationCode           string                          `json:"location_code"`
	StatusCode             string                          `json:"status_code"`
	QuarantineQty          string                          `json:"quarantine_qty"`
	DisposedQty            string                          `json:"disposed_qty"`
	UOMID                  string                          `json:"uom_id"`
	OpenedAt               time.Time                       `json:"opened_at"`
	ClosedAt               *time.Time                      `json:"closed_at"`
	Notes                  *string                         `json:"notes"`
	VersionNo              int64                           `json:"version_no"`
	Dispositions           []QuarantineDispositionResponse `json:"dispositions"`
}

type InboundExceptionResponse struct {
	ID                string    `json:"inbound_exception_id"`
	OwnerID           string    `json:"owner_id"`
	WarehouseID       string    `json:"warehouse_id"`
	SourceDocumentID  string    `json:"source_document_id"`
	SourceLineID      *string   `json:"source_line_id"`
	ExceptionTypeCode string    `json:"exception_type_code"`
	ExpectedQty       *string   `json:"expected_qty"`
	ActualQty         *string   `json:"actual_qty"`
	VarianceQty       *string   `json:"variance_qty"`
	Notes             *string   `json:"notes"`
	CreatedAt         time.Time `json:"created_at"`
	CreatedBy         string    `json:"created_by"`
}

type ReworkTaskResponse struct {
	ID                      string     `json:"rework_task_id"`
	QuarantineDispositionID string     `json:"quarantine_disposition_id"`
	QuarantineCaseID        string     `json:"quarantine_case_id"`
	SourceBalanceID         string     `json:"source_balance_id"`
	OwnerID                 string     `json:"owner_id"`
	WarehouseID             string     `json:"warehouse_id"`
	ItemID                  string     `json:"item_id"`
	ItemCode                string     `json:"item_code"`
	TaskStatusCode          string     `json:"task_status_code"`
	TaskPriorityCode        string     `json:"task_priority_code"`
	PlannedQty              string     `json:"planned_qty"`
	CompletedQty            string     `json:"completed_qty"`
	UOMID                   string     `json:"uom_id"`
	AssignedTo              *string    `json:"assigned_to"`
	WorkInstructions        *string    `json:"work_instructions"`
	ResultNotes             *string    `json:"result_notes"`
	StartedAt               *time.Time `json:"started_at"`
	CompletedAt             *time.Time `json:"completed_at"`
	ReinspectionID          *string    `json:"reinspection_id"`
	VersionNo               int64      `json:"version_no"`
	CreatedAt               time.Time  `json:"created_at"`
}
