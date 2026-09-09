package inbound

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	inventory "wms-api/repository/inventory"
	master "wms-api/repository/master"
)

var (
	ErrNotFound        = errors.New("inbound document not found")
	ErrConflict        = errors.New("inbound document conflicts with existing data")
	ErrConstraint      = errors.New("inbound document violates a data constraint")
	ErrConcurrentWrite = errors.New("inbound document version changed")
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
	db                *gorm.DB
	PurchaseOrder     *PurchaseOrderRepository
	PurchaseOrderLine *PurchaseOrderLineRepository
	InboundOrder      *InboundOrderRepository
	InboundOrderLine  *InboundOrderLineRepository
	Receipt           *ReceiptRepository
	ReceiptLine       *ReceiptLineRepository
	ReceiptInventory  *ReceiptInventoryRepository
	ReceiptLineSerial *ReceiptLineSerialRepository
	QualityInspection *QualityInspectionRepository
	PutawayTask       *PutawayTaskRepository
	QuarantineType    *QuarantineDispositionTypeRepository
	QuarantineCase    *QuarantineCaseRepository
	Disposition       *QuarantineDispositionRepository
	Exception         *InboundExceptionRepository
	ReworkTask        *ReworkTaskRepository
	Scope             *InboundScopeRepository
	Master            *master.OperationalRepositories
	Inventory         *inventory.Repositories
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		db: db, PurchaseOrder: NewPurchaseOrderRepository(db), PurchaseOrderLine: NewPurchaseOrderLineRepository(db),
		InboundOrder: NewInboundOrderRepository(db), InboundOrderLine: NewInboundOrderLineRepository(db),
		Receipt: NewReceiptRepository(db), ReceiptLine: NewReceiptLineRepository(db), ReceiptInventory: NewReceiptInventoryRepository(db),
		ReceiptLineSerial: NewReceiptLineSerialRepository(db), Master: master.NewOperationalRepositories(db), Inventory: inventory.NewRepositories(db),
		QualityInspection: NewQualityInspectionRepository(db), PutawayTask: NewPutawayTaskRepository(db), QuarantineType: NewQuarantineDispositionTypeRepository(db),
		QuarantineCase: NewQuarantineCaseRepository(db), Disposition: NewQuarantineDispositionRepository(db),
		Exception: NewInboundExceptionRepository(db), ReworkTask: NewReworkTaskRepository(db),
		Scope: NewInboundScopeRepository(db),
	}
}

func (r *Repositories) Transaction(ctx context.Context, work func(*Repositories) error) error {
	return Error(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return work(NewRepositories(tx)) }))
}
