package outbound

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/outbound"
)

type OutboundCheckResultRepository struct{ db *gorm.DB }

func NewOutboundCheckResultRepository(db *gorm.DB) *OutboundCheckResultRepository {
	return &OutboundCheckResultRepository{db: db}
}
func (r *OutboundCheckResultRepository) ByCode(ctx context.Context, code string) (model.OutboundCheckResult, error) {
	var v model.OutboundCheckResult
	err := r.db.WithContext(ctx).Where("code=? AND is_active", code).Take(&v).Error
	return v, Error(err)
}
func (r *OutboundCheckResultRepository) Seed(ctx context.Context, v *model.OutboundCheckResult) error {
	return Error(r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "description", "is_pass", "requires_note", "is_active"})}).Create(v).Error)
}
