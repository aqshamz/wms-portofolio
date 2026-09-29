package stockcontrol

import "time"

type CreateAdjustmentLineRequest struct {
	BalanceID       string  `json:"balance_id" binding:"required,max=160"`
	Quantity        string  `json:"quantity" binding:"required,max=30"`
	SerialID        *string `json:"serial_id" binding:"omitempty,max=160"`
	ExpectedVersion int64   `json:"expected_balance_version" binding:"required,min=1"`
}
type CreateAdjustmentRequest struct {
	BusinessDate string                        `json:"business_date" binding:"required"`
	Direction    string                        `json:"direction" binding:"required,oneof=INCREASE DECREASE"`
	ReasonCode   string                        `json:"reason_code" binding:"required,max=40"`
	Notes        *string                       `json:"notes" binding:"omitempty,max=4000"`
	Lines        []CreateAdjustmentLineRequest `json:"lines" binding:"required,min=1,max=100,dive"`
}
type AdjustmentLineSelectionRequest struct {
	ExpectedVersion int64    `json:"expected_version" binding:"required,min=1"`
	LineIDs         []string `json:"line_ids" binding:"required,min=1,max=100,dive,max=180"`
}
type RejectAdjustmentLinesRequest struct {
	ExpectedVersion int64    `json:"expected_version" binding:"required,min=1"`
	LineIDs         []string `json:"line_ids" binding:"required,min=1,max=100,dive,max=180"`
	Reason          string   `json:"reason" binding:"required,max=4000"`
}
type CancelAdjustmentRequest struct {
	ExpectedVersion int64  `json:"expected_version" binding:"required,min=1"`
	Reason          string `json:"reason" binding:"required,max=4000"`
}

type AdjustmentLineResponse struct {
	ID                      string     `json:"inventory_adjustment_line_id"`
	LineNo                  int        `json:"line_no"`
	DecisionCode            string     `json:"decision_code"`
	BalanceID               string     `json:"balance_id"`
	PlannedBalanceVersionNo int64      `json:"planned_balance_version_no"`
	CurrentBalanceVersionNo *int64     `json:"current_balance_version_no,omitempty"`
	ItemID                  string     `json:"item_id"`
	ItemCode                string     `json:"item_code"`
	ItemName                string     `json:"item_name"`
	LotID                   *string    `json:"lot_id"`
	LotNumber               *string    `json:"lot_number"`
	SerialID                *string    `json:"serial_id"`
	SerialNumber            *string    `json:"serial_number"`
	HandlingUnitID          *string    `json:"handling_unit_id"`
	HandlingUnitBarcode     *string    `json:"handling_unit_barcode"`
	LocationID              string     `json:"location_id"`
	LocationCode            string     `json:"location_code"`
	InventoryStatusID       string     `json:"inventory_status_id"`
	InventoryStatusCode     string     `json:"inventory_status_code"`
	Quantity                string     `json:"quantity"`
	UOMID                   string     `json:"uom_id"`
	UOMCode                 string     `json:"uom_code"`
	ApprovedAt              *time.Time `json:"approved_at"`
	ApprovedBy              *string    `json:"approved_by"`
	ApprovedByDisplayName   *string    `json:"approved_by_display_name"`
	RejectedAt              *time.Time `json:"rejected_at"`
	RejectedBy              *string    `json:"rejected_by"`
	RejectedByDisplayName   *string    `json:"rejected_by_display_name"`
	RejectionReason         *string    `json:"rejection_reason"`
	CancelledAt             *time.Time `json:"cancelled_at"`
	CancelledBy             *string    `json:"cancelled_by"`
	CancellationReason      *string    `json:"cancellation_reason"`
	InventoryMovementID     *string    `json:"inventory_movement_id"`
	ResultingBalanceID      *string    `json:"resulting_balance_id"`
	VersionNo               int64      `json:"version_no"`
}
type AdjustmentResponse struct {
	ID                     string                   `json:"inventory_adjustment_id"`
	StatusCode             string                   `json:"status_code"`
	OwnerID                string                   `json:"owner_id"`
	OwnerCode              string                   `json:"owner_code"`
	OwnerName              string                   `json:"owner_name"`
	WarehouseID            string                   `json:"warehouse_id"`
	WarehouseCode          string                   `json:"warehouse_code"`
	WarehouseName          string                   `json:"warehouse_name"`
	BusinessDate           string                   `json:"business_date"`
	Direction              string                   `json:"direction"`
	ReasonCode             string                   `json:"reason_code"`
	ReasonName             string                   `json:"reason_name"`
	Notes                  *string                  `json:"notes"`
	TotalLines             int64                    `json:"total_lines"`
	PendingLines           int64                    `json:"pending_lines"`
	PostedLines            int64                    `json:"posted_lines"`
	RejectedLines          int64                    `json:"rejected_lines"`
	CancelledLines         int64                    `json:"cancelled_lines"`
	Lines                  []AdjustmentLineResponse `json:"lines"`
	CompletedAt            *time.Time               `json:"completed_at"`
	CancelledAt            *time.Time               `json:"cancelled_at"`
	CancelledBy            *string                  `json:"cancelled_by"`
	CancelledByDisplayName *string                  `json:"cancelled_by_display_name"`
	CancellationReason     *string                  `json:"cancellation_reason"`
	CreatedAt              time.Time                `json:"created_at"`
	CreatedBy              string                   `json:"created_by"`
	CreatedByUsername      string                   `json:"created_by_username"`
	CreatedByDisplayName   string                   `json:"created_by_display_name"`
	VersionNo              int64                    `json:"version_no"`
}
