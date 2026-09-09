package outbound

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/outbound"
)

type OutboundCheckExceptionStatusRepository struct{ db *gorm.DB }

func NewOutboundCheckExceptionStatusRepository(db *gorm.DB) *OutboundCheckExceptionStatusRepository {
	return &OutboundCheckExceptionStatusRepository{db: db}
}
func (r *OutboundCheckExceptionStatusRepository) ByCode(ctx context.Context, code string) (model.OutboundCheckExceptionStatus, error) {
	var v model.OutboundCheckExceptionStatus
	err := r.db.WithContext(ctx).Where("code=? AND is_active", code).Take(&v).Error
	return v, Error(err)
}
func (r *OutboundCheckExceptionStatusRepository) Seed(ctx context.Context, v *model.OutboundCheckExceptionStatus) error {
	return Error(r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "is_final", "is_active"})}).Create(v).Error)
}
