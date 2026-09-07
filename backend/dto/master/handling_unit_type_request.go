package master

type CreateHandlingUnitTypeRequest struct {
	Code      string  `json:"code" binding:"required,max=40"`
	Name      string  `json:"name" binding:"required,max=100"`
	MaxWeight *string `json:"max_weight"`
	MaxVolume *string `json:"max_volume"`
}
type UpdateHandlingUnitTypeRequest struct {
	Name      string  `json:"name" binding:"required,max=100"`
	MaxWeight *string `json:"max_weight"`
	MaxVolume *string `json:"max_volume"`
	IsActive  *bool   `json:"is_active" binding:"required"`
}
