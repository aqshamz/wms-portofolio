package master

type ItemCategory struct {
	ID               string  `gorm:"column:category_id;type:uuid;default:gen_random_uuid();primaryKey"`
	OwnerID          string  `gorm:"column:owner_id;type:uuid;not null"`
	ParentCategoryID *string `gorm:"column:parent_category_id;type:uuid"`
	Code             string  `gorm:"column:code;size:40;not null"`
	Name             string  `gorm:"column:name;size:100;not null"`
	IsActive         bool    `gorm:"column:is_active;not null;default:true"`
}

func (ItemCategory) TableName() string { return "item_category" }
