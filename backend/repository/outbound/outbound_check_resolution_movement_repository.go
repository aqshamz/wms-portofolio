package outbound

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type OutboundCheckResolutionMovementRepository struct{ db *gorm.DB }

func NewOutboundCheckResolutionMovementRepository(db *gorm.DB) *OutboundCheckResolutionMovementRepository {
	return &OutboundCheckResolutionMovementRepository{db: db}
}
func (r *OutboundCheckResolutionMovementRepository) Create(ctx context.Context, v *model.OutboundCheckResolutionMovement) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *OutboundCheckResolutionMovementRepository) Exists(ctx context.Context, id string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("inventory_movement").Where("movement_id=?", id).Count(&n).Error
	return n > 0, Error(err)
}
