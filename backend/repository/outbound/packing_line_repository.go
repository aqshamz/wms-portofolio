package outbound

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type PackingLineRow struct {
	model.PackingLine
	PickedSerialID                                           *string
	SerialNo                                                 string
	HandlingUnitBarcode                                      string
	OutboundID, OutboundLineID, ItemCode, LotNumber, UOMCode string
	SourceBalanceVersion                                     int64
	ExpectedQty                                              string
}
type PackingLineRepository struct{ db *gorm.DB }

func NewPackingLineRepository(db *gorm.DB) *PackingLineRepository {
	return &PackingLineRepository{db: db}
}
func packingLineQuery(db *gorm.DB, packingID string) *gorm.DB {
	return db.Table("outbound_check_line cl").Select("COALESCE(pl.packing_line_id,'') packing_line_id,? packing_id,cl.outbound_check_line_id,pt.pick_task_id,sl.staging_balance_id source_balance_id,COALESCE(pl.packing_balance_id,'') packing_balance_id,COALESCE(pl.handling_unit_id,b.handling_unit_id) handling_unit_id,COALESCE(hu.barcode,'') handling_unit_barcode,COALESCE(pl.packed_qty,0)::text packed_qty,cl.uom_id,COALESCE(pl.movement_id,'') movement_id,COALESCE(pl.created_at,cl.created_at) created_at,COALESCE(pl.created_by,cl.created_by) created_by,p.outbound_id,pt.outbound_line_id,i.code item_code,COALESCE(lot.lot_number,'') lot_number,u.code uom_code,b.version_no source_balance_version,cl.checked_qty::text expected_qty,pm.serial_id picked_serial_id,COALESCE(sn.serial_no,'') serial_no", packingID).Joins("JOIN outbound_check c ON c.outbound_check_id=cl.outbound_check_id").Joins("JOIN document_status cs ON cs.status_id=c.status_id AND cs.code='PASSED'").Joins("JOIN packing p ON p.outbound_id=(SELECT st.outbound_id FROM outbound_staging st WHERE st.staging_id=c.staging_id)").Joins("JOIN outbound_staging_line sl ON sl.staging_line_id=cl.staging_line_id").Joins("JOIN pick_execution pe ON pe.pick_execution_id=sl.pick_execution_id").Joins("JOIN pick_task pt ON pt.pick_task_id=pe.pick_task_id").Joins("JOIN outbound_order_line ol ON ol.outbound_line_id=pt.outbound_line_id").Joins("JOIN item i ON i.item_id=ol.item_id").Joins("JOIN inventory_balance b ON b.balance_id=sl.staging_balance_id").Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=b.lot_id").Joins("JOIN uom u ON u.uom_id=cl.uom_id").Joins("LEFT JOIN packing_line pl ON pl.outbound_check_line_id=cl.outbound_check_line_id AND pl.packing_id=p.packing_id").Where("cl.checked_qty>0").Joins("LEFT JOIN handling_unit hu ON hu.handling_unit_id=COALESCE(pl.handling_unit_id,b.handling_unit_id)").Joins("LEFT JOIN inventory_movement pm ON pm.movement_id=pe.movement_id").Joins("LEFT JOIN serial_number sn ON sn.serial_id=pm.serial_id")
}
func (r *PackingLineRepository) Candidates(ctx context.Context, packingID string) ([]PackingLineRow, error) {
	var v []PackingLineRow
	err := packingLineQuery(r.db.WithContext(ctx), packingID).Where("p.packing_id=? AND p.outbound_check_id=c.outbound_check_id", packingID).Order("cl.line_no").Find(&v).Error
	return v, Error(err)
}
func (r *PackingLineRepository) GetCandidate(ctx context.Context, packingID, checkLineID string) (PackingLineRow, error) {
	var v PackingLineRow
	err := packingLineQuery(r.db.WithContext(ctx), packingID).Where("p.packing_id=? AND p.outbound_check_id=c.outbound_check_id AND cl.outbound_check_line_id=?", packingID, checkLineID).Take(&v).Error
	return v, Error(err)
}
func (r *PackingLineRepository) Create(ctx context.Context, v *model.PackingLine) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *PackingLineRepository) MissingCount(ctx context.Context, packingID string) (int64, error) {
	var n int64
	err := packingLineQuery(r.db.WithContext(ctx), packingID).Where("p.packing_id=? AND p.outbound_check_id=c.outbound_check_id AND pl.packing_line_id IS NULL", packingID).Count(&n).Error
	return n, Error(err)
}
func (r *PackingLineRepository) Count(ctx context.Context, packingID string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.PackingLine{}).Where("packing_id=?", packingID).Count(&n).Error
	return n, Error(err)
}
