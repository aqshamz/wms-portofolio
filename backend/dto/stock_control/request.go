package stockcontrol

type CommandBase struct {
	OperationKey     string  `json:"operation_key" binding:"required,max=160"`
	BusinessDate     string  `json:"business_date" binding:"required"`
	SourceDocumentID string  `json:"source_document_id" binding:"required,max=140"`
	SourceLineID     *string `json:"source_line_id" binding:"omitempty,max=160"`
	ReasonCode       *string `json:"reason_code" binding:"omitempty,max=40"`
	Notes            *string `json:"notes"`
}
type InternalMoveRequest struct {
	CommandBase
	SourceBalanceID  string   `json:"source_balance_id" binding:"required,max=160"`
	TargetLocationID string   `json:"target_location_id" binding:"required,uuid"`
	Quantity         string   `json:"quantity" binding:"required"`
	SerialIDs        []string `json:"serial_ids"`
	ExpectedVersion  int64    `json:"expected_version" binding:"min=1"`
}
type StatusChangeRequest struct {
	CommandBase
	SourceBalanceID         string   `json:"source_balance_id" binding:"required,max=160"`
	TargetInventoryStatusID string   `json:"target_inventory_status_id" binding:"required,uuid"`
	Quantity                string   `json:"quantity" binding:"required"`
	SerialIDs               []string `json:"serial_ids"`
	ExpectedVersion         int64    `json:"expected_version" binding:"min=1"`
}
type AdjustmentRequest struct {
	CommandBase
	BalanceID       string   `json:"balance_id" binding:"required,max=160"`
	Direction       string   `json:"direction" binding:"required,oneof=INCREASE DECREASE"`
	Quantity        string   `json:"quantity" binding:"required"`
	SerialIDs       []string `json:"serial_ids"`
	ExpectedVersion int64    `json:"expected_version" binding:"min=1"`
}
type StockCountReconcileRequest struct {
	CommandBase
	BalanceID       string   `json:"balance_id" binding:"required,max=160"`
	CountedQty      string   `json:"counted_qty" binding:"required"`
	SerialIDs       []string `json:"serial_ids"`
	ExpectedVersion int64    `json:"expected_version" binding:"min=1"`
}
type WarehouseTransferRequest struct {
	CommandBase
	SourceBalanceID         string   `json:"source_balance_id" binding:"required,max=160"`
	TargetWarehouseID       string   `json:"target_warehouse_id" binding:"required,uuid"`
	TargetLocationID        string   `json:"target_location_id" binding:"required,uuid"`
	TargetInventoryStatusID string   `json:"target_inventory_status_id" binding:"required,uuid"`
	Quantity                string   `json:"quantity" binding:"required"`
	SerialIDs               []string `json:"serial_ids"`
	ExpectedVersion         int64    `json:"expected_version" binding:"min=1"`
}
