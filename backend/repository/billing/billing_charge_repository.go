package billing

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/billing"
)

type BillingChargeRepository struct{ db *gorm.DB }

func NewBillingChargeRepository(db *gorm.DB) *BillingChargeRepository {
	return &BillingChargeRepository{db: db}
}
func (r *BillingChargeRepository) Create(ctx context.Context, v *model.BillingCharge) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *BillingChargeRepository) List(ctx context.Context, run string) ([]model.BillingCharge, error) {
	v := []model.BillingCharge{}
	e := r.db.WithContext(ctx).Where("billing_run_id=?", run).Order("created_at,billing_charge_id").Find(&v).Error
	return v, Error(e)
}
func (r *BillingChargeRepository) DeleteByRun(ctx context.Context, run string) error {
	return Error(r.db.WithContext(ctx).Where("billing_run_id=?", run).Delete(&model.BillingCharge{}).Error)
}
