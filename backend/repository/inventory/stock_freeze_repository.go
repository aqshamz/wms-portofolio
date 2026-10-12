package inventory

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	model "wms-api/models/inventory"
)

var ErrStockFrozen = errors.New("stock is frozen for grand stock opname")

type countReconciliationKey struct{}

// WithCountReconciliation is an internal exception for an approved correction
// belonging to the freezing document. Never derive it from a request reference.
func WithCountReconciliation(ctx context.Context, countID string) context.Context {
	return context.WithValue(ctx, countReconciliationKey{}, countID)
}

func countScopeKey(ownerID, warehouseID string) string {
	return "inventory-count|" + strings.ToLower(ownerID) + "|" + strings.ToLower(warehouseID)
}

// LockCountScope must be held in the transaction that snapshots and creates the
// grand opname. Writers hold its shared counterpart until their transaction ends.
func (r *InventoryBalanceRepository) LockCountScope(ctx context.Context, ownerID, warehouseID string) error {
	return Error(r.db.WithContext(ctx).Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", countScopeKey(ownerID, warehouseID)).Error)
}

func (r *InventoryBalanceRepository) CheckStockFreeze(ctx context.Context, ownerID, warehouseID string) error {
	if err := r.db.WithContext(ctx).Exec("SELECT pg_advisory_xact_lock_shared(hashtextextended(?, 0))", countScopeKey(ownerID, warehouseID)).Error; err != nil {
		return Error(err)
	}
	var freeze model.StockFreeze
	err := r.db.WithContext(ctx).Where("owner_id=? AND warehouse_id=?", ownerID, warehouseID).Limit(1).Find(&freeze).Error
	if err != nil {
		return Error(err)
	}
	if freeze.CycleCountID == "" {
		return nil
	}
	if approvedCount, _ := ctx.Value(countReconciliationKey{}).(string); approvedCount == freeze.CycleCountID {
		return nil
	}
	return fmt.Errorf("%w %s; finish or cancel that document before changing this customer's stock in this warehouse", ErrStockFrozen, freeze.CycleCountID)
}

func (r *InventoryBalanceRepository) checkBalanceFreeze(ctx context.Context, balanceID string) error {
	var balance model.InventoryBalance
	if err := r.db.WithContext(ctx).Select("owner_id,warehouse_id").Where("balance_id=?", balanceID).Take(&balance).Error; err != nil {
		return Error(err)
	}
	return r.CheckStockFreeze(ctx, balance.OwnerID, balance.WarehouseID)
}

func (r *InventoryBalanceRepository) BeginCountFreeze(ctx context.Context, ownerID, warehouseID, countID string) error {
	return Error(r.db.WithContext(ctx).Create(&model.StockFreeze{OwnerID: ownerID, WarehouseID: warehouseID, CycleCountID: countID}).Error)
}

func (r *InventoryBalanceRepository) HasReservedStock(ctx context.Context, ownerID, warehouseID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.InventoryBalance{}).
		Where("owner_id=? AND warehouse_id=? AND reserved_qty>0", ownerID, warehouseID).Count(&count).Error
	return count > 0, Error(err)
}

func (r *InventoryBalanceRepository) ReleaseCountFreeze(ctx context.Context, countID string) error {
	return Error(r.db.WithContext(ctx).Where("cycle_count_id=?", countID).Delete(&model.StockFreeze{}).Error)
}

func MigrateStockFreezes(db *gorm.DB) error {
	return db.AutoMigrate(&model.StockFreeze{})
}
