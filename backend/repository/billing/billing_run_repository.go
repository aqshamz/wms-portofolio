package billing

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/billing"
)

type BillingRunRow struct {
	model.BillingRun
	StatusCode, OwnerCode, WarehouseCode string
}
type BillingRunRepository struct{ db *gorm.DB }

func NewBillingRunRepository(db *gorm.DB) *BillingRunRepository { return &BillingRunRepository{db: db} }
func (r *BillingRunRepository) Create(ctx context.Context, v *model.BillingRun) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *BillingRunRepository) Lock(ctx context.Context, id string) (model.BillingRun, error) {
	var v model.BillingRun
	e := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("billing_run_id=?", id).Take(&v).Error
	return v, Error(e)
}
func runQuery(db *gorm.DB) *gorm.DB {
	return db.Table("billing_run r").Select("r.*,s.code status_code,o.code owner_code,w.code warehouse_code").Joins("JOIN document_status s ON s.status_id=r.status_id").Joins("JOIN organization o ON o.organization_id=r.owner_id").Joins("JOIN warehouse w ON w.warehouse_id=r.warehouse_id")
}
func (r *BillingRunRepository) Get(ctx context.Context, id string) (BillingRunRow, error) {
	var v BillingRunRow
	e := runQuery(r.db.WithContext(ctx)).Where("r.billing_run_id=?", id).Take(&v).Error
	return v, Error(e)
}
func (r *BillingRunRepository) List(ctx context.Context, f ListFilter) ([]BillingRunRow, int64, error) {
	q := runQuery(r.db.WithContext(ctx)).Where("r.owner_id=?", f.OwnerID)
	if f.WarehouseID != "" {
		q = q.Where("r.warehouse_id=?", f.WarehouseID)
	}
	if f.StatusCode != "" {
		q = q.Where("s.code=?", f.StatusCode)
	}
	if f.Search != "" {
		x := "%" + f.Search + "%"
		q = q.Where("r.billing_run_id ILIKE ?", x)
	}
	var n int64
	if e := q.Session(&gorm.Session{}).Count(&n).Error; e != nil {
		return nil, 0, Error(e)
	}
	v := []BillingRunRow{}
	e := q.Order("r.business_date DESC,r.billing_run_id DESC").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&v).Error
	return v, n, Error(e)
}
func (r *BillingRunRepository) SetStatusAndTotals(ctx context.Context, id, status, actor string, version int64, subtotal, tax, total string) error {
	z := r.db.WithContext(ctx).Model(&model.BillingRun{}).Where("billing_run_id=? AND version_no=?", id, version).Updates(map[string]interface{}{"status_id": status, "subtotal_amount": subtotal, "tax_amount": tax, "total_amount": total, "updated_by": actor, "updated_at": gorm.Expr("clock_timestamp()"), "version_no": gorm.Expr("version_no+1")})
	if z.Error != nil {
		return Error(z.Error)
	}
	if z.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
func (r *BillingRunRepository) SetStatus(ctx context.Context, id, status, actor string, version int64) error {
	z := r.db.WithContext(ctx).Model(&model.BillingRun{}).Where("billing_run_id=? AND version_no=?", id, version).Updates(map[string]interface{}{"status_id": status, "updated_by": actor, "updated_at": gorm.Expr("clock_timestamp()"), "version_no": gorm.Expr("version_no+1")})
	if z.Error != nil {
		return Error(z.Error)
	}
	if z.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
