package master

type PickingStrategy struct {
	ID          string  `gorm:"column:picking_strategy_id;type:uuid;default:gen_random_uuid();primaryKey"`
	OwnerID     *string `gorm:"column:owner_id;type:uuid"`
	WarehouseID *string `gorm:"column:warehouse_id;type:uuid"`
	Code        string  `gorm:"column:code;size:40;not null"`
	Name        string  `gorm:"column:name;size:100;not null"`
	Description *string `gorm:"column:description;type:text"`
	IsActive    bool    `gorm:"column:is_active;not null;default:true"`
}

func (PickingStrategy) TableName() string { return "picking_strategy" }
