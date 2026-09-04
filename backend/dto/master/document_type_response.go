package master

type DocumentTypeResponse struct {
	ID          string  `json:"document_type_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	ModuleCode  string  `json:"module_code"`
	Description *string `json:"description"`
	IsActive    bool    `json:"is_active"`
}
