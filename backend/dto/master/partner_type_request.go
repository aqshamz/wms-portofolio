package master

type CreatePartnerTypeRequest struct {
	Code        string  `json:"code" binding:"required,max=40"`
	Name        string  `json:"name" binding:"required,max=100"`
	Description *string `json:"description"`
}

type UpdatePartnerTypeRequest struct {
	Name        string  `json:"name" binding:"required,max=100"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active" binding:"required"`
}
