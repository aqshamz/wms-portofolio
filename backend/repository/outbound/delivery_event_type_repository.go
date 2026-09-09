package outbound

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/outbound"
)

type DeliveryEventTypeRepository struct{ db *gorm.DB }

func NewDeliveryEventTypeRepository(db *gorm.DB) *DeliveryEventTypeRepository {
	return &DeliveryEventTypeRepository{db: db}
}
func (r *DeliveryEventTypeRepository) ByCode(ctx context.Context, code string) (model.DeliveryEventType, error) {
	var v model.DeliveryEventType
	err := r.db.WithContext(ctx).Where("code=? AND is_active", code).Take(&v).Error
	return v, Error(err)
}
func (r *DeliveryEventTypeRepository) Seed(ctx context.Context, v *model.DeliveryEventType) error {
	return Error(r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "description", "marks_delivered", "marks_failed", "marks_returned", "is_active"})}).Create(v).Error)
}
