package master

type PickingSortMethod struct {
	ID          string  `gorm:"column:picking_sort_method_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code        string  `gorm:"column:code;size:40;not null"`
	Name        string  `gorm:"column:name;size:100;not null"`
	Description *string `gorm:"column:description;type:text"`
	IsActive    bool    `gorm:"column:is_active;not null;default:true"`
}

func (PickingSortMethod) TableName() string { return "picking_sort_method" }
