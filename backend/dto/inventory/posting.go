package inventory

// PostingRequest is an internal service contract for stock-control, inbound and
// outbound. It is intentionally not bound to a generic public HTTP endpoint.
type PostingRequest struct {
	OperationKey               string
	MovementTypeCode           string
	OwnerID                    string
	WarehouseID                string
	BusinessDate               string
	ItemID                     string
	LotID                      *string
	HandlingUnitID             *string
	From                       *BalanceDimension
	To                         *BalanceDimension
	Quantity                   string
	SerialIDs                  []string
	ExpectedSourceVersion      *int64
	ExpectedDestinationVersion *int64
	SourceDocumentID           string
	SourceLineID               *string
	ReasonCodeID               *string
	Notes                      *string
	RelocateHandlingUnit       bool
}
type BalanceDimension struct {
	LocationID        string
	InventoryStatusID string
}
type PostingResult struct {
	Movement         MovementResponse
	FromBalance      *BalanceResponse
	ToBalance        *BalanceResponse
	IdempotentReplay bool
}
