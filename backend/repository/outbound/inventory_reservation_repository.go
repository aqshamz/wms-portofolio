package outbound

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/outbound"
)

type AllocationCandidate struct {
	BalanceID, OwnerID, WarehouseID, LocationID, LocationCode, ItemID, UOMID, InventoryStatusID, InventoryStatusCode, OnHandQty, ReservedQty, AvailableQty string
	LotID, LotNumber, HandlingUnitID                                                                                                                       *string
	VersionNo                                                                                                                                              int64
}
type ReservationRow struct {
	model.InventoryReservation
	StatusCode, OutboundID, ItemCode, LocationCode, LotNumber string
}
type InventoryReservationRepository struct{ db *gorm.DB }

func NewInventoryReservationRepository(db *gorm.DB) *InventoryReservationRepository {
	return &InventoryReservationRepository{db: db}
}
func (r *InventoryReservationRepository) Candidates(ctx context.Context, lineID string, strategyID *string) ([]AllocationCandidate, error) {
	var v []AllocationCandidate
	q := r.db.WithContext(ctx).Table("outbound_order_line l").Select("b.balance_id,b.owner_id,b.warehouse_id,b.location_id,loc.code location_code,b.item_id,b.uom_id,b.inventory_status_id,st.code inventory_status_code,b.on_hand_qty,b.reserved_qty,(b.on_hand_qty-b.reserved_qty) available_qty,b.lot_id,lot.lot_number,b.handling_unit_id,b.version_no").Joins("JOIN outbound_order o ON o.outbound_id=l.outbound_id").Joins("JOIN inventory_balance b ON b.owner_id=o.owner_id AND b.warehouse_id=o.warehouse_id AND b.item_id=l.item_id AND b.uom_id=l.uom_id").Joins("JOIN inventory_status st ON st.inventory_status_id=b.inventory_status_id AND st.is_active AND st.is_allocatable AND st.is_pickable").Joins("JOIN warehouse_location loc ON loc.location_id=b.location_id AND loc.is_active AND NOT loc.is_locked").Joins("JOIN location_type lt ON lt.location_type_id=loc.location_type_id AND lt.is_active AND lt.allows_picking").Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=b.lot_id").Where("l.outbound_line_id=? AND b.on_hand_qty>b.reserved_qty AND (l.requested_lot_no IS NULL OR lot.lot_number=l.requested_lot_no)", lineID)
	if strategyID != nil {
		q = q.Joins("LEFT JOIN picking_strategy_rule pr ON pr.picking_strategy_id=? AND pr.is_active AND (pr.inventory_status_id IS NULL OR pr.inventory_status_id=b.inventory_status_id) AND (pr.zone_id IS NULL OR pr.zone_id=loc.zone_id)", *strategyID).Joins("LEFT JOIN picking_sort_method sm ON sm.picking_sort_method_id=pr.picking_sort_method_id").Order("pr.sequence_no NULLS LAST,CASE WHEN sm.code='FEFO' THEN lot.expiry_date END NULLS LAST,CASE WHEN sm.code='FIFO' THEN lot.created_at END NULLS LAST,CASE WHEN sm.code='LOCATION' THEN loc.pick_sequence END NULLS LAST")
	}
	err := q.Order("loc.pick_sequence,loc.code,b.balance_id").Find(&v).Error
	return v, Error(err)
}
func (r *InventoryReservationRepository) Create(ctx context.Context, v *model.InventoryReservation) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *InventoryReservationRepository) Lock(ctx context.Context, id string) (model.InventoryReservation, error) {
	var v model.InventoryReservation
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("reservation_id=?", id).Take(&v).Error
	return v, Error(err)
}

