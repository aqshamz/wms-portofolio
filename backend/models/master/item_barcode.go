package master

type ItemBarcode struct {
	ID        string  `gorm:"column:item_barcode_id;type:uuid;default:gen_random_uuid();primaryKey"`
	ItemID    string  `gorm:"column:item_id;type:uuid;not null"`
	UOMID     *string `gorm:"column:uom_id;type:uuid"`
	Barcode   string  `gorm:"column:barcode;size:100;not null"`
	IsPrimary bool    `gorm:"column:is_primary;not null;default:false"`
	IsActive  bool    `gorm:"column:is_active;not null;default:true"`
}

func (ItemBarcode) TableName() string { return "item_barcode" }
