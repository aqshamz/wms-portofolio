package stockcontrol

import "time"

type CreateWarehouseTransferRequest struct {
	BusinessDate      string   `json:"business_date" binding:"required"`
	SourceBalanceID   string   `json:"source_balance_id" binding:"required,max=160"`
	TargetWarehouseID string   `json:"target_warehouse_id" binding:"required,uuid"`
	Quantity          string   `json:"quantity" binding:"required,max=30"`
	SerialIDs         []string `json:"serial_ids" binding:"omitempty,max=1,dive,max=160"`
	Notes             *string  `json:"notes" binding:"omitempty,max=4000"`
}
type WarehouseTransferTransitionRequest struct {
	ExpectedVersion int64 `json:"expected_version" binding:"required,min=1"`
}
type CancelWarehouseTransferRequest struct {
	ExpectedVersion int64  `json:"expected_version" binding:"required,min=1"`
	Reason          string `json:"reason" binding:"required,max=4000"`
}
type ReceiveWarehouseTransferRequest struct {
	ExpectedVersion         int64  `json:"expected_version" binding:"required,min=1"`
	BusinessDate            string `json:"business_date" binding:"required"`
	ReceiptLocationID       string `json:"receipt_location_id" binding:"required,uuid"`
	PutawayTargetLocationID string `json:"putaway_target_location_id" binding:"required,uuid"`
}
type PutawayWarehouseTransferRequest struct {
	ExpectedVersion        int64  `json:"expected_version" binding:"required,min=1"`
	ExpectedBalanceVersion int64  `json:"expected_balance_version" binding:"required,min=1"`
	BusinessDate           string `json:"business_date" binding:"required"`
}
type WarehouseTransferLineResponse struct {
	ID                        string     `json:"warehouse_transfer_line_id"`
	LineNo                    int        `json:"line_no"`
	SourceBalanceID           string     `json:"source_balance_id"`
	SourceBalanceVersionNo    int64      `json:"source_balance_version_no"`
	ItemID                    string     `json:"item_id"`
	ItemCode                  string     `json:"item_code"`
	ItemName                  string     `json:"item_name"`
	LotID                     *string    `json:"lot_id"`
	LotNumber                 *string    `json:"lot_number"`
	SerialID                  *string    `json:"serial_id"`
	SerialNumber              *string    `json:"serial_number"`
	HandlingUnitID            *string    `json:"handling_unit_id"`
	HandlingUnitBarcode       *string    `json:"handling_unit_barcode"`
	SourceLocationID          string     `json:"source_location_id"`
	SourceLocationCode        string     `json:"source_location_code"`
	SourceInventoryStatusID   string     `json:"source_inventory_status_id"`
	SourceInventoryStatusCode string     `json:"source_inventory_status_code"`
	UOMID                     string     `json:"uom_id"`
	UOMCode                   string     `json:"uom_code"`
	Quantity                  string     `json:"quantity"`
	DispatchMovementID        *string    `json:"dispatch_movement_id"`
	ReceiptLocationID         *string    `json:"receipt_location_id"`
	ReceiptLocationCode       *string    `json:"receipt_location_code"`
	ReceiptMovementID         *string    `json:"receipt_movement_id"`
	ReceivedBalanceID         *string    `json:"received_balance_id"`
	ReceivedBalanceVersionNo  *int64     `json:"received_balance_version_no"`
	PutawayTargetLocationID   *string    `json:"putaway_target_location_id"`
	PutawayTargetLocationCode *string    `json:"putaway_target_location_code"`
	PutawayMovementID         *string    `json:"putaway_movement_id"`
	PutawayResultBalanceID    *string    `json:"putaway_result_balance_id"`
	PutawayCompletedAt        *time.Time `json:"putaway_completed_at"`
}
type WarehouseTransferResponse struct {
	ID                   string                          `json:"warehouse_transfer_id"`
	StatusCode           string                          `json:"status_code"`
	OwnerID              string                          `json:"owner_id"`
	OwnerCode            string                          `json:"owner_code"`
	OwnerName            string                          `json:"owner_name"`
	SourceWarehouseID    string                          `json:"source_warehouse_id"`
	SourceWarehouseCode  string                          `json:"source_warehouse_code"`
	SourceWarehouseName  string                          `json:"source_warehouse_name"`
	TargetWarehouseID    string                          `json:"target_warehouse_id"`
	TargetWarehouseCode  string                          `json:"target_warehouse_code"`
	TargetWarehouseName  string                          `json:"target_warehouse_name"`
	BusinessDate         string                          `json:"business_date"`
	Notes                *string                         `json:"notes"`
	Lines                []WarehouseTransferLineResponse `json:"lines"`
	ApprovedAt           *time.Time                      `json:"approved_at"`
	DispatchedAt         *time.Time                      `json:"dispatched_at"`
	ReceivedAt           *time.Time                      `json:"received_at"`
	CancelledAt          *time.Time                      `json:"cancelled_at"`
	CancellationReason   *string                         `json:"cancellation_reason"`
	CreatedAt            time.Time                       `json:"created_at"`
	CreatedBy            string                          `json:"created_by"`
	CreatedByDisplayName string                          `json:"created_by_display_name"`
	VersionNo            int64                           `json:"version_no"`
}
