package stockcontrol

import "time"

type CreateReplenishmentRequest struct {
	SourceBalanceID  string  `json:"source_balance_id" binding:"required,max=160"`
	TargetLocationID string  `json:"target_location_id" binding:"required,uuid"`
	SerialID         *string `json:"serial_id" binding:"omitempty,max=160"`
	Quantity         string  `json:"quantity" binding:"required,max=30"`
	PriorityCode     string  `json:"priority_code" binding:"required,max=40"`
	Notes            *string `json:"notes" binding:"omitempty,max=4000"`
	ExpectedVersion  int64   `json:"expected_balance_version" binding:"required,min=1"`
}

type ReplenishmentTransitionRequest struct {
	ExpectedVersion int64 `json:"expected_version" binding:"required,min=1"`
}

type AssignReplenishmentRequest struct {
	ExpectedVersion int64  `json:"expected_version" binding:"required,min=1"`
	AccountID       string `json:"account_id" binding:"required,uuid"`
}

type CompleteReplenishmentRequest struct {
	ExpectedVersion        int64  `json:"expected_version" binding:"required,min=1"`
	ExpectedBalanceVersion int64  `json:"expected_balance_version" binding:"required,min=1"`
	BusinessDate           string `json:"business_date" binding:"required"`
}

type CancelReplenishmentRequest struct {
	ExpectedVersion        int64  `json:"expected_version" binding:"required,min=1"`
	ExpectedBalanceVersion int64  `json:"expected_balance_version" binding:"required,min=1"`
	Reason                 string `json:"reason" binding:"required,max=4000"`
}

type ReplenishmentTaskResponse struct {
	ID                     string     `json:"replenishment_task_id"`
	SourceBalanceID        string     `json:"source_balance_id"`
	SourceBalanceVersionNo int64      `json:"source_balance_version_no"`
	OwnerID                string     `json:"owner_id"`
	WarehouseID            string     `json:"warehouse_id"`
	ItemID                 string     `json:"item_id"`
	ItemCode               string     `json:"item_code"`
	ItemName               string     `json:"item_name"`
	LotID                  *string    `json:"lot_id"`
	LotNumber              *string    `json:"lot_number"`
	SerialID               *string    `json:"serial_id"`
	SerialNumber           *string    `json:"serial_number"`
	HandlingUnitID         *string    `json:"handling_unit_id"`
	HandlingUnitBarcode    *string    `json:"handling_unit_barcode"`
	InventoryStatusID      string     `json:"inventory_status_id"`
	InventoryStatusCode    string     `json:"inventory_status_code"`
	SourceLocationID       string     `json:"source_location_id"`
	SourceLocationCode     string     `json:"source_location_code"`
	TargetLocationID       string     `json:"target_location_id"`
	TargetLocationCode     string     `json:"target_location_code"`
	PlannedQty             string     `json:"planned_qty"`
	CompletedQty           string     `json:"completed_qty"`
	UOMID                  string     `json:"uom_id"`
	UOMCode                string     `json:"uom_code"`
	TaskStatusCode         string     `json:"task_status_code"`
	TaskPriorityCode       string     `json:"task_priority_code"`
	AssignedTo             *string    `json:"assigned_to"`
	AssignedUsername       *string    `json:"assigned_username"`
	AssignedDisplayName    *string    `json:"assigned_display_name"`
	Notes                  *string    `json:"notes"`
	StartedAt              *time.Time `json:"started_at"`
	CompletedAt            *time.Time `json:"completed_at"`
	CancelledAt            *time.Time `json:"cancelled_at"`
	CancelledBy            *string    `json:"cancelled_by"`
	CancellationReason     *string    `json:"cancellation_reason"`
	InventoryMovementID    *string    `json:"inventory_movement_id"`
	ResultingBalanceID     *string    `json:"resulting_balance_id"`
	VersionNo              int64      `json:"version_no"`
	CreatedAt              time.Time  `json:"created_at"`
}

type ReplenishmentAssigneeResponse struct {
	AccountID   string `json:"account_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

type ReplenishmentTargetResponse struct {
	LocationID       string `json:"location_id"`
	Code             string `json:"code"`
	ZoneCode         string `json:"zone_code"`
	LocationTypeCode string `json:"location_type_code"`
}

type PageResponse[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}
