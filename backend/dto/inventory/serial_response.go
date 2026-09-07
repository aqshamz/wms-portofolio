package inventory

import "time"

type SerialResponse struct {
	ID        string    `json:"serial_id"`
	OwnerID   string    `json:"owner_id"`
	ItemID    string    `json:"item_id"`
	SerialNo  string    `json:"serial_no"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy *string   `json:"created_by"`
}
