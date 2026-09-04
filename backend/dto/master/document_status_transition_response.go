package master

type DocumentStatusTransitionResponse struct {
	ID                   string  `json:"transition_id"`
	DocumentTypeID       string  `json:"document_type_id"`
	FromStatusID         string  `json:"from_status_id"`
	ToStatusID           string  `json:"to_status_id"`
	RequiredPermissionID *string `json:"required_permission_id"`
	IsActive             bool    `json:"is_active"`
}
