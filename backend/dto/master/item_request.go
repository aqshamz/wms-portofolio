package master

import "time"

type CreateItemRequest struct {
	OwnerID            string  `json:"owner_id" binding:"required,uuid"`
	CategoryID         *string `json:"category_id" binding:"omitempty,uuid"`
	Code               string  `json:"code" binding:"required,max=60"`
	Name               string  `json:"name" binding:"required,max=200"`
	Description        *string `json:"description"`
	BaseUOMID          string  `json:"base_uom_id" binding:"required,uuid"`
	Weight             *string `json:"weight"`
	Volume             *string `json:"volume"`
	LotControlled      bool    `json:"lot_controlled"`
	SerialControlled   bool    `json:"serial_controlled"`
	ShelfLifeDays      *int    `json:"shelf_life_days" binding:"omitempty,min=0"`
	MinimumReceiveDays *int    `json:"minimum_receive_days" binding:"omitempty,min=0"`
}

type UpdateItemRequest struct {
	CategoryID         *string   `json:"category_id" binding:"omitempty,uuid"`
	Name               string    `json:"name" binding:"required,max=200"`
	Description        *string   `json:"description"`
	Weight             *string   `json:"weight"`
	Volume             *string   `json:"volume"`
	LotControlled      bool      `json:"lot_controlled"`
	SerialControlled   bool      `json:"serial_controlled"`
	ShelfLifeDays      *int      `json:"shelf_life_days" binding:"omitempty,min=0"`
	MinimumReceiveDays *int      `json:"minimum_receive_days" binding:"omitempty,min=0"`
	IsActive           *bool     `json:"is_active" binding:"required"`
	ExpectedUpdatedAt  time.Time `json:"expected_updated_at" binding:"required"`
}
