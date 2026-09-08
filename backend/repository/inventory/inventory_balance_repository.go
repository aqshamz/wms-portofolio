package inventory

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/inventory"
)

type BalanceIdentity struct {
	OwnerID, WarehouseID, LocationID, ItemID string
	LotID, HandlingUnitID                    *string
	InventoryStatusID                        string
}
type BalanceRow struct {
	model.InventoryBalance
	LocationCode, ItemCode, ItemName, InventoryStatusCode, UOMCode string
	LotNumber                                                      *string
	AvailableQty                                                   string
}
type BalanceFilter struct {
	OwnerID, WarehouseID, LocationID, ItemID, LotID, HandlingUnitID, InventoryStatusID, Search string
	IncludeZero                                                                                bool
	Page, PageSize                                                                             int
}
type InventoryBalanceRepository struct{ db *gorm.DB }

func NewInventoryBalanceRepository(db *gorm.DB) *InventoryBalanceRepository {
	return &InventoryBalanceRepository{db: db}
}
func (r *InventoryBalanceRepository) LockStockKey(ctx context.Context, ownerID, warehouseID, itemID string) error {
	return Error(r.db.WithContext(ctx).Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", ownerID+"|"+warehouseID+"|"+itemID).Error)
}
func balanceIdentity(q *gorm.DB, k BalanceIdentity) *gorm.DB {
	q = q.Where("owner_id=? AND warehouse_id=? AND location_id=? AND item_id=? AND inventory_status_id=?", k.OwnerID, k.WarehouseID, k.LocationID, k.ItemID, k.InventoryStatusID)
	if k.LotID == nil {
		q = q.Where("lot_id IS NULL")
	} else {
		q = q.Where("lot_id=?", *k.LotID)
	}
	if k.HandlingUnitID == nil {
		q = q.Where("handling_unit_id IS NULL")
	} else {
		q = q.Where("handling_unit_id=?", *k.HandlingUnitID)
	}
	return q
}
func (r *InventoryBalanceRepository) FindLocked(ctx context.Context, k BalanceIdentity) (model.InventoryBalance, error) {
	var v model.InventoryBalance
	err := balanceIdentity(r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}), k).Take(&v).Error
	return v, Error(err)
}
func (r *InventoryBalanceRepository) Create(ctx context.Context, v *model.InventoryBalance) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *InventoryBalanceRepository) SetOnHand(ctx context.Context, id, quantity string) error {
	result := r.db.WithContext(ctx).Model(&model.InventoryBalance{}).Where("balance_id=?", id).Updates(map[string]interface{}{"on_hand_qty": quantity, "version_no": gorm.Expr("version_no+1"), "updated_at": gorm.Expr("clock_timestamp()")})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}

func (r *InventoryBalanceRepository) SetQuantities(ctx context.Context, id, onHand, reserved string) error {
	result := r.db.WithContext(ctx).Model(&model.InventoryBalance{}).Where("balance_id=?", id).Updates(map[string]interface{}{"on_hand_qty": onHand, "reserved_qty": reserved, "version_no": gorm.Expr("version_no+1"), "updated_at": gorm.Expr("clock_timestamp()")})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}

func (r *InventoryBalanceRepository) CountPositiveForHandlingUnit(ctx context.Context, handlingUnitID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.InventoryBalance{}).Where("handling_unit_id=? AND on_hand_qty>0", handlingUnitID).Count(&count).Error
	return count, Error(err)
}

const balanceSelect = `b.balance_id,b.owner_id,b.warehouse_id,b.location_id,l.code location_code,b.item_id,
 i.code item_code,i.name item_name,b.lot_id,lot.lot_number,b.handling_unit_id,b.inventory_status_id,
 s.code inventory_status_code,b.on_hand_qty,b.reserved_qty,(b.on_hand_qty-b.reserved_qty) available_qty,
 b.uom_id,u.code uom_code,b.version_no,b.updated_at`

func balanceQuery(db *gorm.DB) *gorm.DB {
	return db.Table("inventory_balance b").Select(balanceSelect).
		Joins("JOIN warehouse_location l ON l.location_id=b.location_id").Joins("JOIN item i ON i.item_id=b.item_id").
		Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=b.lot_id").Joins("JOIN inventory_status s ON s.inventory_status_id=b.inventory_status_id").
		Joins("JOIN uom u ON u.uom_id=b.uom_id")
}
func (r *InventoryBalanceRepository) Get(ctx context.Context, id string) (BalanceRow, error) {
	var v BalanceRow
	err := balanceQuery(r.db.WithContext(ctx)).Where("b.balance_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *InventoryBalanceRepository) List(ctx context.Context, f BalanceFilter) ([]BalanceRow, int64, error) {
	q := balanceQuery(r.db.WithContext(ctx))
	for col, val := range map[string]string{"b.owner_id": f.OwnerID, "b.warehouse_id": f.WarehouseID, "b.location_id": f.LocationID, "b.item_id": f.ItemID, "b.lot_id": f.LotID, "b.handling_unit_id": f.HandlingUnitID, "b.inventory_status_id": f.InventoryStatusID} {
		if val != "" {
			q = q.Where(col+"=?", val)
		}
	}
	if !f.IncludeZero {
		q = q.Where("b.on_hand_qty>0")
	}
	if f.Search != "" {
		wild := "%" + f.Search + "%"
		q = q.Where("(i.code ILIKE ? OR i.name ILIKE ? OR l.code ILIKE ? OR lot.lot_number ILIKE ?)", wild, wild, wild, wild)
	}
	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, Error(err)
	}
	rows := make([]BalanceRow, 0)
	err := q.Order("i.code,l.code,b.balance_id").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&rows).Error
	return rows, total, Error(err)
}
