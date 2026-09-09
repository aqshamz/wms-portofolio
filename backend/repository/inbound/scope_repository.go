package inbound

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type InboundScopeRepository struct{ db *gorm.DB }

func NewInboundScopeRepository(db *gorm.DB) *InboundScopeRepository {
	return &InboundScopeRepository{db: db}
}

func (r *InboundScopeRepository) Allowed(ctx context.Context, accountID, ownerID, warehouseID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("account_owner_access ao").
		Joins("JOIN account_warehouse_access aw ON aw.account_id=ao.account_id AND aw.warehouse_id=?", warehouseID).
		Joins("JOIN warehouse_owner wo ON wo.owner_id=ao.owner_id AND wo.warehouse_id=aw.warehouse_id AND wo.is_active").
		Where("ao.account_id=? AND ao.owner_id=?", accountID, ownerID).Count(&count).Error
	return count > 0, Error(err)
}

func (r *InboundScopeRepository) ResourceScope(ctx context.Context, kind, id string) (string, string, error) {
	queries := map[string]string{
		"purchase-orders":     "SELECT owner_id::text,warehouse_id::text FROM purchase_order WHERE purchase_order_id=?",
		"orders":              "SELECT owner_id::text,warehouse_id::text FROM inbound_order WHERE inbound_id=?",
		"receipts":            "SELECT owner_id::text,warehouse_id::text FROM receipt WHERE receipt_id=?",
		"receipt-inventories": "SELECT receipt.owner_id::text,receipt.warehouse_id::text FROM receipt_inventory batch JOIN receipt_line line ON line.receipt_line_id=batch.receipt_line_id JOIN receipt ON receipt.receipt_id=line.receipt_id WHERE batch.receipt_inventory_id=?",
		"quality-inspections": "SELECT receipt.owner_id::text,receipt.warehouse_id::text FROM quality_inspection inspection JOIN receipt_inventory batch ON batch.receipt_inventory_id=inspection.receipt_inventory_id JOIN receipt_line line ON line.receipt_line_id=batch.receipt_line_id JOIN receipt ON receipt.receipt_id=line.receipt_id WHERE inspection.inspection_id=?",
		"putaway-tasks":       "SELECT owner_id::text,warehouse_id::text FROM putaway_task WHERE putaway_task_id=?",
		"quarantine-cases":    "SELECT owner_id::text,warehouse_id::text FROM quarantine_case WHERE quarantine_case_id=?",
		"exceptions":          "SELECT owner_id::text,warehouse_id::text FROM inbound_exception WHERE inbound_exception_id=?",
		"rework-tasks":        "SELECT owner_id::text,warehouse_id::text FROM rework_task WHERE rework_task_id=?",
	}
	query, ok := queries[kind]
	if !ok {
		return "", "", fmt.Errorf("unknown inbound resource kind")
	}
	var value struct{ OwnerID, WarehouseID string }
	err := r.db.WithContext(ctx).Raw(query, id).Take(&value).Error
	return value.OwnerID, value.WarehouseID, Error(err)
}
