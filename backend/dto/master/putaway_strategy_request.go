package master

type CreatePutawayStrategyRequest struct {
	OwnerID     *string `json:"owner_id" binding:"omitempty,uuid"`
	WarehouseID *string `json:"warehouse_id" binding:"omitempty,uuid"`
	Code        string  `json:"code" binding:"required,max=40"`
	Name        string  `json:"name" binding:"required,max=100"`
	Description *string `json:"description"`
}

type UpdatePutawayStrategyRequest struct {
	Name        string  `json:"name" binding:"required,max=100"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active" binding:"required"`
}
