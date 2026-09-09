package outbound

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
	model "wms-api/models/outbound"
)

type PackingRow struct {
	model.Packing
	StatusCode, ClientDeliveryOrderNo, OwnerID, PackingLocationCode string
	LineCount                                                       int64
}
type PackingRepository struct{ db *gorm.DB }

func NewPackingRepository(db *gorm.DB) *PackingRepository { return &PackingRepository{db: db} }
func packingQuery(db *gorm.DB) *gorm.DB {
	return db.Table("packing p").Select("p.*,s.code status_code,o.client_delivery_order_no,o.owner_id,COALESCE(l.code,'') packing_location_code,(SELECT count(*) FROM packing_line pl WHERE pl.packing_id=p.packing_id) line_count").Joins("JOIN document_status s ON s.status_id=p.status_id").Joins("JOIN outbound_order o ON o.outbound_id=p.outbound_id").Joins("LEFT JOIN warehouse_location l ON l.location_id=p.packing_location_id")
}
func (r *PackingRepository) Create(ctx context.Context, v *model.Packing) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *PackingRepository) HasActiveByOutbound(ctx context.Context, outboundID string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("packing p").Joins("JOIN document_status s ON s.status_id=p.status_id").
		Where("p.outbound_id=? AND NOT s.is_cancelled", outboundID).Count(&n).Error
	return n > 0, Error(err)
}
func (r *PackingRepository) Get(ctx context.Context, id string) (PackingRow, error) {
	var v PackingRow
	err := packingQuery(r.db.WithContext(ctx)).Where("p.packing_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *PackingRepository) Lock(ctx context.Context, id string) (model.Packing, error) {
	var v model.Packing
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("packing_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *PackingRepository) List(ctx context.Context, f ListFilter) ([]PackingRow, int64, error) {
	q := packingQuery(r.db.WithContext(ctx)).Where("o.owner_id=?", f.OwnerID)
	if f.WarehouseID != "" {
		q = q.Where("p.warehouse_id=?", f.WarehouseID)
	}
	if f.StatusCode != "" {
		q = q.Where("s.code=?", f.StatusCode)
	}
	if f.Search != "" {
		x := "%" + f.Search + "%"
		q = q.Where("p.packing_id ILIKE ? OR o.client_delivery_order_no ILIKE ?", x, x)
	}
	var n int64
	if err := q.Session(&gorm.Session{}).Count(&n).Error; err != nil {
		return nil, 0, Error(err)
	}
	var v []PackingRow
	err := q.Order("p.created_at DESC,p.packing_id").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&v).Error
	return v, n, Error(err)
}
func (r *PackingRepository) Complete(ctx context.Context, id, statusID, actor string) error {
	res := r.db.WithContext(ctx).Model(&model.Packing{}).Where("packing_id=?", id).Updates(map[string]interface{}{"status_id": statusID, "packed_at": time.Now(), "packed_by": actor})
	if res.Error != nil {
		return Error(res.Error)
	}
	if res.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
func (r *PackingRepository) Cancel(ctx context.Context, id, statusID string) error {
	result := r.db.WithContext(ctx).Model(&model.Packing{}).Where("packing_id=?", id).Update("status_id", statusID)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
