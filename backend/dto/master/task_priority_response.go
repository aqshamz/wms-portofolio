package master

type TaskPriorityResponse struct {
	ID            string `json:"task_priority_id"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	PriorityValue int    `json:"priority_value"`
	IsActive      bool   `json:"is_active"`
}
