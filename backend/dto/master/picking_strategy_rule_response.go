package master

type PickingStrategyRuleResponse struct {
	ID                  string  `json:"rule_id"`
	PickingStrategyID   string  `json:"picking_strategy_id"`
	SequenceNo          int     `json:"sequence_no"`
	InventoryStatusID   *string `json:"inventory_status_id"`
	ZoneID              *string `json:"zone_id"`
	PickingSortMethodID string  `json:"picking_sort_method_id"`
	IsActive            bool    `json:"is_active"`
}
