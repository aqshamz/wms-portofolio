package master

type ItemUOM struct {
	ID               string  `gorm:"column:item_uom_id;type:uuid;default:gen_random_uuid();primaryKey"`
	ItemID           string  `gorm:"column:item_id;type:uuid;not null"`
	UOMID            string  `gorm:"column:uom_id;type:uuid;not null"`
	ConversionToBase string  `gorm:"column:conversion_to_base;type:numeric(20,6);not null"`
	Length           *string `gorm:"column:length;type:numeric(20,6)"`
	Width            *string `gorm:"column:width;type:numeric(20,6)"`
	Height           *string `gorm:"column:height;type:numeric(20,6)"`
	Weight           *string `gorm:"column:weight;type:numeric(20,6)"`
	IsReceivingUOM   bool    `gorm:"column:is_receiving_uom;not null;default:true"`
	IsPickingUOM     bool    `gorm:"column:is_picking_uom;not null;default:true"`
	IsActive         bool    `gorm:"column:is_active;not null;default:true"`
}

func (ItemUOM) TableName() string { return "item_uom" }
