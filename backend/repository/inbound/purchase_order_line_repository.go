package inbound

import (
	"context"

	model "wms-api/models/inbound"

	"gorm.io/gorm"
)

type PurchaseOrderLineRow struct {
	model.PurchaseOrderLine
	ItemCode, ItemName, UOMCode       string
	ScheduledQty, CompletedReceiptQty string
}

type PurchaseOrderLineRepository struct{ db *gorm.DB }

func NewPurchaseOrderLineRepository(db *gorm.DB) *PurchaseOrderLineRepository {
	return &PurchaseOrderLineRepository{db: db}
}

func (r *PurchaseOrderLineRepository) Create(ctx context.Context, value *model.PurchaseOrderLine) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}

func (r *PurchaseOrderLineRepository) Get(ctx context.Context, id string) (model.PurchaseOrderLine, error) {
	var value model.PurchaseOrderLine
	err := r.db.WithContext(ctx).Where("purchase_order_line_id=?", id).Take(&value).Error
	return value, Error(err)
}

func purchaseOrderLineQuery(db *gorm.DB) *gorm.DB {
	return db.Table("purchase_order_line line").Select(`line.*,item.code item_code,item.name item_name,uom.code uom_code,
 COALESCE((SELECT sum(CASE WHEN inbound_status.code='CANCELLED' THEN COALESCE((SELECT sum(cancelled_receipt_line.received_qty) FROM receipt_line cancelled_receipt_line JOIN receipt cancelled_receipt ON cancelled_receipt.receipt_id=cancelled_receipt_line.receipt_id JOIN document_status cancelled_receipt_status ON cancelled_receipt_status.status_id=cancelled_receipt.status_id WHERE cancelled_receipt_line.inbound_line_id=inbound_line.inbound_line_id AND cancelled_receipt_status.code='COMPLETED'),0) ELSE inbound_line.expected_qty END) FROM inbound_order_line inbound_line JOIN inbound_order inbound ON inbound.inbound_id=inbound_line.inbound_id JOIN document_status inbound_status ON inbound_status.status_id=inbound.status_id WHERE inbound_line.purchase_order_line_id=line.purchase_order_line_id),0) scheduled_qty,
 COALESCE((SELECT sum(receipt_line.received_qty) FROM inbound_order_line inbound_line JOIN receipt_line ON receipt_line.inbound_line_id=inbound_line.inbound_line_id JOIN receipt ON receipt.receipt_id=receipt_line.receipt_id JOIN document_status receipt_status ON receipt_status.status_id=receipt.status_id WHERE inbound_line.purchase_order_line_id=line.purchase_order_line_id AND receipt_status.code='COMPLETED'),0) completed_receipt_qty`).
		Joins("JOIN item ON item.item_id=line.item_id").Joins("JOIN uom ON uom.uom_id=line.uom_id")
}

func (r *PurchaseOrderLineRepository) List(ctx context.Context, purchaseOrderID string) ([]PurchaseOrderLineRow, error) {
	rows := make([]PurchaseOrderLineRow, 0)
	err := purchaseOrderLineQuery(r.db.WithContext(ctx)).Where("line.purchase_order_id=?", purchaseOrderID).Order("line.line_no").Find(&rows).Error
	return rows, Error(err)
}

func (r *PurchaseOrderLineRepository) Count(ctx context.Context, purchaseOrderID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PurchaseOrderLine{}).Where("purchase_order_id=?", purchaseOrderID).Count(&count).Error
	return count, Error(err)
}
