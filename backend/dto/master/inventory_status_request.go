package master

type CreateInventoryStatusRequest struct {
	Code          string  `json:"code" binding:"required,max=40"`
	Name          string  `json:"name" binding:"required,max=100"`
	Description   *string `json:"description"`
	IsAllocatable bool    `json:"is_allocatable"`
	IsPickable    bool    `json:"is_pickable"`
}

type UpdateInventoryStatusRequest struct {
	Name          string  `json:"name" binding:"required,max=100"`
	Description   *string `json:"description"`
	IsAllocatable bool    `json:"is_allocatable"`
	IsPickable    bool    `json:"is_pickable"`
	IsActive      *bool   `json:"is_active" binding:"required"`
}
