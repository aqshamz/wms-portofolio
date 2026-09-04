package master

type CreatePutawayStrategyRuleRequest struct {
	SequenceNo          int     `json:"sequence_no" binding:"min=1,max=2147483647"`
	CategoryID          *string `json:"category_id" binding:"omitempty,uuid"`
	LocationTypeID      *string `json:"location_type_id" binding:"omitempty,uuid"`
	ZoneID              *string `json:"zone_id" binding:"omitempty,uuid"`
	MinimumEmptyPercent *string `json:"minimum_empty_percent"`
}

type UpdatePutawayStrategyRuleRequest struct {
	SequenceNo          int     `json:"sequence_no" binding:"min=1,max=2147483647"`
	CategoryID          *string `json:"category_id" binding:"omitempty,uuid"`
	LocationTypeID      *string `json:"location_type_id" binding:"omitempty,uuid"`
	ZoneID              *string `json:"zone_id" binding:"omitempty,uuid"`
	MinimumEmptyPercent *string `json:"minimum_empty_percent"`
	IsActive            *bool   `json:"is_active" binding:"required"`
}
