package master

type PutawayStrategyRule struct {
	ID                  string  `gorm:"column:rule_id;type:uuid;default:gen_random_uuid();primaryKey"`
	PutawayStrategyID   string  `gorm:"column:putaway_strategy_id;type:uuid;not null"`
	SequenceNo          int     `gorm:"column:sequence_no;not null"`
	CategoryID          *string `gorm:"column:category_id;type:uuid"`
	LocationTypeID      *string `gorm:"column:location_type_id;type:uuid"`
	ZoneID              *string `gorm:"column:zone_id;type:uuid"`
	MinimumEmptyPercent *string `gorm:"column:minimum_empty_percent;type:numeric(7,4)"`
	IsActive            bool    `gorm:"column:is_active;not null;default:true"`
}

func (PutawayStrategyRule) TableName() string { return "putaway_strategy_rule" }
