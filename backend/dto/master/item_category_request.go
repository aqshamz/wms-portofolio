package master

type CreateItemCategoryRequest struct {
	OwnerID          string  `json:"owner_id" binding:"required,uuid"`
	ParentCategoryID *string `json:"parent_category_id" binding:"omitempty,uuid"`
	Code             string  `json:"code" binding:"required,max=40"`
	Name             string  `json:"name" binding:"required,max=100"`
}

type UpdateItemCategoryRequest struct {
	ParentCategoryID *string `json:"parent_category_id" binding:"omitempty,uuid"`
	Name             string  `json:"name" binding:"required,max=100"`
	IsActive         *bool   `json:"is_active" binding:"required"`
}