func (r *InventoryReservationRepository) ReplacementEligible(ctx context.Context, id, outboundLineID, stagingID string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("inventory_reservation r").
		Joins("JOIN document_status s ON s.status_id=r.status_id AND s.code='CONSUMED'").
		Joins("JOIN pick_task p ON p.reservation_id=r.reservation_id").
		Joins("JOIN pick_execution e ON e.pick_task_id=p.pick_task_id").
		Joins("JOIN outbound_staging_line l ON l.pick_execution_id=e.pick_execution_id AND l.staging_id=?", stagingID).
		Where("r.reservation_id=? AND r.outbound_line_id=?", id, outboundLineID).Count(&n).Error
	return n == 1, Error(err)
}
func (r *InventoryReservationRepository) List(ctx context.Context, f ListFilter) ([]ReservationRow, int64, error) {
	q := r.db.WithContext(ctx).Table("inventory_reservation r").Select("r.*,s.code status_code,o.outbound_id,i.code item_code,loc.code location_code,COALESCE(lot.lot_number,'') lot_number").Joins("JOIN document_status s ON s.status_id=r.status_id").Joins("JOIN outbound_order_line ol ON ol.outbound_line_id=r.outbound_line_id").Joins("JOIN outbound_order o ON o.outbound_id=ol.outbound_id").Joins("JOIN inventory_balance b ON b.balance_id=r.balance_id").Joins("JOIN item i ON i.item_id=ol.item_id").Joins("JOIN warehouse_location loc ON loc.location_id=b.location_id").Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=b.lot_id").Where("o.owner_id=?", f.OwnerID)
	if f.WarehouseID != "" {
		q = q.Where("o.warehouse_id=?", f.WarehouseID)
	}
	if f.StatusCode != "" {
		q = q.Where("s.code=?", f.StatusCode)
	}
	if f.Search != "" {
		x := "%" + f.Search + "%"
		q = q.Where("r.reservation_id ILIKE ? OR o.outbound_id ILIKE ? OR i.code ILIKE ?", x, x, x)
	}
	var n int64
	if err := q.Session(&gorm.Session{}).Count(&n).Error; err != nil {
		return nil, 0, Error(err)
	}
	var v []ReservationRow
	err := q.Order("r.reserved_at DESC,r.reservation_id").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&v).Error
	return v, n, Error(err)
}
func (r *InventoryReservationRepository) ListActiveByOrder(ctx context.Context, id string) ([]ReservationRow, error) {
	var v []ReservationRow
	err := r.db.WithContext(ctx).Table("inventory_reservation r").Select("r.*,s.code status_code,o.outbound_id").Joins("JOIN document_status s ON s.status_id=r.status_id").Joins("JOIN outbound_order_line l ON l.outbound_line_id=r.outbound_line_id").Joins("JOIN outbound_order o ON o.outbound_id=l.outbound_id").Where("o.outbound_id=? AND s.code IN ('ACTIVE','PARTIALLY_PICKED')", id).Order("r.reservation_id").Find(&v).Error
	return v, Error(err)
}
func (r *InventoryReservationRepository) ListByOrder(ctx context.Context, id string) ([]ReservationRow, error) {
	var v []ReservationRow
	err := r.db.WithContext(ctx).Table("inventory_reservation r").Select("r.*,s.code status_code,o.outbound_id,i.code item_code,loc.code location_code,COALESCE(lot.lot_number,'') lot_number").Joins("JOIN document_status s ON s.status_id=r.status_id").Joins("JOIN outbound_order_line ol ON ol.outbound_line_id=r.outbound_line_id").Joins("JOIN outbound_order o ON o.outbound_id=ol.outbound_id").Joins("JOIN inventory_balance b ON b.balance_id=r.balance_id").Joins("JOIN item i ON i.item_id=ol.item_id").Joins("JOIN warehouse_location loc ON loc.location_id=b.location_id").Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=b.lot_id").Where("o.outbound_id=?", id).Order("r.reserved_at,r.reservation_id").Find(&v).Error
	return v, Error(err)
}
func (r *InventoryReservationRepository) UpdatePicked(ctx context.Context, id, statusID, qty string) error {
	return Error(r.db.WithContext(ctx).Model(&model.InventoryReservation{}).Where("reservation_id=?", id).Updates(map[string]interface{}{"picked_qty": gorm.Expr("picked_qty+?", qty), "status_id": statusID}).Error)
}
func (r *InventoryReservationRepository) Release(ctx context.Context, id, statusID string) error {
	return Error(r.db.WithContext(ctx).Model(&model.InventoryReservation{}).Where("reservation_id=?", id).Updates(map[string]interface{}{"status_id": statusID, "released_at": gorm.Expr("clock_timestamp()")}).Error)
}
