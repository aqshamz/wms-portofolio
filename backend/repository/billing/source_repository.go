package billing

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type MovementSourceRow struct {
	ID, OwnerID, WarehouseID, SourceDocumentID, Quantity, UOMID string
	SourceLineID                                                *string
	BusinessDate                                                time.Time
}
type BalanceSourceRow struct{ ID, Quantity, UOMID string }
type SourceRepository struct{ db *gorm.DB }

func NewSourceRepository(db *gorm.DB) *SourceRepository { return &SourceRepository{db: db} }
func (r *SourceRepository) Movements(ctx context.Context, owner, warehouse, movementType string, from, until time.Time) ([]MovementSourceRow, error) {
	v := []MovementSourceRow{}
	e := r.db.WithContext(ctx).Table("inventory_movement").Select("movement_id id,owner_id,warehouse_id,source_document_id,source_line_id,business_date,quantity,uom_id").Where("owner_id=? AND warehouse_id=? AND movement_type_id=? AND business_date BETWEEN ? AND ?", owner, warehouse, movementType, from, until).Order("business_date,movement_id").Find(&v).Error
	return v, Error(e)
}
func (r *SourceRepository) StorageBalances(ctx context.Context, owner, warehouse string) ([]BalanceSourceRow, error) {
	rows := []BalanceSourceRow{}
	err := r.db.WithContext(ctx).Table("inventory_balance").Select("balance_id id,on_hand_qty quantity,uom_id").Where("owner_id=? AND warehouse_id=? AND on_hand_qty>0", owner, warehouse).Order("balance_id").Find(&rows).Error
	return rows, Error(err)
}
