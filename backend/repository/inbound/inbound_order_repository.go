package inbound

import (
	"context"

	model "wms-api/models/inbound"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InboundOrderRow struct {
	model.InboundOrder
	StatusCode, OwnerCode, VendorCode, VendorName, WarehouseCode string
}

type InboundOrderRepository struct{ db *gorm.DB }

func NewInboundOrderRepository(db *gorm.DB) *InboundOrderRepository {
	return &InboundOrderRepository{db: db}
}
func (r *InboundOrderRepository) Create(ctx context.Context, value *model.InboundOrder) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}
func (r *InboundOrderRepository) Lock(ctx context.Context, id string) (model.InboundOrder, error) {
	var value model.InboundOrder
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("inbound_id=?", id).Take(&value).Error
	return value, Error(err)
}
func inboundOrderQuery(db *gorm.DB) *gorm.DB {
	return db.Table("inbound_order inbound").Select("inbound.*,status.code status_code,owner.code owner_code,vendor.code vendor_code,vendor.name vendor_name,warehouse.code warehouse_code").
		Joins("JOIN document_status status ON status.status_id=inbound.status_id").Joins("JOIN organization owner ON owner.organization_id=inbound.owner_id").
		Joins("JOIN business_partner vendor ON vendor.partner_id=inbound.vendor_id").Joins("JOIN warehouse ON warehouse.warehouse_id=inbound.warehouse_id")
}
func (r *InboundOrderRepository) Get(ctx context.Context, id string) (InboundOrderRow, error) {
	var value InboundOrderRow
	err := inboundOrderQuery(r.db.WithContext(ctx)).Where("inbound.inbound_id=?", id).Take(&value).Error
	return value, Error(err)
}
func (r *InboundOrderRepository) List(ctx context.Context, filter ListFilter) ([]InboundOrderRow, int64, error) {
	query := inboundOrderQuery(r.db.WithContext(ctx)).Where("inbound.owner_id=?", filter.OwnerID)
	if filter.WarehouseID != "" {
		query = query.Where("inbound.warehouse_id=?", filter.WarehouseID)
	}
	if filter.StatusCode != "" {
		query = query.Where("status.code=?", filter.StatusCode)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("inbound.inbound_id ILIKE ? OR inbound.external_reference ILIKE ? OR inbound.supplier_reference ILIKE ? OR vendor.name ILIKE ?", like, like, like, like)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, Error(err)
	}
	rows := make([]InboundOrderRow, 0)
	err := query.Order("inbound.business_date DESC,inbound.inbound_id DESC").Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize).Find(&rows).Error
	return rows, total, Error(err)
}
func (r *InboundOrderRepository) SetStatus(ctx context.Context, id, statusID, actor string, expectedVersion int64) error {
	result := r.db.WithContext(ctx).Model(&model.InboundOrder{}).Where("inbound_id=? AND version_no=?", id, expectedVersion).
		Updates(map[string]interface{}{"status_id": statusID, "updated_by": actor, "updated_at": gorm.Expr("clock_timestamp()"), "version_no": gorm.Expr("version_no+1")})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}

func (r *InboundOrderRepository) CountReceiptsWithStatus(ctx context.Context, id string, statusCodes ...string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Table("receipt").Joins("JOIN document_status status ON status.status_id=receipt.status_id").Where("receipt.inbound_id=?", id)
	if len(statusCodes) != 0 {
		query = query.Where("status.code IN ?", statusCodes)
	}
	err := query.Count(&count).Error
	return count, Error(err)
}
