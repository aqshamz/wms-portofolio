package outbound

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type ValidationRuleRow struct {
	model.OutboundValidationRule
	SeverityCode     string
	BlocksProcessing bool
}
type OutboundValidationRuleRepository struct{ db *gorm.DB }

func NewOutboundValidationRuleRepository(db *gorm.DB) *OutboundValidationRuleRepository {
	return &OutboundValidationRuleRepository{db: db}
}
func (r *OutboundValidationRuleRepository) ListActive(ctx context.Context) ([]ValidationRuleRow, error) {
	var v []ValidationRuleRow
	err := r.db.WithContext(ctx).Table("outbound_validation_rule r").Select("r.*,s.code severity_code,s.blocks_processing").Joins("JOIN validation_severity s ON s.validation_severity_id=r.validation_severity_id").Where("r.is_active").Order("r.display_order,r.code").Find(&v).Error
	return v, Error(err)
}
