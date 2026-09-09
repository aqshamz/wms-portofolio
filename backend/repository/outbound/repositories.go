package outbound

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	inventory "wms-api/repository/inventory"
	master "wms-api/repository/master"
)

var (
	ErrNotFound        = errors.New("outbound document not found")
	ErrConflict        = errors.New("outbound data conflicts with existing data")
	ErrConstraint      = errors.New("outbound data violates a constraint")
	ErrConcurrentWrite = errors.New("outbound version changed")
)

func Error(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		switch pg.Code {
		case "23505":
			return ErrConflict
		case "23503", "23514", "22001", "22P02":
			return ErrConstraint
		}
	}
	return err
}

type ListFilter struct {
	OwnerID, WarehouseID, StatusCode, Search string
	Page, PageSize                           int
}

type Repositories struct {
	db                    *gorm.DB
	Scope                 *OutboundScopeRepository
	Order                 *OutboundOrderRepository
	Line                  *OutboundOrderLineRepository
	ValidationRun         *OutboundValidationRunRepository
	ValidationResult      *OutboundValidationResultRepository
	ValidationRule        *OutboundValidationRuleRepository
	WaveType              *OutboundWaveTypeRepository
	Wave                  *OutboundWaveRepository
	WaveOrder             *OutboundWaveOrderRepository
	Reservation           *InventoryReservationRepository
	PickTask              *PickTaskRepository
	PickExecution         *PickExecutionRepository
	Staging               *OutboundStagingRepository
	StagingLine           *OutboundStagingLineRepository
	CheckResult           *OutboundCheckResultRepository
	ExceptionStatus       *OutboundCheckExceptionStatusRepository
	ResolutionType        *OutboundCheckResolutionTypeRepository
	Check                 *OutboundCheckRepository
	CheckLine             *OutboundCheckLineRepository
	CheckException        *OutboundCheckExceptionRepository
	CheckResolution       *OutboundCheckResolutionRepository
	ResolutionMovement    *OutboundCheckResolutionMovementRepository
	ResolutionReservation *OutboundCheckResolutionReservationRepository
	Packing               *PackingRepository
	PackingLine           *PackingLineRepository
	Carrier               *CarrierRepository
	CarrierService        *CarrierServiceRepository
	CarrierDriver         *CarrierDriverRepository
	Shipment              *ShipmentRepository
	ShipmentDriver        *ShipmentDriverRepository
	ShipmentOrder         *ShipmentOrderRepository
	ShipmentPacking       *ShipmentPackingRepository
	ShipmentLine          *ShipmentLineRepository
	DeliveryEventType     *DeliveryEventTypeRepository
	DeliveryFailureReason *DeliveryFailureReasonRepository
	Delivery              *DeliveryRepository
	DeliveryLine          *DeliveryLineRepository
	DeliveryEvent         *DeliveryEventRepository
	DeliveryEventLine     *DeliveryEventLineRepository
	ReturnPolicy          *OutboundReturnPolicyRepository
	DeliveryReturnLine    *DeliveryReturnLineRepository
	Master                *master.OperationalRepositories
	Inventory             *inventory.Repositories
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{db: db, Scope: NewOutboundScopeRepository(db), Order: NewOutboundOrderRepository(db), Line: NewOutboundOrderLineRepository(db),
		ValidationRun: NewOutboundValidationRunRepository(db), ValidationResult: NewOutboundValidationResultRepository(db), ValidationRule: NewOutboundValidationRuleRepository(db),
		WaveType: NewOutboundWaveTypeRepository(db), Wave: NewOutboundWaveRepository(db), WaveOrder: NewOutboundWaveOrderRepository(db),
		Reservation: NewInventoryReservationRepository(db), PickTask: NewPickTaskRepository(db), PickExecution: NewPickExecutionRepository(db),
		Staging: NewOutboundStagingRepository(db), StagingLine: NewOutboundStagingLineRepository(db),
		CheckResult: NewOutboundCheckResultRepository(db), ExceptionStatus: NewOutboundCheckExceptionStatusRepository(db), ResolutionType: NewOutboundCheckResolutionTypeRepository(db),
		Check: NewOutboundCheckRepository(db), CheckLine: NewOutboundCheckLineRepository(db), CheckException: NewOutboundCheckExceptionRepository(db), CheckResolution: NewOutboundCheckResolutionRepository(db),
		ResolutionMovement: NewOutboundCheckResolutionMovementRepository(db), ResolutionReservation: NewOutboundCheckResolutionReservationRepository(db),
		Packing: NewPackingRepository(db), PackingLine: NewPackingLineRepository(db), Carrier: NewCarrierRepository(db), CarrierService: NewCarrierServiceRepository(db), CarrierDriver: NewCarrierDriverRepository(db), Shipment: NewShipmentRepository(db), ShipmentDriver: NewShipmentDriverRepository(db), ShipmentOrder: NewShipmentOrderRepository(db), ShipmentPacking: NewShipmentPackingRepository(db), ShipmentLine: NewShipmentLineRepository(db),
		DeliveryEventType: NewDeliveryEventTypeRepository(db), DeliveryFailureReason: NewDeliveryFailureReasonRepository(db), Delivery: NewDeliveryRepository(db), DeliveryLine: NewDeliveryLineRepository(db), DeliveryEvent: NewDeliveryEventRepository(db), DeliveryEventLine: NewDeliveryEventLineRepository(db), ReturnPolicy: NewOutboundReturnPolicyRepository(db), DeliveryReturnLine: NewDeliveryReturnLineRepository(db),
		Master: master.NewOperationalRepositories(db), Inventory: inventory.NewRepositories(db)}
}

func (r *Repositories) Transaction(ctx context.Context, work func(*Repositories) error) error {
	return Error(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return work(NewRepositories(tx)) }))
}
