package outbound

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type ValidationRunRow struct {
	model.OutboundValidationRun
	StatusCode string
}
type OutboundValidationRunRepository struct{ db *gorm.DB }

func NewOutboundValidationRunRepository(db *gorm.DB) *OutboundValidationRunRepository {
	return &OutboundValidationRunRepository{db: db}
}
func (r *OutboundValidationRunRepository) Create(ctx context.Context, v *model.OutboundValidationRun) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *OutboundValidationRunRepository) Complete(ctx context.Context, id, statusID, actor string) error {
	return Error(r.db.WithContext(ctx).Model(&model.OutboundValidationRun{}).Where("validation_run_id=?", id).Updates(map[string]interface{}{"status_id": statusID, "validated_at": gorm.Expr("clock_timestamp()"), "validated_by": actor}).Error)
}
func (r *OutboundValidationRunRepository) Get(ctx context.Context, id string) (ValidationRunRow, error) {
	var v ValidationRunRow
	err := r.db.WithContext(ctx).Table("outbound_validation_run r").Select("r.*,s.code status_code").Joins("JOIN document_status s ON s.status_id=r.status_id").Where("r.validation_run_id=?", id).Take(&v).Error
	return v, Error(err)
}
