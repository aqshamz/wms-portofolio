package inbound

import (
	"context"

	model "wms-api/models/inbound"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PurchaseOrderRow struct {
	model.PurchaseOrder
	StatusCode, OwnerCode, VendorCode, VendorName, WarehouseCode string
}

type PurchaseOrderRepository struct{ db *gorm.DB }

func NewPurchaseOrderRepository(db *gorm.DB) *PurchaseOrderRepository {
	return &PurchaseOrderRepository{db: db}
}

func (r *PurchaseOrderRepository) Create(ctx context.Context, value *model.PurchaseOrder) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}

func (r *PurchaseOrderRepository) Lock(ctx context.Context, id string) (model.PurchaseOrder, error) {
	var value model.PurchaseOrder
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("purchase_order_id = ?", id).Take(&value).Error
	return value, Error(err)
}

func purchaseOrderQuery(db *gorm.DB) *gorm.DB {
	return db.Table("purchase_order po").Select("po.*, status.code status_code, owner.code owner_code, vendor.code vendor_code, vendor.name vendor_name, warehouse.code warehouse_code").
		Joins("JOIN document_status status ON status.status_id=po.status_id").
		Joins("JOIN organization owner ON owner.organization_id=po.owner_id").
		Joins("JOIN business_partner vendor ON vendor.partner_id=po.vendor_id").
		Joins("JOIN warehouse ON warehouse.warehouse_id=po.warehouse_id")
}

func (r *PurchaseOrderRepository) Get(ctx context.Context, id string) (PurchaseOrderRow, error) {
	var value PurchaseOrderRow
	err := purchaseOrderQuery(r.db.WithContext(ctx)).Where("po.purchase_order_id=?", id).Take(&value).Error
	return value, Error(err)
}

func (r *PurchaseOrderRepository) List(ctx context.Context, filter ListFilter) ([]PurchaseOrderRow, int64, error) {
	query := purchaseOrderQuery(r.db.WithContext(ctx)).Where("po.owner_id=?", filter.OwnerID)
	if filter.WarehouseID != "" {
		query = query.Where("po.warehouse_id=?", filter.WarehouseID)
	}
	if filter.StatusCode != "" {
		query = query.Where("status.code=?", filter.StatusCode)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("po.purchase_order_id ILIKE ? OR po.purchase_order_no ILIKE ? OR vendor.name ILIKE ?", like, like, like)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, Error(err)
	}
	rows := make([]PurchaseOrderRow, 0)
	err := query.Order("po.business_date DESC,po.purchase_order_id DESC").Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize).Find(&rows).Error
	return rows, total, Error(err)
}

func (r *PurchaseOrderRepository) SetStatus(ctx context.Context, id, statusID, actor string, expectedVersion int64) error {
	result := r.db.WithContext(ctx).Model(&model.PurchaseOrder{}).Where("purchase_order_id=? AND version_no=?", id, expectedVersion).
		Updates(map[string]interface{}{"status_id": statusID, "updated_by": actor, "updated_at": gorm.Expr("clock_timestamp()"), "version_no": gorm.Expr("version_no+1")})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}

func (r *PurchaseOrderRepository) CountNonFinalInboundOrders(ctx context.Context, id string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("inbound_order inbound").
		Joins("JOIN document_status status ON status.status_id=inbound.status_id").
		Where(`EXISTS (SELECT 1 FROM inbound_order_line line JOIN purchase_order_line po_line ON po_line.purchase_order_line_id=line.purchase_order_line_id WHERE line.inbound_id=inbound.inbound_id AND po_line.purchase_order_id=?)`, id).
		Where("NOT status.is_final").Count(&count).Error
	return count, Error(err)
}
