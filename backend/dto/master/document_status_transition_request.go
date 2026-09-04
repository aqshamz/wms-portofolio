package master

type CreateDocumentStatusTransitionRequest struct {
	FromStatusID         string  `json:"from_status_id" binding:"required,uuid"`
	ToStatusID           string  `json:"to_status_id" binding:"required,uuid"`
	RequiredPermissionID *string `json:"required_permission_id" binding:"omitempty,uuid"`
}

type UpdateDocumentStatusTransitionRequest struct {
	RequiredPermissionID *string `json:"required_permission_id" binding:"omitempty,uuid"`
	IsActive             *bool   `json:"is_active" binding:"required"`
}
