package outbound

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/outbound"
)

type OutboundCheckResolutionTypeRepository struct{ db *gorm.DB }

func NewOutboundCheckResolutionTypeRepository(db *gorm.DB) *OutboundCheckResolutionTypeRepository {
	return &OutboundCheckResolutionTypeRepository{db: db}
}
func (r *OutboundCheckResolutionTypeRepository) ByCode(ctx context.Context, code string) (model.OutboundCheckResolutionType, error) {
	var v model.OutboundCheckResolutionType
	err := r.db.WithContext(ctx).Where("code=? AND is_active", code).Take(&v).Error
	return v, Error(err)
}
func (r *OutboundCheckResolutionTypeRepository) Seed(ctx context.Context, v *model.OutboundCheckResolutionType) error {
	return Error(r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "description", "counts_as_stock_correction", "counts_as_replacement", "counts_as_short_acceptance", "requires_approval", "is_active"})}).Create(v).Error)
}
