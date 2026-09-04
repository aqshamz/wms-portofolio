package master

import "time"

type Item struct {
	ID                 string    `gorm:"column:item_id;type:uuid;default:gen_random_uuid();primaryKey"`
	OwnerID            string    `gorm:"column:owner_id;type:uuid;not null"`
	CategoryID         *string   `gorm:"column:category_id;type:uuid"`
	Code               string    `gorm:"column:code;size:60;not null"`
	Name               string    `gorm:"column:name;size:200;not null"`
	Description        *string   `gorm:"column:description;type:text"`
	BaseUOMID          string    `gorm:"column:base_uom_id;type:uuid;not null"`
	Weight             *string   `gorm:"column:weight;type:numeric(20,6)"`
	Volume             *string   `gorm:"column:volume;type:numeric(20,6)"`
	LotControlled      bool      `gorm:"column:lot_controlled;not null;default:false"`
	SerialControlled   bool      `gorm:"column:serial_controlled;not null;default:false"`
	ShelfLifeDays      *int      `gorm:"column:shelf_life_days;type:integer"`
	MinimumReceiveDays *int      `gorm:"column:minimum_receive_days;type:integer"`
	IsActive           bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt          time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy          *string   `gorm:"column:created_by;type:uuid"`
	UpdatedAt          time.Time `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy          *string   `gorm:"column:updated_by;type:uuid"`
}

func (Item) TableName() string { return "item" }
