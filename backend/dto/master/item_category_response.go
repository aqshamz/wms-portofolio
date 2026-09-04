package master

// ItemCategoryResponse is the public JSON contract, separate from the GORM entity.
type ItemCategoryResponse struct {
	ID               string  `json:"category_id"`
	OwnerID          string  `json:"owner_id"`
	ParentCategoryID *string `json:"parent_category_id"`
	Code             string  `json:"code"`
	Name             string  `json:"name"`
	IsActive         bool    `json:"is_active"`
}
