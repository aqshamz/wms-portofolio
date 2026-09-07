package inventory

type MovementType struct {
	ID          string  `gorm:"column:movement_type_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code        string  `gorm:"column:code;size:40;not null"`
	Name        string  `gorm:"column:name;size:100;not null"`
	Description *string `gorm:"column:description;type:text"`
	IsActive    bool    `gorm:"column:is_active;not null;default:true"`
}

func (MovementType) TableName() string { return "movement_type" }
