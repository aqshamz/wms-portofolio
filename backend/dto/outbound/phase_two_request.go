package outbound

type CreateCheckRequest struct {
	StagingID     string  `json:"staging_id" binding:"required,max=120"`
	ParentCheckID *string `json:"parent_check_id" binding:"omitempty,max=120"`
	Notes         *string `json:"notes" binding:"omitempty,max=4000"`
}

type RecordCheckLineRequest struct {
	CheckedQty string  `json:"checked_qty" binding:"required,max=30"`
	ResultCode string  `json:"result_code" binding:"required,max=40"`
	Notes      *string `json:"notes" binding:"omitempty,max=4000"`
}

type ResolveCheckExceptionRequest struct {
	ResolutionTypeCode string  `json:"resolution_type_code" binding:"required,max=40"`
	ResolvedQty        string  `json:"resolved_qty" binding:"required,max=30"`
	MovementID         *string `json:"movement_id" binding:"omitempty,max=140"`
	ReservationID      *string `json:"reservation_id" binding:"omitempty,max=140"`
	Approved           bool    `json:"approved"`
	Notes              *string `json:"notes" binding:"omitempty,max=4000"`
}

type CreateReplacementPickRequest struct {
	BalanceID              string  `json:"balance_id" binding:"required,max=160"`
	ReplacementQty         string  `json:"replacement_qty" binding:"required,max=30"`
	ExpectedBalanceVersion int64   `json:"expected_balance_version" binding:"required,min=1"`
	PickingStrategyID      *string `json:"picking_strategy_id" binding:"omitempty,uuid"`
	PriorityCode           string  `json:"priority_code" binding:"required,max=40"`
	Notes                  *string `json:"notes" binding:"omitempty,max=4000"`
}

type CreatePackingRequest struct {
	OutboundCheckID   string `json:"outbound_check_id" binding:"required,max=120"`
	PackingLocationID string `json:"packing_location_id" binding:"required,uuid"`
}

type PackLineRequest struct {
	ExpectedBalanceVersion int64   `json:"expected_balance_version" binding:"required,min=1"`
	HandlingUnitID         *string `json:"handling_unit_id" binding:"omitempty,max=120"`
	BusinessDate           string  `json:"business_date" binding:"required"`
	OperationKey           string  `json:"operation_key" binding:"required,max=160"`
}

type CreateShipmentRequest struct {
	OwnerID          string   `json:"owner_id" binding:"required,uuid"`
	WarehouseID      string   `json:"warehouse_id" binding:"required,uuid"`
	BusinessDate     string   `json:"business_date" binding:"required"`
	PackingIDs       []string `json:"packing_ids" binding:"required,min=1,max=1000,dive,max=120"`
	CarrierServiceID *string  `json:"carrier_service_id" binding:"omitempty,uuid"`
	RouteReference   *string  `json:"route_reference" binding:"omitempty,max=100"`
	TrackingNumber   *string  `json:"tracking_number" binding:"omitempty,max=150"`
	VehicleNumber    *string  `json:"vehicle_number" binding:"omitempty,max=60"`
	SealNumber       *string  `json:"seal_number" binding:"omitempty,max=60"`
	Notes            *string  `json:"notes" binding:"omitempty,max=4000"`
}

type DispatchShipmentLineRequest struct {
	ExpectedBalanceVersion int64  `json:"expected_balance_version" binding:"required,min=1"`
	BusinessDate           string `json:"business_date" binding:"required"`
	OperationKey           string `json:"operation_key" binding:"required,max=160"`
}

type CreateDeliveryRequest struct {
	ShipmentID        string  `json:"shipment_id" binding:"required,max=120"`
	OutboundID        string  `json:"outbound_id" binding:"required,max=120"`
	BusinessDate      string  `json:"business_date" binding:"required"`
	PlannedDeliveryAt *string `json:"planned_delivery_at"`
	Notes             *string `json:"notes" binding:"omitempty,max=4000"`
}

type DeliveryEventRequest struct {
	EventAt   string  `json:"event_at" binding:"required"`
	Latitude  *string `json:"latitude" binding:"omitempty,max=20"`
	Longitude *string `json:"longitude" binding:"omitempty,max=20"`
	Notes     *string `json:"notes" binding:"omitempty,max=4000"`
}

type DeliverLineRequest struct {
	DeliveredQty       string  `json:"delivered_qty" binding:"required,max=30"`
	EventAt            string  `json:"event_at" binding:"required"`
	RecipientName      string  `json:"recipient_name" binding:"required,max=150"`
	RecipientReference *string `json:"recipient_reference" binding:"omitempty,max=100"`
	ProofReference     string  `json:"proof_reference" binding:"required,max=200"`
	ProofURI           *string `json:"proof_uri" binding:"omitempty,max=2000"`
	Latitude           *string `json:"latitude" binding:"omitempty,max=20"`
	Longitude          *string `json:"longitude" binding:"omitempty,max=20"`
	Notes              *string `json:"notes" binding:"omitempty,max=4000"`
}

type CompleteDeliveryRequest struct {
	DeliveredAt string `json:"delivered_at" binding:"required"`
}

type FailDeliveryRequest struct {
	ReasonCode string  `json:"reason_code" binding:"required,max=40"`
	EventAt    string  `json:"event_at" binding:"required"`
	Latitude   *string `json:"latitude" binding:"omitempty,max=20"`
	Longitude  *string `json:"longitude" binding:"omitempty,max=20"`
	Notes      *string `json:"notes" binding:"omitempty,max=4000"`
}

type UpsertReturnPolicyRequest struct {
	OwnerID                 string `json:"owner_id" binding:"required,uuid"`
	WarehouseID             string `json:"warehouse_id" binding:"required,uuid"`
	ReturnLocationID        string `json:"return_location_id" binding:"required,uuid"`
	ReturnInventoryStatusID string `json:"return_inventory_status_id" binding:"required,uuid"`
	IsActive                bool   `json:"is_active"`
}

type ReturnDeliveryLineRequest struct {
	ReturnedQty  string  `json:"returned_qty" binding:"required,max=30"`
	EventAt      string  `json:"event_at" binding:"required"`
	OperationKey string  `json:"operation_key" binding:"required,max=160"`
	Notes        *string `json:"notes" binding:"omitempty,max=4000"`
}
