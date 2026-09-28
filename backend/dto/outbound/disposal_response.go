package outbound

import "time"

type DisposalResponse struct {
	ID                        string     `json:"disposal_id"`
	QuarantineDispositionID   string     `json:"quarantine_disposition_id"`
	QuarantineCaseID          string     `json:"quarantine_case_id"`
	OwnerID                   string     `json:"owner_id"`
	OwnerCode                 string     `json:"owner_code"`
	OwnerName                 string     `json:"owner_name"`
	WarehouseID               string     `json:"warehouse_id"`
	WarehouseCode             string     `json:"warehouse_code"`
	WarehouseName             string     `json:"warehouse_name"`
	BusinessDate              string     `json:"business_date"`
	SourceBalanceID           string     `json:"source_balance_id"`
	SourceBalanceVersionNo    *int64     `json:"source_balance_version_no,omitempty"`
	AvailableQty              string     `json:"available_qty,omitempty"`
	PlannedBalanceVersionNo   int64      `json:"planned_balance_version_no"`
	ItemID                    string     `json:"item_id"`
	ItemCode                  string     `json:"item_code"`
	ItemName                  string     `json:"item_name"`
	LotID                     *string    `json:"lot_id"`
	LotNumber                 string     `json:"lot_number,omitempty"`
	SerialID                  *string    `json:"serial_id"`
	SerialNo                  string     `json:"serial_no,omitempty"`
	HandlingUnitID            *string    `json:"handling_unit_id"`
	HandlingUnitBarcode       string     `json:"handling_unit_barcode,omitempty"`
	SourceLocationID          string     `json:"source_location_id"`
	SourceLocationCode        string     `json:"source_location_code"`
	SourceInventoryStatusID   string     `json:"source_inventory_status_id"`
	SourceInventoryStatusCode string     `json:"source_inventory_status_code"`
	Quantity                  string     `json:"quantity"`
	UOMID                     string     `json:"uom_id"`
	UOMCode                   string     `json:"uom_code"`
	StatusCode                string     `json:"status_code"`
	Notes                     *string    `json:"notes"`
	PlannedAt                 time.Time  `json:"planned_at"`
	CompletedAt               *time.Time `json:"completed_at"`
	CompletedBy               *string    `json:"completed_by"`
	CompletedByDisplayName    *string    `json:"completed_by_display_name,omitempty"`
	CancelledAt               *time.Time `json:"cancelled_at"`
	CancelledBy               *string    `json:"cancelled_by"`
	CancelledByDisplayName    *string    `json:"cancelled_by_display_name,omitempty"`
	CancellationReason        *string    `json:"cancellation_reason"`
	InventoryMovementID       *string    `json:"inventory_movement_id"`
	CreatedAt                 time.Time  `json:"created_at"`
	CreatedBy                 string     `json:"created_by"`
	CreatedByDisplayName      string     `json:"created_by_display_name"`
	VersionNo                 int64      `json:"version_no"`
}
