package inbound

import (
	"context"

	model "wms-api/models/inbound"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReceiptRow struct {
	model.Receipt
	StatusCode, OwnerCode, WarehouseCode string
}

type ReceiptRepository struct{ db *gorm.DB }

func NewReceiptRepository(db *gorm.DB) *ReceiptRepository { return &ReceiptRepository{db: db} }
func (r *ReceiptRepository) Create(ctx context.Context, value *model.Receipt) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}
func (r *ReceiptRepository) Lock(ctx context.Context, id string) (model.Receipt, error) {
	var value model.Receipt
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("receipt_id=?", id).Take(&value).Error
	return value, Error(err)
}
func receiptQuery(db *gorm.DB) *gorm.DB {
	return db.Table("receipt receipt").Select("receipt.*,status.code status_code,owner.code owner_code,warehouse.code warehouse_code").
		Joins("JOIN document_status status ON status.status_id=receipt.status_id").Joins("JOIN organization owner ON owner.organization_id=receipt.owner_id").Joins("JOIN warehouse ON warehouse.warehouse_id=receipt.warehouse_id")
}
func (r *ReceiptRepository) Get(ctx context.Context, id string) (ReceiptRow, error) {
	var value ReceiptRow
	err := receiptQuery(r.db.WithContext(ctx)).Where("receipt.receipt_id=?", id).Take(&value).Error
	return value, Error(err)
}
func (r *ReceiptRepository) List(ctx context.Context, filter ListFilter) ([]ReceiptRow, int64, error) {
	query := receiptQuery(r.db.WithContext(ctx)).Where("receipt.owner_id=?", filter.OwnerID)
	if filter.WarehouseID != "" {
		query = query.Where("receipt.warehouse_id=?", filter.WarehouseID)
	}
	if filter.StatusCode != "" {
		query = query.Where("status.code=?", filter.StatusCode)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("receipt.receipt_id ILIKE ? OR receipt.delivery_note_no ILIKE ? OR receipt.vehicle_number ILIKE ?", like, like, like)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, Error(err)
	}
	rows := make([]ReceiptRow, 0)
	err := query.Order("receipt.business_date DESC,receipt.receipt_id DESC").Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize).Find(&rows).Error
	return rows, total, Error(err)
}
func (r *ReceiptRepository) SetStatus(ctx context.Context, id, statusID, actor string, expectedVersion int64) error {
	result := r.db.WithContext(ctx).Model(&model.Receipt{}).Where("receipt_id=? AND version_no=?", id, expectedVersion).
		Updates(map[string]interface{}{"status_id": statusID, "updated_by": actor, "updated_at": gorm.Expr("clock_timestamp()"), "version_no": gorm.Expr("version_no+1")})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
