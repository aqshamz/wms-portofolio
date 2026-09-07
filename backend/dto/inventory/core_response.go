package inventory

import "time"

type BalanceResponse struct {
	ID                  string    `json:"balance_id"`
	OwnerID             string    `json:"owner_id"`
	WarehouseID         string    `json:"warehouse_id"`
	LocationID          string    `json:"location_id"`
	LocationCode        string    `json:"location_code"`
	ItemID              string    `json:"item_id"`
	ItemCode            string    `json:"item_code"`
	ItemName            string    `json:"item_name"`
	LotID               *string   `json:"lot_id"`
	LotNumber           *string   `json:"lot_number"`
	HandlingUnitID      *string   `json:"handling_unit_id"`
	InventoryStatusID   string    `json:"inventory_status_id"`
	InventoryStatusCode string    `json:"inventory_status_code"`
	OnHandQty           string    `json:"on_hand_qty"`
	ReservedQty         string    `json:"reserved_qty"`
	AvailableQty        string    `json:"available_qty"`
	UOMID               string    `json:"uom_id"`
	UOMCode             string    `json:"uom_code"`
	VersionNo           int64     `json:"version_no"`
	UpdatedAt           time.Time `json:"updated_at"`
}
type MovementResponse struct {
	ID               string    `json:"movement_id"`
	MovementTypeID   string    `json:"movement_type_id"`
	MovementTypeCode string    `json:"movement_type_code"`
	OwnerID          string    `json:"owner_id"`
	WarehouseID      string    `json:"warehouse_id"`
	BusinessDate     string    `json:"business_date"`
	OccurredAt       time.Time `json:"occurred_at"`
	ItemID           string    `json:"item_id"`
	ItemCode         string    `json:"item_code"`
	LotID            *string   `json:"lot_id"`
	LotNumber        *string   `json:"lot_number"`
	SerialID         *string   `json:"serial_id"`
	HandlingUnitID   *string   `json:"handling_unit_id"`
	FromLocationID   *string   `json:"from_location_id"`
	FromLocationCode *string   `json:"from_location_code"`
	ToLocationID     *string   `json:"to_location_id"`
	ToLocationCode   *string   `json:"to_location_code"`
	FromStatusID     *string   `json:"from_status_id"`
	FromStatusCode   *string   `json:"from_status_code"`
	ToStatusID       *string   `json:"to_status_id"`
	ToStatusCode     *string   `json:"to_status_code"`
	Quantity         string    `json:"quantity"`
	UOMID            string    `json:"uom_id"`
	UOMCode          string    `json:"uom_code"`
	SourceDocumentID string    `json:"source_document_id"`
	SourceLineID     *string   `json:"source_line_id"`
	ReasonCodeID     *string   `json:"reason_code_id"`
	Notes            *string   `json:"notes"`
	OperationKey     *string   `json:"operation_key,omitempty"`
	CreatedBy        string    `json:"created_by"`
}
type SerialStateResponse struct {
	SerialID  string          `json:"serial_id"`
	SerialNo  string          `json:"serial_no"`
	Balance   BalanceResponse `json:"balance"`
	VersionNo int64           `json:"version_no"`
	UpdatedAt time.Time       `json:"updated_at"`
}
type MovementTypeResponse struct {
	ID          string  `json:"movement_type_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IsActive    bool    `json:"is_active"`
}
