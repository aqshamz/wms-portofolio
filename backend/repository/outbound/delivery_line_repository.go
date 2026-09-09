package outbound

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/outbound"
)

type DeliveryLineRow struct {
	model.DeliveryLine
	ItemCode, LotNumber, UOMCode, OutboundLineID, SourceBalanceID, ShipmentMovementID string
}
type DeliverySourceLine struct{ ShipmentLineID, OutboundLineID, SourceBalanceID, PlannedQty, UOMID string }
type DeliveryLineRepository struct{ db *gorm.DB }

func NewDeliveryLineRepository(db *gorm.DB) *DeliveryLineRepository {
	return &DeliveryLineRepository{db: db}
}
func deliveryLineQuery(db *gorm.DB) *gorm.DB {
	return db.Table("delivery_line dl").Select("dl.*,i.code item_code,COALESCE(lot.lot_number,'') lot_number,u.code uom_code,pt.outbound_line_id,sl.source_balance_id,sl.movement_id shipment_movement_id").
		Joins("JOIN shipment_line sl ON sl.shipment_line_id=dl.shipment_line_id").Joins("JOIN packing_line pl ON pl.packing_line_id=sl.packing_line_id").
		Joins("JOIN pick_task pt ON pt.pick_task_id=pl.pick_task_id").Joins("JOIN outbound_order_line ol ON ol.outbound_line_id=pt.outbound_line_id").
		Joins("JOIN item i ON i.item_id=ol.item_id").Joins("JOIN inventory_balance b ON b.balance_id=sl.source_balance_id").
		Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=b.lot_id").Joins("JOIN uom u ON u.uom_id=dl.uom_id")
}
func (r *DeliveryLineRepository) CreateBatch(ctx context.Context, v []model.DeliveryLine) error {
	return Error(r.db.WithContext(ctx).Create(&v).Error)
}
func (r *DeliveryLineRepository) SourceLines(ctx context.Context, shipmentID, outboundID string) ([]DeliverySourceLine, error) {
	var v []DeliverySourceLine
	err := r.db.WithContext(ctx).Table("shipment_line sl").Select("sl.shipment_line_id,pt.outbound_line_id,sl.source_balance_id,sl.shipped_qty::text planned_qty,sl.uom_id").
		Joins("JOIN packing_line pl ON pl.packing_line_id=sl.packing_line_id").Joins("JOIN packing p ON p.packing_id=pl.packing_id").
		Joins("JOIN pick_task pt ON pt.pick_task_id=pl.pick_task_id").Where("sl.shipment_id=? AND p.outbound_id=?", shipmentID, outboundID).Order("sl.shipment_line_id").Find(&v).Error
	return v, Error(err)
}
func (r *DeliveryLineRepository) List(ctx context.Context, id string) ([]DeliveryLineRow, error) {
	var v []DeliveryLineRow
	err := deliveryLineQuery(r.db.WithContext(ctx)).Where("dl.delivery_id=?", id).Order("dl.delivery_line_id").Find(&v).Error
	return v, Error(err)
}
func (r *DeliveryLineRepository) Get(ctx context.Context, id string) (DeliveryLineRow, error) {
	var v DeliveryLineRow
	err := deliveryLineQuery(r.db.WithContext(ctx)).Where("dl.delivery_line_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *DeliveryLineRepository) Lock(ctx context.Context, id string) (model.DeliveryLine, error) {
	var v model.DeliveryLine
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("delivery_line_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *DeliveryLineRepository) AddDelivered(ctx context.Context, id, qty string) error {
	res := r.db.WithContext(ctx).Model(&model.DeliveryLine{}).Where("delivery_line_id=?", id).Update("delivered_qty", gorm.Expr("delivered_qty+?", qty))
	if res.Error != nil {
		return Error(res.Error)
	}
	if res.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
func (r *DeliveryLineRepository) AddReturned(ctx context.Context, id, qty string) error {
	res := r.db.WithContext(ctx).Model(&model.DeliveryLine{}).Where("delivery_line_id=?", id).Update("returned_qty", gorm.Expr("returned_qty+?", qty))
	if res.Error != nil {
		return Error(res.Error)
	}
	if res.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
func (r *DeliveryLineRepository) IncompleteCount(ctx context.Context, id string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.DeliveryLine{}).Where("delivery_id=? AND delivered_qty<planned_qty", id).Count(&n).Error
	return n, Error(err)
}
func (r *DeliveryLineRepository) UnaccountedCount(ctx context.Context, id string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.DeliveryLine{}).Where("delivery_id=? AND delivered_qty+returned_qty<planned_qty", id).Count(&n).Error
	return n, Error(err)
}
func (r *DeliveryLineRepository) DeliveredTotal(ctx context.Context, id string) (string, error) {
	var v struct{ Total string }
	err := r.db.WithContext(ctx).Model(&model.DeliveryLine{}).Select("COALESCE(sum(delivered_qty),0)::text total").Where("delivery_id=?", id).Scan(&v).Error
	return v.Total, Error(err)
}
