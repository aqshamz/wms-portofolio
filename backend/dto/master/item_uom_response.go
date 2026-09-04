package master

// ItemUOMResponse is the public JSON contract, separate from the GORM entity.
type ItemUOMResponse struct {
	ID               string  `json:"item_uom_id"`
	ItemID           string  `json:"item_id"`
	UOMID            string  `json:"uom_id"`
	ConversionToBase string  `json:"conversion_to_base"`
	Length           *string `json:"length"`
	Width            *string `json:"width"`
	Height           *string `json:"height"`
	Weight           *string `json:"weight"`
	IsReceivingUOM   bool    `json:"is_receiving_uom"`
	IsPickingUOM     bool    `json:"is_picking_uom"`
	IsActive         bool    `json:"is_active"`
}
