package outbound

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type ShipmentLineRow struct {
	model.ShipmentLine
	OutboundID, OutboundLineID, ItemCode, LotNumber, UOMCode string
	SourceBalanceVersion                                     int64
}
type ShipmentLineRepository struct{ db *gorm.DB }

func NewShipmentLineRepository(db *gorm.DB) *ShipmentLineRepository {
	return &ShipmentLineRepository{db: db}
}
func shipmentLineQuery(db *gorm.DB, shipmentID string) *gorm.DB {
	return db.Table("packing_line pl").Select("COALESCE(sl.shipment_line_id,'') shipment_line_id,? shipment_id,pl.packing_line_id,pl.packing_balance_id source_balance_id,COALESCE(sl.shipped_qty,pl.packed_qty)::text shipped_qty,pl.uom_id,COALESCE(sl.movement_id,'') movement_id,COALESCE(sl.created_at,pl.created_at) created_at,COALESCE(sl.created_by,pl.created_by) created_by,p.outbound_id,pt.outbound_line_id,i.code item_code,COALESCE(lot.lot_number,'') lot_number,u.code uom_code,b.version_no source_balance_version", shipmentID).Joins("JOIN packing p ON p.packing_id=pl.packing_id").Joins("JOIN shipment_packing sp ON sp.packing_id=p.packing_id").Joins("JOIN inventory_balance b ON b.balance_id=pl.packing_balance_id").Joins("JOIN outbound_check_line cl ON cl.outbound_check_line_id=pl.outbound_check_line_id").Joins("JOIN outbound_staging_line osl ON osl.staging_line_id=cl.staging_line_id").Joins("JOIN pick_execution pe ON pe.pick_execution_id=osl.pick_execution_id").Joins("JOIN pick_task pt ON pt.pick_task_id=pe.pick_task_id").Joins("JOIN outbound_order_line ol ON ol.outbound_line_id=pt.outbound_line_id").Joins("JOIN item i ON i.item_id=ol.item_id").Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=b.lot_id").Joins("JOIN uom u ON u.uom_id=pl.uom_id").Joins("LEFT JOIN shipment_line sl ON sl.packing_line_id=pl.packing_line_id")
}
func (r *ShipmentLineRepository) Candidates(ctx context.Context, id string) ([]ShipmentLineRow, error) {
	var v []ShipmentLineRow
	err := shipmentLineQuery(r.db.WithContext(ctx), id).Where("sp.shipment_id=?", id).Order("p.outbound_id,pl.packing_line_id").Find(&v).Error
	return v, Error(err)
}
func (r *ShipmentLineRepository) GetCandidate(ctx context.Context, shipmentID, packingLineID string) (ShipmentLineRow, error) {
	var v ShipmentLineRow
	err := shipmentLineQuery(r.db.WithContext(ctx), shipmentID).Where("sp.shipment_id=? AND pl.packing_line_id=?", shipmentID, packingLineID).Take(&v).Error
	return v, Error(err)
}
func (r *ShipmentLineRepository) Create(ctx context.Context, v *model.ShipmentLine) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *ShipmentLineRepository) MissingCount(ctx context.Context, id string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("shipment_packing sp").Joins("JOIN packing_line pl ON pl.packing_id=sp.packing_id").Joins("LEFT JOIN shipment_line sl ON sl.packing_line_id=pl.packing_line_id").Where("sp.shipment_id=? AND sl.shipment_line_id IS NULL", id).Count(&n).Error
	return n, Error(err)
}
func (r *ShipmentLineRepository) Count(ctx context.Context, id string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.ShipmentLine{}).Where("shipment_id=?", id).Count(&n).Error
	return n, Error(err)
}
