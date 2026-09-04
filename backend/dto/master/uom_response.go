package master

// UOMResponse is the public JSON contract, separate from the GORM entity.
type UOMResponse struct {
	ID           string `json:"uom_id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	DecimalScale int16  `json:"decimal_scale"`
	IsActive     bool   `json:"is_active"`
}
