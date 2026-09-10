package billing

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/billing"
)

type InvoiceRow struct {
	model.Invoice
	StatusCode, OwnerCode, WarehouseCode string
}
type InvoiceRepository struct{ db *gorm.DB }

func NewInvoiceRepository(db *gorm.DB) *InvoiceRepository { return &InvoiceRepository{db: db} }
func (r *InvoiceRepository) Create(ctx context.Context, v *model.Invoice) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *InvoiceRepository) Lock(ctx context.Context, id string) (model.Invoice, error) {
	var v model.Invoice
	e := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("invoice_id=?", id).Take(&v).Error
	return v, Error(e)
}
func invoiceQuery(db *gorm.DB) *gorm.DB {
	return db.Table("invoice i").Select("i.*,s.code status_code,o.code owner_code,w.code warehouse_code").Joins("JOIN document_status s ON s.status_id=i.status_id").Joins("JOIN organization o ON o.organization_id=i.owner_id").Joins("JOIN warehouse w ON w.warehouse_id=i.warehouse_id")
}
func (r *InvoiceRepository) Get(ctx context.Context, id string) (InvoiceRow, error) {
	var v InvoiceRow
	e := invoiceQuery(r.db.WithContext(ctx)).Where("i.invoice_id=?", id).Take(&v).Error
	return v, Error(e)
}
func (r *InvoiceRepository) ByRun(ctx context.Context, run string) (InvoiceRow, error) {
	var v InvoiceRow
	e := invoiceQuery(r.db.WithContext(ctx)).Where("i.billing_run_id=?", run).Take(&v).Error
	return v, Error(e)
}
func (r *InvoiceRepository) List(ctx context.Context, f ListFilter) ([]InvoiceRow, int64, error) {
	q := invoiceQuery(r.db.WithContext(ctx)).Where("i.owner_id=?", f.OwnerID)
	if f.WarehouseID != "" {
		q = q.Where("i.warehouse_id=?", f.WarehouseID)
	}
	if f.StatusCode != "" {
		q = q.Where("s.code=?", f.StatusCode)
	}
	if f.Search != "" {
		x := "%" + f.Search + "%"
		q = q.Where("i.invoice_id ILIKE ? OR i.billing_run_id ILIKE ?", x, x)
	}
	var n int64
	if e := q.Session(&gorm.Session{}).Count(&n).Error; e != nil {
		return nil, 0, Error(e)
	}
	v := []InvoiceRow{}
	e := q.Order("i.issue_date DESC,i.invoice_id DESC").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&v).Error
	return v, n, Error(e)
}
func (r *InvoiceRepository) UpdateDraft(ctx context.Context, id, actor string, version int64, values map[string]interface{}) error {
	values["updated_by"] = actor
	values["updated_at"] = gorm.Expr("clock_timestamp()")
	values["version_no"] = gorm.Expr("version_no+1")
	z := r.db.WithContext(ctx).Model(&model.Invoice{}).Where("invoice_id=? AND version_no=? AND status_id IN (SELECT status_id FROM document_status WHERE document_type_id=invoice.document_type_id AND code='DRAFT')", id, version).Updates(values)
	if z.Error != nil {
		return Error(z.Error)
	}
	if z.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
func (r *InvoiceRepository) SetFinancialsAndStatus(ctx context.Context, id, status, actor string, version int64, credit, total, paid string) error {
	z := r.db.WithContext(ctx).Model(&model.Invoice{}).Where("invoice_id=? AND version_no=?", id, version).Updates(map[string]interface{}{"status_id": status, "credit_amount": credit, "total_amount": total, "paid_amount": paid, "updated_by": actor, "updated_at": gorm.Expr("clock_timestamp()"), "version_no": gorm.Expr("version_no+1")})
	if z.Error != nil {
		return Error(z.Error)
	}
	if z.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
func (r *InvoiceRepository) SetStatus(ctx context.Context, id, status, actor string, version int64) error {
	z := r.db.WithContext(ctx).Model(&model.Invoice{}).Where("invoice_id=? AND version_no=?", id, version).Updates(map[string]interface{}{"status_id": status, "updated_by": actor, "updated_at": gorm.Expr("clock_timestamp()"), "version_no": gorm.Expr("version_no+1")})
	if z.Error != nil {
		return Error(z.Error)
	}
	if z.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
