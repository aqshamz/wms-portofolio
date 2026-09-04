package master

import "time"

type DocumentDailyCounterResponse struct {
	DocumentTypeID string    `json:"document_type_id"`
	BusinessDate   string    `json:"business_date"`
	LastNumber     string    `json:"last_number"`
	UpdatedAt      time.Time `json:"updated_at"`
}
