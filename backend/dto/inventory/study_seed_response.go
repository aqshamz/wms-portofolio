package inventory

type StudySeedResponse struct {
	OwnerID             string             `json:"owner_id"`
	PrimaryWarehouseID  string             `json:"primary_warehouse_id"`
	TransferWarehouseID string             `json:"transfer_warehouse_id"`
	Locations           map[string]string  `json:"locations"`
	Items               map[string]string  `json:"items"`
	Lots                map[string]string  `json:"lots"`
	HandlingUnits       map[string]string  `json:"handling_units"`
	Serials             map[string]string  `json:"serials"`
	Balances            []BalanceResponse  `json:"balances"`
	Movements           []MovementResponse `json:"opening_movements"`
	CreatedIdentities   int                `json:"created_identities"`
	ReusedIdentities    int                `json:"reused_identities"`
	PostedMovements     int                `json:"posted_movements"`
	ReplayedMovements   int                `json:"replayed_movements"`
}
