package outbound

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/outbound"
)

type OutboundWaveRow struct {
	model.OutboundWave
	StatusCode, WaveTypeCode, OwnerCode, WarehouseCode string
	OrderCount, PickTaskCount                          int64
}
type OutboundWaveRepository struct{ db *gorm.DB }

func NewOutboundWaveRepository(db *gorm.DB) *OutboundWaveRepository {
	return &OutboundWaveRepository{db: db}
}
func (r *OutboundWaveRepository) Create(ctx context.Context, v *model.OutboundWave) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *OutboundWaveRepository) Lock(ctx context.Context, id string) (model.OutboundWave, error) {
	var v model.OutboundWave
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("wave_id=?", id).Take(&v).Error
	return v, Error(err)
}
func waveQuery(db *gorm.DB) *gorm.DB {
	return db.Table("outbound_wave w").Select("w.*,s.code status_code,t.code wave_type_code,o.code owner_code,wh.code warehouse_code,(SELECT count(*) FROM outbound_wave_order wo WHERE wo.wave_id=w.wave_id) order_count,(SELECT count(*) FROM pick_task p WHERE p.wave_id=w.wave_id) pick_task_count").Joins("JOIN document_status s ON s.status_id=w.status_id").Joins("JOIN outbound_wave_type t ON t.outbound_wave_type_id=w.outbound_wave_type_id").Joins("JOIN organization o ON o.organization_id=w.owner_id").Joins("JOIN warehouse wh ON wh.warehouse_id=w.warehouse_id")
}
func (r *OutboundWaveRepository) Get(ctx context.Context, id string) (OutboundWaveRow, error) {
	var v OutboundWaveRow
	err := waveQuery(r.db.WithContext(ctx)).Where("w.wave_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *OutboundWaveRepository) List(ctx context.Context, f ListFilter) ([]OutboundWaveRow, int64, error) {
	q := waveQuery(r.db.WithContext(ctx)).Where("w.owner_id=?", f.OwnerID)
	if f.WarehouseID != "" {
		q = q.Where("w.warehouse_id=?", f.WarehouseID)
	}
	if f.StatusCode != "" {
		q = q.Where("s.code=?", f.StatusCode)
	}
	if f.Search != "" {
		q = q.Where("w.wave_id ILIKE ?", "%"+f.Search+"%")
	}
	var n int64
	if err := q.Session(&gorm.Session{}).Count(&n).Error; err != nil {
		return nil, 0, Error(err)
	}
	var v []OutboundWaveRow
	err := q.Order("w.business_date DESC,w.wave_id DESC").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&v).Error
	return v, n, Error(err)
}
func (r *OutboundWaveRepository) SetStatus(ctx context.Context, id, statusID, actor string, version int64, extra map[string]interface{}) error {
	v := map[string]interface{}{"status_id": statusID, "updated_by": actor, "updated_at": gorm.Expr("clock_timestamp()"), "version_no": gorm.Expr("version_no+1")}
	for k, x := range extra {
		v[k] = x
	}
	res := r.db.WithContext(ctx).Model(&model.OutboundWave{}).Where("wave_id=? AND version_no=?", id, version).Updates(v)
	if res.Error != nil {
		return Error(res.Error)
	}
	if res.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
func (r *OutboundWaveRepository) SetStatusLocked(ctx context.Context, id, statusID, actor string, extra map[string]interface{}) error {
	v := map[string]interface{}{"status_id": statusID, "updated_by": actor, "updated_at": gorm.Expr("clock_timestamp()"), "version_no": gorm.Expr("version_no+1")}
	for k, x := range extra {
		v[k] = x
	}
	return Error(r.db.WithContext(ctx).Model(&model.OutboundWave{}).Where("wave_id=?", id).Updates(v).Error)
}
