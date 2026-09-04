package master

// InventoryStatusResponse is the public JSON contract, separate from the GORM entity.
type InventoryStatusResponse struct {
	ID            string  `json:"inventory_status_id"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Description   *string `json:"description"`
	IsAllocatable bool    `json:"is_allocatable"`
	IsPickable    bool    `json:"is_pickable"`
	IsActive      bool    `json:"is_active"`
}
