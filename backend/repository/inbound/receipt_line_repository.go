package inbound

import (
	"context"

	model "wms-api/models/inbound"

	"gorm.io/gorm"
)

type ReceiptLineRow struct {
	model.ReceiptLine
	ItemCode, ItemName, UOMCode, AcceptedQty, BatchedQty string
}

type ReceiptLineRepository struct{ db *gorm.DB }

func NewReceiptLineRepository(db *gorm.DB) *ReceiptLineRepository {
	return &ReceiptLineRepository{db: db}
}
func (r *ReceiptLineRepository) Create(ctx context.Context, value *model.ReceiptLine) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}
func (r *ReceiptLineRepository) DeleteByReceipt(ctx context.Context, receiptID string) error {
	return Error(r.db.WithContext(ctx).Where("receipt_id=?", receiptID).Delete(&model.ReceiptLine{}).Error)
}
func (r *ReceiptLineRepository) Get(ctx context.Context, id string) (model.ReceiptLine, error) {
	var value model.ReceiptLine
	err := r.db.WithContext(ctx).Where("receipt_line_id=?", id).Take(&value).Error
	return value, Error(err)
}
func (r *ReceiptLineRepository) List(ctx context.Context, receiptID string) ([]ReceiptLineRow, error) {
	rows := make([]ReceiptLineRow, 0)
	err := r.db.WithContext(ctx).Table("receipt_line line").Select(`line.*,item.code item_code,item.name item_name,uom.code uom_code,(line.received_qty-line.rejected_qty)::text accepted_qty,COALESCE((SELECT sum(batch.source_qty) FROM receipt_inventory batch WHERE batch.receipt_line_id=line.receipt_line_id),0)::text batched_qty`).Joins("JOIN item ON item.item_id=line.item_id").Joins("JOIN uom ON uom.uom_id=line.uom_id").Where("line.receipt_id=?", receiptID).Order("line.line_no").Find(&rows).Error
	return rows, Error(err)
}
func (r *ReceiptLineRepository) ReceivedForInboundLine(ctx context.Context, inboundLineID string) (string, error) {
	var total string
	err := r.db.WithContext(ctx).Table("receipt_line line").Select("COALESCE(sum(line.received_qty),0)::text").Joins("JOIN receipt ON receipt.receipt_id=line.receipt_id").Joins("JOIN document_status status ON status.status_id=receipt.status_id").Where("line.inbound_line_id=? AND status.code IN ('OPEN','COMPLETED')", inboundLineID).Scan(&total).Error
	return total, Error(err)
}

func (r *ReceiptLineRepository) CompletedBeforeReceipt(ctx context.Context, inboundLineID, receiptID string) (string, error) {
	var total string
	err := r.db.WithContext(ctx).Table("receipt_line line").Select("COALESCE(sum(line.received_qty),0)::text").Joins("JOIN receipt ON receipt.receipt_id=line.receipt_id").Joins("JOIN document_status status ON status.status_id=receipt.status_id").Where("line.inbound_line_id=? AND receipt.receipt_id<>? AND status.code='COMPLETED'", inboundLineID, receiptID).Scan(&total).Error
	return total, Error(err)
}
