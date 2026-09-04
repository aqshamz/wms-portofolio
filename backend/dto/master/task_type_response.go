package master

type TaskTypeResponse struct {
	ID          string  `json:"task_type_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IsActive    bool    `json:"is_active"`
}
