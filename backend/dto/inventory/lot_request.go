package inventory

type CreateLotRequest struct {
	OwnerID         string  `json:"owner_id" binding:"required,uuid"`
	ItemID          string  `json:"item_id" binding:"required,uuid"`
	LotNumber       string  `json:"lot_number" binding:"required,max=100"`
	ManufactureDate *string `json:"manufacture_date"`
	ExpiryDate      *string `json:"expiry_date"`
	QualityStatusID *string `json:"quality_status_id" binding:"omitempty,uuid"`
}
