package stockcontrol

import inventory "wms-api/dto/inventory"

type CommandResponse struct {
	Operation          string                       `json:"operation"`
	Movements          []inventory.MovementResponse `json:"movements"`
	SourceBalance      *inventory.BalanceResponse   `json:"source_balance,omitempty"`
	DestinationBalance *inventory.BalanceResponse   `json:"destination_balance,omitempty"`
	IdempotentReplay   bool                         `json:"idempotent_replay"`
	NoVariance         bool                         `json:"no_variance,omitempty"`
}
