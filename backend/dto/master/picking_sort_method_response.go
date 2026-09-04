package master

type PickingSortMethodResponse struct {
	ID          string  `json:"picking_sort_method_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IsActive    bool    `json:"is_active"`
}
