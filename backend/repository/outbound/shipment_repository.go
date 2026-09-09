package outbound

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
	model "wms-api/models/outbound"
)

type ShipmentRow struct {
	model.Shipment
	StatusCode, CarrierCode, CarrierServiceCode string
	OrderCount, LineCount                       int64
}
type ShipmentRepository struct{ db *gorm.DB }

func NewShipmentRepository(db *gorm.DB) *ShipmentRepository { return &ShipmentRepository{db: db} }
func shipmentQuery(db *gorm.DB) *gorm.DB {
	return db.Table("shipment s").Select("s.*,ds.code status_code,COALESCE(c.code,'') carrier_code,COALESCE(cs.code,'') carrier_service_code,(SELECT count(*) FROM shipment_order so WHERE so.shipment_id=s.shipment_id AND so.removed_at IS NULL) order_count,(SELECT count(*) FROM shipment_line sl WHERE sl.shipment_id=s.shipment_id) line_count").Joins("JOIN document_status ds ON ds.status_id=s.status_id").Joins("LEFT JOIN carrier_service cs ON cs.carrier_service_id=s.carrier_service_id").Joins("LEFT JOIN carrier c ON c.carrier_id=cs.carrier_id")
}
func (r *ShipmentRepository) Create(ctx context.Context, v *model.Shipment) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *ShipmentRepository) Get(ctx context.Context, id string) (ShipmentRow, error) {
	var v ShipmentRow
	err := shipmentQuery(r.db.WithContext(ctx)).Where("s.shipment_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *ShipmentRepository) Lock(ctx context.Context, id string) (model.Shipment, error) {
	var v model.Shipment
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("shipment_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *ShipmentRepository) List(ctx context.Context, f ListFilter) ([]ShipmentRow, int64, error) {
	q := shipmentQuery(r.db.WithContext(ctx)).Where("s.owner_id=?", f.OwnerID)
	if f.WarehouseID != "" {
		q = q.Where("s.warehouse_id=?", f.WarehouseID)
	}
	if f.StatusCode != "" {
		q = q.Where("ds.code=?", f.StatusCode)
	}
	if f.Search != "" {
		x := "%" + f.Search + "%"
		q = q.Where("s.shipment_id ILIKE ? OR s.tracking_number ILIKE ? OR s.vehicle_number ILIKE ?", x, x, x)
	}
	var n int64
	if err := q.Session(&gorm.Session{}).Count(&n).Error; err != nil {
		return nil, 0, Error(err)
	}
	var v []ShipmentRow
	err := q.Order("s.business_date DESC,s.shipment_id DESC").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&v).Error
	return v, n, Error(err)
}
func (r *ShipmentRepository) Complete(ctx context.Context, id, statusID, actor string) error {
	res := r.db.WithContext(ctx).Model(&model.Shipment{}).Where("shipment_id=?", id).Updates(map[string]interface{}{"status_id": statusID, "shipped_at": time.Now(), "dispatched_by": actor})
	if res.Error != nil {
		return Error(res.Error)
	}
	if res.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
func (r *ShipmentRepository) Cancel(ctx context.Context, id, statusID string) error {
	result := r.db.WithContext(ctx).Model(&model.Shipment{}).Where("shipment_id=?", id).Update("status_id", statusID)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
