package master

import (
	"context"
	"time"

	model "wms-api/models/master"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AccountWarehouseAccessDetail struct {
	AccountID     string
	WarehouseID   string
	WarehouseCode string
	WarehouseName string
	GrantedBy     *string
	GrantedAt     time.Time
}

type AccountWarehouseAccessRepository struct{ db *gorm.DB }

func NewAccountWarehouseAccessRepository(db *gorm.DB) *AccountWarehouseAccessRepository {
	return &AccountWarehouseAccessRepository{db: db}
}

func (r *AccountWarehouseAccessRepository) Grant(ctx context.Context, value *model.AccountWarehouseAccess) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "account_id"}, {Name: "warehouse_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"granted_by": value.GrantedBy,
			"granted_at": gorm.Expr("clock_timestamp()"),
		}),
	}).Create(value).Error
}

func (r *AccountWarehouseAccessRepository) List(
	ctx context.Context, accountID string,
) ([]AccountWarehouseAccessDetail, error) {
	var rows []AccountWarehouseAccessDetail
	err := r.db.WithContext(ctx).Table("account_warehouse_access AS access").
		Select(`access.account_id, access.warehouse_id, warehouse.code AS warehouse_code,
			warehouse.name AS warehouse_name, access.granted_by, access.granted_at`).
		Joins("JOIN warehouse ON warehouse.warehouse_id = access.warehouse_id").
		Where("access.account_id = ?", accountID).Order("warehouse.name").Scan(&rows).Error
	return rows, err
}

func (r *AccountWarehouseAccessRepository) Revoke(
	ctx context.Context, accountID, warehouseID string,
) error {
	result := r.db.WithContext(ctx).
		Where("account_id = ? AND warehouse_id = ?", accountID, warehouseID).
		Delete(&model.AccountWarehouseAccess{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
