package authentication

import (
	"context"
	model "wms-api/models/authentication"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AccountStatusRepository struct{ db *gorm.DB }

func NewAccountStatusRepository(db *gorm.DB) *AccountStatusRepository {
	return &AccountStatusRepository{db: db}
}

func (r *AccountStatusRepository) Seed(ctx context.Context, statuses []model.AccountStatus) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}},
		DoNothing: true,
	}).Create(&statuses).Error
}

func (r *AccountStatusRepository) FindByCode(ctx context.Context, code string) (model.AccountStatus, error) {
	var status model.AccountStatus
	err := r.db.WithContext(ctx).Where("code = ? AND is_active", code).First(&status).Error
	return status, err
}
