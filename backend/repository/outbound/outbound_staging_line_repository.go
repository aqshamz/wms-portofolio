package outbound

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type StagingLineRow struct {
	model.OutboundStagingLine
	PickTaskID, ItemCode, UOMCode string
}
type OutboundStagingLineRepository struct{ db *gorm.DB }

func NewOutboundStagingLineRepository(db *gorm.DB) *OutboundStagingLineRepository {
	return &OutboundStagingLineRepository{db: db}
}
func (r *OutboundStagingLineRepository) CreateBatch(ctx context.Context, v []model.OutboundStagingLine) error {
	return Error(r.db.WithContext(ctx).Create(&v).Error)
}
func (r *OutboundStagingLineRepository) Create(ctx context.Context, v *model.OutboundStagingLine) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *OutboundStagingLineRepository) List(ctx context.Context, id string) ([]StagingLineRow, error) {
	var v []StagingLineRow
	err := r.db.WithContext(ctx).Table("outbound_staging_line s").Select("s.*,p.pick_task_id,i.code item_code,u.code uom_code").Joins("JOIN pick_execution e ON e.pick_execution_id=s.pick_execution_id").Joins("JOIN pick_task p ON p.pick_task_id=e.pick_task_id").Joins("JOIN outbound_order_line l ON l.outbound_line_id=p.outbound_line_id").Joins("JOIN item i ON i.item_id=l.item_id").Joins("JOIN uom u ON u.uom_id=s.uom_id").Where("s.staging_id=?", id).Order("s.staging_line_id").Find(&v).Error
	return v, Error(err)
}
func (r *OutboundStagingLineRepository) AddRemoved(ctx context.Context, id, qty string) error {
	result := r.db.WithContext(ctx).Model(&model.OutboundStagingLine{}).Where("staging_line_id=? AND staged_qty-removed_qty>=?", id, qty).Update("removed_qty", gorm.Expr("removed_qty+?", qty))
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConstraint
	}
	return nil
}
