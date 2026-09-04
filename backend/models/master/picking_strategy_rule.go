package master

type PickingStrategyRule struct {
	ID                  string  `gorm:"column:rule_id;type:uuid;default:gen_random_uuid();primaryKey"`
	PickingStrategyID   string  `gorm:"column:picking_strategy_id;type:uuid;not null"`
	SequenceNo          int     `gorm:"column:sequence_no;not null"`
	InventoryStatusID   *string `gorm:"column:inventory_status_id;type:uuid"`
	ZoneID              *string `gorm:"column:zone_id;type:uuid"`
	PickingSortMethodID string  `gorm:"column:picking_sort_method_id;type:uuid;not null"`
	IsActive            bool    `gorm:"column:is_active;not null;default:true"`
}

func (PickingStrategyRule) TableName() string { return "picking_strategy_rule" }
