package master

type InventoryStatus struct {
	ID            string  `gorm:"column:inventory_status_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code          string  `gorm:"column:code;size:40;not null"`
	Name          string  `gorm:"column:name;size:100;not null"`
	Description   *string `gorm:"column:description;type:text"`
	IsAllocatable bool    `gorm:"column:is_allocatable;not null;default:false"`
	IsPickable    bool    `gorm:"column:is_pickable;not null;default:false"`
	IsActive      bool    `gorm:"column:is_active;not null;default:true"`
}

func (InventoryStatus) TableName() string { return "inventory_status" }
