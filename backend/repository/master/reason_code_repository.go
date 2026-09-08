package master

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/master"
)

type ReasonCodeRepository struct{ db *gorm.DB }

func NewReasonCodeRepository(db *gorm.DB) *ReasonCodeRepository { return &ReasonCodeRepository{db: db} }
func (r *ReasonCodeRepository) ByModuleCode(ctx context.Context, module, code string) (model.ReasonCode, error) {
	var v model.ReasonCode
	err := r.db.WithContext(ctx).Where("module_code=? AND code=? AND is_active", module, code).Take(&v).Error
	return v, err
}

func (r *ReasonCodeRepository) Seed(ctx context.Context, values []model.ReasonCode) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "module_code"}, {Name: "code"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "description", "requires_note", "is_active"})}).Create(&values).Error
}
