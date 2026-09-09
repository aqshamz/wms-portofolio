package outbound

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
	model "wms-api/models/outbound"
)

type CheckExceptionRow struct {
	model.OutboundCheckException
	CheckID, StagingID, StagingLineID, StagingBalanceID                         string
	ResultCode, StatusCode, OutboundLineID, ItemID, UOMID, OwnerID, WarehouseID string
	ResolvedQty                                                                 string
}
type OutboundCheckExceptionRepository struct{ db *gorm.DB }

func NewOutboundCheckExceptionRepository(db *gorm.DB) *OutboundCheckExceptionRepository {
	return &OutboundCheckExceptionRepository{db: db}
}
func exceptionQuery(db *gorm.DB) *gorm.DB {
	return db.Table("outbound_check_exception e").
		Select("e.*,l.outbound_check_id check_id,c.staging_id,l.staging_line_id,sl.staging_balance_id,r.code result_code,s.code status_code,p.outbound_line_id,ol.item_id,ol.uom_id,o.owner_id,o.warehouse_id,COALESCE((SELECT sum(x.resolved_qty) FROM outbound_check_resolution x WHERE x.outbound_check_exception_id=e.outbound_check_exception_id),0) resolved_qty").
		Joins("JOIN outbound_check_line l ON l.outbound_check_line_id=e.outbound_check_line_id").
		Joins("JOIN outbound_check c ON c.outbound_check_id=l.outbound_check_id").
		Joins("JOIN outbound_staging_line sl ON sl.staging_line_id=l.staging_line_id").
		Joins("JOIN pick_execution pe ON pe.pick_execution_id=sl.pick_execution_id").
		Joins("JOIN pick_task p ON p.pick_task_id=pe.pick_task_id").
		Joins("JOIN outbound_order_line ol ON ol.outbound_line_id=p.outbound_line_id").
		Joins("JOIN outbound_order o ON o.outbound_id=ol.outbound_id").
		Joins("JOIN outbound_check_result r ON r.outbound_check_result_id=l.outbound_check_result_id").
		Joins("JOIN outbound_check_exception_status s ON s.outbound_check_exception_status_id=e.status_id")
}
func (r *OutboundCheckExceptionRepository) Create(ctx context.Context, v *model.OutboundCheckException) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *OutboundCheckExceptionRepository) Get(ctx context.Context, id string) (CheckExceptionRow, error) {
	var v CheckExceptionRow
	err := exceptionQuery(r.db.WithContext(ctx)).Where("e.outbound_check_exception_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *OutboundCheckExceptionRepository) Lock(ctx context.Context, id string) (model.OutboundCheckException, error) {
	var v model.OutboundCheckException
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("outbound_check_exception_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *OutboundCheckExceptionRepository) ListByCheck(ctx context.Context, id string) ([]CheckExceptionRow, error) {
	var v []CheckExceptionRow
	err := exceptionQuery(r.db.WithContext(ctx)).Where("l.outbound_check_id=?", id).Order("e.opened_at,e.outbound_check_exception_id").Find(&v).Error
	return v, Error(err)
}
func (r *OutboundCheckExceptionRepository) Resolve(ctx context.Context, id, statusID, actor string) error {
	res := r.db.WithContext(ctx).Model(&model.OutboundCheckException{}).Where("outbound_check_exception_id=?", id).Updates(map[string]interface{}{"status_id": statusID, "resolved_at": time.Now(), "resolved_by": actor})
	if res.Error != nil {
		return Error(res.Error)
	}
	if res.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}

func (r *OutboundCheckExceptionRepository) SetStatus(ctx context.Context, id, statusID string) error {
	result := r.db.WithContext(ctx).Model(&model.OutboundCheckException{}).Where("outbound_check_exception_id=?", id).Update("status_id", statusID)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
