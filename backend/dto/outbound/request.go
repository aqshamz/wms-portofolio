package outbound

type OutboundOrderLineRequest struct {
	ItemID                string  `json:"item_id" binding:"required,uuid"`
	OrderedQty            string  `json:"ordered_qty" binding:"required,max=30"`
	RequestedLotNo        *string `json:"requested_lot_no" binding:"omitempty,max=100"`
	CustomerLineReference *string `json:"customer_line_reference" binding:"omitempty,max=100"`
	Notes                 *string `json:"notes" binding:"omitempty,max=4000"`
}
type CreateOutboundOrderRequest struct {
	OwnerID               string                     `json:"owner_id" binding:"required,uuid"`
	CustomerID            string                     `json:"customer_id" binding:"required,uuid"`
	ShipToPartnerID       string                     `json:"ship_to_partner_id" binding:"required,uuid"`
	WarehouseID           string                     `json:"warehouse_id" binding:"required,uuid"`
	BusinessDate          string                     `json:"business_date" binding:"required"`
	ClientDeliveryOrderNo string                     `json:"client_delivery_order_no" binding:"required,max=120"`
	CustomerOrderNo       *string                    `json:"customer_order_no" binding:"omitempty,max=120"`
	ExternalReference     *string                    `json:"external_reference" binding:"omitempty,max=120"`
	RequestedShipAt       string                     `json:"requested_ship_at" binding:"required"`
	Notes                 *string                    `json:"notes" binding:"omitempty,max=4000"`
	Lines                 []OutboundOrderLineRequest `json:"lines" binding:"required,min=1,max=500,dive"`
}
type UpdateOutboundOrderRequest struct {
	ExpectedVersion   int64   `json:"expected_version" binding:"required,min=1"`
	ShipToPartnerID   string  `json:"ship_to_partner_id" binding:"required,uuid"`
	RequestedShipAt   string  `json:"requested_ship_at" binding:"required"`
	CustomerOrderNo   *string `json:"customer_order_no" binding:"omitempty,max=120"`
	ExternalReference *string `json:"external_reference" binding:"omitempty,max=120"`
	Notes             *string `json:"notes" binding:"omitempty,max=4000"`
}
type AddOutboundOrderLineRequest struct {
	ExpectedVersion int64 `json:"expected_version" binding:"required,min=1"`
	OutboundOrderLineRequest
}
type UpdateOutboundOrderLineRequest struct {
	ExpectedVersion       int64   `json:"expected_version" binding:"required,min=1"`
	OrderedQty            string  `json:"ordered_qty" binding:"required,max=30"`
	RequestedLotNo        *string `json:"requested_lot_no" binding:"omitempty,max=100"`
	CustomerLineReference *string `json:"customer_line_reference" binding:"omitempty,max=100"`
	Notes                 *string `json:"notes" binding:"omitempty,max=4000"`
}
type DeleteOutboundOrderLineRequest struct {
	ExpectedVersion int64 `json:"expected_version" binding:"required,min=1"`
}
type TransitionRequest struct {
	ExpectedVersion int64 `json:"expected_version" binding:"required,min=1"`
}
type CancelOrderRequest struct {
	ExpectedVersion int64  `json:"expected_version" binding:"required,min=1"`
	Reason          string `json:"reason" binding:"required,max=4000"`
}
type ValidateOutboundRequest struct {
	ExpectedVersion int64   `json:"expected_version" binding:"required,min=1"`
	Notes           *string `json:"notes" binding:"omitempty,max=4000"`
}
type AllocateOutboundRequest struct {
	ExpectedVersion   int64   `json:"expected_version" binding:"required,min=1"`
	PickingStrategyID *string `json:"picking_strategy_id" binding:"omitempty,uuid"`
	AllowPartial      bool    `json:"allow_partial"`
}
type ReleaseReservationRequest struct {
	Reason string `json:"reason" binding:"required,max=4000"`
}
type CreateWaveRequest struct {
	OwnerID           string   `json:"owner_id" binding:"required,uuid"`
	WarehouseID       string   `json:"warehouse_id" binding:"required,uuid"`
	BusinessDate      string   `json:"business_date" binding:"required"`
	WaveTypeCode      string   `json:"wave_type_code" binding:"required,max=40"`
	PickingStrategyID *string  `json:"picking_strategy_id" binding:"omitempty,uuid"`
	PlannedReleaseAt  *string  `json:"planned_release_at"`
	OutboundIDs       []string `json:"outbound_ids" binding:"required,min=1,max=1000,dive,max=120"`
	Notes             *string  `json:"notes" binding:"omitempty,max=4000"`
}
type ReleaseWaveRequest struct {
	ExpectedVersion   int64  `json:"expected_version" binding:"required,min=1"`
	StagingLocationID string `json:"staging_location_id" binding:"required,uuid"`
	PriorityCode      string `json:"priority_code" binding:"required,max=40"`
}
type StartPickRequest struct{}
type ConfirmPickRequest struct {
	PickedQty              string `json:"picked_qty" binding:"required,max=30"`
	ExpectedBalanceVersion int64  `json:"expected_balance_version" binding:"required,min=1"`
	BusinessDate           string `json:"business_date" binding:"required"`
	OperationKey           string `json:"operation_key" binding:"required,max=160"`
}
type CloseShortPickRequest struct {
	ReasonCode string  `json:"reason_code" binding:"required,max=40"`
	Notes      *string `json:"notes" binding:"omitempty,max=4000"`
}
type CreateStagingRequest struct {
	OutboundID string  `json:"outbound_id" binding:"required,max=120"`
	WaveID     string  `json:"wave_id" binding:"required,max=120"`
	Notes      *string `json:"notes" binding:"omitempty,max=4000"`
}
type CompleteStagingRequest struct{}
