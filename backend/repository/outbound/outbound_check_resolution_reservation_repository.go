package outbound

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type OutboundCheckResolutionReservationRepository struct{ db *gorm.DB }

type ReplacementWorkRow struct {
	ExceptionID, ResolutionID, StagingID, OutboundLineID, ResolvedQty string
}

func NewOutboundCheckResolutionReservationRepository(db *gorm.DB) *OutboundCheckResolutionReservationRepository {
	return &OutboundCheckResolutionReservationRepository{db: db}
}
func (r *OutboundCheckResolutionReservationRepository) Create(ctx context.Context, v *model.OutboundCheckResolutionReservation) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}

func (r *OutboundCheckResolutionReservationRepository) ReplacementByReservation(ctx context.Context, reservationID string) (ReplacementWorkRow, error) {
	var value ReplacementWorkRow
	err := r.db.WithContext(ctx).Table("outbound_check_resolution_reservation rr").
		Select("r.outbound_check_exception_id exception_id,r.outbound_check_resolution_id resolution_id,c.staging_id,p.outbound_line_id,r.resolved_qty::text resolved_qty").
		Joins("JOIN outbound_check_resolution r ON r.outbound_check_resolution_id=rr.outbound_check_resolution_id").
		Joins("JOIN outbound_check_resolution_type t ON t.outbound_check_resolution_type_id=r.outbound_check_resolution_type_id AND t.counts_as_replacement").
		Joins("JOIN outbound_check_exception e ON e.outbound_check_exception_id=r.outbound_check_exception_id").
		Joins("JOIN outbound_check_line l ON l.outbound_check_line_id=e.outbound_check_line_id").
		Joins("JOIN outbound_check c ON c.outbound_check_id=l.outbound_check_id").
		Joins("JOIN inventory_reservation p ON p.reservation_id=rr.reservation_id").
		Where("rr.reservation_id=?", reservationID).Take(&value).Error
	return value, Error(err)
}
func (r *OutboundCheckResolutionReservationRepository) Exists(ctx context.Context, id string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("inventory_reservation").Where("reservation_id=?", id).Count(&n).Error
	return n > 0, Error(err)
}
