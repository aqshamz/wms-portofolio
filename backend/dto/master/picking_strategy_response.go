package master

type PickingStrategyResponse struct {
	ID          string  `json:"picking_strategy_id"`
	OwnerID     *string `json:"owner_id"`
	WarehouseID *string `json:"warehouse_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IsActive    bool    `json:"is_active"`
}
