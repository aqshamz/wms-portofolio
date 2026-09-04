package master

type AssignWarehouseOwnerRequest struct {
	OwnerID string `json:"owner_id" binding:"required"`
}

type CreateLocationTypeRequest struct {
	Code            string  `json:"code" binding:"required,max=40"`
	Name            string  `json:"name" binding:"required,max=100"`
	Description     *string `json:"description"`
	AllowsReceiving bool    `json:"allows_receiving"`
	AllowsStorage   bool    `json:"allows_storage"`
	AllowsPicking   bool    `json:"allows_picking"`
	AllowsShipping  bool    `json:"allows_shipping"`
}

type UpdateLocationTypeRequest struct {
	Name            string  `json:"name" binding:"required,max=100"`
	Description     *string `json:"description"`
	AllowsReceiving bool    `json:"allows_receiving"`
	AllowsStorage   bool    `json:"allows_storage"`
	AllowsPicking   bool    `json:"allows_picking"`
	AllowsShipping  bool    `json:"allows_shipping"`
	IsActive        bool    `json:"is_active"`
}

type CreateWarehouseZoneRequest struct {
	Code        string  `json:"code" binding:"required,max=40"`
	Name        string  `json:"name" binding:"required,max=100"`
	Description *string `json:"description"`
}

type UpdateWarehouseZoneRequest struct {
	Name        string  `json:"name" binding:"required,max=100"`
	Description *string `json:"description"`
	IsActive    bool    `json:"is_active"`
}

type CreateWarehouseLocationRequest struct {
	ZoneID         string  `json:"zone_id" binding:"required"`
	LocationTypeID string  `json:"location_type_id" binding:"required"`
	Code           string  `json:"code" binding:"required,max=60"`
	Barcode        *string `json:"barcode" binding:"omitempty,max=100"`
	Aisle          *string `json:"aisle" binding:"omitempty,max=20"`
	Bay            *string `json:"bay" binding:"omitempty,max=20"`
	LevelNo        *string `json:"level_no" binding:"omitempty,max=20"`
	PositionNo     *string `json:"position_no" binding:"omitempty,max=20"`
	PickSequence   int     `json:"pick_sequence" binding:"min=0"`
	MaxWeight      *string `json:"max_weight"`
	MaxVolume      *string `json:"max_volume"`
	IsPickFace     bool    `json:"is_pick_face"`
}

type UpdateWarehouseLocationRequest struct {
	ZoneID         string  `json:"zone_id" binding:"required"`
	LocationTypeID string  `json:"location_type_id" binding:"required"`
	Barcode        *string `json:"barcode" binding:"omitempty,max=100"`
	Aisle          *string `json:"aisle" binding:"omitempty,max=20"`
	Bay            *string `json:"bay" binding:"omitempty,max=20"`
	LevelNo        *string `json:"level_no" binding:"omitempty,max=20"`
	PositionNo     *string `json:"position_no" binding:"omitempty,max=20"`
	PickSequence   int     `json:"pick_sequence" binding:"min=0"`
	MaxWeight      *string `json:"max_weight"`
	MaxVolume      *string `json:"max_volume"`
	IsPickFace     bool    `json:"is_pick_face"`
	IsLocked       bool    `json:"is_locked"`
	IsActive       bool    `json:"is_active"`
}

type GrantAccountAccessRequest struct {
	AccountID string `json:"account_id" binding:"required"`
}
