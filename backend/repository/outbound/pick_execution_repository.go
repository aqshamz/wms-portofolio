package outbound

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type PickExecutionRepository struct{ db *gorm.DB }

func NewPickExecutionRepository(db *gorm.DB) *PickExecutionRepository {
	return &PickExecutionRepository{db: db}
}
func (r *PickExecutionRepository) Create(ctx context.Context, v *model.PickExecution) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *PickExecutionRepository) GetByMovement(ctx context.Context, id string) (model.PickExecution, error) {
	var v model.PickExecution
	err := r.db.WithContext(ctx).Where("movement_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *PickExecutionRepository) ListForOrderWave(ctx context.Context, outboundID, waveID string) ([]model.PickExecution, error) {
	var v []model.PickExecution
	err := r.db.WithContext(ctx).Table("pick_execution e").Joins("JOIN pick_task p ON p.pick_task_id=e.pick_task_id").Joins("JOIN outbound_order_line l ON l.outbound_line_id=p.outbound_line_id").Where("l.outbound_id=? AND p.wave_id=?", outboundID, waveID).Order("e.picked_at,e.pick_execution_id").Find(&v).Error
	return v, Error(err)
}
