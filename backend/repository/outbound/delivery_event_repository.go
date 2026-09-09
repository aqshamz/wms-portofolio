package outbound

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type DeliveryEventRow struct {
	model.DeliveryEvent
	EventTypeCode, FailureReasonCode string
}
type DeliveryEventRepository struct{ db *gorm.DB }

func NewDeliveryEventRepository(db *gorm.DB) *DeliveryEventRepository {
	return &DeliveryEventRepository{db: db}
}
func (r *DeliveryEventRepository) Create(ctx context.Context, v *model.DeliveryEvent) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *DeliveryEventRepository) List(ctx context.Context, id string) ([]DeliveryEventRow, error) {
	var v []DeliveryEventRow
	err := r.db.WithContext(ctx).Table("delivery_event e").Select("e.*,t.code event_type_code,COALESCE(fr.code,'') failure_reason_code").Joins("JOIN delivery_event_type t ON t.delivery_event_type_id=e.delivery_event_type_id").Joins("LEFT JOIN delivery_failure_reason fr ON fr.delivery_failure_reason_id=e.delivery_failure_reason_id").Where("e.delivery_id=?", id).Order("e.event_at,e.delivery_event_id").Find(&v).Error
	return v, Error(err)
}
func (r *DeliveryEventRepository) Count(ctx context.Context, id string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.DeliveryEvent{}).Where("delivery_id=?", id).Count(&n).Error
	return n, Error(err)
}
