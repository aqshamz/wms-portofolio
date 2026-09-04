package master

import "time"

type WarehouseOwnerResponse struct {
	WarehouseID string    `json:"warehouse_id"`
	OwnerID     string    `json:"owner_id"`
	OwnerCode   string    `json:"owner_code"`
	OwnerName   string    `json:"owner_name"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type WarehouseZoneResponse struct {
	ID            string    `json:"zone_id"`
	WarehouseID   string    `json:"warehouse_id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Description   *string   `json:"description,omitempty"`
	IsActive      bool      `json:"is_active"`
	LocationCount int64     `json:"location_count"`
	CreatedAt     time.Time `json:"created_at"`
}

type WarehouseLocationResponse struct {
	ID               string    `json:"location_id"`
	WarehouseID      string    `json:"warehouse_id"`
	WarehouseCode    string    `json:"warehouse_code,omitempty"`
	WarehouseName    string    `json:"warehouse_name,omitempty"`
	ZoneID           string    `json:"zone_id"`
	ZoneCode         string    `json:"zone_code,omitempty"`
	ZoneName         string    `json:"zone_name,omitempty"`
	LocationTypeID   string    `json:"location_type_id"`
	LocationTypeCode string    `json:"location_type_code,omitempty"`
	LocationTypeName string    `json:"location_type_name,omitempty"`
	Code             string    `json:"code"`
	Barcode          *string   `json:"barcode,omitempty"`
	Aisle            *string   `json:"aisle,omitempty"`
	Bay              *string   `json:"bay,omitempty"`
	LevelNo          *string   `json:"level_no,omitempty"`
	PositionNo       *string   `json:"position_no,omitempty"`
	PickSequence     int       `json:"pick_sequence"`
	MaxWeight        *string   `json:"max_weight,omitempty"`
	MaxVolume        *string   `json:"max_volume,omitempty"`
	IsPickFace       bool      `json:"is_pick_face"`
	IsLocked         bool      `json:"is_locked"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
}

type AccountOwnerAccessResponse struct {
	AccountID string    `json:"account_id"`
	OwnerID   string    `json:"owner_id"`
	OwnerCode string    `json:"owner_code"`
	OwnerName string    `json:"owner_name"`
	GrantedBy *string   `json:"granted_by,omitempty"`
	GrantedAt time.Time `json:"granted_at"`
}

type AccountWarehouseAccessResponse struct {
	AccountID     string    `json:"account_id"`
	WarehouseID   string    `json:"warehouse_id"`
	WarehouseCode string    `json:"warehouse_code"`
	WarehouseName string    `json:"warehouse_name"`
	GrantedBy     *string   `json:"granted_by,omitempty"`
	GrantedAt     time.Time `json:"granted_at"`
}
