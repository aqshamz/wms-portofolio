package inventory

type CreateHandlingUnitRequest struct {
	WarehouseID          string  `json:"warehouse_id" binding:"required,uuid"`
	OwnerID              string  `json:"owner_id" binding:"required,uuid"`
	HandlingUnitTypeID   string  `json:"handling_unit_type_id" binding:"required,uuid"`
	ParentHandlingUnitID *string `json:"parent_handling_unit_id" binding:"omitempty,max=120"`
	CurrentLocationID    *string `json:"current_location_id" binding:"omitempty,uuid"`
	Barcode              string  `json:"barcode" binding:"required,max=120"`
}
