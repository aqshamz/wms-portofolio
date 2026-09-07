package inventory

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	master "wms-api/repository/master"
)

var (
	ErrNotFound   = errors.New("inventory identity not found")
	ErrConflict   = errors.New("inventory identity already exists")
	ErrConstraint = errors.New("inventory identity violates a data constraint")
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

type Filter struct {
	OwnerID, ItemID, WarehouseID, LocationID, ParentID, Search, Number string
	Closed                                                             *bool
	Page, PageSize                                                     int
}
type identityTable[T any] struct {
	db           *gorm.DB
	key, label   string
	handlingUnit bool
}

func (r *identityTable[T]) Create(ctx context.Context, v *T) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *identityTable[T]) Get(ctx context.Context, id string) (T, error) {
	var v T
	err := r.db.WithContext(ctx).Where(r.key+" = ?", id).Take(&v).Error
	return v, Error(err)
}
func (r *identityTable[T]) GetShared(ctx context.Context, id string) (T, error) {
	var v T
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Where(r.key+" = ?", id).Take(&v).Error
	return v, Error(err)
}
func (r *identityTable[T]) List(ctx context.Context, f Filter) ([]T, int64, error) {
	q := r.db.WithContext(ctx).Model(new(T))
	if f.OwnerID != "" {
		q = q.Where("owner_id = ?", f.OwnerID)
	}
	if r.handlingUnit {
		if f.WarehouseID != "" {
			q = q.Where("warehouse_id = ?", f.WarehouseID)
		}
		if f.LocationID != "" {
			q = q.Where("current_location_id = ?", f.LocationID)
		}
		if f.ParentID != "" {
			q = q.Where("parent_handling_unit_id = ?", f.ParentID)
		}
		if f.Closed != nil {
			q = q.Where("is_closed = ?", *f.Closed)
		}
	} else if f.ItemID != "" {
		q = q.Where("item_id = ?", f.ItemID)
	}
	if f.Number != "" {
		q = q.Where(r.label+" = ?", f.Number)
	}
	if f.Search != "" {
		q = q.Where("("+r.key+" ILIKE ? OR "+r.label+" ILIKE ?)", "%"+f.Search+"%", "%"+f.Search+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, Error(err)
	}
	rows := make([]T, 0)
	err := q.Order(r.key).Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&rows).Error
	return rows, total, Error(err)
}

type Repositories struct {
	db             *gorm.DB
	Lot            *InventoryLotRepository
	Serial         *SerialNumberRepository
	HandlingUnit   *HandlingUnitRepository
	Balance        *InventoryBalanceRepository
	Movement       *InventoryMovementRepository
	SerialState    *SerialInventoryRepository
	MovementType   *MovementTypeRepository
	Catalog        *master.CatalogRepositories
	Owner          *master.OrganizationRepository
	Warehouse      *master.WarehouseRepository
	WarehouseOwner *master.WarehouseOwnerRepository
	Location       *master.WarehouseLocationRepository
	Zone           *master.WarehouseZoneRepository
	LocationType   *master.LocationTypeRepository
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{db: db, Lot: NewInventoryLotRepository(db), Serial: NewSerialNumberRepository(db), HandlingUnit: NewHandlingUnitRepository(db),
		Balance: NewInventoryBalanceRepository(db), Movement: NewInventoryMovementRepository(db), SerialState: NewSerialInventoryRepository(db), MovementType: NewMovementTypeRepository(db),
		Catalog: master.NewCatalogRepositories(db), Owner: master.NewOrganizationRepository(db), Warehouse: master.NewWarehouseRepository(db),
		WarehouseOwner: master.NewWarehouseOwnerRepository(db), Location: master.NewWarehouseLocationRepository(db),
		Zone: master.NewWarehouseZoneRepository(db), LocationType: master.NewLocationTypeRepository(db)}
}
func (r *Repositories) Transaction(ctx context.Context, work func(*Repositories) error) error {
	return Error(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return work(NewRepositories(tx)) }))
}
