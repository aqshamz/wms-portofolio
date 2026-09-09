package outbound

import "time"

type CheckLineResponse struct {
	ID            string     `json:"outbound_check_line_id"`
	StagingLineID string     `json:"staging_line_id"`
	ItemCode      string     `json:"item_code"`
	LotNumber     string     `json:"lot_number,omitempty"`
	LineNo        int        `json:"line_no"`
	ExpectedQty   string     `json:"expected_qty"`
	CheckedQty    *string    `json:"checked_qty"`
	ExceptionQty  string     `json:"exception_qty"`
	ResultCode    string     `json:"result_code,omitempty"`
	Notes         *string    `json:"notes"`
	CheckedAt     *time.Time `json:"checked_at"`
}

type CheckResponse struct {
	ID                    string              `json:"outbound_check_id"`
	ParentCheckID         *string             `json:"parent_check_id"`
	StagingID             string              `json:"staging_id"`
	OutboundID            string              `json:"outbound_id"`
	ClientDeliveryOrderNo string              `json:"client_delivery_order_no"`
	OwnerID               string              `json:"owner_id"`
	WarehouseID           string              `json:"warehouse_id"`
	StatusCode            string              `json:"status_code"`
	CheckedAt             *time.Time          `json:"checked_at"`
	Notes                 *string             `json:"notes"`
	LineCount             int64               `json:"line_count"`
	Lines                 []CheckLineResponse `json:"lines,omitempty"`
}

type CheckExceptionResponse struct {
	ID           string     `json:"outbound_check_exception_id"`
	CheckLineID  string     `json:"outbound_check_line_id"`
	ResultCode   string     `json:"result_code"`
	StatusCode   string     `json:"status_code"`
	ExceptionQty string     `json:"exception_qty"`
	ResolvedQty  string     `json:"resolved_qty"`
	OpenedAt     time.Time  `json:"opened_at"`
	ResolvedAt   *time.Time `json:"resolved_at"`
	Notes        *string    `json:"notes"`
}

type ReplacementPickResponse struct {
	ExceptionID   string           `json:"outbound_check_exception_id"`
	ResolutionID  string           `json:"outbound_check_resolution_id"`
	WaveID        string           `json:"wave_id"`
	ReservationID string           `json:"reservation_id"`
	Task          PickTaskResponse `json:"task"`
}

type PackingLineResponse struct {
	ID                   string  `json:"packing_line_id"`
	OutboundCheckLineID  string  `json:"outbound_check_line_id"`
	PickTaskID           string  `json:"pick_task_id"`
	ItemCode             string  `json:"item_code"`
	LotNumber            string  `json:"lot_number,omitempty"`
	SourceBalanceID      string  `json:"source_balance_id"`
	SourceBalanceVersion int64   `json:"source_balance_version"`
	PackingBalanceID     string  `json:"packing_balance_id,omitempty"`
	HandlingUnitID       *string `json:"handling_unit_id"`
	PackedQty            string  `json:"packed_qty"`
	UOMCode              string  `json:"uom_code"`
	MovementID           string  `json:"movement_id,omitempty"`
}

type PackingResponse struct {
	ID                    string                `json:"packing_id"`
	OutboundID            string                `json:"outbound_id"`
	ClientDeliveryOrderNo string                `json:"client_delivery_order_no"`
	OwnerID               string                `json:"owner_id"`
	WarehouseID           string                `json:"warehouse_id"`
	PackingLocationID     *string               `json:"packing_location_id"`
	PackingLocationCode   string                `json:"packing_location_code,omitempty"`
	StatusCode            string                `json:"status_code"`
	PackedAt              *time.Time            `json:"packed_at"`
	LineCount             int64                 `json:"line_count"`
	Lines                 []PackingLineResponse `json:"lines,omitempty"`
}

type ShipmentLineResponse struct {
	ID                   string `json:"shipment_line_id,omitempty"`
	PackingLineID        string `json:"packing_line_id"`
	OutboundID           string `json:"outbound_id"`
	ItemCode             string `json:"item_code"`
	LotNumber            string `json:"lot_number,omitempty"`
	SourceBalanceID      string `json:"source_balance_id"`
	SourceBalanceVersion int64  `json:"source_balance_version"`
	ShippedQty           string `json:"shipped_qty"`
	UOMCode              string `json:"uom_code"`
	MovementID           string `json:"movement_id,omitempty"`
}

