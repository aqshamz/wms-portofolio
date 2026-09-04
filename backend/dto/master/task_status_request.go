package master

type CreateTaskStatusRequest struct {
	Code        string `json:"code" binding:"required,max=40"`
	Name        string `json:"name" binding:"required,max=100"`
	IsInitial   bool   `json:"is_initial"`
	IsFinal     bool   `json:"is_final"`
	IsCancelled bool   `json:"is_cancelled"`
}

type UpdateTaskStatusRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	IsInitial   bool   `json:"is_initial"`
	IsFinal     bool   `json:"is_final"`
	IsCancelled bool   `json:"is_cancelled"`
	IsActive    *bool  `json:"is_active" binding:"required"`
}
