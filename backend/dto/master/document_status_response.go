package master

type DocumentStatusResponse struct {
	ID             string  `json:"status_id"`
	DocumentTypeID string  `json:"document_type_id"`
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	Description    *string `json:"description"`
	IsInitial      bool    `json:"is_initial"`
	IsFinal        bool    `json:"is_final"`
	IsCancelled    bool    `json:"is_cancelled"`
	DisplayOrder   int     `json:"display_order"`
	IsActive       bool    `json:"is_active"`
}