type ShipmentResponse struct {
	ID                 string                   `json:"shipment_id"`
	OwnerID            string                   `json:"owner_id"`
	WarehouseID        string                   `json:"warehouse_id"`
	CarrierServiceID   *string                  `json:"carrier_service_id"`
	CarrierCode        string                   `json:"carrier_code,omitempty"`
	CarrierServiceCode string                   `json:"carrier_service_code,omitempty"`
	BusinessDate       string                   `json:"business_date"`
	StatusCode         string                   `json:"status_code"`
	RouteReference     *string                  `json:"route_reference"`
	TrackingNumber     *string                  `json:"tracking_number"`
	VehicleNumber      *string                  `json:"vehicle_number"`
	SealNumber         *string                  `json:"seal_number"`
	ShippedAt          *time.Time               `json:"shipped_at"`
	Notes              *string                  `json:"notes"`
	OrderCount         int64                    `json:"order_count"`
	LineCount          int64                    `json:"line_count"`
	Lines              []ShipmentLineResponse   `json:"lines,omitempty"`
	Drivers            []ShipmentDriverResponse `json:"drivers,omitempty"`
}

type DeliveryLineResponse struct {
	ID             string `json:"delivery_line_id"`
	ShipmentLineID string `json:"shipment_line_id"`
	ItemCode       string `json:"item_code"`
	LotNumber      string `json:"lot_number,omitempty"`
	PlannedQty     string `json:"planned_qty"`
	DeliveredQty   string `json:"delivered_qty"`
	ReturnedQty    string `json:"returned_qty"`
	UOMCode        string `json:"uom_code"`
}

type DeliveryEventResponse struct {
	ID                string    `json:"delivery_event_id"`
	EventTypeCode     string    `json:"event_type_code"`
	FailureReasonCode string    `json:"failure_reason_code,omitempty"`
	EventAt           time.Time `json:"event_at"`
	RecipientName     *string   `json:"recipient_name"`
	ProofReference    *string   `json:"proof_reference"`
	ProofURI          *string   `json:"proof_uri"`
	Notes             *string   `json:"notes"`
	Latitude          *string   `json:"latitude"`
	Longitude         *string   `json:"longitude"`
}

type DeliveryResponse struct {
	ID                    string                  `json:"delivery_id"`
	ShipmentID            string                  `json:"shipment_id"`
	OutboundID            string                  `json:"outbound_id"`
	ClientDeliveryOrderNo string                  `json:"client_delivery_order_no"`
	OwnerID               string                  `json:"owner_id"`
	WarehouseID           string                  `json:"warehouse_id"`
	BusinessDate          string                  `json:"business_date"`
	StatusCode            string                  `json:"status_code"`
	PlannedDeliveryAt     *time.Time              `json:"planned_delivery_at"`
	ArrivedAt             *time.Time              `json:"arrived_at"`
	DeliveredAt           *time.Time              `json:"delivered_at"`
	RecipientName         *string                 `json:"recipient_name"`
	RecipientReference    *string                 `json:"recipient_reference"`
	ProofReference        *string                 `json:"proof_reference"`
	ProofURI              *string                 `json:"proof_uri"`
	Notes                 *string                 `json:"notes"`
	Lines                 []DeliveryLineResponse  `json:"lines,omitempty"`
	Events                []DeliveryEventResponse `json:"events,omitempty"`
}

type ReturnPolicyResponse struct {
	OwnerID                   string `json:"owner_id"`
	WarehouseID               string `json:"warehouse_id"`
	ReturnLocationID          string `json:"return_location_id"`
	ReturnLocationCode        string `json:"return_location_code"`
	ReturnInventoryStatusID   string `json:"return_inventory_status_id"`
	ReturnInventoryStatusCode string `json:"return_inventory_status_code"`
	IsActive                  bool   `json:"is_active"`
}

type ReturnLineResponse struct {
	ID                string `json:"delivery_return_line_id"`
	DeliveryLineID    string `json:"delivery_line_id"`
	ReturnedBalanceID string `json:"returned_balance_id"`
	ReturnedQty       string `json:"returned_qty"`
	MovementID        string `json:"movement_id"`
}
