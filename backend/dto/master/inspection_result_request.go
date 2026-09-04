package master

type CreateInspectionResultRequest struct {
	Code        string  `json:"code" binding:"required,max=40"`
	Name        string  `json:"name" binding:"required,max=100"`
	Description *string `json:"description"`
	IsAccepted  bool    `json:"is_accepted"`
}

type UpdateInspectionResultRequest struct {
	Name        string  `json:"name" binding:"required,max=100"`
	Description *string `json:"description"`
	IsAccepted  bool    `json:"is_accepted"`
	IsActive    *bool   `json:"is_active" binding:"required"`
}
