package master

type CreateItemBarcodeRequest struct {
	UOMID     *string `json:"uom_id" binding:"omitempty,uuid"`
	Barcode   string  `json:"barcode" binding:"required,max=100"`
	IsPrimary bool    `json:"is_primary"`
}

type UpdateItemBarcodeRequest struct {
	UOMID    *string `json:"uom_id" binding:"omitempty,uuid"`
	Barcode  string  `json:"barcode" binding:"required,max=100"`
	IsActive *bool   `json:"is_active" binding:"required"`
}
