package billing

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/billing"
)

type InvoiceLineRepository struct{ db *gorm.DB }

func NewInvoiceLineRepository(db *gorm.DB) *InvoiceLineRepository {
	return &InvoiceLineRepository{db: db}
}
func (r *InvoiceLineRepository) Create(ctx context.Context, v *model.InvoiceLine) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *InvoiceLineRepository) List(ctx context.Context, invoice string) ([]model.InvoiceLine, error) {
	v := []model.InvoiceLine{}
	e := r.db.WithContext(ctx).Where("invoice_id=?", invoice).Order("line_no").Find(&v).Error
	return v, Error(e)
}
