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

func (r *AccountStatusRepository) FindByID(ctx context.Context, id string) (model.AccountStatus, error) {
	var status model.AccountStatus
	err := r.db.WithContext(ctx).Where("account_status_id = ? AND is_active", id).First(&status).Error
	return status, err
}

func (r *AccountStatusRepository) List(ctx context.Context) ([]model.AccountStatus, error) {
	rows := make([]model.AccountStatus, 0)
	err := r.db.WithContext(ctx).Where("is_active").Order("code").Find(&rows).Error
	return rows, err
}
