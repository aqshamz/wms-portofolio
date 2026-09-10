package billing

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/billing"
)

type RateCardLineRow struct {
	model.RateCardLine
	MovementTypeCode *string
}
type RateCardLineRepository struct{ db *gorm.DB }

func NewRateCardLineRepository(db *gorm.DB) *RateCardLineRepository {
	return &RateCardLineRepository{db: db}
}
func (r *RateCardLineRepository) Create(ctx context.Context, v *model.RateCardLine) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *RateCardLineRepository) Get(ctx context.Context, id string) (RateCardLineRow, error) {
	var v RateCardLineRow
	e := r.db.WithContext(ctx).Table("rate_card_line l").Select("l.*,m.code movement_type_code").Joins("LEFT JOIN movement_type m ON m.movement_type_id=l.movement_type_id").Where("l.rate_card_line_id=?", id).Take(&v).Error
	return v, Error(e)
}
func (r *RateCardLineRepository) List(ctx context.Context, card string, activeOnly bool) ([]RateCardLineRow, error) {
	q := r.db.WithContext(ctx).Table("rate_card_line l").Select("l.*,m.code movement_type_code").Joins("LEFT JOIN movement_type m ON m.movement_type_id=l.movement_type_id").Where("l.rate_card_id=?", card)
	if activeOnly {
		q = q.Where("l.is_active")
	}
	v := []RateCardLineRow{}
	e := q.Order("l.service_code,l.rate_card_line_id").Find(&v).Error
	return v, Error(e)
}
func (r *RateCardLineRepository) Update(ctx context.Context, id, actor string, values map[string]interface{}) error {
	values["updated_by"] = actor
	values["updated_at"] = gorm.Expr("clock_timestamp()")
	return Error(r.db.WithContext(ctx).Model(&model.RateCardLine{}).Where("rate_card_line_id=?", id).Updates(values).Error)
}
func (r *RateCardLineRepository) Delete(ctx context.Context, id string) error {
	return Error(r.db.WithContext(ctx).Delete(&model.RateCardLine{}, "rate_card_line_id=?", id).Error)
}
