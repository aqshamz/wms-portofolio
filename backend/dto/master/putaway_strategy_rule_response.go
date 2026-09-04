package master

type PutawayStrategyRuleResponse struct {
	ID                  string  `json:"rule_id"`
	PutawayStrategyID   string  `json:"putaway_strategy_id"`
	SequenceNo          int     `json:"sequence_no"`
	CategoryID          *string `json:"category_id"`
	LocationTypeID      *string `json:"location_type_id"`
	ZoneID              *string `json:"zone_id"`
	MinimumEmptyPercent *string `json:"minimum_empty_percent"`
	IsActive            bool    `json:"is_active"`
}
