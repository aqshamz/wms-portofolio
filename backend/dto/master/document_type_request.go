package master

type CreateDocumentTypeRequest struct {
	Code        string  `json:"code" binding:"required,max=40"`
	Name        string  `json:"name" binding:"required,max=100"`
	ModuleCode  string  `json:"module_code" binding:"required,max=50"`
	Description *string `json:"description"`
}

type UpdateDocumentTypeRequest struct {
	Name        string  `json:"name" binding:"required,max=100"`
	ModuleCode  string  `json:"module_code" binding:"required,max=50"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active" binding:"required"`
}
