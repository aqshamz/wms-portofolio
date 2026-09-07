package inventory

type CreateSerialRequest struct {
	OwnerID  string `json:"owner_id" binding:"required,uuid"`
	ItemID   string `json:"item_id" binding:"required,uuid"`
	SerialNo string `json:"serial_no" binding:"required,max=120"`
}
