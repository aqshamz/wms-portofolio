package master

type CreateItemUOMRequest struct {
	UOMID            string  `json:"uom_id" binding:"required,uuid"`
	ConversionToBase string  `json:"conversion_to_base" binding:"required"`
	Length           *string `json:"length"`
	Width            *string `json:"width"`
	Height           *string `json:"height"`
	Weight           *string `json:"weight"`
	IsReceivingUOM   *bool   `json:"is_receiving_uom"`
	IsPickingUOM     *bool   `json:"is_picking_uom"`
}

type UpdateItemUOMRequest struct {
	ConversionToBase string  `json:"conversion_to_base" binding:"required"`
	Length           *string `json:"length"`
	Width            *string `json:"width"`
	Height           *string `json:"height"`
	Weight           *string `json:"weight"`
	IsReceivingUOM   *bool   `json:"is_receiving_uom"`
	IsPickingUOM     *bool   `json:"is_picking_uom"`
	IsActive         *bool   `json:"is_active" binding:"required"`
}
