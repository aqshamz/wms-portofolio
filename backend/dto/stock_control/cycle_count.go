package stockcontrol

import "time"

type CreateCycleCountRequest struct {
	BusinessDate string   `json:"business_date" binding:"required"`
	ToleranceQty string   `json:"tolerance_quantity" binding:"omitempty,max=30"`
	Notes        *string  `json:"notes" binding:"omitempty,max=4000"`
	BalanceIDs   []string `json:"balance_ids" binding:"required,min=1,max=100,dive,max=160"`
}
type CreateGrandStockOpnameRequest struct {
	OwnerID      string  `json:"owner_id" binding:"required,max=36"`
	WarehouseID  string  `json:"warehouse_id" binding:"required,max=36"`
	BusinessDate string  `json:"business_date" binding:"required"`
	ToleranceQty string  `json:"tolerance_quantity" binding:"omitempty,max=30"`
	Notes        *string `json:"notes" binding:"omitempty,max=4000"`
}
type CycleCountEntryRequest struct {
	LineID     string  `json:"line_id" binding:"required,max=180"`
	CountedQty string  `json:"counted_quantity" binding:"required,max=30"`
	Notes      *string `json:"notes" binding:"omitempty,max=2000"`
}
type AddFoundStockRequest struct {
	ExpectedVersion   int64   `json:"expected_version" binding:"required,min=1"`
	ItemID            string  `json:"item_id" binding:"required,uuid"`
	LocationID        string  `json:"location_id" binding:"required,uuid"`
	InventoryStatusID string  `json:"inventory_status_id" binding:"required,uuid"`
	LotID             *string `json:"lot_id" binding:"omitempty,max=120"`
	HandlingUnitID    *string `json:"handling_unit_id" binding:"omitempty,max=120"`
	CountedQty        string  `json:"counted_quantity" binding:"required,max=30"`
	Notes             *string `json:"notes" binding:"required,max=2000"`
}
type FoundStockItemOption struct {
	ID               string `json:"item_id"`
	Code             string `json:"code"`
	Name             string `json:"name"`
	UOMCode          string `json:"uom_code"`
	LotControlled    bool   `json:"lot_controlled"`
	SerialControlled bool   `json:"serial_controlled"`
}
type FoundStockOption struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}
type FoundStockOptionsResponse struct {
	Items         []FoundStockItemOption `json:"items"`
	Locations     []FoundStockOption     `json:"locations"`
	Statuses      []FoundStockOption     `json:"statuses"`
	Lots          []FoundStockOption     `json:"lots"`
	HandlingUnits []FoundStockOption     `json:"handling_units"`
}
type RecordCycleCountRequest struct {
	ExpectedVersion int64                    `json:"expected_version" binding:"required,min=1"`
	Lines           []CycleCountEntryRequest `json:"lines" binding:"required,min=1,max=5000,dive"`
}
type CycleCountDecisionRequest struct {
	ExpectedVersion int64    `json:"expected_version" binding:"required,min=1"`
	LineIDs         []string `json:"line_ids" binding:"required,min=1,max=5000,dive,max=180"`
}
type RejectCycleCountRequest struct {
	ExpectedVersion int64    `json:"expected_version" binding:"required,min=1"`
	LineIDs         []string `json:"line_ids" binding:"required,min=1,max=5000,dive,max=180"`
	Reason          string   `json:"reason" binding:"required,max=4000"`
}
type CancelCycleCountRequest struct {
	ExpectedVersion int64  `json:"expected_version" binding:"required,min=1"`
	Reason          string `json:"reason" binding:"required,max=4000"`
}
type ReconcileCycleCountSerialsRequest struct {
	ExpectedVersion    int64    `json:"expected_version" binding:"required,min=1"`
	LineID             string   `json:"line_id" binding:"required,max=180"`
	MissingSerialIDs   []string `json:"missing_serial_ids" binding:"max=5000,dive,max=160"`
	FoundSerialNumbers []string `json:"found_serial_numbers" binding:"max=5000,dive,max=120"`
}
type CycleCountSerialResponse struct {
	SerialID string `json:"serial_id"`
	SerialNo string `json:"serial_no"`
}
type CycleCountSerialReconciliationResponse struct {
	ID                   string    `json:"cycle_count_serial_reconciliation_id"`
	SerialID             string    `json:"serial_id"`
	SerialNo             string    `json:"serial_no"`
	DirectionCode        string    `json:"direction_code"`
	InventoryMovementID  string    `json:"inventory_movement_id"`
	CreatedAt            time.Time `json:"created_at"`
	CreatedByDisplayName string    `json:"created_by_display_name"`
}
type CycleCountLineResponse struct {
	ID                      string                                   `json:"cycle_count_line_id"`
	LineNo                  int                                      `json:"line_no"`
	DecisionCode            string                                   `json:"decision_code"`
	BalanceID               *string                                  `json:"balance_id"`
	FoundStock              bool                                     `json:"found_stock"`
	CreatedBy               string                                   `json:"created_by"`
	CountedBy               *string                                  `json:"counted_by"`
	CreatedByDisplayName    string                                   `json:"created_by_display_name"`
	DiscoveryNotes          *string                                  `json:"discovery_notes"`
	SnapshotVersionNo       int64                                    `json:"snapshot_version_no"`
	CurrentBalanceVersionNo *int64                                   `json:"current_balance_version_no,omitempty"`
	ItemCode                string                                   `json:"item_code"`
	ItemName                string                                   `json:"item_name"`
	LocationCode            string                                   `json:"location_code"`
	InventoryStatusCode     string                                   `json:"inventory_status_code"`
	UOMCode                 string                                   `json:"uom_code"`
	LotNumber               *string                                  `json:"lot_number"`
	HandlingUnitBarcode     *string                                  `json:"handling_unit_barcode"`
	SerialControlled        bool                                     `json:"serial_controlled"`
	SystemQty               *string                                  `json:"system_quantity,omitempty"`
	CountedQty              *string                                  `json:"counted_quantity"`
	VarianceQty             *string                                  `json:"variance_quantity"`
	CountAttempts           int                                      `json:"count_attempts"`
	RequiresRecount         bool                                     `json:"requires_recount"`
	CountedAt               *time.Time                               `json:"counted_at"`
	CountedByDisplayName    *string                                  `json:"counted_by_display_name"`
	CountNotes              *string                                  `json:"count_notes"`
	DecidedAt               *time.Time                               `json:"decided_at"`
	DecidedByDisplayName    *string                                  `json:"decided_by_display_name"`
	DecisionReason          *string                                  `json:"decision_reason"`
	InventoryMovementID     *string                                  `json:"inventory_movement_id"`
	ResultingBalanceID      *string                                  `json:"resulting_balance_id"`
	CurrentSerials          []CycleCountSerialResponse               `json:"current_serials"`
	SerialReconciliations   []CycleCountSerialReconciliationResponse `json:"serial_reconciliations"`
}
type CycleCountResponse struct {
	StockFrozen          bool                     `json:"stock_frozen"`
	ID                   string                   `json:"cycle_count_id"`
	CountTypeCode        string                   `json:"count_type_code"`
	StatusCode           string                   `json:"status_code"`
	OwnerID              string                   `json:"owner_id"`
	OwnerCode            string                   `json:"owner_code"`
	OwnerName            string                   `json:"owner_name"`
	WarehouseID          string                   `json:"warehouse_id"`
	WarehouseCode        string                   `json:"warehouse_code"`
	WarehouseName        string                   `json:"warehouse_name"`
	BusinessDate         string                   `json:"business_date"`
	ToleranceQty         string                   `json:"tolerance_quantity"`
	BlindCount           bool                     `json:"blind_count"`
	Notes                *string                  `json:"notes"`
	TotalLines           int64                    `json:"total_lines"`
	OpenLines            int64                    `json:"open_lines"`
	CountedLines         int64                    `json:"counted_lines"`
	RecountLines         int64                    `json:"recount_lines"`
	FinalLines           int64                    `json:"final_lines"`
	Lines                []CycleCountLineResponse `json:"lines"`
	CompletedAt          *time.Time               `json:"completed_at"`
	CancelledAt          *time.Time               `json:"cancelled_at"`
	CancellationReason   *string                  `json:"cancellation_reason"`
	CreatedAt            time.Time                `json:"created_at"`
	CreatedBy            string                   `json:"created_by"`
	CreatedByDisplayName string                   `json:"created_by_display_name"`
	VersionNo            int64                    `json:"version_no"`
}
