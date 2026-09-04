package master

type CreateDocumentStatusRequest struct {
	Code         string  `json:"code" binding:"required,max=40"`
	Name         string  `json:"name" binding:"required,max=100"`
	Description  *string `json:"description"`
	IsInitial    bool    `json:"is_initial"`
	IsFinal      bool    `json:"is_final"`
	IsCancelled  bool    `json:"is_cancelled"`
	DisplayOrder int     `json:"display_order" binding:"min=0,max=2147483647"`
}

type UpdateDocumentStatusRequest struct {
	Name         string  `json:"name" binding:"required,max=100"`
	Description  *string `json:"description"`
	IsInitial    bool    `json:"is_initial"`
	IsFinal      bool    `json:"is_final"`
	IsCancelled  bool    `json:"is_cancelled"`
	DisplayOrder int     `json:"display_order" binding:"min=0,max=2147483647"`
	IsActive     *bool   `json:"is_active" binding:"required"`
}
