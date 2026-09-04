package master

type CreatePickingStrategyRuleRequest struct {
	SequenceNo          int     `json:"sequence_no" binding:"min=1,max=2147483647"`
	InventoryStatusID   *string `json:"inventory_status_id" binding:"omitempty,uuid"`
	ZoneID              *string `json:"zone_id" binding:"omitempty,uuid"`
	PickingSortMethodID string  `json:"picking_sort_method_id" binding:"required,uuid"`
}

type UpdatePickingStrategyRuleRequest struct {
	SequenceNo          int     `json:"sequence_no" binding:"min=1,max=2147483647"`
	InventoryStatusID   *string `json:"inventory_status_id" binding:"omitempty,uuid"`
	ZoneID              *string `json:"zone_id" binding:"omitempty,uuid"`
	PickingSortMethodID string  `json:"picking_sort_method_id" binding:"required,uuid"`
	IsActive            *bool   `json:"is_active" binding:"required"`
}
