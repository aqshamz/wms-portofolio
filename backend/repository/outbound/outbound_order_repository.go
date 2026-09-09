package outbound

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/outbound"
)

type OutboundOrderRow struct {
	model.OutboundOrder
	StatusCode, OwnerCode, CustomerCode, CustomerName, ShipToCode, WarehouseCode string
}
type OutboundOrderRepository struct{ db *gorm.DB }

func NewOutboundOrderRepository(db *gorm.DB) *OutboundOrderRepository {
	return &OutboundOrderRepository{db: db}
}
func (r *OutboundOrderRepository) Create(ctx context.Context, value *model.OutboundOrder) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}
func (r *OutboundOrderRepository) UpdateDraft(ctx context.Context, id, actor string, version int64, values map[string]interface{}) error {
	values["updated_by"] = actor
	values["updated_at"] = gorm.Expr("clock_timestamp()")
	values["version_no"] = gorm.Expr("version_no+1")
	result := r.db.WithContext(ctx).Model(&model.OutboundOrder{}).
		Where("outbound_id=? AND version_no=? AND status_id IN (SELECT status_id FROM document_status WHERE document_type_id=outbound_order.document_type_id AND code='DRAFT')", id, version).
		Updates(values)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
func (r *OutboundOrderRepository) Lock(ctx context.Context, id string) (model.OutboundOrder, error) {
	var v model.OutboundOrder
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("outbound_id=?", id).Take(&v).Error
	return v, Error(err)
}
func outboundOrderQuery(db *gorm.DB) *gorm.DB {
	return db.Table("outbound_order o").Select("o.*,s.code status_code,owner.code owner_code,c.code customer_code,c.name customer_name,COALESCE(ship.code,'') ship_to_code,w.code warehouse_code").Joins("JOIN document_status s ON s.status_id=o.status_id").Joins("JOIN organization owner ON owner.organization_id=o.owner_id").Joins("JOIN business_partner c ON c.partner_id=o.customer_id").Joins("LEFT JOIN business_partner ship ON ship.partner_id=o.ship_to_partner_id").Joins("JOIN warehouse w ON w.warehouse_id=o.warehouse_id")
}
func (r *OutboundOrderRepository) Get(ctx context.Context, id string) (OutboundOrderRow, error) {
	var v OutboundOrderRow
	err := outboundOrderQuery(r.db.WithContext(ctx)).Where("o.outbound_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *OutboundOrderRepository) List(ctx context.Context, f ListFilter) ([]OutboundOrderRow, int64, error) {
	q := outboundOrderQuery(r.db.WithContext(ctx)).Where("o.owner_id=?", f.OwnerID)
	if f.WarehouseID != "" {
		q = q.Where("o.warehouse_id=?", f.WarehouseID)
	}
	if f.StatusCode != "" {
		q = q.Where("s.code=?", f.StatusCode)
	}
	if f.Search != "" {
		x := "%" + f.Search + "%"
		q = q.Where("o.outbound_id ILIKE ? OR o.client_delivery_order_no ILIKE ? OR c.name ILIKE ?", x, x, x)
	}
	var n int64
	if err := q.Session(&gorm.Session{}).Count(&n).Error; err != nil {
		return nil, 0, Error(err)
	}
	rows := make([]OutboundOrderRow, 0)
	err := q.Order("o.business_date DESC,o.outbound_id DESC").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&rows).Error
	return rows, n, Error(err)
}
func (r *OutboundOrderRepository) SetStatus(ctx context.Context, id, statusID, actor string, version int64, extra map[string]interface{}) error {
	values := map[string]interface{}{"status_id": statusID, "updated_by": actor, "updated_at": gorm.Expr("clock_timestamp()"), "version_no": gorm.Expr("version_no+1")}
	for k, v := range extra {
		values[k] = v
	}
	res := r.db.WithContext(ctx).Model(&model.OutboundOrder{}).Where("outbound_id=? AND version_no=?", id, version).Updates(values)
	if res.Error != nil {
		return Error(res.Error)
	}
	if res.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
func (r *OutboundOrderRepository) SetStatusLocked(ctx context.Context, id, statusID, actor string, extra map[string]interface{}) error {
	values := map[string]interface{}{"status_id": statusID, "updated_by": actor, "updated_at": gorm.Expr("clock_timestamp()"), "version_no": gorm.Expr("version_no+1")}
	for k, v := range extra {
		values[k] = v
	}
	return Error(r.db.WithContext(ctx).Model(&model.OutboundOrder{}).Where("outbound_id=?", id).Updates(values).Error)
}
func (r *OutboundOrderRepository) TouchAllocated(ctx context.Context, id, statusID, actor string, complete bool) error {
	v := map[string]interface{}{"status_id": statusID, "updated_by": actor, "updated_at": gorm.Expr("clock_timestamp()"), "version_no": gorm.Expr("version_no+1")}
	if complete {
		v["allocated_at"] = time.Now()
	}
	return Error(r.db.WithContext(ctx).Model(&model.OutboundOrder{}).Where("outbound_id=?", id).Updates(v).Error)
}
