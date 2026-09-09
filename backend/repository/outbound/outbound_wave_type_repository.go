package outbound

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/outbound"
)

type OutboundWaveTypeRepository struct{ db *gorm.DB }

func NewOutboundWaveTypeRepository(db *gorm.DB) *OutboundWaveTypeRepository {
	return &OutboundWaveTypeRepository{db: db}
}
func (r *OutboundWaveTypeRepository) ByCode(ctx context.Context, code string) (model.OutboundWaveType, error) {
	var v model.OutboundWaveType
	err := r.db.WithContext(ctx).Where("code=? AND is_active", code).Take(&v).Error
	return v, Error(err)
}
func (r *OutboundWaveTypeRepository) Seed(ctx context.Context, v *model.OutboundWaveType) error {
	return Error(r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "description", "is_active"})}).Create(v).Error)
}
