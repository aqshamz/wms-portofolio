package master

import "time"

// ItemResponse is the public JSON contract, separate from the GORM entity.
type ItemResponse struct {
	ID                 string    `json:"item_id"`
	OwnerID            string    `json:"owner_id"`
	CategoryID         *string   `json:"category_id"`
	Code               string    `json:"code"`
	Name               string    `json:"name"`
	Description        *string   `json:"description"`
	BaseUOMID          string    `json:"base_uom_id"`
	Weight             *string   `json:"weight"`
	Volume             *string   `json:"volume"`
	LotControlled      bool      `json:"lot_controlled"`
	SerialControlled   bool      `json:"serial_controlled"`
	ShelfLifeDays      *int      `json:"shelf_life_days"`
	MinimumReceiveDays *int      `json:"minimum_receive_days"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	CreatedBy          *string   `json:"created_by"`
	UpdatedAt          time.Time `json:"updated_at"`
	UpdatedBy          *string   `json:"updated_by"`
}
