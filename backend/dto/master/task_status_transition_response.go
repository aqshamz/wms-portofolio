package master

type TaskStatusTransitionResponse struct {
	ID                   string  `json:"task_status_transition_id"`
	FromStatusID         string  `json:"from_status_id"`
	ToStatusID           string  `json:"to_status_id"`
	RequiredPermissionID *string `json:"required_permission_id"`
	IsActive             bool    `json:"is_active"`
}
