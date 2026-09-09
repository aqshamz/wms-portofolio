package outbound

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type WaveOrderRow struct {
	model.OutboundWaveOrder
	ClientDeliveryOrderNo, StatusCode string
}
type OutboundWaveOrderRepository struct{ db *gorm.DB }

func NewOutboundWaveOrderRepository(db *gorm.DB) *OutboundWaveOrderRepository {
	return &OutboundWaveOrderRepository{db: db}
}
func (r *OutboundWaveOrderRepository) CreateBatch(ctx context.Context, v []model.OutboundWaveOrder) error {
	return Error(r.db.WithContext(ctx).Create(&v).Error)
}
func (r *OutboundWaveOrderRepository) List(ctx context.Context, id string) ([]WaveOrderRow, error) {
	var v []WaveOrderRow
	err := r.db.WithContext(ctx).Table("outbound_wave_order wo").Select("wo.*,o.client_delivery_order_no,s.code status_code").Joins("JOIN outbound_order o ON o.outbound_id=wo.outbound_id").Joins("JOIN document_status s ON s.status_id=o.status_id").Where("wo.wave_id=?", id).Order("wo.outbound_id").Find(&v).Error
	return v, Error(err)
}
func (r *OutboundWaveOrderRepository) HasActiveWave(ctx context.Context, id string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("outbound_wave_order wo").Joins("JOIN outbound_wave w ON w.wave_id=wo.wave_id").Joins("JOIN document_status s ON s.status_id=w.status_id").Where("wo.outbound_id=? AND NOT s.is_final", id).Count(&n).Error
	return n > 0, Error(err)
}
