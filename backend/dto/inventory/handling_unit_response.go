package inventory

import "time"

type HandlingUnitResponse struct {
	ID                   string    `json:"handling_unit_id"`
	WarehouseID          string    `json:"warehouse_id"`
	OwnerID              string    `json:"owner_id"`
	HandlingUnitTypeID   string    `json:"handling_unit_type_id"`
	ParentHandlingUnitID *string   `json:"parent_handling_unit_id"`
	CurrentLocationID    *string   `json:"current_location_id"`
	Barcode              string    `json:"barcode"`
	IsClosed             bool      `json:"is_closed"`
	CreatedAt            time.Time `json:"created_at"`
	CreatedBy            *string   `json:"created_by"`
}
