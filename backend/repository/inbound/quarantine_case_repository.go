package inbound

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/inbound"
)

type QuarantineCaseRow struct {
	model.QuarantineCase
	StatusCode, ItemID, ItemCode, LotNumber, LocationCode, DisposedQty string
}

type QuarantineCaseRepository struct{ db *gorm.DB }

func NewQuarantineCaseRepository(db *gorm.DB) *QuarantineCaseRepository {
	return &QuarantineCaseRepository{db: db}
}
func (r *QuarantineCaseRepository) Create(ctx context.Context, value *model.QuarantineCase) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}
func (r *QuarantineCaseRepository) Lock(ctx context.Context, id string) (model.QuarantineCase, error) {
	var value model.QuarantineCase
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("quarantine_case_id=?", id).Take(&value).Error
	return value, Error(err)
}
func quarantineCaseQuery(db *gorm.DB) *gorm.DB {
	return db.Table("quarantine_case qc").Select(`qc.*,status.code status_code,batch.item_id,item.code item_code,COALESCE(lot.lot_number,'') lot_number,location.code location_code,COALESCE((SELECT sum(d.disposition_qty) FROM quarantine_disposition d JOIN document_status ds ON ds.status_id=d.status_id WHERE d.quarantine_case_id=qc.quarantine_case_id AND ds.code='PROCESSED'),0)::text disposed_qty`).
		Joins("JOIN document_status status ON status.status_id=qc.status_id").Joins("JOIN receipt_inventory batch ON batch.receipt_inventory_id=qc.receipt_inventory_id").Joins("JOIN item ON item.item_id=batch.item_id").Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=batch.lot_id").Joins("JOIN inventory_balance balance ON balance.balance_id=qc.quarantine_balance_id").Joins("JOIN warehouse_location location ON location.location_id=balance.location_id")
}
func (r *QuarantineCaseRepository) Get(ctx context.Context, id string) (QuarantineCaseRow, error) {
	var value QuarantineCaseRow
	err := quarantineCaseQuery(r.db.WithContext(ctx)).Where("qc.quarantine_case_id=?", id).Take(&value).Error
	return value, Error(err)
}
func (r *QuarantineCaseRepository) GetByInspection(ctx context.Context, inspectionID string) (QuarantineCaseRow, error) {
	var value QuarantineCaseRow
	err := quarantineCaseQuery(r.db.WithContext(ctx)).Where("qc.inspection_id=?", inspectionID).Take(&value).Error
	return value, Error(err)
}
func (r *QuarantineCaseRepository) List(ctx context.Context, filter ListFilter) ([]QuarantineCaseRow, int64, error) {
	query := quarantineCaseQuery(r.db.WithContext(ctx)).Where("qc.owner_id=?", filter.OwnerID)
	if filter.WarehouseID != "" {
		query = query.Where("qc.warehouse_id=?", filter.WarehouseID)
	}
	if filter.StatusCode != "" {
		query = query.Where("status.code=?", filter.StatusCode)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("qc.quarantine_case_id ILIKE ? OR item.code ILIKE ? OR lot.lot_number ILIKE ?", like, like, like)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, Error(err)
	}
	rows := make([]QuarantineCaseRow, 0)
	err := query.Order("qc.opened_at DESC,qc.quarantine_case_id DESC").Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize).Find(&rows).Error
	return rows, total, Error(err)
}
func (r *QuarantineCaseRepository) SetStatus(ctx context.Context, id, statusID, actor string, close bool, expectedVersion int64) error {
	changes := map[string]interface{}{"status_id": statusID, "updated_at": gorm.Expr("clock_timestamp()"), "updated_by": actor, "version_no": gorm.Expr("version_no+1")}
	if close {
		changes["closed_at"] = time.Now()
	}
	result := r.db.WithContext(ctx).Model(&model.QuarantineCase{}).Where("quarantine_case_id=? AND version_no=?", id, expectedVersion).Updates(changes)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
