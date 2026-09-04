package master

type CreateTaskPriorityRequest struct {
	Code          string `json:"code" binding:"required,max=40"`
	Name          string `json:"name" binding:"required,max=100"`
	PriorityValue int    `json:"priority_value" binding:"min=0,max=2147483647"`
}

type UpdateTaskPriorityRequest struct {
	Name          string `json:"name" binding:"required,max=100"`
	PriorityValue int    `json:"priority_value" binding:"min=0,max=2147483647"`
	IsActive      *bool  `json:"is_active" binding:"required"`
}
