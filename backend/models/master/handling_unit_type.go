package master

type HandlingUnitType struct {
	ID        string  `gorm:"column:handling_unit_type_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code      string  `gorm:"column:code;size:40;not null"`
	Name      string  `gorm:"column:name;size:100;not null"`
	MaxWeight *string `gorm:"column:max_weight;type:numeric(20,6)"`
	MaxVolume *string `gorm:"column:max_volume;type:numeric(20,6)"`
	IsActive  bool    `gorm:"column:is_active;not null;default:true"`
}

func (HandlingUnitType) TableName() string { return "handling_unit_type" }
