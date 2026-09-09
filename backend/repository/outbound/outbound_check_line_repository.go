package outbound

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
	model "wms-api/models/outbound"
)

type CheckLineRow struct {
	model.OutboundCheckLine
	ItemCode, LotNumber, ResultCode, PickTaskID, StagingBalanceID, OutboundLineID string
	BalanceVersion                                                                int64
}

type OutboundCheckLineRepository struct{ db *gorm.DB }

func NewOutboundCheckLineRepository(db *gorm.DB) *OutboundCheckLineRepository {
	return &OutboundCheckLineRepository{db: db}
}
func checkLineQuery(db *gorm.DB) *gorm.DB {
	return db.Table("outbound_check_line l").Select("l.*,i.code item_code,COALESCE(lot.lot_number,'') lot_number,COALESCE(r.code,'') result_code,p.pick_task_id,sl.staging_balance_id,p.outbound_line_id,b.version_no balance_version").Joins("JOIN outbound_staging_line sl ON sl.staging_line_id=l.staging_line_id").Joins("JOIN pick_execution e ON e.pick_execution_id=sl.pick_execution_id").Joins("JOIN pick_task p ON p.pick_task_id=e.pick_task_id").Joins("JOIN outbound_order_line ol ON ol.outbound_line_id=p.outbound_line_id").Joins("JOIN item i ON i.item_id=ol.item_id").Joins("JOIN inventory_balance b ON b.balance_id=sl.staging_balance_id").Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=b.lot_id").Joins("LEFT JOIN outbound_check_result r ON r.outbound_check_result_id=l.outbound_check_result_id")
}
func (r *OutboundCheckLineRepository) CreateBatch(ctx context.Context, v []model.OutboundCheckLine) error {
	return Error(r.db.WithContext(ctx).Create(&v).Error)
}
func (r *OutboundCheckLineRepository) List(ctx context.Context, id string) ([]CheckLineRow, error) {
	var v []CheckLineRow
	err := checkLineQuery(r.db.WithContext(ctx)).Where("l.outbound_check_id=?", id).Order("l.line_no").Find(&v).Error
	return v, Error(err)
}
func (r *OutboundCheckLineRepository) Get(ctx context.Context, id string) (CheckLineRow, error) {
	var v CheckLineRow
	err := checkLineQuery(r.db.WithContext(ctx)).Where("l.outbound_check_line_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *OutboundCheckLineRepository) Lock(ctx context.Context, id string) (model.OutboundCheckLine, error) {
	var v model.OutboundCheckLine
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("outbound_check_line_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *OutboundCheckLineRepository) Record(ctx context.Context, id, resultID, qty, exception string, notes *string, actor string) error {
	res := r.db.WithContext(ctx).Model(&model.OutboundCheckLine{}).Where("outbound_check_line_id=? AND checked_qty IS NULL", id).Updates(map[string]interface{}{"checked_qty": qty, "exception_qty": exception, "outbound_check_result_id": resultID, "notes": notes, "checked_at": time.Now(), "checked_by": actor})
	if res.Error != nil {
		return Error(res.Error)
	}
	if res.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
func (r *OutboundCheckLineRepository) Unrecorded(ctx context.Context, id string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.OutboundCheckLine{}).Where("outbound_check_id=? AND checked_qty IS NULL", id).Count(&n).Error
	return n, Error(err)
}
func (r *OutboundCheckLineRepository) FailureCount(ctx context.Context, id string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("outbound_check_line l").Joins("JOIN outbound_check_result r ON r.outbound_check_result_id=l.outbound_check_result_id").Where("l.outbound_check_id=? AND (NOT r.is_pass OR l.checked_qty<>l.expected_qty)", id).Count(&n).Error
	return n, Error(err)
}
