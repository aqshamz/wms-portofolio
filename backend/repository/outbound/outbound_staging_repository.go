package outbound

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type StagingRow struct {
	model.OutboundStaging
	StatusCode, ClientDeliveryOrderNo, WarehouseCode, StagingLocationCode string
	LineCount                                                             int64
}
type OutboundStagingRepository struct{ db *gorm.DB }

func NewOutboundStagingRepository(db *gorm.DB) *OutboundStagingRepository {
	return &OutboundStagingRepository{db: db}
}
func (r *OutboundStagingRepository) Create(ctx context.Context, v *model.OutboundStaging) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func stagingQuery(db *gorm.DB) *gorm.DB {
	return db.Table("outbound_staging x").Select("x.*,s.code status_code,o.client_delivery_order_no,w.code warehouse_code,l.code staging_location_code,(SELECT count(*) FROM outbound_staging_line sl WHERE sl.staging_id=x.staging_id) line_count").Joins("JOIN document_status s ON s.status_id=x.status_id").Joins("JOIN outbound_order o ON o.outbound_id=x.outbound_id").Joins("JOIN warehouse w ON w.warehouse_id=x.warehouse_id").Joins("JOIN warehouse_location l ON l.location_id=x.staging_location_id")
}
func (r *OutboundStagingRepository) Get(ctx context.Context, id string) (StagingRow, error) {
	var v StagingRow
	err := stagingQuery(r.db.WithContext(ctx)).Where("x.staging_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *OutboundStagingRepository) List(ctx context.Context, f ListFilter) ([]StagingRow, int64, error) {
	q := stagingQuery(r.db.WithContext(ctx)).Where("o.owner_id=?", f.OwnerID)
	if f.WarehouseID != "" {
		q = q.Where("x.warehouse_id=?", f.WarehouseID)
	}
	if f.StatusCode != "" {
		q = q.Where("s.code=?", f.StatusCode)
	}
	if f.Search != "" {
		x := "%" + f.Search + "%"
		q = q.Where("x.staging_id ILIKE ? OR o.client_delivery_order_no ILIKE ?", x, x)
	}
	var n int64
	if err := q.Session(&gorm.Session{}).Count(&n).Error; err != nil {
		return nil, 0, Error(err)
	}
	var v []StagingRow
	err := q.Order("x.created_at DESC,x.staging_id").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&v).Error
	return v, n, Error(err)
}
func (r *OutboundStagingRepository) Complete(ctx context.Context, id, statusID, actor string) error {
	res := r.db.WithContext(ctx).Model(&model.OutboundStaging{}).Where("staging_id=?", id).Updates(map[string]interface{}{"status_id": statusID, "staged_at": gorm.Expr("clock_timestamp()"), "staged_by": actor})
	if res.Error != nil {
		return Error(res.Error)
	}
	if res.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
func (r *OutboundStagingRepository) ExistsForOrderWave(ctx context.Context, outboundID, waveID string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.OutboundStaging{}).Where("outbound_id=? AND wave_id=?", outboundID, waveID).Count(&n).Error
	return n > 0, Error(err)
}
