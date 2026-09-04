package master

// QualityStatusResponse is the public JSON contract, separate from the GORM entity.
type QualityStatusResponse struct {
	ID          string  `json:"quality_status_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IsActive    bool    `json:"is_active"`
}
