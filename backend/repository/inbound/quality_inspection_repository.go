package inbound

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/inbound"
)

type QualityInspectionRow struct {
	model.QualityInspection
	QualityStatusCode, InspectionResultCode, ReceiptID, OwnerID, WarehouseID string
	ItemID, ItemCode, LotNumber, LocationCode                                string
}

type QualityInspectionRepository struct{ db *gorm.DB }

func NewQualityInspectionRepository(db *gorm.DB) *QualityInspectionRepository {
	return &QualityInspectionRepository{db: db}
}
func (r *QualityInspectionRepository) Create(ctx context.Context, value *model.QualityInspection) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}
func (r *QualityInspectionRepository) Lock(ctx context.Context, id string) (model.QualityInspection, error) {
	var value model.QualityInspection
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("inspection_id=?", id).Take(&value).Error
	return value, Error(err)
}
func qualityInspectionQuery(db *gorm.DB) *gorm.DB {
	return db.Table("quality_inspection inspection").Select(`inspection.*,quality.code quality_status_code,result.code inspection_result_code,receipt.receipt_id,receipt.owner_id,receipt.warehouse_id,batch.item_id,item.code item_code,COALESCE(lot.lot_number,'') lot_number,location.code location_code`).
		Joins("JOIN quality_status quality ON quality.quality_status_id=inspection.quality_status_id").Joins("LEFT JOIN inspection_result result ON result.inspection_result_id=inspection.inspection_result_id").
		Joins("JOIN receipt_inventory batch ON batch.receipt_inventory_id=inspection.receipt_inventory_id").Joins("JOIN receipt_line line ON line.receipt_line_id=batch.receipt_line_id").Joins("JOIN receipt ON receipt.receipt_id=line.receipt_id").
		Joins("JOIN item ON item.item_id=batch.item_id").Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=batch.lot_id").Joins("JOIN warehouse_location location ON location.location_id=batch.received_location_id")
}
func (r *QualityInspectionRepository) Get(ctx context.Context, id string) (QualityInspectionRow, error) {
	var value QualityInspectionRow
	err := qualityInspectionQuery(r.db.WithContext(ctx)).Where("inspection.inspection_id=?", id).Take(&value).Error
	return value, Error(err)
}
func (r *QualityInspectionRepository) List(ctx context.Context, filter ListFilter) ([]QualityInspectionRow, int64, error) {
	query := qualityInspectionQuery(r.db.WithContext(ctx)).Where("receipt.owner_id=?", filter.OwnerID)
	if filter.WarehouseID != "" {
		query = query.Where("receipt.warehouse_id=?", filter.WarehouseID)
	}
	if filter.StatusCode != "" {
		query = query.Where("quality.code=?", filter.StatusCode)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("inspection.inspection_id ILIKE ? OR receipt.receipt_id ILIKE ? OR item.code ILIKE ? OR lot.lot_number ILIKE ?", like, like, like, like)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, Error(err)
	}
	rows := make([]QualityInspectionRow, 0)
	err := query.Order("inspection.created_at DESC,inspection.inspection_id DESC").Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize).Find(&rows).Error
	return rows, total, Error(err)
}
func (r *QualityInspectionRepository) Complete(ctx context.Context, id, qualityStatusID, resultID, passedQty, failedQty, actor string, expectedVersion int64, notes *string) error {
	now := time.Now()
	changes := map[string]interface{}{"quality_status_id": qualityStatusID, "inspection_result_id": resultID, "passed_qty": passedQty, "failed_qty": failedQty, "inspected_at": now, "inspected_by": actor, "updated_at": gorm.Expr("clock_timestamp()"), "updated_by": actor, "version_no": gorm.Expr("version_no+1")}
	if notes != nil {
		changes["notes"] = notes
	}
	result := r.db.WithContext(ctx).Model(&model.QualityInspection{}).Where("inspection_id=? AND version_no=? AND inspected_at IS NULL", id, expectedVersion).Updates(changes)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}

func (r *QualityInspectionRepository) CountByBatch(ctx context.Context, batchID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.QualityInspection{}).Where("receipt_inventory_id=?", batchID).Count(&count).Error
	return count, Error(err)
}

func (r *QualityInspectionRepository) GetChild(ctx context.Context, parentID string) (QualityInspectionRow, error) {
	var value QualityInspectionRow
	err := qualityInspectionQuery(r.db.WithContext(ctx)).Where("inspection.parent_inspection_id=?", parentID).Order("inspection.created_at DESC").Take(&value).Error
	return value, Error(err)
}

func (r *QualityInspectionRepository) Cancel(ctx context.Context, id, waivedStatusID, actor, reason string, expectedVersion int64) error {
	result := r.db.WithContext(ctx).Model(&model.QualityInspection{}).Where("inspection_id=? AND version_no=? AND inspected_at IS NULL AND cancelled_at IS NULL", id, expectedVersion).Updates(map[string]interface{}{
		"quality_status_id": waivedStatusID, "cancelled_at": time.Now(), "cancelled_by": actor, "cancellation_reason": reason, "updated_at": gorm.Expr("clock_timestamp()"), "updated_by": actor, "version_no": gorm.Expr("version_no+1"),
	})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
