package outbound

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/outbound"
)

type DeliveryRow struct {
	model.Delivery
	StatusCode, ClientDeliveryOrderNo, OwnerID, WarehouseID string
}
type DeliveryRepository struct{ db *gorm.DB }

func NewDeliveryRepository(db *gorm.DB) *DeliveryRepository { return &DeliveryRepository{db: db} }
func deliveryQuery(db *gorm.DB) *gorm.DB {
	return db.Table("delivery d").Select("d.*,s.code status_code,o.client_delivery_order_no,o.owner_id,o.warehouse_id").Joins("JOIN document_status s ON s.status_id=d.status_id").Joins("JOIN outbound_order o ON o.outbound_id=d.outbound_id")
}
func (r *DeliveryRepository) Create(ctx context.Context, v *model.Delivery) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *DeliveryRepository) Get(ctx context.Context, id string) (DeliveryRow, error) {
	var v DeliveryRow
	err := deliveryQuery(r.db.WithContext(ctx)).Where("d.delivery_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *DeliveryRepository) Lock(ctx context.Context, id string) (model.Delivery, error) {
	var v model.Delivery
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("delivery_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *DeliveryRepository) List(ctx context.Context, f ListFilter) ([]DeliveryRow, int64, error) {
	q := deliveryQuery(r.db.WithContext(ctx)).Where("o.owner_id=?", f.OwnerID)
	if f.WarehouseID != "" {
		q = q.Where("o.warehouse_id=?", f.WarehouseID)
	}
	if f.StatusCode != "" {
		q = q.Where("s.code=?", f.StatusCode)
	}
	if f.Search != "" {
		x := "%" + f.Search + "%"
		q = q.Where("d.delivery_id ILIKE ? OR o.client_delivery_order_no ILIKE ?", x, x)
	}
	var n int64
	if err := q.Session(&gorm.Session{}).Count(&n).Error; err != nil {
		return nil, 0, Error(err)
	}
	var v []DeliveryRow
	err := q.Order("d.business_date DESC,d.delivery_id DESC").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&v).Error
	return v, n, Error(err)
}
func (r *DeliveryRepository) Update(ctx context.Context, id string, values map[string]interface{}) error {
	res := r.db.WithContext(ctx).Model(&model.Delivery{}).Where("delivery_id=?", id).Updates(values)
	if res.Error != nil {
		return Error(res.Error)
	}
	if res.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
