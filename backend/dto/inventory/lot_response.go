package inventory

import "time"

type LotResponse struct {
	ID              string    `json:"lot_id"`
	OwnerID         string    `json:"owner_id"`
	ItemID          string    `json:"item_id"`
	LotNumber       string    `json:"lot_number"`
	ManufactureDate *string   `json:"manufacture_date"`
	ExpiryDate      *string   `json:"expiry_date"`
	QualityStatusID *string   `json:"quality_status_id"`
	CreatedAt       time.Time `json:"created_at"`
	CreatedBy       *string   `json:"created_by"`
}
