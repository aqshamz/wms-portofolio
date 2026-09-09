package inbound

import (
	"context"

	model "wms-api/models/inbound"

	"gorm.io/gorm"
)

type InboundOrderLineRow struct {
	model.InboundOrderLine
	ItemCode, ItemName, UOMCode string
	CompletedReceiptQty         string
}

type InboundOrderLineRepository struct{ db *gorm.DB }

func NewInboundOrderLineRepository(db *gorm.DB) *InboundOrderLineRepository {
	return &InboundOrderLineRepository{db: db}
}
func (r *InboundOrderLineRepository) Create(ctx context.Context, value *model.InboundOrderLine) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}
func (r *InboundOrderLineRepository) NextLineNo(ctx context.Context, inboundID string) (int, error) {
	var value struct{ Next int }
	err := r.db.WithContext(ctx).Model(&model.InboundOrderLine{}).Select("COALESCE(max(line_no),0)+1 next").Where("inbound_id=?", inboundID).Scan(&value).Error
	return value.Next, Error(err)
}
func (r *InboundOrderLineRepository) UpdateDraft(ctx context.Context, id, inboundID string, values map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&model.InboundOrderLine{}).Where("inbound_line_id=? AND inbound_id=?", id, inboundID).Updates(values)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
func (r *InboundOrderLineRepository) DeleteDraft(ctx context.Context, id, inboundID string) error {
	result := r.db.WithContext(ctx).Where("inbound_line_id=? AND inbound_id=?", id, inboundID).Delete(&model.InboundOrderLine{})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
func (r *InboundOrderLineRepository) Get(ctx context.Context, id string) (model.InboundOrderLine, error) {
	var value model.InboundOrderLine
	err := r.db.WithContext(ctx).Where("inbound_line_id=?", id).Take(&value).Error
	return value, Error(err)
}
func inboundOrderLineQuery(db *gorm.DB) *gorm.DB {
	return db.Table("inbound_order_line line").Select(`line.*,item.code item_code,item.name item_name,uom.code uom_code,
 COALESCE((SELECT sum(receipt_line.received_qty) FROM receipt_line JOIN receipt ON receipt.receipt_id=receipt_line.receipt_id JOIN document_status receipt_status ON receipt_status.status_id=receipt.status_id WHERE receipt_line.inbound_line_id=line.inbound_line_id AND receipt_status.code='COMPLETED'),0) completed_receipt_qty`).
		Joins("JOIN item ON item.item_id=line.item_id").Joins("JOIN uom ON uom.uom_id=line.uom_id")
}
func (r *InboundOrderLineRepository) List(ctx context.Context, inboundID string) ([]InboundOrderLineRow, error) {
	rows := make([]InboundOrderLineRow, 0)
	err := inboundOrderLineQuery(r.db.WithContext(ctx)).Where("line.inbound_id=?", inboundID).Order("line.line_no").Find(&rows).Error
	return rows, Error(err)
}
func (r *InboundOrderLineRepository) Count(ctx context.Context, inboundID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.InboundOrderLine{}).Where("inbound_id=?", inboundID).Count(&count).Error
	return count, Error(err)
}
func (r *InboundOrderLineRepository) ScheduledForPurchaseOrderLine(ctx context.Context, purchaseOrderLineID string) (string, error) {
	var total string
	err := r.db.WithContext(ctx).Table("inbound_order_line line").Select(`COALESCE(sum(CASE WHEN status.code='CANCELLED' THEN COALESCE((SELECT sum(receipt_line.received_qty) FROM receipt_line JOIN receipt ON receipt.receipt_id=receipt_line.receipt_id JOIN document_status receipt_status ON receipt_status.status_id=receipt.status_id WHERE receipt_line.inbound_line_id=line.inbound_line_id AND receipt_status.code='COMPLETED'),0) ELSE line.expected_qty END),0)::text`).Joins("JOIN inbound_order inbound ON inbound.inbound_id=line.inbound_id").Joins("JOIN document_status status ON status.status_id=inbound.status_id").Where("line.purchase_order_line_id=?", purchaseOrderLineID).Scan(&total).Error
	return total, Error(err)
}

func (r *InboundOrderLineRepository) ScheduledExcept(ctx context.Context, purchaseOrderLineID, excludedInboundLineID string) (string, error) {
	var total string
	err := r.db.WithContext(ctx).Table("inbound_order_line line").Select(`COALESCE(sum(CASE WHEN status.code='CANCELLED' THEN COALESCE((SELECT sum(receipt_line.received_qty) FROM receipt_line JOIN receipt ON receipt.receipt_id=receipt_line.receipt_id JOIN document_status receipt_status ON receipt_status.status_id=receipt.status_id WHERE receipt_line.inbound_line_id=line.inbound_line_id AND receipt_status.code='COMPLETED'),0) ELSE line.expected_qty END),0)::text`).Joins("JOIN inbound_order inbound ON inbound.inbound_id=line.inbound_id").Joins("JOIN document_status status ON status.status_id=inbound.status_id").Where("line.purchase_order_line_id=? AND line.inbound_line_id<>?", purchaseOrderLineID, excludedInboundLineID).Scan(&total).Error
	return total, Error(err)
}
