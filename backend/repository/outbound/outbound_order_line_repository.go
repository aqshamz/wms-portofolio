package outbound

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type OutboundOrderLineRow struct {
	model.OutboundOrderLine
	ItemCode, ItemName, UOMCode string
}
type OutboundOrderLineRepository struct{ db *gorm.DB }

func NewOutboundOrderLineRepository(db *gorm.DB) *OutboundOrderLineRepository {
	return &OutboundOrderLineRepository{db: db}
}
func (r *OutboundOrderLineRepository) CreateBatch(ctx context.Context, v []model.OutboundOrderLine) error {
	return Error(r.db.WithContext(ctx).Create(&v).Error)
}
func (r *OutboundOrderLineRepository) Create(ctx context.Context, v *model.OutboundOrderLine) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *OutboundOrderLineRepository) NextLineNo(ctx context.Context, outboundID string) (int, error) {
	var value struct{ Next int }
	err := r.db.WithContext(ctx).Model(&model.OutboundOrderLine{}).
		Select("COALESCE(max(line_no),0)+1 next").Where("outbound_id=?", outboundID).Scan(&value).Error
	return value.Next, Error(err)
}
func (r *OutboundOrderLineRepository) UpdateDraft(ctx context.Context, id, outboundID string, values map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&model.OutboundOrderLine{}).
		Where("outbound_line_id=? AND outbound_id=?", id, outboundID).Updates(values)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
func (r *OutboundOrderLineRepository) DeleteDraft(ctx context.Context, id, outboundID string) error {
	result := r.db.WithContext(ctx).Where("outbound_line_id=? AND outbound_id=?", id, outboundID).
		Delete(&model.OutboundOrderLine{})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
func (r *OutboundOrderLineRepository) List(ctx context.Context, id string) ([]OutboundOrderLineRow, error) {
	rows := make([]OutboundOrderLineRow, 0)
	err := r.db.WithContext(ctx).Table("outbound_order_line l").Select("l.*,i.code item_code,i.name item_name,u.code uom_code").Joins("JOIN item i ON i.item_id=l.item_id").Joins("JOIN uom u ON u.uom_id=l.uom_id").Where("l.outbound_id=?", id).Order("l.line_no").Find(&rows).Error
	return rows, Error(err)
}
func (r *OutboundOrderLineRepository) Get(ctx context.Context, id string) (OutboundOrderLineRow, error) {
	var v OutboundOrderLineRow
	err := r.db.WithContext(ctx).Table("outbound_order_line l").Select("l.*,i.code item_code,i.name item_name,u.code uom_code").Joins("JOIN item i ON i.item_id=l.item_id").Joins("JOIN uom u ON u.uom_id=l.uom_id").Where("l.outbound_line_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *OutboundOrderLineRepository) AddAllocated(ctx context.Context, id, qty string) error {
	return Error(r.db.WithContext(ctx).Model(&model.OutboundOrderLine{}).Where("outbound_line_id=?", id).Update("allocated_qty", gorm.Expr("allocated_qty+?", qty)).Error)
}
func (r *OutboundOrderLineRepository) SubtractAllocated(ctx context.Context, id, qty string) error {
	return Error(r.db.WithContext(ctx).Model(&model.OutboundOrderLine{}).Where("outbound_line_id=? AND allocated_qty>=?", id, qty).Update("allocated_qty", gorm.Expr("allocated_qty-?", qty)).Error)
}
func (r *OutboundOrderLineRepository) AddPicked(ctx context.Context, id, qty string) error {
	return Error(r.db.WithContext(ctx).Model(&model.OutboundOrderLine{}).Where("outbound_line_id=?", id).Update("picked_qty", gorm.Expr("picked_qty+?", qty)).Error)
}
func (r *OutboundOrderLineRepository) AddChecked(ctx context.Context, id, qty string) error {
	return Error(r.db.WithContext(ctx).Model(&model.OutboundOrderLine{}).Where("outbound_line_id=?", id).Update("checked_qty", gorm.Expr("checked_qty+?", qty)).Error)
}
func (r *OutboundOrderLineRepository) AddPacked(ctx context.Context, id, qty string) error {
	return Error(r.db.WithContext(ctx).Model(&model.OutboundOrderLine{}).Where("outbound_line_id=?", id).Update("packed_qty", gorm.Expr("packed_qty+?", qty)).Error)
}
func (r *OutboundOrderLineRepository) AddShipped(ctx context.Context, id, qty string) error {
	return Error(r.db.WithContext(ctx).Model(&model.OutboundOrderLine{}).Where("outbound_line_id=?", id).Update("shipped_qty", gorm.Expr("shipped_qty+?", qty)).Error)
}
func (r *OutboundOrderLineRepository) AddDelivered(ctx context.Context, id, qty string) error {
	return Error(r.db.WithContext(ctx).Model(&model.OutboundOrderLine{}).Where("outbound_line_id=?", id).Update("delivered_qty", gorm.Expr("delivered_qty+?", qty)).Error)
}
func (r *OutboundOrderLineRepository) AddRejected(ctx context.Context, id, qty string) error {
	return Error(r.db.WithContext(ctx).Model(&model.OutboundOrderLine{}).Where("outbound_line_id=?", id).Update("rejected_qty", gorm.Expr("rejected_qty+?", qty)).Error)
}
func (r *OutboundOrderLineRepository) AcceptShort(ctx context.Context, id, qty string) error {
	return Error(r.db.WithContext(ctx).Model(&model.OutboundOrderLine{}).Where("outbound_line_id=?", id).Update("short_accepted_qty", gorm.Expr("short_accepted_qty+?", qty)).Error)
}
func (r *OutboundOrderLineRepository) AllAllocated(ctx context.Context, id string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.OutboundOrderLine{}).Where("outbound_id=? AND allocated_qty<ordered_qty", id).Count(&n).Error
	return n == 0, Error(err)
}
func (r *OutboundOrderLineRepository) HasAllocation(ctx context.Context, id string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.OutboundOrderLine{}).Where("outbound_id=? AND allocated_qty>0", id).Count(&n).Error
	return n > 0, Error(err)
}
