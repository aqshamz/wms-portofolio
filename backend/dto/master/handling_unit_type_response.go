package master

type HandlingUnitTypeResponse struct {
	ID        string  `json:"handling_unit_type_id"`
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	MaxWeight *string `json:"max_weight"`
	MaxVolume *string `json:"max_volume"`
	IsActive  bool    `json:"is_active"`
}
