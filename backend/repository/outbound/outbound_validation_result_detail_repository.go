package outbound

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type ValidationResultRow struct {
	model.OutboundValidationResultDetail
	RuleCode, RuleName, SeverityCode string
}
type OutboundValidationResultRepository struct{ db *gorm.DB }

func NewOutboundValidationResultRepository(db *gorm.DB) *OutboundValidationResultRepository {
	return &OutboundValidationResultRepository{db: db}
}
func (r *OutboundValidationResultRepository) CreateBatch(ctx context.Context, v []model.OutboundValidationResultDetail) error {
	return Error(r.db.WithContext(ctx).Create(&v).Error)
}
func (r *OutboundValidationResultRepository) List(ctx context.Context, id string) ([]ValidationResultRow, error) {
	var v []ValidationResultRow
	err := r.db.WithContext(ctx).Table("outbound_validation_result_detail d").Select("d.*,r.code rule_code,r.name rule_name,s.code severity_code").Joins("JOIN outbound_validation_rule r ON r.outbound_validation_rule_id=d.outbound_validation_rule_id").Joins("JOIN validation_severity s ON s.validation_severity_id=r.validation_severity_id").Where("d.validation_run_id=?", id).Order("r.display_order").Find(&v).Error
	return v, Error(err)
}
