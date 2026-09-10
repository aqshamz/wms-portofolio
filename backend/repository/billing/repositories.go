package billing

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	inventory "wms-api/repository/inventory"
	master "wms-api/repository/master"
)

var (
	ErrNotFound        = errors.New("billing data not found")
	ErrConflict        = errors.New("billing data conflicts with existing data")
	ErrConstraint      = errors.New("billing data violates a constraint")
	ErrConcurrentWrite = errors.New("billing version changed")
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
	DateFrom, DateUntil                      string
	Page, PageSize                           int
}

type Repositories struct {
	db           *gorm.DB
	Scope        *ScopeRepository
	Contract     *BillingContractRepository
	RateCard     *RateCardRepository
	RateCardLine *RateCardLineRepository
	Event        *BillableEventRepository
	Run          *BillingRunRepository
	Charge       *BillingChargeRepository
	Invoice      *InvoiceRepository
	InvoiceLine  *InvoiceLineRepository
	CreditNote   *CreditNoteRepository
	Payment      *PaymentRepository
	Source       *SourceRepository
	Master       *master.OperationalRepositories
	Inventory    *inventory.Repositories
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{db: db, Scope: NewScopeRepository(db), Contract: NewBillingContractRepository(db), RateCard: NewRateCardRepository(db), RateCardLine: NewRateCardLineRepository(db), Event: NewBillableEventRepository(db), Run: NewBillingRunRepository(db), Charge: NewBillingChargeRepository(db), Invoice: NewInvoiceRepository(db), InvoiceLine: NewInvoiceLineRepository(db), CreditNote: NewCreditNoteRepository(db), Payment: NewPaymentRepository(db), Source: NewSourceRepository(db), Master: master.NewOperationalRepositories(db), Inventory: inventory.NewRepositories(db)}
}
func (r *Repositories) Transaction(ctx context.Context, work func(*Repositories) error) error {
	return Error(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return work(NewRepositories(tx)) }))
}
