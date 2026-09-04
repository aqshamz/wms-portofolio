package master

import "time"

type WarehouseLocation struct {
	ID             string    `gorm:"column:location_id;type:uuid;default:gen_random_uuid();primaryKey;uniqueIndex:uq_location_warehouse,priority:1" json:"location_id"`
	WarehouseID    string    `gorm:"column:warehouse_id;type:uuid;not null;uniqueIndex:uq_warehouse_location_code,priority:1;uniqueIndex:uq_warehouse_location_barcode,priority:1;uniqueIndex:uq_location_warehouse,priority:2" json:"warehouse_id"`
	ZoneID         string    `gorm:"column:zone_id;type:uuid;not null;index" json:"zone_id"`
	LocationTypeID string    `gorm:"column:location_type_id;type:uuid;not null;index" json:"location_type_id"`
	Code           string    `gorm:"column:code;size:60;not null;uniqueIndex:uq_warehouse_location_code,priority:2" json:"code"`
	Barcode        *string   `gorm:"column:barcode;size:100;uniqueIndex:uq_warehouse_location_barcode,priority:2" json:"barcode,omitempty"`
	Aisle          *string   `gorm:"column:aisle;size:20" json:"aisle,omitempty"`
	Bay            *string   `gorm:"column:bay;size:20" json:"bay,omitempty"`
	LevelNo        *string   `gorm:"column:level_no;size:20" json:"level_no,omitempty"`
	PositionNo     *string   `gorm:"column:position_no;size:20" json:"position_no,omitempty"`
	PickSequence   int       `gorm:"column:pick_sequence;not null;default:0" json:"pick_sequence"`
	MaxWeight      *string   `gorm:"column:max_weight;type:numeric(20,6)" json:"max_weight,omitempty"`
	MaxVolume      *string   `gorm:"column:max_volume;type:numeric(20,6)" json:"max_volume,omitempty"`
	IsPickFace     bool      `gorm:"column:is_pick_face;not null;default:false" json:"is_pick_face"`
	IsLocked       bool      `gorm:"column:is_locked;not null;default:false" json:"is_locked"`
	IsActive       bool      `gorm:"column:is_active;not null;default:true" json:"is_active"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:clock_timestamp()" json:"created_at"`
	CreatedBy      *string   `gorm:"column:created_by;type:uuid" json:"created_by,omitempty"`
}

func (WarehouseLocation) TableName() string { return "warehouse_location" }
