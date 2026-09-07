package stockcontrol

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/master"
)

type ReasonCodeRepository struct{ db *gorm.DB }

func NewReasonCodeRepository(db *gorm.DB) *ReasonCodeRepository { return &ReasonCodeRepository{db: db} }
func (r *ReasonCodeRepository) ByCode(ctx context.Context, module, code string) (model.ReasonCode, error) {
	var v model.ReasonCode
	err := r.db.WithContext(ctx).Where("module_code=? AND code=?", module, code).Take(&v).Error
	return v, err
}
func (r *ReasonCodeRepository) List(ctx context.Context, module string, active *bool) ([]model.ReasonCode, error) {
	q := r.db.WithContext(ctx).Where("module_code=?", module)
	if active != nil {
		q = q.Where("is_active=?", *active)
	}
	rows := make([]model.ReasonCode, 0)
	err := q.Order("code").Find(&rows).Error
	return rows, err
}
func (r *ReasonCodeRepository) Seed(ctx context.Context, values []model.ReasonCode) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "module_code"}, {Name: "code"}}, DoNothing: true}).Create(&values).Error
}
