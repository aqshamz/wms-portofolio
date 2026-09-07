package stockcontrol

type ReasonCodeResponse struct {
	ID           string  `json:"reason_code_id"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	RequiresNote bool    `json:"requires_note"`
	IsActive     bool    `json:"is_active"`
}
