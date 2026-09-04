package master

type AppModuleResponse struct {
	ID           string `json:"module_id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	DisplayOrder int    `json:"display_order"`
	IsActive     bool   `json:"is_active"`
}
