package inventory

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"wms-api/requestscope"
)

type InventoryScopeRepository struct{ db *gorm.DB }

func NewInventoryScopeRepository(db *gorm.DB) *InventoryScopeRepository {
	return &InventoryScopeRepository{db: db}
}

func (r *InventoryScopeRepository) OwnerAllowed(ctx context.Context, accountID, ownerID string) (bool, error) {
	if requestscope.IsUnrestricted(ctx, accountID) {
		return true, nil
	}
	var count int64
	err := r.db.WithContext(ctx).Table("account_owner_access").
		Where("account_id=? AND owner_id=?", accountID, ownerID).
		Count(&count).Error
	return count > 0, Error(err)
}

func (r *InventoryScopeRepository) Allowed(ctx context.Context, accountID, ownerID, warehouseID string) (bool, error) {
	if requestscope.IsUnrestricted(ctx, accountID) {
		return true, nil
	}
	var count int64
	err := r.db.WithContext(ctx).Table("account_owner_access ao").
		Joins("JOIN account_warehouse_access aw ON aw.account_id=ao.account_id AND aw.warehouse_id=?", warehouseID).
		Joins("JOIN warehouse_owner wo ON wo.owner_id=ao.owner_id AND wo.warehouse_id=aw.warehouse_id AND wo.is_active").
		Where("ao.account_id=? AND ao.owner_id=?", accountID, ownerID).Count(&count).Error
	return count > 0, Error(err)
}

func (r *InventoryScopeRepository) ResourceScope(ctx context.Context, resource, id string) (string, string, error) {
	queries := map[string]string{
		"balance":       "SELECT owner_id::text, warehouse_id::text FROM inventory_balance WHERE balance_id=?",
		"movement":      "SELECT owner_id::text, warehouse_id::text FROM inventory_movement WHERE movement_id=?",
		"serial-state":  "SELECT state.owner_id::text, balance.warehouse_id::text FROM serial_inventory state JOIN inventory_balance balance ON balance.balance_id=state.balance_id WHERE state.serial_id=?",
		"lot":           "SELECT owner_id::text, '' warehouse_id FROM inventory_lot WHERE lot_id=?",
		"serial":        "SELECT owner_id::text, '' warehouse_id FROM serial_number WHERE serial_id=?",
		"handling-unit": "SELECT owner_id::text, warehouse_id::text FROM handling_unit WHERE handling_unit_id=?",
	}
	query, exists := queries[resource]
	if !exists {
		return "", "", fmt.Errorf("unknown scoped inventory resource")
	}
	var value struct {
		OwnerID     string
		WarehouseID string
	}
	err := r.db.WithContext(ctx).Raw(query, id).Take(&value).Error
	return value.OwnerID, value.WarehouseID, Error(err)
}
