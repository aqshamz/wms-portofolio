package master

type AppModule struct {
	ID           string `gorm:"column:module_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code         string `gorm:"column:code;size:50;not null"`
	Name         string `gorm:"column:name;size:100;not null"`
	DisplayOrder int    `gorm:"column:display_order;not null;default:0"`
	IsActive     bool   `gorm:"column:is_active;not null;default:true"`
}

func (AppModule) TableName() string { return "app_module" }
