package master

// InspectionResultResponse is the public JSON contract, separate from the GORM entity.
type InspectionResultResponse struct {
	ID          string  `json:"inspection_result_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IsAccepted  bool    `json:"is_accepted"`
	IsActive    bool    `json:"is_active"`
}
