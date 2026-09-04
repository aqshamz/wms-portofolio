package master

type CreateAppModuleRequest struct {
	Code         string `json:"code" binding:"required,max=50"`
	Name         string `json:"name" binding:"required,max=100"`
	DisplayOrder int    `json:"display_order" binding:"min=0,max=2147483647"`
}

type UpdateAppModuleRequest struct {
	Name         string `json:"name" binding:"required,max=100"`
	DisplayOrder int    `json:"display_order" binding:"min=0,max=2147483647"`
	IsActive     *bool  `json:"is_active" binding:"required"`
}
