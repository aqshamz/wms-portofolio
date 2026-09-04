package master

// ItemBarcodeResponse is the public JSON contract, separate from the GORM entity.
type ItemBarcodeResponse struct {
	ID        string  `json:"item_barcode_id"`
	ItemID    string  `json:"item_id"`
	UOMID     *string `json:"uom_id"`
	Barcode   string  `json:"barcode"`
	IsPrimary bool    `json:"is_primary"`
	IsActive  bool    `json:"is_active"`
}
