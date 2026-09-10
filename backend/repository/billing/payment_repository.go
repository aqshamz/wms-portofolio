package billing

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/billing"
)

type PaymentRow struct {
	model.Payment
	StatusCode string
}
type PaymentRepository struct{ db *gorm.DB }

func NewPaymentRepository(db *gorm.DB) *PaymentRepository { return &PaymentRepository{db: db} }
func (r *PaymentRepository) Create(ctx context.Context, v *model.Payment) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *PaymentRepository) List(ctx context.Context, invoice string) ([]PaymentRow, error) {
	v := []PaymentRow{}
	e := r.db.WithContext(ctx).Table("payment p").Select("p.*,s.code status_code").Joins("JOIN document_status s ON s.status_id=p.status_id").Where("p.invoice_id=?", invoice).Order("p.created_at").Find(&v).Error
	return v, Error(e)
}
