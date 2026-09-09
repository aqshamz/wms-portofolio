package outbound

import "time"

type PageResponse[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int64 `json:"total_pages"`
}
type OutboundOrderLineResponse struct {
	ID                    string  `json:"outbound_line_id"`
	LineNo                int     `json:"line_no"`
	ItemID                string  `json:"item_id"`
	ItemCode              string  `json:"item_code"`
	ItemName              string  `json:"item_name"`
	OrderedQty            string  `json:"ordered_qty"`
	AllocatedQty          string  `json:"allocated_qty"`
	PickedQty             string  `json:"picked_qty"`
	CheckedQty            string  `json:"checked_qty"`
	RejectedQty           string  `json:"rejected_qty"`
	ShortAcceptedQty      string  `json:"short_accepted_qty"`
	PackedQty             string  `json:"packed_qty"`
	ShippedQty            string  `json:"shipped_qty"`
	DeliveredQty          string  `json:"delivered_qty"`
	UOMID                 string  `json:"uom_id"`
	UOMCode               string  `json:"uom_code"`
	RequestedLotNo        *string `json:"requested_lot_no"`
	CustomerLineReference *string `json:"customer_line_reference"`
	Notes                 *string `json:"notes"`
}
type OutboundOrderResponse struct {
	ID                    string                      `json:"outbound_id"`
	OwnerID               string                      `json:"owner_id"`
	OwnerCode             string                      `json:"owner_code"`
	CustomerID            string                      `json:"customer_id"`
	CustomerCode          string                      `json:"customer_code"`
	CustomerName          string                      `json:"customer_name"`
	ShipToPartnerID       *string                     `json:"ship_to_partner_id"`
	ShipToCode            string                      `json:"ship_to_code"`
	WarehouseID           string                      `json:"warehouse_id"`
	WarehouseCode         string                      `json:"warehouse_code"`
	BusinessDate          string                      `json:"business_date"`
	RequestedShipAt       *time.Time                  `json:"requested_ship_at"`
	ClientDeliveryOrderNo string                      `json:"client_delivery_order_no"`
	CustomerOrderNo       *string                     `json:"customer_order_no"`
	ExternalReference     *string                     `json:"external_reference"`
	ShipToName            string                      `json:"ship_to_name"`
	ShipToAddress1        string                      `json:"ship_to_address_1"`
	ShipToAddress2        *string                     `json:"ship_to_address_2"`
	ShipToCity            *string                     `json:"ship_to_city"`
	ShipToProvince        *string                     `json:"ship_to_province"`
	ShipToPostalCode      *string                     `json:"ship_to_postal_code"`
	ShipToCountryCode     *string                     `json:"ship_to_country_code"`
	StatusCode            string                      `json:"status_code"`
	Notes                 *string                     `json:"notes"`
	VersionNo             int64                       `json:"version_no"`
	CreatedAt             time.Time                   `json:"created_at"`
	Lines                 []OutboundOrderLineResponse `json:"lines,omitempty"`
}
type ValidationResultResponse struct {
	RuleCode      string  `json:"rule_code"`
	RuleName      string  `json:"rule_name"`
	SeverityCode  string  `json:"severity_code"`
	Passed        bool    `json:"passed"`
	ResultMessage *string `json:"result_message"`
}
type ValidationRunResponse struct {
	ID          string                     `json:"validation_run_id"`
	OutboundID  string                     `json:"outbound_id"`
	StatusCode  string                     `json:"status_code"`
	ValidatedAt *time.Time                 `json:"validated_at"`
	Notes       *string                    `json:"notes"`
	Results     []ValidationResultResponse `json:"results"`
}
type ReservationResponse struct {
	ID             string    `json:"reservation_id"`
	OutboundID     string    `json:"outbound_id"`
	OutboundLineID string    `json:"outbound_line_id"`
	BalanceID      string    `json:"balance_id"`
	ItemCode       string    `json:"item_code"`
	LocationCode   string    `json:"location_code"`
	LotNumber      string    `json:"lot_number,omitempty"`
	ReservedQty    string    `json:"reserved_qty"`
	PickedQty      string    `json:"picked_qty"`
	UOMID          string    `json:"uom_id"`
	StatusCode     string    `json:"status_code"`
	ReservedAt     time.Time `json:"reserved_at"`
}
type AllocationResponse struct {
	Order        OutboundOrderResponse `json:"order"`
	Reservations []ReservationResponse `json:"reservations"`
	RequestedQty string                `json:"requested_qty"`
	AllocatedQty string                `json:"allocated_qty"`
	ShortageQty  string                `json:"shortage_qty"`
}
type WaveOrderResponse struct {
	OutboundID            string `json:"outbound_id"`
	ClientDeliveryOrderNo string `json:"client_delivery_order_no"`
	StatusCode            string `json:"status_code"`
}
type WaveResponse struct {
	ID                string              `json:"wave_id"`
	OwnerID           string              `json:"owner_id"`
	OwnerCode         string              `json:"owner_code"`
	WarehouseID       string              `json:"warehouse_id"`
	WarehouseCode     string              `json:"warehouse_code"`
	WaveTypeCode      string              `json:"wave_type_code"`
	PickingStrategyID *string             `json:"picking_strategy_id"`
	BusinessDate      string              `json:"business_date"`
	PlannedReleaseAt  *time.Time          `json:"planned_release_at"`
	ReleasedAt        *time.Time          `json:"released_at"`
	CompletedAt       *time.Time          `json:"completed_at"`
	StatusCode        string              `json:"status_code"`
	Notes             *string             `json:"notes"`
	VersionNo         int64               `json:"version_no"`
	OrderCount        int64               `json:"order_count"`
	PickTaskCount     int64               `json:"pick_task_count"`
	Orders            []WaveOrderResponse `json:"orders,omitempty"`
}
type PickTaskResponse struct {
	ID                    string     `json:"pick_task_id"`
	WaveID                string     `json:"wave_id"`
	ReservationID         string     `json:"reservation_id"`
	OutboundID            string     `json:"outbound_id"`
	ClientDeliveryOrderNo string     `json:"client_delivery_order_no"`
	OutboundLineID        string     `json:"outbound_line_id"`
	ItemCode              string     `json:"item_code"`
	SourceLocationID      string     `json:"source_location_id"`
	SourceLocationCode    string     `json:"source_location_code"`
	TargetLocationID      *string    `json:"target_location_id"`
	TargetLocationCode    string     `json:"target_location_code"`
	LotNumber             string     `json:"lot_number,omitempty"`
	PlannedQty            string     `json:"planned_qty"`
	PickedQty             string     `json:"picked_qty"`
	ShortQty              string     `json:"short_qty"`
	UOMID                 string     `json:"uom_id"`
	StatusCode            string     `json:"status_code"`
	PriorityCode          string     `json:"priority_code"`
	AssignedTo            *string    `json:"assigned_to"`
	BalanceID             string     `json:"balance_id"`
	BalanceVersion        int64      `json:"balance_version"`
	StartedAt             *time.Time `json:"started_at"`
	CompletedAt           *time.Time `json:"completed_at"`
}
type PickConfirmationResponse struct {
	ExecutionID      string           `json:"pick_execution_id"`
	MovementID       string           `json:"movement_id"`
	StagingBalanceID string           `json:"staging_balance_id"`
	Task             PickTaskResponse `json:"task"`
}
type StagingLineResponse struct {
	ID               string `json:"staging_line_id"`
	PickExecutionID  string `json:"pick_execution_id"`
	PickTaskID       string `json:"pick_task_id"`
	StagingBalanceID string `json:"staging_balance_id"`
	ItemCode         string `json:"item_code"`
	StagedQty        string `json:"staged_qty"`
	RemovedQty       string `json:"removed_qty"`
	UOMCode          string `json:"uom_code"`
}
type StagingResponse struct {
	ID                    string                `json:"staging_id"`
	OutboundID            string                `json:"outbound_id"`
	ClientDeliveryOrderNo string                `json:"client_delivery_order_no"`
	WaveID                string                `json:"wave_id"`
	WarehouseID           string                `json:"warehouse_id"`
	WarehouseCode         string                `json:"warehouse_code"`
	StagingLocationID     string                `json:"staging_location_id"`
	StagingLocationCode   string                `json:"staging_location_code"`
	StatusCode            string                `json:"status_code"`
	StagedAt              *time.Time            `json:"staged_at"`
	Notes                 *string               `json:"notes"`
	LineCount             int64                 `json:"line_count"`
	Lines                 []StagingLineResponse `json:"lines,omitempty"`
}
