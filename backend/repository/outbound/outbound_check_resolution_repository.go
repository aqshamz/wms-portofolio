package outbound

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type OutboundCheckResolutionRepository struct{ db *gorm.DB }

type CheckResolutionTotals struct {
	CorrectedQty   string
	AcceptedQty    string
	ReplacementQty string
}

func NewOutboundCheckResolutionRepository(db *gorm.DB) *OutboundCheckResolutionRepository {
	return &OutboundCheckResolutionRepository{db: db}
}
func (r *OutboundCheckResolutionRepository) Create(ctx context.Context, v *model.OutboundCheckResolution) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *OutboundCheckResolutionRepository) Total(ctx context.Context, id string) (string, error) {
	var v struct{ Total string }
	err := r.db.WithContext(ctx).Table("outbound_check_resolution").Select("COALESCE(sum(resolved_qty),0)::text total").Where("outbound_check_exception_id=?", id).Scan(&v).Error
	return v.Total, Error(err)
}

func (r *OutboundCheckResolutionRepository) Totals(ctx context.Context, id string) (CheckResolutionTotals, error) {
	var v CheckResolutionTotals
	err := r.db.WithContext(ctx).Table("outbound_check_resolution r").
		Select("COALESCE(sum(CASE WHEN t.counts_as_stock_correction THEN r.resolved_qty ELSE 0 END),0)::text corrected_qty,COALESCE(sum(CASE WHEN t.counts_as_short_acceptance THEN r.resolved_qty ELSE 0 END),0)::text accepted_qty,COALESCE(sum(CASE WHEN t.counts_as_replacement THEN r.resolved_qty ELSE 0 END),0)::text replacement_qty").
		Joins("JOIN outbound_check_resolution_type t ON t.outbound_check_resolution_type_id=r.outbound_check_resolution_type_id").
		Where("r.outbound_check_exception_id=?", id).Scan(&v).Error
	return v, Error(err)
}
