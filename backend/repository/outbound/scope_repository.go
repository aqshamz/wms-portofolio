package outbound

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type OutboundScopeRepository struct{ db *gorm.DB }

func NewOutboundScopeRepository(db *gorm.DB) *OutboundScopeRepository {
	return &OutboundScopeRepository{db: db}
}

func (r *OutboundScopeRepository) Allowed(ctx context.Context, accountID, ownerID, warehouseID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("account_owner_access ao").
		Joins("JOIN account_warehouse_access aw ON aw.account_id=ao.account_id AND aw.warehouse_id=?", warehouseID).
		Joins("JOIN warehouse_owner wo ON wo.owner_id=ao.owner_id AND wo.warehouse_id=aw.warehouse_id AND wo.is_active").
		Where("ao.account_id=? AND ao.owner_id=?", accountID, ownerID).Count(&count).Error
	return count > 0, Error(err)
}

func (r *OutboundScopeRepository) ResourceScope(ctx context.Context, kind, id string) (string, string, error) {
	queries := map[string]string{
		"orders":           "SELECT o.owner_id::text owner_id,o.warehouse_id::text warehouse_id FROM outbound_order o WHERE o.outbound_id=?",
		"validation-runs":  "SELECT o.owner_id::text owner_id,o.warehouse_id::text warehouse_id FROM outbound_validation_run v JOIN outbound_order o ON o.outbound_id=v.outbound_id WHERE v.validation_run_id=?",
		"reservations":     "SELECT o.owner_id::text owner_id,o.warehouse_id::text warehouse_id FROM inventory_reservation r JOIN outbound_order_line l ON l.outbound_line_id=r.outbound_line_id JOIN outbound_order o ON o.outbound_id=l.outbound_id WHERE r.reservation_id=?",
		"waves":            "SELECT w.owner_id::text owner_id,w.warehouse_id::text warehouse_id FROM outbound_wave w WHERE w.wave_id=?",
		"pick-tasks":       "SELECT o.owner_id::text owner_id,o.warehouse_id::text warehouse_id FROM pick_task p JOIN outbound_order_line l ON l.outbound_line_id=p.outbound_line_id JOIN outbound_order o ON o.outbound_id=l.outbound_id WHERE p.pick_task_id=?",
		"stagings":         "SELECT o.owner_id::text owner_id,s.warehouse_id::text warehouse_id FROM outbound_staging s JOIN outbound_order o ON o.outbound_id=s.outbound_id WHERE s.staging_id=?",
		"checks":           "SELECT o.owner_id::text owner_id,s.warehouse_id::text warehouse_id FROM outbound_check c JOIN outbound_staging s ON s.staging_id=c.staging_id JOIN outbound_order o ON o.outbound_id=s.outbound_id WHERE c.outbound_check_id=?",
		"check-exceptions": "SELECT o.owner_id::text owner_id,o.warehouse_id::text warehouse_id FROM outbound_check_exception e JOIN outbound_check_line l ON l.outbound_check_line_id=e.outbound_check_line_id JOIN outbound_check c ON c.outbound_check_id=l.outbound_check_id JOIN outbound_staging s ON s.staging_id=c.staging_id JOIN outbound_order o ON o.outbound_id=s.outbound_id WHERE e.outbound_check_exception_id=?",
		"packings":         "SELECT o.owner_id::text owner_id,p.warehouse_id::text warehouse_id FROM packing p JOIN outbound_order o ON o.outbound_id=p.outbound_id WHERE p.packing_id=?",
		"shipments":        "SELECT s.owner_id::text owner_id,s.warehouse_id::text warehouse_id FROM shipment s WHERE s.shipment_id=?",
		"deliveries":       "SELECT o.owner_id::text owner_id,o.warehouse_id::text warehouse_id FROM delivery d JOIN outbound_order o ON o.outbound_id=d.outbound_id WHERE d.delivery_id=?",
	}
	query, ok := queries[kind]
	if !ok {
		return "", "", fmt.Errorf("unknown outbound resource kind")
	}
	var value struct{ OwnerID, WarehouseID string }
	err := r.db.WithContext(ctx).Raw(query, id).Take(&value).Error
	return value.OwnerID, value.WarehouseID, Error(err)
}
