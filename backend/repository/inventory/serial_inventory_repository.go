package inventory

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/inventory"
)

type SerialStateRow struct {
	model.SerialInventory
	SerialNo string
	Balance  BalanceRow `gorm:"-"`
}
type SerialStateFilter struct {
	OwnerID, WarehouseID, LocationID, ItemID, LotID, HandlingUnitID, InventoryStatusID, Search string
	Page, PageSize                                                                             int
}
type SerialInventoryRepository struct{ db *gorm.DB }

func NewSerialInventoryRepository(db *gorm.DB) *SerialInventoryRepository {
	return &SerialInventoryRepository{db: db}
}
func (r *SerialInventoryRepository) GetLocked(ctx context.Context, serialID string) (model.SerialInventory, error) {
	var v model.SerialInventory
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("serial_id=?", serialID).Take(&v).Error
	return v, Error(err)
}
func (r *SerialInventoryRepository) Create(ctx context.Context, v *model.SerialInventory) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *SerialInventoryRepository) Move(ctx context.Context, serialID, balanceID string) error {
	result := r.db.WithContext(ctx).Model(&model.SerialInventory{}).Where("serial_id=?", serialID).Updates(map[string]interface{}{"balance_id": balanceID, "version_no": gorm.Expr("version_no+1"), "updated_at": gorm.Expr("clock_timestamp()")})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}

func (r *SerialInventoryRepository) IDsByBalance(ctx context.Context, balanceID string) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).Model(&model.SerialInventory{}).Where("balance_id=?", balanceID).Order("serial_id").Pluck("serial_id", &ids).Error
	return ids, Error(err)
}

func (r *SerialInventoryRepository) MoveBalance(ctx context.Context, sourceBalanceID, targetBalanceID string) error {
	return Error(r.db.WithContext(ctx).Model(&model.SerialInventory{}).Where("balance_id=?", sourceBalanceID).Updates(map[string]interface{}{"balance_id": targetBalanceID, "version_no": gorm.Expr("version_no+1"), "updated_at": gorm.Expr("clock_timestamp()")}).Error)
}

func (r *SerialInventoryRepository) DeleteBalance(ctx context.Context, sourceBalanceID string) error {
	return Error(r.db.WithContext(ctx).Where("balance_id=?", sourceBalanceID).Delete(&model.SerialInventory{}).Error)
}
func (r *SerialInventoryRepository) Delete(ctx context.Context, serialID string) error {
	result := r.db.WithContext(ctx).Where("serial_id=?", serialID).Delete(&model.SerialInventory{})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
func serialStateQuery(db *gorm.DB) *gorm.DB {
	return db.Table("serial_inventory si").Select("si.*,sn.serial_no").Joins("JOIN serial_number sn ON sn.serial_id=si.serial_id").Joins("JOIN inventory_balance b ON b.balance_id=si.balance_id")
}
func (r *SerialInventoryRepository) Get(ctx context.Context, serialID string) (SerialStateRow, error) {
	var v SerialStateRow
	err := serialStateQuery(r.db.WithContext(ctx)).Where("si.serial_id=?", serialID).Take(&v).Error
	if err != nil {
		return v, Error(err)
	}
	v.Balance, err = NewInventoryBalanceRepository(r.db).Get(ctx, v.BalanceID)
	return v, err
}
func (r *SerialInventoryRepository) List(ctx context.Context, f SerialStateFilter) ([]SerialStateRow, int64, error) {
	q := serialStateQuery(r.db.WithContext(ctx))
	for col, val := range map[string]string{"si.owner_id": f.OwnerID, "si.item_id": f.ItemID, "b.warehouse_id": f.WarehouseID, "b.location_id": f.LocationID, "b.lot_id": f.LotID, "b.handling_unit_id": f.HandlingUnitID, "b.inventory_status_id": f.InventoryStatusID} {
		if val != "" {
			q = q.Where(col+"=?", val)
		}
	}
	if f.Search != "" {
		wild := "%" + f.Search + "%"
		q = q.Where("(sn.serial_no ILIKE ? OR si.serial_id ILIKE ?)", wild, wild)
	}
	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, Error(err)
	}
	rows := make([]SerialStateRow, 0)
	if err := q.Order("sn.serial_no").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&rows).Error; err != nil {
		return nil, 0, Error(err)
	}
	balances := NewInventoryBalanceRepository(r.db)
	for i := range rows {
		b, err := balances.Get(ctx, rows[i].BalanceID)
		if err != nil {
			return nil, 0, err
		}
		rows[i].Balance = b
	}
	return rows, total, nil
}
