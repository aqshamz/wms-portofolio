package billing

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/billing"
)

type CreditNoteRow struct {
	model.CreditNote
	StatusCode string
}
type CreditNoteRepository struct{ db *gorm.DB }

func NewCreditNoteRepository(db *gorm.DB) *CreditNoteRepository { return &CreditNoteRepository{db: db} }
func (r *CreditNoteRepository) Create(ctx context.Context, v *model.CreditNote) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *CreditNoteRepository) List(ctx context.Context, invoice string) ([]CreditNoteRow, error) {
	v := []CreditNoteRow{}
	e := r.db.WithContext(ctx).Table("credit_note c").Select("c.*,s.code status_code").Joins("JOIN document_status s ON s.status_id=c.status_id").Where("c.invoice_id=?", invoice).Order("c.created_at").Find(&v).Error
	return v, Error(e)
}
