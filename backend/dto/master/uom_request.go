package master

type CreateUOMRequest struct {
	Code         string `json:"code" binding:"required,max=20"`
	Name         string `json:"name" binding:"required,max=100"`
	DecimalScale int16  `json:"decimal_scale" binding:"min=0,max=6"`
}

type UpdateUOMRequest struct {
	Name         string `json:"name" binding:"required,max=100"`
	DecimalScale int16  `json:"decimal_scale" binding:"min=0,max=6"`
	IsActive     *bool  `json:"is_active" binding:"required"`
}
