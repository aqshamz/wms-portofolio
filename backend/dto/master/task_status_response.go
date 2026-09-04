package master

type TaskStatusResponse struct {
	ID          string `json:"task_status_id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	IsInitial   bool   `json:"is_initial"`
	IsFinal     bool   `json:"is_final"`
	IsCancelled bool   `json:"is_cancelled"`
	IsActive    bool   `json:"is_active"`
}
