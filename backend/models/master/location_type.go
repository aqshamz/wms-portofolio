package master

type LocationType struct {
	ID              string  `gorm:"column:location_type_id;type:uuid;default:gen_random_uuid();primaryKey" json:"location_type_id"`
	Code            string  `gorm:"column:code;size:40;not null;uniqueIndex" json:"code"`
	Name            string  `gorm:"column:name;size:100;not null" json:"name"`
	Description     *string `gorm:"column:description;type:text" json:"description,omitempty"`
	AllowsReceiving bool    `gorm:"column:allows_receiving;not null;default:false" json:"allows_receiving"`
	AllowsStorage   bool    `gorm:"column:allows_storage;not null;default:false" json:"allows_storage"`
	AllowsPicking   bool    `gorm:"column:allows_picking;not null;default:false" json:"allows_picking"`
	AllowsShipping  bool    `gorm:"column:allows_shipping;not null;default:false" json:"allows_shipping"`
	IsActive        bool    `gorm:"column:is_active;not null;default:true" json:"is_active"`
}

func (LocationType) TableName() string { return "location_type" }
